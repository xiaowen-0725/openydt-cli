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
