// Command e2e exercises every included interface against the TEST environment
// and writes TEST_REPORT.md. Dependent interfaces are driven through business
// scenarios (billing / month-ticket / blacklist+visitor / coupon) so that a
// prior call's output (chargeBillToken, monthTicketConfigId, special-car-type id,
// trader/template code) feeds the next call's input. Everything else gets a
// best-effort standalone call using the catalog's sample body.
//
// Outcomes: PASS (status=1) | BIZFAIL (status=2, reached business) | ERROR
// (transport/sign/auth) | SKIP (prerequisite missing). Nothing is silently
// dropped — failures are recorded with their reason.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xiaowen-0725/openydt-cli/internal/catalog"
	"github.com/xiaowen-0725/openydt-cli/internal/client"
	"github.com/xiaowen-0725/openydt-cli/internal/config"
	"github.com/xiaowen-0725/openydt-cli/internal/sign"
)

const (
	catalogPath = "catalog/catalog.json"
	reportPath  = "TEST_REPORT.md"

	writePark = "1ZS7H5PQH9" // zhoujw 测试车场:可写, 可查费
	dataPark  = "PTD2YBBZ"   // 有存量数据:适合查记录/在场
	mainCar   = "粤EJW962"    // writePark 上可查费的车牌
)

const (
	PASS     = "PASS"
	BIZFAIL  = "BIZFAIL"
	ERROR    = "ERROR"
	SKIP     = "SKIP"
	NODEPLOY = "NODEPLOY" // 测试环境未部署该接口(接口不存在)——非 CLI 问题
)

type Result struct {
	Cmd, Domain, RW, Scenario, Outcome, Message, Note string
	HTTP, Status, ResultCode                          int
}

type Runner struct {
	cli     *client.Client
	cat     *catalog.Catalog
	results []Result
	done    map[string]bool
	ctx     map[string]string
	calls   int
}

func main() {
	cat, err := catalog.Load(catalogPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load catalog:", err)
		os.Exit(1)
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}
	env := os.Getenv("OPENYDT_ENV")
	if env == "" {
		env = "test"
	}
	r, err := cfg.Resolve("", env, "v2")
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve profile (run: openydt config set ...):", err)
		os.Exit(1)
	}
	if r.Env != "test" {
		fmt.Fprintf(os.Stderr, "拒绝在非 test 环境跑 E2E(当前 %s)\n", r.Env)
		os.Exit(1)
	}
	run := &Runner{
		cli:  client.New(r.BaseURL, r.Key, r.Secret, sign.V2, "openydt-cli/e2e"),
		cat:  cat,
		done: map[string]bool{},
		ctx:  map[string]string{"writePark": writePark, "dataPark": dataPark, "mainCar": mainCar},
	}

	fmt.Printf("E2E against %s (env=%s, %d included interfaces)\n", r.BaseURL, r.Env, len(cat.Included()))
	run.scenarioBilling()
	run.scenarioMonthTicket()
	run.scenarioList()
	run.scenarioCoupon()
	run.genericPass()
	run.writeReport()
	fmt.Printf("done: %d calls, report -> %s\n", run.calls, reportPath)
}

// ---- core call/record ----

func (r *Runner) call(cmd, body string) (*client.Response, error) {
	time.Sleep(250 * time.Millisecond) // ~4 req/s, under the 5/s limit
	r.calls++
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	return r.cli.Call(ctx, cmd, body)
}

func classify(resp *client.Response, err error) (string, string) {
	if err != nil {
		return ERROR, err.Error()
	}
	if resp.Status == 9 || strings.Contains(resp.Message, "接口不存在") {
		return NODEPLOY, fmt.Sprintf("接口不存在(测试环境未部署) | status=%d %s", resp.Status, resp.Message)
	}
	switch resp.Status {
	case client.StatusSuccess:
		return PASS, resp.Message
	case client.StatusBizFail:
		return BIZFAIL, fmt.Sprintf("resultCode=%d %s | %s", resp.ResultCode, client.ResultText(resp.ResultCode), resp.Message)
	default:
		return ERROR, fmt.Sprintf("status=%d %s | %s", resp.Status, client.StatusText(resp.Status), resp.Message)
	}
}

