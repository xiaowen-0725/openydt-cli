package selfupdate

import (
	"testing"
	"time"
)

func withTriggerSeams(t *testing.T) *int {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, key := range []string{
		"CI", "GITHUB_ACTIONS", "BUILD_NUMBER", "GITLAB_CI",
		"OPENYDT_NO_UPDATE_CHECK", "OPENYDT_UPDATE_CHECK_CHILD",
	} {
		t.Setenv(key, "")
	}
	calls := 0
	spawnFunc = func() error { calls++; return nil }
	nowFunc = func() time.Time { return time.Unix(1_000_000, 0).UTC() }
	t.Cleanup(func() {
		spawnFunc = spawnBackgroundCheck
		nowFunc = time.Now
		setPending(nil)
	})
	return &calls
}

func TestMaybeTriggerColdStartSpawnsWithoutBlocking(t *testing.T) {
	calls := withTriggerSeams(t)
	MaybeTrigger("0.4.1")
	if *calls != 1 {
		t.Fatalf("spawn calls = %d, want 1", *calls)
	}
	if Pending() != nil {
		t.Fatal("cold state must not produce an empty update notice")
	}
	MaybeTrigger("0.4.1")
	if *calls != 1 {
		t.Fatalf("recent attempt should debounce, calls = %d", *calls)
	}
}

func TestMaybeTriggerNotifiesOncePerLatestVersion(t *testing.T) {
	withTriggerSeams(t)
	if err := WriteState(State{
		LatestVersion: "0.4.2",
		CheckedAt:     nowFunc().Format(time.RFC3339),
	}); err != nil {
		t.Fatal(err)
	}

	MaybeTrigger("0.4.1")
	if notice := Pending(); notice == nil || notice.LatestVersion != "0.4.2" {
		t.Fatalf("expected update notice, got %+v", notice)
	}
	MaybeTrigger("0.4.1")
	if Pending() != nil {
		t.Fatal("same version notice must not repeat")
	}
}

func TestMaybeTriggerSkipsRecentCheck(t *testing.T) {
	calls := withTriggerSeams(t)
	if err := WriteState(State{CheckedAt: nowFunc().Add(-time.Hour).Format(time.RFC3339)}); err != nil {
		t.Fatal(err)
	}
	MaybeTrigger("0.4.1")
	if *calls != 0 {
		t.Fatalf("recent check spawned %d times", *calls)
	}
}

func TestMaybeTriggerHonorsOptOut(t *testing.T) {
	calls := withTriggerSeams(t)
	t.Setenv("OPENYDT_NO_UPDATE_CHECK", "1")
	MaybeTrigger("0.4.1")
	if *calls != 0 {
		t.Fatalf("opt-out spawned %d times", *calls)
	}
}
