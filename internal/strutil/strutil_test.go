package strutil

import "testing"

func TestSubCmd(t *testing.T) {
	cases := []struct {
		domain, cmd, want string
	}{
		// evcharge cmds repeat the domain name -> strip the redundant prefix.
		{"evcharge", "evchargeStationList", "station-list"},
		{"evcharge", "evchargeStationDetail", "station-detail"},
		{"evcharge", "evchargeOrderList", "order-list"},
		{"evcharge", "evchargeOrderDetail", "order-detail"},
		{"evcharge", "evchargePileList", "pile-list"},
		{"evcharge", "evchargeStationStatistics", "station-statistics"},

		// Verb-prefixed cmds of other domains must be returned unchanged: the
		// kebab does not start with "<domain>-", so nothing is stripped.
		{"trade", "getParkFee", "get-park-fee"},
		{"trade", "payParkFee", "pay-park-fee"},
		{"parking", "supplementParkingRecordIn", "supplement-parking-record-in"},
		{"coupon", "createCouponTemplate", "create-coupon-template"},
		{"data", "getBillSummary", "get-bill-summary"},

		// Guards: empty domain, and a cmd equal to the domain (no remainder).
		{"", "evchargeStationList", "evcharge-station-list"},
		{"evcharge", "evcharge", "evcharge"},
	}
	for _, c := range cases {
		if got := SubCmd(c.domain, c.cmd); got != c.want {
			t.Errorf("SubCmd(%q, %q) = %q, want %q", c.domain, c.cmd, got, c.want)
		}
	}
}