// run calls cmd with body, records the result, and returns the response (or nil).
func (r *Runner) run(cmd, scenario, body string) *client.Response {
	it, _ := r.cat.Find(cmd)
	resp, err := r.call(cmd, body)
	outcome, msg := classify(resp, err)
	res := Result{Cmd: cmd, Domain: it.Domain, RW: it.ReadWrite, Scenario: scenario, Outcome: outcome, Message: msg}
	if resp != nil {
		res.HTTP, res.Status, res.ResultCode = resp.HTTPStatus, resp.Status, resp.ResultCode
	}
	r.results = append(r.results, res)
	r.done[cmd] = true
	tag := outcome
	if outcome == PASS {
		tag = "✓ " + outcome
	}
	fmt.Printf("  [%-7s] %-34s %s\n", tag, cmd, clip(msg, 70))
	return resp
}

func (r *Runner) skip(cmd, scenario, reason string) {
	it, _ := r.cat.Find(cmd)
	r.results = append(r.results, Result{Cmd: cmd, Domain: it.Domain, RW: it.ReadWrite, Scenario: scenario, Outcome: SKIP, Note: reason})
	r.done[cmd] = true
	fmt.Printf("  [%-7s] %-34s %s\n", SKIP, cmd, reason)
}

// ---- scenarios ----

// sample builds a body from cmd's catalog sampleBody, overlaying overrides.
func (r *Runner) sample(cmd string, override map[string]any) string {
	return mergeSample(r.cat, cmd, override)
}

func (r *Runner) scenarioBilling() {
	fmt.Println("· 场景①: 停车缴费")
	const s = "billing"
	plate := genPlate()
	// 1. 进车补录(用样例字段名: carCodeColor 等)
	r.run("supplementParkingRecordIn", s, r.sample("supplementParkingRecordIn", map[string]any{
		"parkCode": writePark, "carCode": plate, "enterTime": now(), "carCodeColor": 1,
	}))
	// 2. 在场确认(parkCodeList 为数组)
	r.run("getParkOnSiteCar", s, r.sample("getParkOnSiteCar", map[string]any{
		"parkCodeList": []any{writePark}, "pageNum": 1, "pageSize": 10,
	}))
	// 3. 查费(既有可查费车牌, 拿 token + parkingCode)
	resp := r.run("getParkFee", s, r.sample("getParkFee", map[string]any{"parkCode": writePark, "carCode": mainCar}))
	token := digStr(resp, "otherAttr", "chargeBillToken")
	billNo := digStr(resp, "otherAttr", "chargeBillNumber")
	parkingCode := digStr(resp, "parkingCode")
	should := digStr(resp, "shouldPayValue")
	r.ctx["parkingCode"] = parkingCode
	// 4. 缴费(token + parkingCode)
	if token != "" {
		r.run("payParkFee", s, r.sample("payParkFee", map[string]any{
			"parkCode": writePark, "carCode": mainCar, "parkingCode": parkingCode,
			"chargeBillToken": token, "chargeBillNumber": billNo,
			"paidValue": should, "tradeNo": "e2e" + uniq(),
		}))
	} else {
		r.skip("payParkFee", s, "前置 getParkFee 未返回 chargeBillToken")
	}
	// 5. 查订单/记录(用 parkingCode)
	r.run("getPayBill", s, r.sample("getPayBill", map[string]any{"parkCode": writePark, "parkingCode": parkingCode}))
	r.run("getParkDetail", s, r.sample("getParkDetail", map[string]any{"parkCode": writePark, "parkingCode": parkingCode, "carCode": mainCar}))
	r.run("getPaymentRecordDetailList", s, r.sample("getPaymentRecordDetailList", map[string]any{"parkCode": dataPark, "pageNum": 1, "pageSize": 10}))
}

