package config

import (
	"bytes"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

func TestSetDefaultUpdatesCurrentProfile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("OPENYDT_PROFILE", "")
	seed := &config.Config{CurrentProfile: "test"}
	seed.Upsert(config.Profile{Name: "test", Key: "k", Secret: "s", Env: "test"})
	if err := seed.Save(); err != nil {
		t.Fatal(err)
	}

	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	cmd := New(f)
	cmd.SetArgs([]string{"set-default", "--park", "PTD2YBBZ", "--car-no", "桂566666"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("set-default: %v", err)
	}

	got, _ := config.Load()
	p, _ := got.Find("test")
	if p.DefaultPark != "PTD2YBBZ" || p.DefaultCarNo != "桂566666" {
		t.Fatalf("defaults not saved: %+v", p)
	}
	if p.Key != "k" || p.Secret != "s" {
		t.Fatalf("creds clobbered: %+v", p)
	}
}

func TestSetDefaultRequiresAtLeastOne(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	seed := &config.Config{CurrentProfile: "test"}
	seed.Upsert(config.Profile{Name: "test", Key: "k", Secret: "s"})
	_ = seed.Save()
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	cmd := New(f)
	cmd.SetArgs([]string{"set-default"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error when neither --park nor --car-no given")
	}
}

func TestConfigSetPreservesDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("OPENYDT_PROFILE", "")
	seed := &config.Config{CurrentProfile: "test"}
	seed.Upsert(config.Profile{Name: "test", Key: "k", Secret: "s", Env: "test", DefaultPark: "PTD2YBBZ", DefaultCarNo: "桂566666"})
	if err := seed.Save(); err != nil {
		t.Fatal(err)
	}

	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	cmd := New(f)
	cmd.SetArgs([]string{"set", "--profile", "test", "--key", "newk", "--secret", "news", "--env", "test"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config set: %v", err)
	}

	got, _ := config.Load()
	p, _ := got.Find("test")
	if p.Key != "newk" || p.Secret != "news" {
		t.Fatalf("creds not updated: %+v", p)
	}
	if p.DefaultPark != "PTD2YBBZ" || p.DefaultCarNo != "桂566666" {
		t.Fatalf("config set clobbered defaults: %+v", p)
	}
}
