package skillsync

import (
	"strings"
	"testing"
)

func TestRunSyncArgs(t *testing.T) {
	var captured []string
	runnerOverride = func(args ...string) *RunResult {
		captured = args
		return &RunResult{}
	}
	defer func() { runnerOverride = nil }()

	RunSync(false, "0.4.2")
	want := "-y skills add https://github.com/xiaowen-0725/openydt-cli/tree/v0.4.2 -g --copy -y"
	if got := strings.Join(captured, " "); got != want {
		t.Fatalf("RunSync(false) args = %q, want %q", got, want)
	}

	RunSync(true, "0.4.2")
	wantForce := "-y skills add https://github.com/xiaowen-0725/openydt-cli/tree/v0.4.2 -g --copy --all"
	if got := strings.Join(captured, " "); got != wantForce {
		t.Fatalf("RunSync(true) args = %q, want %q", got, wantForce)
	}
}

func TestSourceConst(t *testing.T) {
	if selectSource("0.4.2", "/package/skills") != "/package/skills" {
		t.Fatalf("bundled Skills must take precedence")
	}
	if sourceWithoutBundle("0.4.2") != "https://github.com/xiaowen-0725/openydt-cli/tree/v0.4.2" {
		t.Fatalf("sourceWithoutBundle() = %q", sourceWithoutBundle("0.4.2"))
	}
	if sourceWithoutBundle("dev") != "xiaowen-0725/openydt-cli" {
		t.Fatalf("dev sourceWithoutBundle() = %q", sourceWithoutBundle("dev"))
	}
}
