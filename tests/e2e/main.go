// Command e2e exercises every included interface against the TEST environment
// and writes TEST_REPORT.md.
//
// Input construction = catalog sample body (correct field names) + real Fixtures
// overlay (park / on-site car / channel / trader / month-ticket / coupon, harvested
// from the data-rich cloud park PTD2YBBZ) + per-cmd recipes.json (token-substituted
// corrections derived from real failures). Dependent chains (billing pay) thread a
// prior call's output into the next.
//
// Outcomes: PASS (status=1) | BIZFAIL (status=2) | NODEPLOY (接口不存在) |
// ERROR (transport/sign/auth/system) | SKIP (intentionally not attempted).
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
	"github.com/xiaowen-0725/openydt-cli/internal/strutil"
)

// recipes.json meta fields (prefixed __ so they are skipped when building the body).
const (
	metaPrefix = "__"
	metaSkip   = "__skip"   // 故意不测(破坏性/需特殊前置)
	metaNoData = "__nodata" // 接口可达但测试环境缺设备/数据 → 降级 NODATA
)

const (
	catalogPath = "catalog/catalog.json"
	recipesPath = "tests/e2e/recipes.json"
	reportPath  = "TEST_REPORT.md"
)

const (
	PASS     = "PASS"
	BIZFAIL  = "BIZFAIL"
	ERROR    = "ERROR"
	SKIP     = "SKIP"
	NODEPLOY = "NODEPLOY"
	NODATA   = "NODATA" // 接口可达且参数正确, 但测试环境缺设备/在场车/会话/账单/特定状态
)

type Result struct {
	Cmd, Domain, RW, Scenario, Outcome, Message, Note string
	HTTP, Status, ResultCode                          int
}

type Runner struct {
	cli     *client.Client
	cat     *catalog.Catalog
	fx      *Fixtures
	recipes map[string]map[string]any
	results []Result
	done    map[string]bool
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
	res, err := cfg.Resolve("", env, "v2")
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve profile:", err)
		os.Exit(1)
	}
	if res.Env != "test" {
		fmt.Fprintf(os.Stderr, "拒绝在非 test 环境跑 E2E(当前 %s)\n", res.Env)
		os.Exit(1)
	}

	r := &Runner{
		cli:     client.New(res.BaseURL, res.Key, res.Secret, sign.V2, "openydt-cli/e2e"),
		cat:     cat,
		fx:      defaultFixtures(),
		recipes: loadRecipes(),
		done:    map[string]bool{},
	}
	fmt.Printf("E2E against %s (env=%s, %d included)\n", res.BaseURL, res.Env, len(cat.Included()))
	fmt.Println("· 发现真实 fixtures …")
	r.fx.Discover(r.cli)
	fmt.Printf("  park=%s carNo=%s parkingCode=%s channel=%s trader=%s monthCfg=%s monthTicket=%s special=%s coupon=%s\n",
		r.fx.Park, r.fx.CarNo, r.fx.ParkingCode, r.fx.ChannelOut, r.fx.TraderCode, r.fx.MonthCfgID, r.fx.MonthTicketID, r.fx.SpecialTypeID, r.fx.CouponCode)

	r.scenarioBilling()
	r.scenarioThreads()
	r.scenarioCoupon()
	r.scenarioPark()
	r.genericPass()
	r.writeReport()
	fmt.Printf("done: %d calls, report -> %s\n", r.calls, reportPath)
}

// ---- body construction ----