func (r *Runner) scenarioMonthTicket() {
	fmt.Println("· 场景②: 月票")
	const s = "monthticket"
	// 特殊车辆类型(供黑名单/访客复用)
	resp := r.run("addSpecialCarType", s, r.sample("addSpecialCarType", map[string]any{
		"parkCode": writePark, "parkCodes": writePark, "carTypeName": "e2e类型" + uniq(),
	}))
	if id := firstNonEmpty(digStr(resp, "carTypeId"), digStr(resp, "specialCarTypeId"), digStr(resp, "id")); id != "" {
		r.ctx["specialCarTypeId"] = id
	}
	// 1. 创建月票类型
	resp = r.run("addOnlineMonthTicketType", s, r.sample("addOnlineMonthTicketType", map[string]any{
		"parkCodes": writePark, "ticketName": "e2e月票" + uniq(), "price": 1.00,
	}))
	cfgID := digStr(resp, "monthTicketConfigId")
	if cfgID != "" {
		r.ctx["monthTicketConfigId"] = cfgID
	}
	// 2. 开通月票(用 monthTicketConfigId, 样例提供 userName 等)
	if cfgID != "" {
		r.run("addOnlineMonthTicket", s, r.sample("addOnlineMonthTicket", map[string]any{
			"parkCode": writePark, "monthTicketConfigId": cfgID, "carCode": genPlate(),
		}))
	} else {
		r.skip("addOnlineMonthTicket", s, "前置 addOnlineMonthTicketType 未返回 monthTicketConfigId")
	}
	// 3. 查询(parkCodeList 数组)
	r.run("getOnlineMonthTicketList", s, r.sample("getOnlineMonthTicketList", map[string]any{
		"parkCodeList": []any{writePark}, "pageNum": 1, "pageSize": 10,
	}))
}

func (r *Runner) scenarioList() {
	fmt.Println("· 场景③: 黑名单 / 访客")
	const s = "list"
	typeID := r.ctx["specialCarTypeId"]
	if typeID == "" {
		resp := r.run("addSpecialCarType", s, r.sample("addSpecialCarType", map[string]any{
			"parkCode": writePark, "parkCodes": writePark, "carTypeName": "e2e名单" + uniq(),
		}))
		typeID = firstNonEmpty(digStr(resp, "carTypeId"), digStr(resp, "specialCarTypeId"), digStr(resp, "id"))
	}
	r.run("getSpecialCarTypeList", s, r.sample("getSpecialCarTypeList", map[string]any{
		"parkCode": writePark, "parkCodeList": []any{writePark}, "vipGroupType": 1,
	}))
	if typeID != "" {
		r.run("addBlackListCar", s, r.sample("addBlackListCar", map[string]any{
			"parkCode": writePark, "carCode": genPlate(), "specialCarTypeId": typeID, "carType": typeID,
		}))
		r.run("addVisitorCarNew", s, r.sample("addVisitorCarNew", map[string]any{
			"parkCode": writePark, "carCode": genPlate(), "specialCarTypeId": typeID, "carType": typeID,
		}))
	} else {
		r.skip("addBlackListCar", s, "无特殊车辆类型ID")
		r.skip("addVisitorCarNew", s, "无特殊车辆类型ID")
	}
	r.run("getParkBlackList", s, r.sample("getParkBlackList", map[string]any{
		"parkCode": writePark, "parkCodeList": []any{writePark},
	}))
}

