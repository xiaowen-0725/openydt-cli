package cmdutil

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

func TestInjectDefaultsPure(t *testing.T) {
	hasPark := func(n string) bool { return n == "parkCode" }

	out, inj := injectDefaults(`{}`, "PTD2YBBZ", "", hasPark)
	if !strings.Contains(out, `"parkCode":"PTD2YBBZ"`) || len(inj) != 1 {
		t.Fatalf("expected parkCode injected, got %s %v", out, inj)
	}

	out, inj = injectDefaults(`{"parkCode":"X"}`, "PTD2YBBZ", "", hasPark)
	if out != `{"parkCode":"X"}` || len(inj) != 0 {
		t.Fatalf("must not overwrite existing, got %s %v", out, inj)
	}

	out, inj = injectDefaults(`{}`, "PTD2YBBZ", "", func(string) bool { return false })
	if out != `{}` || len(inj) != 0 {
		t.Fatalf("must not inject when param absent, got %s %v", out, inj)
	}

	out, _ = injectDefaults(`{}`, "", "", hasPark)
	if out != `{}` {
		t.Fatalf("no defaults => unchanged, got %s", out)
	}

	hasCar := func(n string) bool { return n == "carCode" }
	out, inj = injectDefaults(`{}`, "", "桂566666", hasCar)
	if !strings.Contains(out, `"carCode":"桂566666"`) || len(inj) != 1 {
		t.Fatalf("expected carCode injected, got %s %v", out, inj)
	}
}

func TestApplyDefaultsWithRealCatalog(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("OPENYDT_PROFILE", "")
	cfg := &config.Config{CurrentProfile: "test"}
	cfg.Upsert(config.Profile{Name: "test", Key: "k", Secret: "s", Env: "test", DefaultPark: "PTD2YBBZ"})
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	f := &Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

	if got := f.applyDefaults("getBillSummary", "{}"); !strings.Contains(got, `"parkCode":"PTD2YBBZ"`) {
		t.Fatalf("expected parkCode injected for getBillSummary, got %s", got)
	}
	if got := f.applyDefaults("getBillSummary", `{"parkCode":"OTHER"}`); !strings.Contains(got, `"OTHER"`) || strings.Contains(got, "PTD2YBBZ") {
		t.Fatalf("must not overwrite explicit parkCode, got %s", got)
	}
	if got := f.applyDefaults("no_such_cmd_xyz", "{}"); got != "{}" {
		t.Fatalf("unknown cmd => unchanged, got %s", got)
	}
}
