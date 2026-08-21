package cmdutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/client"
	"github.com/xiaowen-0725/openydt-cli/internal/sign"
)

func TestWriteGuard(t *testing.T) {
	cases := []struct {
		name                string
		write, yes, dry, ro bool
		wantErr             bool
		wantReadOnly        bool // err 是只读拒绝
	}{
		{"read passes", false, false, false, false, false, false},
		{"write no-yes blocked", true, false, false, false, true, false},
		{"write yes ok", true, true, false, false, false, false},
		{"write dry-run ok", true, false, true, false, false, false},
		{"readonly blocks write even with yes", true, true, false, true, true, true},
		{"readonly allows read", false, false, false, true, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := writeGuard("someCmd", c.write, c.yes, c.dry, c.ro)
			if (err != nil) != c.wantErr {
				t.Fatalf("err=%v want %v", err, c.wantErr)
			}
			if c.wantReadOnly && err != nil && !isReadOnlyErr(err) {
				t.Fatalf("expected read-only error, got %v", err)
			}
		})
	}
}

func TestRunCallAllPagesUsesCatalogMaximumAndStreamsRecords(t *testing.T) {
	var requestedPages []int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		page := int(body["pageNum"].(float64))
		requestedPages = append(requestedPages, page)
		if got := int(body["pageSize"].(float64)); got != 100 {
			t.Fatalf("pageSize=%d want catalog max 100", got)
		}
		_, _ = w.Write([]byte(`{"status":1,"resultCode":0,"message":"ok","data":{"total":2,"recordList":[{"id":1},{"id":2}]}}`))
	}))
	defer server.Close()

	var out, stderr bytes.Buffer
	f := &Factory{Out: &out, Err: &stderr, AllPages: true}
	f.clientFactory = func() (*client.Client, error) {
		return client.New(server.URL, "key", "secret", sign.V2, "test"), nil
	}
	if err := f.RunCall("getCarOutList", `{"parkCode":"P","pageNum":9,"pageSize":1}`); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "{\"id\":1}\n{\"id\":2}\n"; got != want {
		t.Fatalf("output=%q want %q", got, want)
	}
	if len(requestedPages) != 1 || requestedPages[0] != 1 {
		t.Fatalf("requested pages=%v", requestedPages)
	}
}

func TestRunCallAllPagesWritesPrivateNDJSONFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":1,"resultCode":0,"message":"ok","data":{"total":1,"recordList":[{"id":1}]}}`))
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "records.ndjson")
	f := &Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}, AllPages: true, OutFile: path}
	f.clientFactory = func() (*client.Client, error) {
		return client.New(server.URL, "key", "secret", sign.V2, "test"), nil
	}
	if err := f.RunCall("getCarOutList", `{"parkCode":"P"}`); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "{\"id\":1}\n"; got != want {
		t.Fatalf("file=%q want %q", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("mode=%o want 600", got)
	}
}

func TestRunCallRejectsOutWithoutAllPages(t *testing.T) {
	f := &Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}, OutFile: "records.ndjson"}
	err := f.RunCall("getCarOutList", `{"parkCode":"P"}`)
	if err == nil || !strings.Contains(err.Error(), "--out 仅与 --all-pages") {
		t.Fatalf("err=%v", err)
	}
}

// TestConfirmWriteReadOnly verifies that ConfirmWrite returns a read-only error
// when the factory is in read-only mode, regardless of --yes.
func TestConfirmWriteReadOnly(t *testing.T) {
	// ReadOnly=true → should get read-only error immediately.
	f := &Factory{ReadOnly: true}
	err := f.ConfirmWrite("x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !isReadOnlyErr(err) {
		t.Fatalf("expected read-only error, got %v", err)
	}

	// ReadOnly=true + Yes=true → still read-only (yes cannot override read-only).
	f2 := &Factory{ReadOnly: true, Yes: true}
	err2 := f2.ConfirmWrite("x")
	if err2 == nil {
		t.Fatal("expected error even with --yes, got nil")
	}
	if !isReadOnlyErr(err2) {
		t.Fatalf("expected read-only error with --yes, got %v", err2)
	}
}

// TestBuildErrorInfoMissingParam verifies the "必填集对比" branch:
// status=7 + body missing a required scalar field → ei.Field set to that param name.
// Uses payParkFee which has parkingCode as its first required non-group scalar param.
func TestBuildErrorInfoMissingParam(t *testing.T) {
	// body that has no required fields at all
	emptyBody := `{}`
	resp := &client.Response{
		Status:  client.StatusBadParams,
		Message: "请求参数不完整",
	}
	ei := buildErrorInfo("payParkFee", emptyBody, resp)
	if ei.Field == "" {
		t.Fatal("status=7 + body 缺必填 → ei.Field 应被定位到第一个缺失必填参数,got empty")
	}
	// parkingCode is the first required non-group scalar for payParkFee
	if ei.Field != "parkingCode" {
		t.Fatalf("expected ei.Field=parkingCode, got %q", ei.Field)
	}
	if !ei.FieldRequired {
		t.Fatal("ei.FieldRequired 应为 true")
	}
}

func TestBuildErrorInfoSessionExpiredHint(t *testing.T) {
	// 908「会话已过期」: resultCode 太泛(其它错误),应回退到 MessageHint 给出针对性提示
	resp := &client.Response{
		Status:     client.StatusBizFail,
		ResultCode: 908,
		Message:    "会话已过期",
	}
	ei := buildErrorInfo("correctCarOnChannel", `{"parkCode":"P","channelCode":"C"}`, resp)
	if ei.Hint == "" {
		t.Fatal("908 会话已过期 应有文案级 hint, got empty")
	}
	if !strings.Contains(ei.Hint, "channel-snap") {
		t.Fatalf("hint 应指向 channel-snap, got %q", ei.Hint)
	}
}
