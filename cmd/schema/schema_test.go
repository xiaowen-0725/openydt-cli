package schema

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
)

func TestSchemaShowsSupplementalResponseEnums(t *testing.T) {
	var stdout, stderr bytes.Buffer
	f := &cmdutil.Factory{Out: &stdout, Err: &stderr}
	cmd := New(f)
	cmd.SetArgs([]string{"getCarOutList"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"响应枚举", "leaveType", "4 遥控开闸", "11 欠费放行",
		"98 盘点离场", "业务语义", "离场记录事件全集",
		"6 可疑跟车 [物理离场, 逃费]",
		"97 重复进场在场车辆强制离场 [逻辑闭环, 经营统计排除]",
	} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("schema output missing %q:\n%s", expected, stdout.String())
		}
	}
	if strings.Contains(stdout.String(), "needsSettle") {
		t.Fatalf("internal settlement field leaked into schema:\n%s", stdout.String())
	}
}

func TestSchemaJSONIncludesSupplementalResponseEnums(t *testing.T) {
	var stdout, stderr bytes.Buffer
	f := &cmdutil.Factory{Out: &stdout, Err: &stderr}
	cmd := New(f)
	cmd.SetArgs([]string{"getCarOutList", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var result struct {
		ResponseEnums []struct {
			Fields []string `json:"fields"`
			Values []struct {
				Code int    `json:"code"`
				Name string `json:"name"`
			} `json:"values"`
		} `json:"responseEnums"`
		DomainSemantics struct {
			RecordSet     string `json:"recordSet"`
			Deduplication struct {
				Key     string `json:"key"`
				OrderBy string `json:"orderBy"`
				Keep    string `json:"keep"`
			} `json:"deduplication"`
			Fields []struct {
				Field  string `json:"field"`
				Values []struct {
					Code              int      `json:"code"`
					EventNature       string   `json:"eventNature"`
					TrafficTreatment  string   `json:"trafficTreatment"`
					DurationTreatment string   `json:"durationTreatment"`
					BusinessTags      []string `json:"businessTags"`
				} `json:"values"`
			} `json:"fields"`
		} `json:"domainSemantics"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.ResponseEnums) != 1 {
		t.Fatalf("responseEnums=%+v", result.ResponseEnums)
	}
	if got := strings.Join(result.ResponseEnums[0].Fields, ","); got != "enterType,leaveType" {
		t.Fatalf("fields=%q", got)
	}
	found := false
	for _, value := range result.ResponseEnums[0].Values {
		if value.Code == 98 && value.Name == "盘点离场" {
			found = true
		}
		if value.Code == 99 || value.Code == 100 {
			t.Fatalf("excluded response enum leaked into schema: %+v", value)
		}
	}
	if !found {
		t.Fatal("response enum missing code 98")
	}
	if result.DomainSemantics.RecordSet != "离场记录事件全集" {
		t.Fatalf("recordSet=%q", result.DomainSemantics.RecordSet)
	}
	if got := result.DomainSemantics.Deduplication; got.Key != "parkingCode" || got.OrderBy != "leaveTime" || got.Keep != "latest" {
		t.Fatalf("deduplication=%+v", got)
	}
	if len(result.DomainSemantics.Fields) != 1 || result.DomainSemantics.Fields[0].Field != "leaveType" {
		t.Fatalf("domain semantics fields=%+v", result.DomainSemantics.Fields)
	}
	semantics := map[int]struct {
		Nature, Traffic, Duration string
		Tags                      []string
	}{}
	for _, value := range result.DomainSemantics.Fields[0].Values {
		semantics[value.Code] = struct {
			Nature, Traffic, Duration string
			Tags                      []string
		}{value.EventNature, value.TrafficTreatment, value.DurationTreatment, value.BusinessTags}
	}
	if got := semantics[97]; got.Nature != "logical_closure" || got.Traffic != "exclude" || got.Duration != "exclude" {
		t.Fatalf("code 97 semantics=%+v", got)
	}
	if got := semantics[6]; got.Nature != "physical_departure" || !contains(got.Tags, "escape") {
		t.Fatalf("code 6 semantics=%+v", got)
	}
	if bytes.Contains(stdout.Bytes(), []byte("needsSettle")) {
		t.Fatalf("internal settlement field leaked into JSON: %s", stdout.String())
	}
}

func TestSchemaDistinguishesGateOperationFromDeparture(t *testing.T) {
	var stdout, stderr bytes.Buffer
	f := &cmdutil.Factory{Out: &stdout, Err: &stderr}
	cmd := New(f)
	cmd.SetArgs([]string{"getAbnormalOpenGateList", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var result struct {
		DomainSemantics struct {
			RecordSet        string `json:"recordSet"`
			ImpliesDeparture bool   `json:"impliesDeparture"`
			Fields           []struct {
				Values []struct {
					Code         int      `json:"code"`
					EventNature  string   `json:"eventNature"`
					BusinessTags []string `json:"businessTags"`
				} `json:"values"`
			} `json:"fields"`
		} `json:"domainSemantics"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.DomainSemantics.RecordSet != "非系统开闸操作流水" || result.DomainSemantics.ImpliesDeparture {
		t.Fatalf("domainSemantics=%+v", result.DomainSemantics)
	}
	if len(result.DomainSemantics.Fields) != 1 {
		t.Fatalf("fields=%+v", result.DomainSemantics.Fields)
	}
	foundRemoteController := false
	for _, value := range result.DomainSemantics.Fields[0].Values {
		if value.Code == 1 && value.EventNature == "gate_operation" && contains(value.BusinessTags, "physical_remote_controller") {
			foundRemoteController = true
		}
	}
	if !foundRemoteController {
		t.Fatalf("missing physical remote controller semantics: %+v", result.DomainSemantics.Fields[0].Values)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
