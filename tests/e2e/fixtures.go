package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xiaowen-0725/openydt-cli/internal/client"
)

// Fixtures are real values from the data-rich cloud park "智汇云测试专用车场123412"
// (PTD2YBBZ). Defaults are harvested via read APIs / SaaS backend; Discover()
// refreshes the volatile ones (on-site car, trader, month-ticket, coupon) at runtime.
type Fixtures struct {
	Park          string // 云车场, 授权商范围内, 有通道/设备/月票/商家
	CarNo         string // 在场车牌
	ParkingCode   string // 在场车停车流水号
	ChannelOut    string // 出口通道 customCode
	ChannelIn     string // 入口通道 customCode
	ChannelID     int    // 通道 id
	TraderCode    string // 商家编码
	TraderAccount string // 商家登录账号
	MonthCfgID    string // 月票类型 id (monthTicketConfigId)
	MonthTicketID string // 已开通月票 id
	VipCarNo      string // 已有月票的车牌
	TopBillCode   string // 月票账单 code
	CouponCode    string // 已发放电子券 code(运行时发现)
	SpecialTypeID string // 特殊车辆类型 id(黑名单/访客用)
	TraderID      int    // 商家 id(Long)
	AreaID        int    // 车场区域 id
	SellBillID    int    // 电子券销售单 id
}

func defaultFixtures() *Fixtures {
	return &Fixtures{
		Park:          "PTD2YBBZ",
		CarNo:         "桂566666",
		ParkingCode:   "2060187075109569365",
		ChannelOut:    "765RMFAFR",
		ChannelIn:     "765RMFAFT",
		ChannelID:     500119,
		TraderCode:    "O563CNQTNZVY",
		TraderAccount: "0508001",
		MonthCfgID:    "1200009646",
		MonthTicketID: "1200093148",
		VipCarNo:      "粤HH7772",
		TopBillCode:   "2026052815334712692521",
		CouponCode:    "GJNOYX7SPAOG",
		TraderID:      509676,
		AreaID:        501768,
		SellBillID:    1500024632,
	}
}

// 查询时间区间字段(query ranges)→ 用近 25 天的紧凑时间;不动业务有效期字段(validFrom/startTime 等)。
var rangeFromKeys = map[string]bool{
	"enterTimeFrom": true, "exitTimeFrom": true, "payTimeFrom": true, "createTimeFrom": true,
	"operateTimeFrom": true, "queryTimeFrom": true, "timeFrom": true, "startTimeFrom": true,
	"outTimeFrom": true, "updateTimeFrom": true, "beginTime": true, "chargeTimeFrom": true,
}
var rangeToKeys = map[string]bool{
	"enterTimeTo": true, "exitTimeTo": true, "payTimeTo": true, "createTimeTo": true,
	"operateTimeTo": true, "queryTimeTo": true, "timeTo": true, "startTimeTo": true,
	"outTimeTo": true, "updateTimeTo": true, "endTime2": true, "chargeTimeTo": true,
}

func tsCompact(t time.Time) string { return t.Format("20060102150405") }

// overlay parses body (usually a Doc sampleBody) and replaces identifier / range-time
// fields with real fixtures, leaving other example fields intact. nested couponList etc.
// keep their sample values.
func (fx *Fixtures) overlay(body string) string {
	var v any
	if json.Unmarshal([]byte(body), &v) != nil || v == nil {
		v = map[string]any{}
	}
	now := time.Now()
	v = fx.walk(v, tsCompact(now.AddDate(0, 0, -25)), tsCompact(now))
	b, _ := json.Marshal(v)
	return string(b)
}

// walk recursively replaces identifier / range-time fields with real fixtures at
// every nesting level (nested objects + arrays-of-objects like couponList).
func (fx *Fixtures) walk(v any, from, to string) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			t[k] = fx.walk(val, from, to) // recurse first
			if nv, ok := fx.fieldValue(k, t[k], from, to); ok {
				t[k] = nv
			}
		}
		return t
	case []any:
		for i, e := range t {
			t[i] = fx.walk(e, from, to)
		}
		return t
	default:
		return v
	}
}

// fieldValue returns the fixture value for a known field key (preserving array vs
// scalar shape from the current value), or (nil,false) if the key is not a fixture.
func (fx *Fixtures) fieldValue(k string, cur any, from, to string) (any, bool) {
	shape := func(s string) any {
		if _, isArr := cur.([]any); isArr {
			return []any{s}
		}
		return s
	}
	switch k {
	case "parkCode", "parkCodes", "parkCodeList", "financialParkCode", "parkCodeArr":
		return shape(fx.Park), true
	case "carCode", "carNo", "plateNo", "carNoList", "carCodeList", "plateNoList":
		return shape(fx.CarNo), true
	case "parkingCode", "parkingCodeList", "parkingCodes":
		return shape(fx.ParkingCode), true
	case "channelCode", "customCode", "channelCustomCode", "outChannelCode":
		return shape(fx.ChannelOut), true
	case "enterChannelCode", "inChannelCode":
		return shape(fx.ChannelIn), true
	case "channelId":
		return fx.ChannelID, true
	case "traderId":
		return fx.TraderID, true
	case "areaId":
		return fx.AreaID, true
	case "sellBillId", "sellBillNumber":
		return fx.SellBillID, true
	case "traderCode", "traderCodeList":
		return shape(fx.TraderCode), true
	case "monthTicketConfigId":
		return shape(fx.MonthCfgID), true
	case "monthTicketId", "monthTicketBillId":
		return shape(fx.MonthTicketID), true
	case "specialCarTypeId", "vipGroupId", "carTypeId":
		if fx.SpecialTypeID != "" {
			return shape(fx.SpecialTypeID), true
		}
	case "couponCode", "couponCodeList":
		if fx.CouponCode != "" {
			return shape(fx.CouponCode), true
		}
	default:
		if rangeFromKeys[k] {
			return from, true
		}
		if rangeToKeys[k] {
			return to, true
		}
	}
	return nil, false
}

