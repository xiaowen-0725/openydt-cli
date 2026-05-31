// Package client is the openydt open-platform HTTP client: it signs requests,
// retries the gateway's intermittent failures, and decodes the response envelope.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/xiaowen-0725/openydt-cli/internal/sign"
)

// signRe matches a sign query-parameter value that looks like a hex digest
// (16+ hex chars). It deliberately avoids matching short algorithm tags such
// as "sign=v2" or "sign=v3", which must be preserved in log output.
var signRe = regexp.MustCompile(`sign=[0-9a-f]{16,}`)

// Client calls the openydt platform. Build one with New.
type Client struct {
	HTTP       *http.Client
	BaseURL    string // e.g. https://openapi-test.yidianting.com.cn
	Key        string
	Secret     string
	Sign       sign.Version
	UserAgent  string
	MaxRetries int           // extra attempts after the first try
	RetryBase  time.Duration // base backoff
	Verbose    bool          // --verbose: log HTTP attempts/status/timing to Log
	Log        io.Writer     // destination for verbose output (typically stderr)
}

// New builds a client with sensible defaults.
func New(baseURL, key, secret string, v sign.Version, userAgent string) *Client {
	return &Client{
		HTTP:       &http.Client{Timeout: 30 * time.Second},
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Key:        key,
		Secret:     secret,
		Sign:       v,
		UserAgent:  userAgent,
		MaxRetries: 3,
		RetryBase:  400 * time.Millisecond,
	}
}

// Response is the platform's standard envelope.
type Response struct {
	Status     int             `json:"status"`
	ResultCode int             `json:"resultCode"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data"`

	HTTPStatus int             `json:"-"`
	Raw        json.RawMessage `json:"-"`
}

// OK reports whether the call succeeded at the business level (status == 1).
func (r *Response) OK() bool { return r.Status == StatusSuccess }

// Prepared captures exactly what was/would be sent (used for --dry-run and verify).
type Prepared struct {
	Method        string
	URL           string
	Authorization string
	Sign          string
	Ts            string
	Body          string // compacted JSON, byte-identical to what is sent and signed
}

// Prepare builds the signed request artifacts for a command without sending it.
func (c *Client) Prepare(cmd, body string) (Prepared, error) {
	compact := sign.CompactBody(body)
	ts := sign.Now()
	s, err := sign.Compute(c.Sign, c.Key, c.Secret, ts, compact)
	if err != nil {
		return Prepared{}, err
	}
	return Prepared{
		Method:        http.MethodPost,
		URL:           fmt.Sprintf("%s/openydt/api/v3/%s?sign=%s", c.BaseURL, cmd, s),
		Authorization: sign.Authorization(c.Key, ts),
		Sign:          s,
		Ts:            ts,
		Body:          compact,
	}, nil
}

// logf writes a verbose HTTP log line when c.Verbose is set and c.Log is non-nil.
// The formatted message is redacted before writing: any sign=<hex digest> (16+
// hex chars) is replaced with sign=*** so that the MD5 hash never leaks into
// stderr even when a *url.Error embeds the full request URL.
func (c *Client) logf(format string, a ...any) {
	if c.Verbose && c.Log != nil {
		msg := fmt.Sprintf("[openydt http] "+format, a...)
		msg = signRe.ReplaceAllString(msg, "sign=***")
		fmt.Fprintln(c.Log, msg)
	}
}

// Call signs and sends cmd with the given JSON body, retrying transient gateway
// failures with exponential backoff. The returned Response carries the business
// envelope even when status != 1; transport-level failures return an error.
func (c *Client) Call(ctx context.Context, cmd, body string) (*Response, error) {
	p, err := c.Prepare(cmd, body)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		if attempt > 0 {
			bo := c.backoff(attempt)
			c.logf("attempt %d/%d backoff %dms (cmd=%s)", attempt+1, c.MaxRetries+1, bo.Milliseconds(), cmd)
			if err := sleep(ctx, bo); err != nil {
				return nil, err
			}
		} else {
			c.logf("attempt 1/%d (cmd=%s)", c.MaxRetries+1, cmd)
		}
		resp, retry, err := c.do(ctx, p)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !retry {
			c.logf("non-retryable error (cmd=%s): %v", cmd, err)
			return nil, err
		}
		c.logf("retryable error (cmd=%s): %v", cmd, err)
	}
	return nil, fmt.Errorf("request failed after %d attempts: %w", c.MaxRetries+1, lastErr)
}

func (c *Client) do(ctx context.Context, p Prepared) (resp *Response, retry bool, err error) {
	req, err := http.NewRequestWithContext(ctx, p.Method, p.URL, bytes.NewReader([]byte(p.Body)))
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("Authorization", p.Authorization)
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	t0 := time.Now()
	httpResp, err := c.HTTP.Do(req)
	if err != nil {
		// network error / connection reset: retryable
		c.logf("-> network error (%dms): %v", time.Since(t0).Milliseconds(), err)
		return nil, true, err
	}
	defer httpResp.Body.Close()
	raw, _ := io.ReadAll(httpResp.Body)
	ms := time.Since(t0).Milliseconds()

	// Gateway (TGW/APISIX) intermittently 404s a valid route, or 5xx/429s.
	if isRetryableStatus(httpResp.StatusCode) {
		c.logf("-> HTTP %d (%dms) [retry]", httpResp.StatusCode, ms)
		return nil, true, fmt.Errorf("gateway HTTP %d: %s", httpResp.StatusCode, snippet(raw))
	}

	var r Response
	if err := json.Unmarshal(raw, &r); err != nil {
		// Not the business envelope (e.g. an HTML error page): not retryable.
		c.logf("-> HTTP %d (%dms) [non-retryable, unexpected body]", httpResp.StatusCode, ms)
		return nil, false, fmt.Errorf("HTTP %d, unexpected body: %s", httpResp.StatusCode, snippet(raw))
	}
	c.logf("-> HTTP %d (%dms) status=%d", httpResp.StatusCode, ms, r.Status)
	r.HTTPStatus = httpResp.StatusCode
	r.Raw = raw
	return &r, false, nil
}

func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusNotFound, // 404 — TGW node routing flake
		http.StatusTooManyRequests,    // 429 — rate limited
		http.StatusBadGateway,         // 502
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout:     // 504
		return true
	}
	return false
}

func (c *Client) backoff(attempt int) time.Duration {
	d := c.RetryBase << (attempt - 1) // 400ms, 800ms, 1600ms, ...
	jitter := time.Duration(rand.Int63n(int64(c.RetryBase)))
	return d + jitter
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func snippet(b []byte) string {
	const max = 200
	s := strings.TrimSpace(string(b))
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
