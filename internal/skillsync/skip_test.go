package skillsync

import "testing"

func TestShouldSkip(t *testing.T) {
	// 清掉可能存在的 CI 变量,保证基线可控
	for _, k := range []string{"CI", "GITHUB_ACTIONS", "BUILD_NUMBER", "GITLAB_CI", "OPENYDT_NO_SKILLS_SYNC", "OPENYDT_SKILL_SYNC_CHILD"} {
		t.Setenv(k, "")
	}

	if shouldSkip("0.1.1") {
		t.Fatalf("release version on clean env should NOT skip")
	}

	t.Setenv("OPENYDT_NO_SKILLS_SYNC", "1")
	if !shouldSkip("0.1.1") {
		t.Fatalf("opt-out env should skip")
	}
	t.Setenv("OPENYDT_NO_SKILLS_SYNC", "")

	t.Setenv("OPENYDT_SKILL_SYNC_CHILD", "1")
	if !shouldSkip("0.1.1") {
		t.Fatalf("child marker should skip")
	}
	t.Setenv("OPENYDT_SKILL_SYNC_CHILD", "")

	t.Setenv("CI", "true")
	if !shouldSkip("0.1.1") {
		t.Fatalf("CI should skip")
	}
	t.Setenv("CI", "")

	for _, v := range []string{"", "dev", "DEV", "0.1.1-3-gabc", "v1.2", "garbage"} {
		if !shouldSkip(v) {
			t.Fatalf("non-release version %q should skip", v)
		}
	}
	for _, v := range []string{"0.1.1", "v0.1.1", "V2.0.0", "10.20.30"} {
		if shouldSkip(v) {
			t.Fatalf("release version %q should NOT skip", v)
		}
	}
}
