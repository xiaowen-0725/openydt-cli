package client

import (
	"bytes"
	"strings"
	"testing"
)

// TestLogfRedactsSign verifies that logf strips hex sign digests from log
// output but leaves short algorithm tags (e.g. "sign=v2") intact.
func TestLogfRedactsSign(t *testing.T) {
	t.Run("redacts hex digest in URL error", func(t *testing.T) {
		var buf bytes.Buffer
		c := &Client{Verbose: true, Log: &buf}

		// Simulate the string produced by *url.Error when the sign query
		// parameter contains a 32-char MD5 hash.
		c.logf("Post \"https://x/api/v3/getX?sign=abcdef0123456789abcdef0123456789\": boom")

		out := buf.String()
		if strings.Contains(out, "abcdef0123456789abcdef0123456789") {
			t.Errorf("log output still contains original hex digest: %q", out)
		}
		if !strings.Contains(out, "sign=***") {
			t.Errorf("log output missing redaction marker sign=***: %q", out)
		}
	})

	t.Run("preserves sign=v2 algorithm tag", func(t *testing.T) {
		var buf bytes.Buffer
		c := &Client{Verbose: true, Log: &buf}

		c.logf("sign=v2 base=https://openapi-test.yidianting.com.cn")

		out := buf.String()
		if !strings.Contains(out, "sign=v2") {
			t.Errorf("log output should preserve sign=v2 tag but got: %q", out)
		}
		if strings.Contains(out, "sign=***") {
			t.Errorf("sign=v2 was incorrectly redacted to sign=***: %q", out)
		}
	})

	t.Run("preserves sign=v3 algorithm tag", func(t *testing.T) {
		var buf bytes.Buffer
		c := &Client{Verbose: true, Log: &buf}

		c.logf("sign=v3 base=https://openapi.example.com")

		out := buf.String()
		if !strings.Contains(out, "sign=v3") {
			t.Errorf("log output should preserve sign=v3 tag but got: %q", out)
		}
	})

	t.Run("no output when Verbose is false", func(t *testing.T) {
		var buf bytes.Buffer
		c := &Client{Verbose: false, Log: &buf}

		c.logf("should not appear sign=abcdef0123456789")

		if buf.Len() != 0 {
			t.Errorf("expected no output when Verbose=false, got: %q", buf.String())
		}
	})

	t.Run("no output when Log is nil", func(t *testing.T) {
		// Must not panic.
		c := &Client{Verbose: true, Log: nil}
		c.logf("no log destination sign=abcdef0123456789")
	})
}
