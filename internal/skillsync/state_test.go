package skillsync

import (
	"os"
	"testing"
	"time"
)

func writeRaw(p string, b []byte) error { return os.WriteFile(p, b, 0o644) }

func TestNormalizeVersion(t *testing.T) {
	cases := map[string]string{
		"v0.1.1": "0.1.1", "V0.1.1": "0.1.1", "0.1.1": "0.1.1", " v0.1.1 ": "0.1.1",
	}
	for in, want := range cases {
		if got := normalizeVersion(in); got != want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStateRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if s, err := ReadState(); err != nil || s != nil {
		t.Fatalf("cold read: got (%v,%v), want (nil,nil)", s, err)
	}

	in := State{Version: "0.1.1", LastAttemptVersion: "0.1.1", Skills: []string{"openydt-billing"}, UpdatedAt: "2026-05-30T00:00:00Z"}
	if err := WriteState(in); err != nil {
		t.Fatalf("WriteState: %v", err)
	}
	got, err := ReadState()
	if err != nil || got == nil {
		t.Fatalf("ReadState after write: (%v,%v)", got, err)
	}
	if got.Version != "0.1.1" || got.LastAttemptVersion != "0.1.1" || len(got.Skills) != 1 {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

func TestReadCorruptStateErrors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := WriteState(State{Version: "0.1.1"}); err != nil {
		t.Fatal(err)
	}
	p, _ := statePath()
	if err := writeRaw(p, []byte("{not json")); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadState(); err == nil {
		t.Fatalf("expected error on corrupt state")
	}
}

func TestRecordSuccess(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	nowFunc = func() time.Time { return time.Unix(0, 0).UTC() }
	defer func() { nowFunc = time.Now }()
	if err := RecordSuccess("v0.1.2"); err != nil {
		t.Fatal(err)
	}
	got, _ := ReadState()
	if got.Version != "v0.1.2" || got.LastAttemptVersion != "v0.1.2" {
		t.Fatalf("RecordSuccess state = %+v", got)
	}
}