func (r *Runner) scenarioCoupon() {
	fmt.Println("· 场景④: 电子券")
	const s = "coupon"
	// 商家(样例含 loginAccount/password/traderName)
	acct := "e2e" + uniq()
	resp := r.run("createTrader", s, r.sample("createTrader", map[string]any{
		"parkCode": writePark, "traderName": "e2e商家" + uniq(),
		"loginAccount": acct, "password": "e2e123456", "traderUserAccount": acct,
	}))
	trader := firstNonEmpty(digStr(resp, "traderCode"), digStr(resp, "loginAccount"), digStr(resp))
	if trader != "" {
		r.ctx["traderCode"] = trader
	}
	// 券模板
	resp = r.run("createCouponTemplate", s, r.sample("createCouponTemplate", map[string]any{
		"parkCode": writePark, "traderCode": trader,
	}))
	tpl := firstNonEmpty(digStr(resp, "couponTemplateCode"), digStr(resp, "templateCode"), digStr(resp, "couponTemplateSn"), digStr(resp))
	if tpl != "" {
		r.ctx["couponTemplateCode"] = tpl
	}
	// 售卖
	if trader != "" && tpl != "" {
		r.run("sellCoupon", s, r.sample("sellCoupon", map[string]any{"traderCode": trader, "couponTemplateCode": tpl}))
	} else {
		r.skip("sellCoupon", s, "缺 traderCode / couponTemplateCode")
	}
	// 发放
	if tpl != "" {
		r.run("sendCoupon", s, r.sample("sendCoupon", map[string]any{
			"parkCode": writePark, "carCode": genPlate(), "couponTemplateCode": tpl,
		}))
	} else {
		r.skip("sendCoupon", s, "缺 couponTemplateCode")
	}
	r.run("getTraderList", s, r.sample("getTraderList", map[string]any{"parkCode": writePark}))
}

// ---- generic pass over remaining included interfaces ----

func (r *Runner) genericPass() {
	fmt.Println("· 通用兜底: 其余纳入接口(用 catalog 示例 body + 测试车场上下文)")
	list := r.cat.Included()
	sort.Slice(list, func(i, j int) bool {
		if list[i].Domain != list[j].Domain {
			return list[i].Domain < list[j].Domain
		}
		return list[i].Cmd < list[j].Cmd
	})
	for _, it := range list {
		if r.done[it.Cmd] {
			continue
		}
		r.run(it.Cmd, "generic", genericBody(it))
	}
}

// genericBody starts from the catalog sample and swaps park/car identifiers to
// our authorized test values so the call has the best chance of reaching business.
func genericBody(it catalog.Iface) string {
	var m map[string]any
	if json.Unmarshal([]byte(it.SampleBody), &m) != nil || m == nil {
		m = map[string]any{}
	}
	setIf := func(k, v string) {
		if cur, ok := m[k]; ok {
			if _, isArr := cur.([]any); isArr {
				m[k] = []any{v}
			} else {
				m[k] = v
			}
		}
	}
	park := dataPark
	if it.ReadWrite == "write" {
		park = writePark
	}
	for _, k := range []string{"parkCode", "parkCodes", "parkCodeList"} {
		setIf(k, park)
	}
	for _, k := range []string{"carCode", "carNo", "plateNo"} {
		setIf(k, mainCar)
	}
	// 补常见必填: 缺 parkCode 但参数要求 / 分页字段
	for _, p := range it.Params {
		if _, has := m[p.Name]; has {
			continue
		}
		switch p.Name {
		case "parkCode":
			m["parkCode"] = park
		case "pageNum", "pageNo", "pageIndex", "currentPage":
			m[p.Name] = 1
		case "pageSize", "pageCount", "limit":
			m[p.Name] = 10
		}
	}
	return obj(m)
}

// mergeSample overlays overrides on top of an interface's catalog sample body.
func mergeSample(cat *catalog.Catalog, cmd string, override map[string]any) string {
	it, _ := cat.Find(cmd)
	var m map[string]any
	if json.Unmarshal([]byte(it.SampleBody), &m) != nil || m == nil {
		m = map[string]any{}
	}
	for k, v := range override {
		m[k] = v
	}
	return obj(m)
}

// ---- report ----

