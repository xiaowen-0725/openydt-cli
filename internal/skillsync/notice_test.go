package skillsync

import (
	"strings"
	"testing"
)

func TestNoticePending(t *testing.T) {
	setPending(nil)
	if Pending() != nil {
		t.Fatalf("expected nil pending")
	}
	setPending(&Notice{Reason: "未检测到 npx/node"})
	n := Pending()
	if n == nil || !strings.Contains(n.Message(), "skills add") {
		t.Fatalf("notice message missing fix command: %+v", n)
	}
	if !strings.Contains(n.Message(), "未检测到 npx/node") {
		t.Fatalf("notice message missing reason: %s", n.Message())
	}
	setPending(nil)
}
