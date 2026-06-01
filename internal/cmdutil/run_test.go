package cmdutil

import (
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/client"
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
