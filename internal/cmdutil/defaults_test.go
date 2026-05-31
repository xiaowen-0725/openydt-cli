package cmdutil

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/catalog"
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

func TestSchemaDefaultParse(t *testing.T) {
	cases := map[string]string{
		"记录类型（0未定义，1有牌车）默认1":       "1",
		"车牌颜色：0其他，1蓝色（默认为1）":       "1",
		"整场或区域：0 整场，1区域":           "", // no documented default
		"每页多少条，最多1000条，默认10":       "10",
		"returnCarImage 0不返回 1返回 默认 0": "0",
	}
	for desc, want := range cases {
		if got := schemaDefault(desc); got != want {
			t.Errorf("schemaDefault(%q) = %q, want %q", desc, got, want)
		}
	}
}

func TestInjectSchemaDefaultsPure(t *testing.T) {
	params := []catalog.Param{
		{Name: "parkCode", Type: "String", Desc: "停车场编号"},
		{Name: "recordType", Type: "Integer", Desc: "记录类型（…5误触发）默认1"},
		{Name: "pageSize", Type: "Integer", Desc: "每页多少条，默认10"},
		{Name: "nested", Type: "String", Desc: "子字段默认9", Group: "couponList"}, // skipped: nested
	}

	out, inj := injectSchemaDefaults(`{}`, params)
	if !strings.Contains(out, `"recordType":1`) || !strings.Contains(out, `"pageSize":10`) {
		t.Fatalf("expected recordType/pageSize defaults, got %s", out)
	}
	if strings.Contains(out, "nested") || strings.Contains(out, "parkCode") {
		t.Fatalf("must not inject nested or no-default fields, got %s", out)
	}
	if len(inj) != 2 {
		t.Fatalf("expected 2 injected, got %v", inj)
	}

	out, inj = injectSchemaDefaults(`{"recordType":2}`, params)
	if !strings.Contains(out, `"recordType":2`) {
		t.Fatalf("must not overwrite explicit recordType, got %s", out)
	}
	for _, n := range inj {
		if n == "recordType" {
			t.Fatalf("recordType must not be reported injected when present, got %v", inj)
		}
	}
}

func TestApplySchemaDefaultsRealCatalog(t *testing.T) {
	f := &Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	got := f.applySchemaDefaults("supplementParkingRecordIn", `{"parkCode":"PTD2YBBZ"}`)
	if !strings.Contains(got, `"recordType":1`) {
		t.Fatalf("expected recordType=1 injected for supplementParkingRecordIn, got %s", got)
	}
	if got := f.applySchemaDefaults("no_such_cmd_xyz", "{}"); got != "{}" {
		t.Fatalf("unknown cmd => unchanged, got %s", got)
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