func (r *Runner) writeReport() {
	counts := map[string]int{}
	for _, x := range r.results {
		counts[x.Outcome]++
	}
	byDomain := map[string][]Result{}
	for _, x := range r.results {
		byDomain[x.Domain] = append(byDomain[x.Domain], x)
	}
	domains := make([]string, 0, len(byDomain))
	for d := range byDomain {
		domains = append(domains, d)
	}
	sort.Strings(domains)

	var b strings.Builder
	fmt.Fprintf(&b, "# openydt-cli 端到端测试报告\n\n")
	fmt.Fprintf(&b, "- 环境: 测试环境 (openapi-test.yidianting.com.cn)\n")
	fmt.Fprintf(&b, "- 接口总数(已尝试): %d\n", len(r.results))
	fmt.Fprintf(&b, "- 结果: ✓PASS=%d, BIZFAIL=%d, NODEPLOY=%d, ERROR=%d, SKIP=%d\n",
		counts[PASS], counts[BIZFAIL], counts[NODEPLOY], counts[ERROR], counts[SKIP])
	reached := counts[PASS] + counts[BIZFAIL]
	fmt.Fprintf(&b, "- 成功调通业务层(PASS+BIZFAIL): %d / %d\n", reached, len(r.results))
	fmt.Fprintf(&b, "- 调用次数: %d\n\n", r.calls)
	fmt.Fprintf(&b, "> PASS=业务成功(status=1);BIZFAIL=接口已调通但业务校验未过(status=2,多因测试数据/入参,附 resultCode);NODEPLOY=测试环境未部署该接口(接口不存在,非 CLI 问题);ERROR=传输/签名/鉴权/系统异常;SKIP=前置依赖缺失未尝试(附原因)。\n\n")

	// 关注清单
	fmt.Fprintf(&b, "## 需关注(非 PASS)\n\n")
	fmt.Fprintf(&b, "| 命令 | 域 | 结果 | 说明 |\n|---|---|---|---|\n")
	for _, d := range domains {
		for _, x := range byDomain[d] {
			if x.Outcome == PASS {
				continue
			}
			detail := x.Message
			if x.Note != "" {
				detail = x.Note
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", x.Cmd, x.Domain, x.Outcome, mdEsc(clip(detail, 90)))
		}
	}

	// 按域明细
	fmt.Fprintf(&b, "\n## 按域明细\n")
	for _, d := range domains {
		rows := byDomain[d]
		sort.Slice(rows, func(i, j int) bool { return rows[i].Cmd < rows[j].Cmd })
		fmt.Fprintf(&b, "\n### %s (%d)\n\n", d, len(rows))
		fmt.Fprintf(&b, "| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |\n|---|---|---|---|---|---|\n")
		for _, x := range rows {
			sc := fmt.Sprintf("%d/%d", x.Status, x.ResultCode)
			if x.Outcome == SKIP {
				sc = "-"
			}
			detail := x.Message
			if x.Note != "" {
				detail = x.Note
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n",
				x.Cmd, x.RW, x.Scenario, x.Outcome, sc, mdEsc(clip(detail, 80)))
		}
	}
	if err := os.WriteFile(reportPath, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write report:", err)
	}
}

// ---- helpers ----

func obj(m map[string]any) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func digStr(resp *client.Response, path ...string) string {
	if resp == nil || len(resp.Data) == 0 {
		return ""
	}
	var v any
	if json.Unmarshal(resp.Data, &v) != nil {
		return ""
	}
	for _, p := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return ""
		}
		v = m[p]
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return ""
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func now() string  { return time.Now().Format("20060102150405") } // 平台时间格式 yyyyMMddHHmmss
func ts14() string { return time.Now().Format("20060102150405") }

// uniq returns a per-call unique suffix so repeated runs don't collide on names.
func uniq() string { return strconv.FormatInt(time.Now().UnixNano()%1000000, 10) }

var plateSeq int

func genPlate() string {
	plateSeq++
	return fmt.Sprintf("粤E%02dT%02d", time.Now().Second(), plateSeq%100)
}

func clip(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func mdEsc(s string) string { return strings.ReplaceAll(s, "|", "\\|") }
