// Package pagination fetches page-number based platform queries sequentially
// and streams their records as NDJSON.
package pagination

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xiaowen-0725/openydt-cli/internal/client"
)

// FetchFunc performs one platform request for the supplied JSON body.
type FetchFunc func(context.Context, string) (*client.Response, error)

// Options is the small interface for a complete paginated query.
type Options struct {
	Body     string
	PageSize int
	Interval time.Duration
	Fetch    FetchFunc
	Out      io.Writer
	Progress io.Writer
	Sleep    func(context.Context, time.Duration) error
}

// Stats describes the completed extraction.
type Stats struct {
	Pages   int
	Records int
	Total   int
}

// PageError preserves a non-success platform response for the CLI renderer.
type PageError struct {
	Response *client.Response
}

func (e *PageError) Error() string {
	return fmt.Sprintf("分页查询失败: status=%d resultCode=%d message=%s", e.Response.Status, e.Response.ResultCode, e.Response.Message)
}

// Run fetches pages from page 1 in order, pausing between pages and writing one
// compact JSON record per line. It stops at total or at a short/empty page.
func Run(ctx context.Context, opts Options) (Stats, error) {
	if opts.Fetch == nil {
		return Stats{}, fmt.Errorf("分页查询缺少 fetch")
	}
	if opts.PageSize <= 0 {
		return Stats{}, fmt.Errorf("pageSize 必须大于 0")
	}
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	if opts.Progress == nil {
		opts.Progress = io.Discard
	}
	if opts.Sleep == nil {
		opts.Sleep = sleep
	}

	var base map[string]any
	if err := json.Unmarshal([]byte(opts.Body), &base); err != nil {
		return Stats{}, fmt.Errorf("分页请求体必须是 JSON 对象: %w", err)
	}
	if base == nil {
		return Stats{}, fmt.Errorf("分页请求体必须是 JSON 对象")
	}

	enc := json.NewEncoder(opts.Out)
	enc.SetEscapeHTML(false)
	stats := Stats{Total: -1}
	for page := 1; ; page++ {
		base["pageNum"] = page
		base["pageSize"] = opts.PageSize
		body, err := json.Marshal(base)
		if err != nil {
			return stats, err
		}
		resp, err := opts.Fetch(ctx, string(body))
		if err != nil {
			return stats, err
		}
		if !resp.OK() {
			return stats, &PageError{Response: resp}
		}

		records, total, err := extractPage(resp.Data)
		if err != nil {
			return stats, err
		}
		for _, record := range records {
			if err := enc.Encode(record); err != nil {
				return stats, err
			}
		}
		stats.Pages++
		stats.Records += len(records)
		if total >= 0 {
			stats.Total = total
		}
		writeProgress(opts.Progress, stats, opts.PageSize)

		if len(records) == 0 || (stats.Total >= 0 && stats.Records >= stats.Total) || (stats.Total < 0 && len(records) < opts.PageSize) {
			if stats.Total < 0 {
				stats.Total = stats.Records
			}
			return stats, nil
		}
		if err := opts.Sleep(ctx, opts.Interval); err != nil {
			return stats, err
		}
	}
}

func extractPage(data json.RawMessage) ([]json.RawMessage, int, error) {
	var direct []json.RawMessage
	if err := json.Unmarshal(data, &direct); err == nil {
		return direct, -1, nil
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, -1, fmt.Errorf("分页响应 data 不是对象: %w", err)
	}
	total := -1
	for _, key := range []string{"total", "count", "totalCount"} {
		if raw, ok := obj[key]; ok {
			if parsed, ok := parseTotal(raw); ok {
				total = parsed
				break
			}
		}
	}
	for _, key := range []string{"recordList", "recordInList", "recordOutList", "monthTicketList", "records", "rows", "list"} {
		raw, ok := obj[key]
		if !ok {
			continue
		}
		var records []json.RawMessage
		if err := json.Unmarshal(raw, &records); err == nil {
			return records, total, nil
		}
	}
	var fallback []json.RawMessage
	found := false
	for key, raw := range obj {
		if strings.Contains(strings.ToLower(key), "image") {
			continue
		}
		var records []json.RawMessage
		if err := json.Unmarshal(raw, &records); err != nil {
			continue
		}
		if found {
			return nil, total, fmt.Errorf("分页响应包含多个未知记录数组")
		}
		fallback = records
		found = true
	}
	if found {
		return fallback, total, nil
	}
	if len(obj) == 0 || total == 0 {
		return []json.RawMessage{}, total, nil
	}
	if total < 0 && !hasPaginationMetadata(obj) {
		return []json.RawMessage{json.RawMessage(data)}, 1, nil
	}
	return nil, total, fmt.Errorf("分页响应未找到记录数组")
}

func parseTotal(raw json.RawMessage) (int, bool) {
	var total int
	if err := json.Unmarshal(raw, &total); err == nil {
		return total, true
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return 0, false
	}
	total, err := strconv.Atoi(strings.TrimSpace(text))
	return total, err == nil
}

func hasPaginationMetadata(obj map[string]json.RawMessage) bool {
	for _, key := range []string{"total", "count", "totalCount", "currCount", "pageNum", "pageSize", "pageCount"} {
		if _, ok := obj[key]; ok {
			return true
		}
	}
	return false
}

func writeProgress(w io.Writer, stats Stats, pageSize int) {
	if stats.Total >= 0 {
		pages := (stats.Total + pageSize - 1) / pageSize
		fmt.Fprintf(w, "[openydt] page %d/%d records %d/%d\n", stats.Pages, pages, stats.Records, stats.Total)
		return
	}
	fmt.Fprintf(w, "[openydt] page %d records %d\n", stats.Pages, stats.Records)
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
