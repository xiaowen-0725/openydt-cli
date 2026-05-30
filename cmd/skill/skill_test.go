package skill

import (
	"bytes"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/skillsync"
)

func TestSkillCommandWiring(t *testing.T) {
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	cmd := New(f)
	if cmd.Use != "skill" {
		t.Fatalf("Use = %q", cmd.Use)
	}
	sub, _, err := cmd.Find([]string{"sync"})
	if err != nil || sub.Use != "sync" {
		t.Fatalf("sync subcommand missing: %v", err)
	}
	// skill 命令必须有空 PersistentPreRun 以退出根命令的自动触发
	if cmd.PersistentPreRun == nil {
		t.Fatalf("skill cmd must define a (no-op) PersistentPreRun to opt out of auto-trigger")
	}
}

func TestSyncSuccess(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	skillsync.SetRunnerOverrideForTest(func(args ...string) *skillsync.RunResult {
		return &skillsync.RunResult{}
	})
	defer skillsync.SetRunnerOverrideForTest(nil)

	errOut := &bytes.Buffer{}
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: errOut}
	cmd := New(f)
	cmd.SetArgs([]string{"sync"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync execute: %v", err)
	}
	st, _ := skillsync.ReadState()
	if st == nil || st.Version != cmdutil.Version {
		t.Fatalf("sync should RecordSuccess; state=%+v", st)
	}
}

func TestSyncFailureReturnsError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	skillsync.SetRunnerOverrideForTest(func(args ...string) *skillsync.RunResult {
		r := &skillsync.RunResult{}
		r.Err = errForTest("boom")
		return r
	})
	defer skillsync.SetRunnerOverrideForTest(nil)

	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	cmd := New(f)
	cmd.SetArgs([]string{"sync", "--quiet"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("sync failure should return error")
	}
}

type errForTest string

func (e errForTest) Error() string { return string(e) }