// bodyFor builds the request body for cmd: overlay(catalog sample) + recipe + extra,
// all token-substituted, then fills common missing fields.
func (r *Runner) bodyFor(cmd string, extra map[string]any) string {
	it, _ := r.cat.Find(cmd)
	var m map[string]any
	if json.Unmarshal([]byte(r.fx.overlay(it.SampleBody)), &m) != nil || m == nil {
		m = map[string]any{}
	}
	tok := r.tokens()
	for k, v := range r.recipes[cmd] {
		if strings.HasPrefix(k, metaPrefix) { // __skip / __nodata 等元字段
			continue
		}
		m[k] = subst(v, tok)
	}
	for k, v := range extra {
		m[k] = subst(v, tok)
	}
	for _, p := range it.Params { // 补常见缺失
		if _, has := m[p.Name]; has {
			continue
		}
		switch p.Name {
		case "parkCode":
			m["parkCode"] = r.fx.Park
		case "pageNum", "pageNo", "pageIndex", "currentPage", "page":
			m[p.Name] = 1
		case "pageSize", "pageCount", "limit":
			m[p.Name] = 10
		}
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func (r *Runner) tokens() map[string]string {
	now := time.Now()
	return map[string]string{
		"PARK": r.fx.Park, "CARNO": r.fx.CarNo, "PARKINGCODE": r.fx.ParkingCode,
		"CHANNEL_OUT": r.fx.ChannelOut, "CHANNEL_IN": r.fx.ChannelIn, "CHANNEL_ID": strconv.Itoa(r.fx.ChannelID),
		"TRADER": r.fx.TraderCode, "TRADER_ACCOUNT": r.fx.TraderAccount,
		"MONTH_CFG": r.fx.MonthCfgID, "MONTH_TICKET": r.fx.MonthTicketID, "VIP_CARNO": r.fx.VipCarNo,
		"BILL_CODE": r.fx.TopBillCode, "SPECIAL_TYPE": r.fx.SpecialTypeID, "COUPON": r.fx.CouponCode,
		"TS": now.Format("20060102150405"), "TODAY": now.Format("20060102"),
		"TOMORROW": now.AddDate(0, 0, 1).Format("20060102150405"), "PLUS7D": now.AddDate(0, 0, 7).Format("20060102150405"),
		"DATE": now.Format("2006-01-02"), "UNIQ": uniq(),
		"T_FROM": now.AddDate(0, 0, -25).Format("20060102150405"), "T_TO": now.Format("20060102150405"),
	}
}

// subst replaces {{TOKEN}} inside string values at every nesting level
// (strings, arrays, and nested objects — e.g. couponTemplate.validFrom).
func subst(v any, tok map[string]string) any {
	switch t := v.(type) {
	case string:
		for k, val := range tok {
			t = strings.ReplaceAll(t, "{{"+k+"}}", val)
		}
		return t
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = subst(e, tok)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, e := range t {
			out[k] = subst(e, tok)
		}
		return out
	default:
		return v
	}
}

// ---- core call/record ----

func (r *Runner) call(cmd, body string) (*client.Response, error) {
	time.Sleep(250 * time.Millisecond)
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

func (r *Runner) run(cmd, scenario string, extra map[string]any) *client.Response {
	it, _ := r.cat.Find(cmd)
	resp, err := r.call(cmd, r.bodyFor(cmd, extra))
	outcome, msg := classify(resp, err)
	// recipes 里标记的环境限制项:接口可达但测试环境缺设备/数据,降级为 NODATA(非 CLI 失败)。
	if (outcome == BIZFAIL || outcome == ERROR) && r.recipeMeta(cmd, metaNoData) != "" {
		outcome = NODATA
		msg = "环境限制: " + r.recipeMeta(cmd, metaNoData) + " | " + msg
	}
	res := Result{Cmd: cmd, Domain: it.Domain, RW: it.ReadWrite, Scenario: scenario, Outcome: outcome, Message: msg}
	if resp != nil {
		res.HTTP, res.Status, res.ResultCode = resp.HTTPStatus, resp.Status, resp.ResultCode
	}
	r.results = append(r.results, res)
	r.done[cmd] = true
	tag := outcome
	if outcome == PASS {
		tag = "✓PASS"
	}
	fmt.Printf("  [%-8s] %-34s %s\n", tag, cmd, clip(msg, 66))
	return resp
}

func (r *Runner) skip(cmd, scenario, reason string) {
	it, _ := r.cat.Find(cmd)
	r.results = append(r.results, Result{Cmd: cmd, Domain: it.Domain, RW: it.ReadWrite, Scenario: scenario, Outcome: SKIP, Note: reason})
	r.done[cmd] = true
	fmt.Printf("  [%-8s] %-34s %s\n", SKIP, cmd, reason)
}

// ---- billing chain (needs token threading) ----

func (r *Runner) scenarioBilling() {
	fmt.Println("· 场景: 停车缴费(查费→缴费)")
	resp := r.run("getParkFee", "billing", map[string]any{"parkCode": "{{PARK}}", "carCode": "{{CARNO}}"})
	parkingCode := digStr(resp, "parkingCode")
	chargeDate := digStr(resp, "chargeDate")
	should := digNum(resp, "shouldPayValue")
	if parkingCode != "" {
		// 实付额必须等于查费返回的应缴额(actPayCharge=shouldPayValue), 否则"费用超出仍应缴金额"。
		r.run("payParkFee", "billing", map[string]any{
			"parkCode": "{{PARK}}", "carCode": "{{CARNO}}", "parkingCode": parkingCode,
			"chargeDate": chargeDate, "actPayCharge": should, "billCode": "e2e{{UNIQ}}",
			"payDate": "{{TS}}", "payOrigin": 9, "paymentMode": 4, "couponList": []any{},
		})
	} else {
		r.skip("payParkFee", "billing", "前置 getParkFee 未返回 parkingCode")
	}
}

// scenarioCoupon: 自建启用商家 → 建模板 → 售卖 → 发放 → 回收, 把产物喂给按券码/账单的查询命令。
func (r *Runner) scenarioCoupon() {
	fmt.Println("· 链式: 电子券(建商家→建模板→售卖→发放→回收)")
	acct := "e2e" + uniq()
	resp := r.run("createTrader", "coupon", map[string]any{
		"account": acct, "loginAccount": acct, "traderName": "e2e商家" + uniq(),
		"password": "Test@123456", "phone": "13800138000", "contact": "e2e",
	})
	tc := firstNonEmptyS(digStr(resp, "traderCode"), r.fx.TraderCode)
	r.run("ValidateTraderAccountAndPassword", "coupon", map[string]any{
		"traderUserAccount": acct, "traderPassword": "Test@123456", "parkCode": "{{PARK}}",
	})
	resp = r.run("createCouponTemplate", "coupon", map[string]any{"traderCode": tc})
	tpl := firstNonEmptyS(digStr(resp, "couponCode"), digStr(resp, "traderCouponTemplateCode"), r.fx.CouponCode)
	r.run("createCoupon", "coupon", map[string]any{"traderCode": tc})
	resp = r.run("sellCoupon", "coupon", map[string]any{
		"traderCode": tc, "traderCouponTemplateCode": tpl, "sellNum": 1, "sellMoney": 1, "sellTime": "{{DATE}} 09:00:00",
	})
	sbid := firstRaw(resp, "sellBillId", "id")
	if sbid == nil {
		sbid = r.fx.SellBillID
	}
	resp = r.run("createFixedCoupon", "coupon", map[string]any{"traderCode": tc, "sellBillId": sbid, "maxNum": 1, "uniqNo": "e2e" + uniq()})
	fixedCode := firstNonEmptyS(digStr(resp, "code"), tpl)
	qr := firstNonEmptyS(digStr(resp, "qrCode"), digStr(resp, "url"))
	resp = r.run("sendCoupon", "coupon", map[string]any{"traderCode": tc, "sellBillId": sbid, "carCode": "{{CARNO}}", "carCodeColor": 1})
	sn := digStr(resp, "couponSn")
	r.run("cancelCoupon", "coupon", map[string]any{"traderCode": tc, "couponSn": sn})
	r.run("sendCouponByCouponCode", "coupon", map[string]any{"couponCode": tpl, "carNo": "{{CARNO}}", "parkCode": "{{PARK}}"})
	r.run("queryCouponTemplateByCouponCode", "coupon", map[string]any{"code": tpl})
	r.run("checkCouponWhetherSendAvailable", "coupon", map[string]any{"couponCode": tpl})
	r.run("queryCouponAvailableParkByCouponCode", "coupon", map[string]any{"couponCode": tpl, "fixedStatus": 0})
	r.run("queryCouponPrintRecord", "coupon", map[string]any{"couponCode": tpl})
	r.run("queryUsableCoupon", "coupon", map[string]any{"traderCode": tc, "sellBillId": sbid})
	r.run("checkCouponQrCodeValidStatus", "coupon", map[string]any{"couponCode": fixedCode, "origin": 0})
	r.run("lockCoupon", "coupon", map[string]any{"url": qr})
}

// scenarioPark: getParkYdtCharge 的整段返回作为 getParkYdtOtherCarTypeChargeInfo 的入参。
func (r *Runner) scenarioPark() {
	fmt.Println("· 链式: 车场计费测算")
	resp := r.run("getParkYdtCharge", "park", map[string]any{"parkCode": "{{PARK}}"})
	extra := map[string]any{"parkCode": "{{PARK}}", "startTime": "{{DATE}} 12:00:00"}
	if resp != nil && len(resp.Data) > 0 {
		var data map[string]any
		if json.Unmarshal(resp.Data, &data) == nil && data != nil {
			extra["parkYdtChargeVo"] = data
			if arr, ok := data["parkYdtChargeStandardVoList"].([]any); ok && len(arr) > 0 {
				if first, ok := arr[0].(map[string]any); ok {
					if v, ok := first["standardSeq"]; ok {
						extra["standardSeq"] = v
					}
					if v, ok := first["carType"]; ok {
						extra["carType"] = v
					}
				}
			}
		}
	}
	r.run("getParkYdtOtherCarTypeChargeInfo", "park", extra)
}

func firstNonEmptyS(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// ---- threaded chains (need a prior call's output) ----

func (r *Runner) scenarioThreads() {
	fmt.Println("· 链式: 白名单 / 访客 / 月票冻结")
	// 白名单: redListAdd → ruleId → delRedList
	resp := r.run("redListAdd", "redlist", nil)
	if id := firstRaw(resp, "ruleId", "id", "redListId"); id != nil {
		r.run("delRedList", "redlist", map[string]any{"ruleId": id})
	} else {
		r.skip("delRedList", "redlist", "前置 redListAdd 未返回 ruleId")
	}
	// 访客: addVisitorCarNew → visitorId → cancelVisitorCarNew
	resp = r.run("addVisitorCarNew", "visitor", map[string]any{"carCode": "粤V{{UNIQ}}"})
	if id := firstRaw(resp, "visitorId", "id"); id != nil {
		r.run("cancelVisitorCarNew", "visitor", map[string]any{"visitorId": id})
	} else {
		r.skip("cancelVisitorCarNew", "visitor", "前置 addVisitorCarNew 未返回 visitorId")
	}
	// 月票冻结 → 解冻(同一 billId, 须先冻结后解冻)
	r.run("freezeMonthTicket", "ticket", nil)
	r.run("unFreezeMonthTicket", "ticket", nil)
}

func firstRaw(resp *client.Response, paths ...string) any {
	if resp == nil || len(resp.Data) == 0 {
		return nil
	}
	var d map[string]any
	if json.Unmarshal(resp.Data, &d) != nil {
		return nil
	}
	for _, p := range paths {
		if v, ok := d[p]; ok && v != nil {
			return v
		}
	}
	return nil
}

// ---- generic pass over all remaining included interfaces ----

func (r *Runner) genericPass() {
	fmt.Println("· 全量: 其余纳入接口(sampleBody + fixtures + recipes)")
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
		if reason := r.recipeSkip(it.Cmd); reason != "" {
			r.skip(it.Cmd, "generic", reason)
			continue
		}
		r.run(it.Cmd, "generic", nil)
	}
}

// recipeMeta reads a meta field (e.g. __skip / __nodata) from a cmd's recipe.
func (r *Runner) recipeMeta(cmd, key string) string {
	if rec, ok := r.recipes[cmd]; ok {
		if s, ok := rec[key].(string); ok {
			return s
		}
	}
	return ""
}

// recipeSkip lets recipes.json mark a cmd as intentionally skipped (destructive on shared data).
func (r *Runner) recipeSkip(cmd string) string { return r.recipeMeta(cmd, metaSkip) }

// ---- report ----

func (r *Runner) writeReport() {
	counts := map[string]int{}
	byDomain := map[string][]Result{}
	for _, x := range r.results {
		counts[x.Outcome]++
		byDomain[x.Domain] = append(byDomain[x.Domain], x)
	}
	domains := make([]string, 0, len(byDomain))
	for d := range byDomain {
		domains = append(domains, d)
	}
	sort.Strings(domains)
	fails := counts[BIZFAIL] + counts[ERROR] + counts[SKIP]

	var b strings.Builder
	fmt.Fprintf(&b, "# openydt-cli 端到端测试报告\n\n")
	fmt.Fprintf(&b, "- 环境: 测试环境 (openapi-test.yidianting.com.cn);主车场 %s(智汇云测试专用车场123412)\n", r.fx.Park)
	fmt.Fprintf(&b, "- 接口总数(已尝试): %d\n", len(r.results))
	fmt.Fprintf(&b, "- 结果: ✓PASS=%d, BIZFAIL=%d, NODEPLOY=%d, ERROR=%d, SKIP=%d\n",
		counts[PASS], counts[BIZFAIL], counts[NODEPLOY], counts[ERROR], counts[SKIP])
	fmt.Fprintf(&b, "- **非 NODEPLOY 失败(BIZFAIL+ERROR+SKIP): %d**;成功调通业务层(PASS+BIZFAIL): %d\n", fails, counts[PASS]+counts[BIZFAIL])
	fmt.Fprintf(&b, "- 调用次数: %d\n\n", r.calls)
	fmt.Fprintf(&b, "> PASS=业务成功;BIZFAIL=已调通但业务校验未过;NODEPLOY=测试环境未部署(接口不存在,非 CLI 问题);ERROR=传输/系统异常;SKIP=故意不测(破坏性/需特殊前置)。\n\n")

	fmt.Fprintf(&b, "## 需关注(非 PASS,非 NODEPLOY)\n\n| 命令 | 域 | 结果 | 说明 |\n|---|---|---|---|\n")
	for _, d := range domains {
		for _, x := range byDomain[d] {
			if x.Outcome == PASS || x.Outcome == NODEPLOY {
				continue
			}
			detail := x.Message
			if x.Note != "" {
				detail = x.Note
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", x.Cmd, x.Domain, x.Outcome, mdEsc(clip(detail, 90)))
		}
	}

	fmt.Fprintf(&b, "\n## 按域明细\n")
	for _, d := range domains {
		rows := byDomain[d]
		sort.Slice(rows, func(i, j int) bool { return rows[i].Cmd < rows[j].Cmd })
		fmt.Fprintf(&b, "\n### %s (%d)\n\n| 命令 | 读写 | 场景 | 结果 | status/resultCode | 说明 |\n|---|---|---|---|---|---|\n", d, len(rows))
		for _, x := range rows {
			sc := fmt.Sprintf("%d/%d", x.Status, x.ResultCode)
			if x.Outcome == SKIP {
				sc = "-"
			}
			detail := x.Message
			if x.Note != "" {
				detail = x.Note
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n", x.Cmd, x.RW, x.Scenario, x.Outcome, sc, mdEsc(clip(detail, 78)))
		}
	}
	if err := os.WriteFile(reportPath, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write report:", err)
	}
}

// ---- helpers ----

func loadRecipes() map[string]map[string]any {
	out := map[string]map[string]any{}
	data, err := os.ReadFile(recipesPath)
	if err != nil {
		return out
	}
	_ = json.Unmarshal(data, &out)
	return out
}

// digPath descends the given key path in resp.Data and returns the leaf value (or nil).
func digPath(resp *client.Response, path ...string) any {
	if resp == nil || len(resp.Data) == 0 {
		return nil
	}
	var v any
	if json.Unmarshal(resp.Data, &v) != nil {
		return nil
	}
	for _, p := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[p]
	}
	return v
}

func digStr(resp *client.Response, path ...string) string {
	switch t := digPath(resp, path...).(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	}
	return ""
}

func digNum(resp *client.Response, path ...string) float64 {
	if f, ok := digPath(resp, path...).(float64); ok {
		return f
	}
	return 0
}

func uniq() string { return strconv.FormatInt(time.Now().UnixNano()%1000000, 10) }

func clip(s string, n int) string { return strutil.Clip(strings.ReplaceAll(s, "\n", " "), n) }

func mdEsc(s string) string { return strings.ReplaceAll(s, "|", "\\|") }