// Discover refreshes volatile fixtures from read APIs (best-effort; defaults remain on failure).
func (fx *Fixtures) Discover(cli *client.Client) {
	call := func(cmd, body string) map[string]any {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		resp, err := cli.Call(ctx, cmd, body)
		if err != nil || resp == nil || resp.Status != client.StatusSuccess {
			return nil
		}
		var d map[string]any
		_ = json.Unmarshal(resp.Data, &d)
		return d
	}
	now := time.Now()
	from := tsCompact(now.AddDate(0, 0, -25))
	to := tsCompact(now)

	// 在场车 → carNo + parkingCode
	if d := call("getParkOnSiteCar", fmt.Sprintf(`{"parkCodeList":["%s"],"enterTimeFrom":"%s","enterTimeTo":"%s","pageNum":1,"pageSize":1}`, fx.Park, from, to)); d != nil {
		if rec := firstRecord(d); rec != nil {
			if v, _ := rec["carNo"].(string); v != "" {
				fx.CarNo = v
			}
			if v, _ := rec["parkingCode"].(string); v != "" {
				fx.ParkingCode = v
			}
			if f, ok := rec["areaId"].(float64); ok && f > 0 {
				fx.AreaID = int(f)
			}
		}
	}
	// 月票类型 + 已开通月票
	if d := call("getMonthTicketConfigDetailList", fmt.Sprintf(`{"parkCodeList":["%s"],"pageNum":1,"pageSize":1}`, fx.Park)); d != nil {
		if rec := firstRecord(d); rec != nil {
			fx.MonthCfgID = numStr(rec["id"], fx.MonthCfgID)
		}
	}
	if d := call("getOnlineMonthTicketList", fmt.Sprintf(`{"parkCodeList":["%s"],"pageNum":1,"pageSize":1}`, fx.Park)); d != nil {
		if rec := firstRecord(d); rec != nil {
			fx.MonthTicketID = numStr(rec["id"], fx.MonthTicketID)
			fx.MonthCfgID = numStr(rec["monthTicketConfigId"], fx.MonthCfgID)
			fx.VipCarNo = strDefault(rec["carNo"], fx.VipCarNo)
			fx.TopBillCode = strDefault(rec["topBillCode"], fx.TopBillCode)
		}
	}
	// 商家
	if d := call("getTraderList", fmt.Sprintf(`{"parkCode":"%s","createTimeFrom":"%s","createTimeTo":"%s","pageNum":1,"pageSize":1}`, fx.Park, from, to)); d != nil {
		if rec := firstRecord(d); rec != nil {
			fx.TraderCode = strDefault(rec["traderCode"], fx.TraderCode)
			fx.TraderAccount = strDefault(rec["account"], fx.TraderAccount)
			if f, ok := rec["traderId"].(float64); ok && f > 0 {
				fx.TraderID = int(f)
			}
		}
	}
	// 特殊车辆类型 id(黑名单/访客复用)
	for _, body := range []string{
		fmt.Sprintf(`{"parkCode":"%s","vipGroupType":1,"pageNum":1,"pageSize":1}`, fx.Park),
		fmt.Sprintf(`{"parkCodeList":["%s"],"vipGroupType":1,"pageNum":1,"pageSize":1}`, fx.Park),
	} {
		if fx.SpecialTypeID != "" {
			break
		}
		if d := call("getSpecialCarTypeList", body); d != nil {
			if rec := firstRecord(d); rec != nil {
				fx.SpecialTypeID = numStr(rec["id"], "")
			}
		}
	}
	// 电子券: 发放记录 → couponCode(字段 code) + sellBillId
	wide := tsCompact(now.AddDate(0, 0, -50))
	if d := call("getTraderCouponGrantRecordList", fmt.Sprintf(`{"parkCodeList":["%s"],"traderId":%d,"beginTime":"%s","endTime":"%s","pageNum":1,"pageSize":1}`, fx.Park, fx.TraderID, wide, to)); d != nil {
		if rec := firstRecord(d); rec != nil {
			fx.CouponCode = strDefault(rec["code"], strDefault(rec["couponCode"], fx.CouponCode))
			if f, ok := rec["sellBillId"].(float64); ok && f > 0 {
				fx.SellBillID = int(f)
			}
		}
	}
}

func firstRecord(d map[string]any) map[string]any {
	for _, key := range []string{"recordList", "list", "records", "rows", "dataList"} {
		if arr, ok := d[key].([]any); ok && len(arr) > 0 {
			if rec, ok := arr[0].(map[string]any); ok {
				return rec
			}
		}
	}
	return nil
}

func numStr(v any, def string) string {
	switch t := v.(type) {
	case float64:
		return fmt.Sprintf("%.0f", t)
	case string:
		if t != "" {
			return t
		}
	}
	return def
}

func strDefault(v any, def string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return def
}
