package skillsync

import (
	"testing"
	"time"
)

// withTriggerSeams 安装可控的 spawn / npx / clock seam,并在测试结束还原。
func withTriggerSeams(t *testing.T, npx bool) *int {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, k := range []string{"CI", "GITHUB_ACTIONS", "BUILD_NUMBER", "GITLAB_CI", "OPENYDT_NO_SKILLS_SYNC", "OPENYDT_SKILL_SYNC_CHILD"} {
		t.Setenv(k, "")
	}
	calls := 0
	spawnFunc = func(string) error { calls++; return nil }
	npxAvailableFunc = func() bool { return npx }
	nowFunc = func() time.Time { return time.Unix(0, 0).UTC() }
	t.Cleanup(func() {
		spawnFunc = spawnBackgroundSync
		npxAvailableFunc = npxAvailable
		nowFunc = time.Now
		setPending(nil)
	})
	return &calls
}

func TestMaybeTriggerColdStartSpawns(t *testing.T) {
	calls := withTriggerSeams(t, true)
	MaybeTrigger("0.1.1")
	if *calls != 1 {
		t.Fatalf("cold start: spawn calls = %d, want 1", *calls)
	}
	// 第二次同版本应被防抖跳过
	MaybeTrigger("0.1.1")
	if *calls != 1 {
		t.Fatalf("debounce: spawn calls = %d, want 1", *calls)
	}
}

func TestMaybeTriggerInSyncNoSpawn(t *testing.T) {
	calls := withTriggerSeams(t, true)
	if err := RecordSuccess("0.1.1"); err != nil {
		t.Fatal(err)
	}
	MaybeTrigger("0.1.1")
	if *calls != 0 {
		t.Fatalf("in-sync: spawn calls = %d, want 0", *calls)
	}
}

func TestMaybeTriggerDriftSpawns(t *testing.T) {
	calls := withTriggerSeams(t, true)
	if err := RecordSuccess("0.1.0"); err != nil {
		t.Fatal(err)
	}
	MaybeTrigger("0.1.1")
	if *calls != 1 {
		t.Fatalf("drift: spawn calls = %d, want 1", *calls)
	}
}

func TestMaybeTriggerNoNpxSetsNotice(t *testing.T) {
	calls := withTriggerSeams(t, false)
	MaybeTrigger("0.1.1")
	if *calls != 0 {
		t.Fatalf("no npx: spawn calls = %d, want 0", *calls)
	}
	if Pending() == nil {
		t.Fatalf("no npx: expected degraded notice")
	}
}

func TestMaybeTriggerSkipsWhenOptOut(t *testing.T) {
	calls := withTriggerSeams(t, true)
	t.Setenv("OPENYDT_NO_SKILLS_SYNC", "1")
	MaybeTrigger("0.1.1")
	if *calls != 0 {
		t.Fatalf("opt-out: spawn calls = %d, want 0", *calls)
	}
}
