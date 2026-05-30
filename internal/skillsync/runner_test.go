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

	RunSync(false)
	want := "-y skills add xiaowen-0725/openydt-cli -g -y"
	if got := strings.Join(captured, " "); got != want {
		t.Fatalf("RunSync(false) args = %q, want %q", got, want)
	}

	RunSync(true)
	wantForce := "-y skills add xiaowen-0725/openydt-cli -g --all"
	if got := strings.Join(captured, " "); got != wantForce {
		t.Fatalf("RunSync(true) args = %q, want %q", got, wantForce)
	}
}

func TestSourceConst(t *testing.T) {
	if Source() != "xiaowen-0725/openydt-cli" {
		t.Fatalf("Source() = %q", Source())
	}
}
