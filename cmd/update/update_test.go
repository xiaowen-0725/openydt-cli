package update

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/selfupdate"
)

func TestCheckJSON(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	check := func(context.Context, string) (selfupdate.Result, error) {
		return selfupdate.Result{
			CurrentVersion:  "0.4.1",
			LatestVersion:   "0.4.2",
			UpdateAvailable: true,
		}, nil
	}
	cmd := newCommand(f, check, nil)
	cmd.SetArgs([]string{"check", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := f.Out.(*bytes.Buffer).String(); !strings.Contains(got, `"latestVersion":"0.4.2"`) {
		t.Fatalf("unexpected JSON: %s", got)
	}
}

func TestUpdateInstallsExactLatestVersion(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	check := func(context.Context, string) (selfupdate.Result, error) {
		return selfupdate.Result{CurrentVersion: "0.4.1", LatestVersion: "0.4.2", UpdateAvailable: true}, nil
	}
	installed := ""
	install := func(_ context.Context, version string, _, _ io.Writer) error {
		installed = version
		return nil
	}
	cmd := newCommand(f, check, install)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if installed != "0.4.2" {
		t.Fatalf("installed version = %q", installed)
	}
}

func TestUpdateDoesNothingWhenCurrent(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	check := func(context.Context, string) (selfupdate.Result, error) {
		return selfupdate.Result{CurrentVersion: "0.4.2", LatestVersion: "0.4.2"}, nil
	}
	called := false
	install := func(_ context.Context, _ string, _, _ io.Writer) error {
		called = true
		return nil
	}
	cmd := newCommand(f, check, install)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("installer must not run when already current")
	}
}
