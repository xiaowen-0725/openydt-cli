package catalog

import "testing"

func TestHints(t *testing.T) {
	read := Iface{Cmd: "getParkFee", ReadWrite: "read"}
	if h := read.Hints(); !h.ReadOnly || h.Destructive || !h.Idempotent {
		t.Fatalf("read hints=%+v", h)
	}
	del := Iface{Cmd: "deleteTrader", ReadWrite: "write"}
	if h := del.Hints(); h.ReadOnly || !h.Destructive {
		t.Fatalf("delete hints=%+v", h)
	}
	pay := Iface{Cmd: "payParkFee", ReadWrite: "write",
		Params: []Param{{Name: "billCode", Required: true}}}
	if h := pay.Hints(); !h.Idempotent || h.IdempotencyKey != "billCode" {
		t.Fatalf("pay hints=%+v", h)
	}
	supp := Iface{Cmd: "supplementParkingRecordIn", ReadWrite: "write"}
	if h := supp.Hints(); h.Idempotent {
		t.Fatalf("无幂等键的写不应标 idempotent: %+v", h)
	}
	unfreeze := Iface{Cmd: "unFreezeMonthTicket", ReadWrite: "write"}
	if h := unfreeze.Hints(); h.Destructive {
		t.Fatalf("unFreezeMonthTicket 是恢复操作,不应标 destructive: %+v", h)
	}
}

func TestPaginationSpec(t *testing.T) {
	it := Iface{Params: []Param{
		{Name: "pageNum", Type: "Integer", Desc: "第几页，从1开始"},
		{Name: "pageSize", Type: "Integer", Desc: "每页多少条，最多100条"},
	}}
	spec, ok := it.Pagination()
	if !ok {
		t.Fatal("expected pageable interface")
	}
	if spec.PageField != "pageNum" || spec.SizeField != "pageSize" || spec.MaxPageSize != 100 {
		t.Fatalf("spec=%+v", spec)
	}

	for _, tc := range []struct {
		desc string
		want int
	}{
		{"每页有多少条(默认10，最大1000)", 1000},
		{"每页多少条，取值范围：1-100", 100},
		{"pageSize 不能大于 50", 50},
	} {
		got := maxPageSize(tc.desc)
		if got != tc.want {
			t.Errorf("maxPageSize(%q)=%d want %d", tc.desc, got, tc.want)
		}
	}

	if _, ok := (Iface{Params: []Param{{Name: "pageSize"}}}).Pagination(); ok {
		t.Fatal("pageNum 缺失时不应标记为分页接口")
	}

	swapped := Iface{Params: []Param{
		{Name: "pageSize", Desc: "第几页，从1开始"},
		{Name: "pageNum", Desc: "每页多少条，最多1000条"},
	}}
	if spec, ok := swapped.Pagination(); !ok || spec.MaxPageSize != 1000 {
		t.Fatalf("swapped doc descriptions should still derive max: %+v ok=%v", spec, ok)
	}
}
