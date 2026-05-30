package config

import "testing"

func TestProfileDefaultsRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := &Config{}
	cfg.Upsert(Profile{Name: "test", Key: "k", Secret: "s", Env: "test", DefaultPark: "PTD2YBBZ", DefaultCarNo: "桂566666"})
	cfg.CurrentProfile = "test"
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	p, ok := got.Find("test")
	if !ok || p.DefaultPark != "PTD2YBBZ" || p.DefaultCarNo != "桂566666" {
		t.Fatalf("defaults round trip mismatch: %+v ok=%v", p, ok)
	}
}

func TestActiveByPrecedence(t *testing.T) {
	t.Setenv("OPENYDT_PROFILE", "")
	cfg := &Config{
		CurrentProfile: "test",
		Profiles: []Profile{
			{Name: "test", DefaultPark: "P_CUR"},
			{Name: "prod", DefaultPark: "P_PROD"},
		},
	}
	if p, ok := cfg.Active("prod"); !ok || p.DefaultPark != "P_PROD" {
		t.Fatalf("flag should win, got %+v", p)
	}
	if p, ok := cfg.Active(""); !ok || p.DefaultPark != "P_CUR" {
		t.Fatalf("should fall back to CurrentProfile, got %+v", p)
	}
	t.Setenv("OPENYDT_PROFILE", "prod")
	if p, ok := cfg.Active(""); !ok || p.DefaultPark != "P_PROD" {
		t.Fatalf("OPENYDT_PROFILE should win over CurrentProfile, got %+v", p)
	}
}

func TestActiveNoneFound(t *testing.T) {
	t.Setenv("OPENYDT_PROFILE", "")
	cfg := &Config{}
	if _, ok := cfg.Active(""); ok {
		t.Fatalf("expected no active profile")
	}
}
