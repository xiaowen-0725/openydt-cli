package pagination

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xiaowen-0725/openydt-cli/internal/client"
)

func TestRunFetchesPagesSequentiallyAndStreamsNDJSON(t *testing.T) {
	var pages []int
	fetch := func(_ context.Context, body string) (*client.Response, error) {
		var request map[string]any
		if err := json.Unmarshal([]byte(body), &request); err != nil {
			t.Fatal(err)
		}
		page := int(request["pageNum"].(float64))
		pages = append(pages, page)

		records := `[{"id":1},{"id":2}]`
		if page == 2 {
			records = `[{"id":3}]`
		}
		return &client.Response{
			Status: client.StatusSuccess,
			Data:   json.RawMessage(`{"total":3,"recordList":` + records + `}`),
		}, nil
	}

	var out, progress bytes.Buffer
	var sleeps []time.Duration
	stats, err := Run(context.Background(), Options{
		Body:     `{"parkCode":"P"}`,
		PageSize: 2,
		Interval: 500 * time.Millisecond,
		Fetch:    fetch,
		Out:      &out,
		Progress: &progress,
		Sleep:    func(_ context.Context, d time.Duration) error { sleeps = append(sleeps, d); return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(pages, []int{1, 2}) {
		t.Fatalf("pages=%v, want [1 2]", pages)
	}
	if !reflect.DeepEqual(sleeps, []time.Duration{500 * time.Millisecond}) {
		t.Fatalf("sleeps=%v, want one 500ms pause", sleeps)
	}
	if got, want := out.String(), "{\"id\":1}\n{\"id\":2}\n{\"id\":3}\n"; got != want {
		t.Fatalf("NDJSON output:\n%s\nwant:\n%s", got, want)
	}
	if stats.Pages != 2 || stats.Records != 3 || stats.Total != 3 {
		t.Fatalf("stats=%+v", stats)
	}
	if !strings.Contains(progress.String(), "page 2/2") || !strings.Contains(progress.String(), "records 3/3") {
		t.Fatalf("progress=%q", progress.String())
	}
}

func TestRunFallsBackToTheOnlyTopLevelArrayAndStopsOnShortPage(t *testing.T) {
	page := 0
	fetch := func(_ context.Context, _ string) (*client.Response, error) {
		page++
		items := `[{"id":1},{"id":2}]`
		if page == 2 {
			items = `[{"id":3}]`
		}
		return &client.Response{
			Status: client.StatusSuccess,
			Data:   json.RawMessage(`{"items":` + items + `,"currCount":` + string(rune('0'+page)) + `}`),
		}, nil
	}

	var out bytes.Buffer
	stats, err := Run(context.Background(), Options{
		Body: `{"parkCode":"P"}`, PageSize: 2, Fetch: fetch, Out: &out,
		Sleep: func(context.Context, time.Duration) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Pages != 2 || stats.Records != 3 || stats.Total != 3 {
		t.Fatalf("stats=%+v", stats)
	}
}

func TestRunTreatsEmptyDataObjectAsNoRecords(t *testing.T) {
	var out bytes.Buffer
	stats, err := Run(context.Background(), Options{
		Body: `{}`, PageSize: 100, Out: &out,
		Fetch: func(context.Context, string) (*client.Response, error) {
			return &client.Response{Status: client.StatusSuccess, Data: json.RawMessage(`{}`)}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Records != 0 || out.Len() != 0 {
		t.Fatalf("stats=%+v output=%q", stats, out.String())
	}
}

func TestRunTreatsDirectDataObjectAsOneRecord(t *testing.T) {
	var out bytes.Buffer
	stats, err := Run(context.Background(), Options{
		Body: `{}`, PageSize: 100, Out: &out,
		Fetch: func(context.Context, string) (*client.Response, error) {
			return &client.Response{Status: client.StatusSuccess, Data: json.RawMessage(`{"parkingCode":"X"}`)}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Records != 1 || out.String() != "{\"parkingCode\":\"X\"}\n" {
		t.Fatalf("stats=%+v output=%q", stats, out.String())
	}
}

func TestExtractPageAcceptsNumericStringTotal(t *testing.T) {
	_, total, err := extractPage(json.RawMessage(`{"total":"10","couponList":[{"id":1}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if total != 10 {
		t.Fatalf("total=%d want 10", total)
	}
}

func TestRunRejectsNonObjectRequestBody(t *testing.T) {
	_, err := Run(context.Background(), Options{
		Body: `null`, PageSize: 10,
		Fetch: func(context.Context, string) (*client.Response, error) {
			t.Fatal("fetch should not run")
			return nil, nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "JSON 对象") {
		t.Fatalf("err=%v", err)
	}
}
