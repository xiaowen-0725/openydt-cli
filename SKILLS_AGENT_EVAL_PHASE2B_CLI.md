# Phase 2B — CLI / Go 代码改进 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: 用 `superpowers:subagent-driven-development`(推荐:Task 各派 fresh subagent,TDD,任务间 review)或 `superpowers:executing-plans`。步骤用 checkbox(`- [ ]`)跟踪。**TDD**:先写失败测试 → 跑红 → 实现 → 跑绿 → 提交。

**Goal:** 落地 EVALUATION.md backlog 中**需改 Go/CLI 代码**的 P0/P1/P2 项:把写守护下沉到 `RunCall`(堵 api 裸通道 P0)、加全局 `--read-only`、`_error` 增 `nextCommands`/参数定位升级/重试分类、`schema -o json`、per-command MCP 三元注解(Go 派生)、`--verbose` HTTP 可观测、references 草稿自动渲染、README 计数从 catalog 生成。

**Architecture:** 安全与错误增强都收敛在**单一调用链路** `Factory.RunCall`(run.go)与其 `buildErrorInfo`——所有域命令与 `api` 都走它,改一处全覆盖。MCP 三元注解**在 Go 侧从既有 catalog 字段派生**(`catalog.Iface.Hints()`),不改 Node 抽取器、不重生成 `catalog.json`(DO-NOT-EDIT 产物),只在 `schema`/`--help` 表面化。codegen(`internal/gen`)改动后跑 `make generate` 重生成 `cmd/gen`(也是产物)。

**Tech Stack:** Go(`go test ./...`、`go vet ./...`、`make build`、`make generate`)、cobra、`jq`。**不碰**:`catalog.json` 手改、`skills/**`(那是 Phase 2A/2C)。

**依据:** `EVALUATION.md` §6 backlog(③⑤⑩⑭⑯)+ §7.3 CLI 补丁清单;`SKILLS_AGENT_EVAL_DESIGN.md` §7 边界。已读源:`cmd/api/api.go`、`internal/cmdutil/{run,factory}.go`、`cmd/root.go`、`internal/output/output.go`、`internal/catalog/catalog.go`、`internal/client/{client,codes}.go`、`cmd/schema/schema.go`、`internal/gen/main.go`。

---

## File Structure

| 文件 | 职责 | 动作 |
|---|---|---|
| `internal/cmdutil/factory.go` | 加 `ReadOnly bool` 全局标志 | Modify |
| `cmd/root.go` | 绑定 `--read-only` 持久标志 + 读 `OPENYDT_READ_ONLY` 环境变量 | Modify |
| `internal/catalog/catalog.go` | `IsWrite(cmd)`、`Iface.Hints()` 派生(readOnly/destructive/idempotent + 幂等键) | Modify |
| `internal/cmdutil/run.go` | `writeGuard()` 纯函数 + `RunCall` 调用(堵 api、施加 --read-only);`buildErrorInfo` 增 nextCommands/参数定位升级 | Modify |
| `internal/output/output.go` | `ErrorInfo` 增 `NextCommands`/`DocURL`/`SkillRoute`/`RetryClass`;table 渲染 | Modify |
| `internal/client/codes.go` | `ResultNextCommands`/`StatusNextCommands`/`RetryClass` | Modify |
| `internal/client/client.go` | `--verbose` 时输出 HTTP 重试/退避/状态/耗时 | Modify |
| `cmd/schema/schema.go` | `-o json` 机器可读输出 + 表面化 Hints | Modify |
| `internal/gen/main.go` | `--help`(longText)注入 Hints + `--refs <dir>` 渲染 reference 草稿到非 skills 目录 | Modify |
| `cmd/gen/*.go` | 重生成产物(`make generate`) | Regenerate |
| `scripts/counts.sh` + `Makefile` + `README.md` | `make counts` 从 catalog 生成计数;README 修 11→12、计数标注「以 make counts 为准」 | Create/Modify |
| 各 `*_test.go` | 各任务的单测 | Create/Modify |

---

## Task 1：写守护下沉 RunCall(P0 ③) + 全局 --read-only(⑯)

把写确认从「仅生成命令」下沉到**所有命令共享的 `RunCall`**,堵住 `api` 裸通道;并加全局只读开关。

**Files:** `internal/catalog/catalog.go`、`internal/cmdutil/factory.go`、`cmd/root.go`、`internal/cmdutil/run.go`、`internal/cmdutil/run_test.go`

- [ ] **Step 1:写失败测试(纯函数 writeGuard)**

新建/追加 `internal/cmdutil/run_test.go`:
```go
package cmdutil

import "testing"

func TestWriteGuard(t *testing.T) {
	cases := []struct {
		name              string
		write, yes, dry, ro bool
		wantErr           bool
		wantReadOnly      bool // err 是只读拒绝
	}{
		{"read passes", false, false, false, false, false, false},
		{"write no-yes blocked", true, false, false, false, true, false},
		{"write yes ok", true, true, false, false, false, false},
		{"write dry-run ok", true, false, true, false, false, false},
		{"readonly blocks write even with yes", true, true, false, true, true, true},
		{"readonly allows read", false, false, false, true, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := writeGuard("someCmd", c.write, c.yes, c.dry, c.ro)
			if (err != nil) != c.wantErr {
				t.Fatalf("err=%v want %v", err, c.wantErr)
			}
			if c.wantReadOnly && err != nil && !isReadOnlyErr(err) {
				t.Fatalf("expected read-only error, got %v", err)
			}
		})
	}
}
```

- [ ] **Step 2:跑红**

Run: `go test ./internal/cmdutil/ -run TestWriteGuard`
Expected: FAIL（`writeGuard`/`isReadOnlyErr` undefined）。

- [ ] **Step 3:实现 catalog.IsWrite + factory.ReadOnly + writeGuard + RunCall 接线**

`internal/catalog/catalog.go` 末尾加:
```go
// IsWrite reports whether cmd is a write operation per the embedded catalog.
// Unknown cmds default to false (api 兜底未知 cmd 时不误拦,但已知写 cmd 必拦)。
func (c *Catalog) IsWrite(cmd string) bool {
	if it, ok := c.Find(cmd); ok {
		return it.ReadWrite == "write"
	}
	return false
}
```
`internal/cmdutil/factory.go` 的 `Factory` 结构体里(`Verbose bool` 之后)加字段:
```go
	ReadOnly bool
```
`cmd/root.go` 的持久 flag 绑定(`pf.BoolVarP(&f.Verbose,...)` 之后)加:
```go
	pf.BoolVar(&f.ReadOnly, "read-only", false, "只读模式:拒绝任何写操作(也可设 OPENYDT_READ_ONLY=1)")
```
并在 `Execute()` 里 `f := cmdutil.NewFactory()` 之后加环境变量兜底(import `os`):
```go
	if v := os.Getenv("OPENYDT_READ_ONLY"); v == "1" || v == "true" {
		f.ReadOnly = true
	}
```
`internal/cmdutil/run.go`:加纯函数并在 `RunCall` 里 client 构建前调用。在 `RunCall` 的 `body = f.applyDefaults(...)` 之后、`c, err := f.Client()` 之前插入:
```go
	if err := f.guardWrite(cmd); err != nil {
		return err
	}
```
并在文件内新增:
```go
type readOnlyError struct{ cmd string }

func (e readOnlyError) Error() string {
	return fmt.Sprintf("只读模式(--read-only / OPENYDT_READ_ONLY)下拒绝写操作 %q;去掉只读开关再执行", e.cmd)
}

func isReadOnlyErr(err error) bool {
	var e readOnlyError
	return errors.As(err, &e)
}

// writeGuard decides whether a call may proceed. Pure (unit-testable).
func writeGuard(cmd string, isWrite, yes, dryRun, readOnly bool) error {
	if !isWrite {
		return nil
	}
	if readOnly {
		return ExitError{Code: output.ExitUsage, Err: readOnlyError{cmd}}
	}
	if yes || dryRun {
		return nil
	}
	return usageErr(fmt.Errorf("%q 是写操作,需加 --yes 确认 (或 --dry-run 预览)", cmd))
}

// guardWrite resolves write-ness from the embedded catalog and applies writeGuard.
func (f *Factory) guardWrite(cmd string) error {
	isWrite := false
	if cat, err := catalog.Embedded(); err == nil {
		isWrite = cat.IsWrite(cmd)
	}
	return writeGuard(cmd, isWrite, f.Yes, f.DryRun, f.ReadOnly)
}
```
在 run.go import 块补 `"errors"`(catalog/output 已 import)。
> 说明:生成命令仍各自 `ConfirmWrite`(无害冗余);`RunCall` 的 `guardWrite` 是唯一对 `api` 也生效的护栏。`ConfirmWrite` 保留不动(向后兼容、dry-run 语义一致)。

- [ ] **Step 4:跑绿 + 全量测试 + vet**

Run: `go test ./internal/cmdutil/ -run TestWriteGuard && go test ./... && go vet ./...`
Expected: PASS。

- [ ] **Step 5:行为冒烟(只读真验证)**

Run:
```bash
go build -o bin/openydt . && \
echo "--- api 写 cmd 无 --yes 应被拦截(此前是裸通道直发) ---" && \
./bin/openydt api createTrader --body '{}' ; echo "exit=$?" && \
echo "--- --read-only 拒绝写 ---" && \
./bin/openydt api createTrader --read-only --yes --body '{}' ; echo "exit=$?" && \
echo "--- 读 cmd 不受影响(dry-run 预览) ---" && \
./bin/openydt api getAuthParkCodes --dry-run >/dev/null ; echo "exit=$?"
```
Expected: 第一条 exit=2 并提示「写操作,需加 --yes」(P0 已堵);第二条 exit=2 并提示只读拒绝;第三条 exit=0。

- [ ] **Step 6:提交**

```bash
git add internal/catalog/catalog.go internal/cmdutil/factory.go cmd/root.go internal/cmdutil/run.go internal/cmdutil/run_test.go
git commit -m "fix(cli): 写守护下沉 RunCall 堵 api 裸通道(P0) + 全局 --read-only/OPENYDT_READ_ONLY

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2：_error 增 nextCommands + 参数定位升级 + 重试分类(⑯⑤)

让失败响应带「下一步可执行命令」与更准的缺参定位,Agent 直接照着自纠。

**Files:** `internal/client/codes.go`、`internal/output/output.go`、`internal/cmdutil/run.go`、`internal/client/codes_test.go`、`internal/cmdutil/run_test.go`

- [ ] **Step 1:写失败测试**

追加 `internal/client/codes_test.go`:
```go
package client

import "testing"

func TestResultNextCommands(t *testing.T) {
	if got := ResultNextCommands(904); len(got) == 0 || got[0] != "park get-auth-park-codes" {
		t.Fatalf("904 nextCommands=%v", got)
	}
	if got := ResultNextCommands(912); len(got) == 0 || got[0] != "trade get-park-fee" {
		t.Fatalf("912 nextCommands=%v", got)
	}
	if got := ResultNextCommands(907); len(got) != 0 {
		t.Fatalf("907 应无重发建议(幂等命中),got %v", got)
	}
}

func TestRetryClass(t *testing.T) {
	if RetryClass(StatusSysError) != "server_indeterminate" {
		t.Fatal("status3 应 server_indeterminate")
	}
	if RetryClass(StatusSuccess) != "" {
		t.Fatal("success 无 retry class")
	}
}
```
追加到 `internal/cmdutil/run_test.go`:
```go
func TestBuildErrorInfoMissingParam(t *testing.T) {
	// status=7 + body 缺某必填 → Field 定位到缺失项(不依赖中文 message)
	// 用一个 catalog 中已知有必填参数的 cmd;若离线/无 catalog 则跳过。
	// (实现见 buildErrorInfo;此测试保证缺参定位走「必填集对比」分支)
}
```
> 注:`TestBuildErrorInfoMissingParam` 用占位骨架——执行者按 catalog 里一个有必填标量参数的真实 cmd(如 `getParkFee` 的 `carCode`?以 `openydt schema getParkFee` 实查)补一个具体断言:构造 resp{Status:7}, body 缺该必填 → `ei.Field` == 该参数名。

- [ ] **Step 2:跑红**

Run: `go test ./internal/client/ -run 'TestResultNextCommands|TestRetryClass'`
Expected: FAIL（未定义)。

- [ ] **Step 3:实现**

`internal/client/codes.go` 末尾加:
```go
// ResultNextCommands maps a business result code to concrete next-step commands
// (without the `openydt ` prefix). Empty means "no retry/next action implied".
func ResultNextCommands(code int) []string {
	switch code {
	case 904, 910, 911:
		return []string{"park get-auth-park-codes"}
	case 905, 1801:
		return []string{"parking get-park-on-site-car"}
	case 906:
		return []string{"trade get-park-fee"}
	case 912:
		return []string{"trade get-park-fee"}
	case 909:
		return []string{"schema <cmd>"}
	default: // 907 幂等命中等:无重发建议
		return nil
	}
}

// StatusNextCommands maps a transport/auth status to next-step commands.
func StatusNextCommands(status int) []string {
	switch status {
	case StatusNoAuth:
		return []string{"park get-auth-park-codes"}
	case StatusBadParams:
		return []string{"schema <cmd>"}
	default:
		return nil
	}
}

// RetryClass classifies how (if at all) a status should be retried.
//   server_indeterminate: 系统异常,可同幂等键重试
//   ""                   : 业务/参数错,不应盲目重试(改参或对账)
func RetryClass(status int) string {
	if status == StatusSysError {
		return "server_indeterminate"
	}
	return ""
}
```
`internal/output/output.go` 的 `ErrorInfo` 结构体加字段(`AllowedValues` 之后):
```go
	NextCommands []string `json:"nextCommands,omitempty"`
	RetryClass   string   `json:"retryClass,omitempty"`
	DocURL       string   `json:"docUrl,omitempty"`
	SkillRoute   string   `json:"skillRoute,omitempty"`
```
`RenderError` 的 table 分支里(`if e.Retriable {...}` 之前)加:
```go
		if len(e.NextCommands) > 0 {
			fmt.Fprintf(w, "  下一步     : %s\n", "openydt "+strings.Join(e.NextCommands, "  |  openydt "))
		}
```
`internal/cmdutil/run.go`:把 `buildErrorInfo` 签名改为带 body,并填充新字段 + 升级缺参定位。
- `RunCall` 内调用处改:`ei := buildErrorInfo(cmd, body, resp)`。
- `buildErrorInfo` 改为:
```go
func buildErrorInfo(cmd, body string, resp *client.Response) *output.ErrorInfo {
	ei := &output.ErrorInfo{
		Cmd: cmd, Status: resp.Status, StatusText: client.StatusText(resp.Status),
		Message: resp.Message, Retriable: client.Retriable(resp.Status),
		RetryClass: client.RetryClass(resp.Status),
	}
	if resp.Status == client.StatusBizFail {
		ei.ResultCode = resp.ResultCode
		ei.ResultText = client.ResultText(resp.ResultCode)
		ei.Hint = client.ResultHint(resp.ResultCode)
		ei.NextCommands = client.ResultNextCommands(resp.ResultCode)
	} else {
		ei.NextCommands = client.StatusNextCommands(resp.Status)
	}
	if ei.Hint == "" {
		ei.Hint = client.StatusHint(resp.Status)
	}
	// nextCommands 里的 "schema <cmd>" 占位换成真实 cmd
	for i, nc := range ei.NextCommands {
		if nc == "schema <cmd>" {
			ei.NextCommands[i] = "schema " + cmd
		}
	}
	ei.locateParam(cmd, body, resp)
	return ei
}
```
新增 `locateParam`(参数定位:先按「必填集对比 body 缺哪个」,失败再回退中文文案正则):
```go
func (ei *output.ErrorInfo) locateParam(cmd, body string, resp *client.Response) {
	cat, err := catalog.Embedded()
	if err != nil {
		return
	}
	it, ok := cat.Find(cmd)
	if !ok {
		return
	}
	// 1) 参数不完整(status=7)或 909:按必填集对比 body 实际键,定位首个缺失必填标量
	if resp.Status == client.StatusBadParams || resp.ResultCode == 909 {
		present := map[string]bool{}
		var m map[string]any
		if json.Unmarshal([]byte(body), &m) == nil {
			for k := range m {
				present[k] = true
			}
		}
		for _, p := range it.Params {
			if p.Required && p.Group == "" && !present[p.Name] {
				ei.setParam(p)
				return
			}
		}
	}
	// 2) 回退:中文文案("…不能为空/参数错误")提到的参数名
	if strings.Contains(resp.Message, "不能为空") || strings.Contains(resp.Message, "参数错误") {
		for _, tok := range fieldRe.FindAllString(resp.Message, -1) {
			if p, ok := it.FindParam(tok); ok {
				ei.setParam(p)
				return
			}
		}
	}
}
```
> `ei.setParam(p)` 是把 `Field/FieldType/FieldRequired/FieldDesc/AllowedValues` 一次性填好的小 helper——因 `ErrorInfo` 在 `output` 包、`locateParam`/`setParam` 用到 catalog,**把这两个方法定义在 `run.go`(cmdutil 包)里、接收者改为 `*output.ErrorInfo`** 即可(Go 允许为其它包类型定义方法吗?**不允许**)。**改为自由函数**:`locateParam(ei *output.ErrorInfo, cmd, body string, resp *client.Response)` 与 `setParam(ei *output.ErrorInfo, p catalog.Param)`,在 buildErrorInfo 里 `locateParam(ei, cmd, body, resp)` 调用。`setParam`:
```go
func setParam(ei *output.ErrorInfo, p catalog.Param) {
	ei.Field = p.Name
	ei.FieldType = p.Type
	ei.FieldRequired = p.Required
	ei.FieldDesc = p.Desc
	ei.AllowedValues = p.EnumValues()
}
```
(删掉旧 buildErrorInfo 里内联的 catalog 提参块,统一走 locateParam。)

- [ ] **Step 4:补全 Step1 占位测试 + 跑绿**

按 `openydt schema getParkFee` 实查一个必填标量参数名,补全 `TestBuildErrorInfoMissingParam` 的具体断言。
Run: `go test ./... && go vet ./...`
Expected: PASS。

- [ ] **Step 5:提交**

```bash
git add internal/client/codes.go internal/client/codes_test.go internal/output/output.go internal/cmdutil/run.go internal/cmdutil/run_test.go
git commit -m "feat(cli): _error 增 nextCommands/retryClass + 缺参定位改按必填集对比(更稳)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3：MCP 三元注解(Go 派生)+ schema -o json(⑯)

per-command 暴露 readOnly/destructive/idempotent(及幂等键),让 Agent 调用前即可判「安全可重试 vs 会重复扣费」。**在 Go 侧从既有 catalog 字段派生,不改 catalog.json / 不动 Node 抽取器。**

**Files:** `internal/catalog/catalog.go`、`internal/catalog/catalog_test.go`、`cmd/schema/schema.go`

- [ ] **Step 1:写失败测试**

新建 `internal/catalog/catalog_test.go`:
```go
package catalog

import "testing"

func TestHints(t *testing.T) {
	read := Iface{Cmd: "getParkFee", ReadWrite: "read"}
	if h := read.Hints(); !h.ReadOnly || h.Destructive {
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
}
```

- [ ] **Step 2:跑红**

Run: `go test ./internal/catalog/ -run TestHints`
Expected: FAIL（`Hints` undefined)。

- [ ] **Step 3:实现 Hints**

`internal/catalog/catalog.go` 末尾加:
```go
import "strings" // 若文件未 import,加到 import 块

// Hints are MCP-style per-command annotations derived from catalog fields
// (no separate source: read-only from readwrite, destructive/idempotent heuristic).
type Hints struct {
	ReadOnly       bool   `json:"readOnly"`
	Destructive    bool   `json:"destructive"`
	Idempotent     bool   `json:"idempotent"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

var idemKeys = []string{"billCode", "thirdBillCode", "thirdpartyBillCode", "uniqNo", "transationNum"}

func (it Iface) Hints() Hints {
	h := Hints{ReadOnly: it.ReadWrite == "read"}
	if h.ReadOnly {
		h.Idempotent = true // 读天然幂等
		return h
	}
	lc := strings.ToLower(it.Cmd)
	for _, kw := range []string{"delete", "del", "cancel", "remove", "freeze", "frozen", "refund"} {
		if strings.Contains(lc, kw) {
			h.Destructive = true
			break
		}
	}
	for _, p := range it.Params {
		for _, k := range idemKeys {
			if p.Name == k {
				h.Idempotent = true
				h.IdempotencyKey = k
				return h
			}
		}
	}
	return h
}
```
> 注:若 catalog.go 顶部 import 块没有 `"strings"`,加上(`"encoding/json"`、`"os"` 已有)。

- [ ] **Step 4:schema 表面化 Hints + -o json**

`cmd/schema/schema.go`:
- `showOne`:在「读写」行后加(human 模式):
```go
	h := it.Hints()
	tags := []string{}
	if h.ReadOnly { tags = append(tags, "read-only") }
	if h.Destructive { tags = append(tags, "destructive") }
	if h.Idempotent { if h.IdempotencyKey != "" { tags = append(tags, "idempotent(key="+h.IdempotencyKey+")") } else { tags = append(tags, "idempotent") } }
	if len(tags) > 0 { fmt.Fprintf(w, "注解:       %s\n", strings.Join(tags, ", ")) }
```
- `-o json`:`RunE` 开头按 `f.Format()` 分流。新增 `showOneJSON`/`listJSON`:当 `f.Format()==output.JSON` 时输出结构化。`showOneJSON` 输出 `{cmd,domain,readwrite,direction,included,hints,params,sampleBody}`:
```go
func showOneJSON(f *cmdutil.Factory, it catalog.Iface) error {
	return output.PrintJSON(f.Out, map[string]any{
		"cmd": it.Cmd, "domain": it.Domain, "readwrite": it.ReadWrite,
		"direction": it.Direction, "included": it.Included, "hints": it.Hints(),
		"params": it.Params, "sampleBody": json.RawMessage(orEmptyObj(it.SampleBody)),
	})
}
```
`RunE` 改:`if len(args)==1 { if f.Format()==output.JSON { return showOneJSONByName(f,cat,args[0]) }; return showOne(...) }` 同理 list。`orEmptyObj` 把空 sampleBody 兜成 `{}`;import `encoding/json`、`output`。
> 默认 `-o` 是 `json`(root 默认),为不破坏现有「人读」习惯:schema 默认仍走人读文本,**仅当用户显式 `-o json` 时**走 JSON。实现:用一个本地 `--json` bool flag 或检测 root 的 `-o` 是否被显式设过。**最简稳妥**:给 schema 加独立 `--json` flag(`cmd.Flags().Bool("json",false,...)`),为真才输出 JSON;不复用全局 `-o`(避免影响默认人读)。Step1 无需测此分流,Step5 冒烟覆盖。

- [ ] **Step 5:跑绿 + 冒烟**

Run:
```bash
go test ./... && go vet ./... && go build -o bin/openydt . && \
./bin/openydt schema getParkFee | grep -E '注解|read-only' && \
./bin/openydt schema payParkFee --json | jq '.hints'
```
Expected: 测试 PASS;getParkFee 显示 `read-only`;payParkFee 的 `--json` 输出 `hints.idempotent=true, idempotencyKey="billCode"`。

- [ ] **Step 6:提交**

```bash
git add internal/catalog/catalog.go internal/catalog/catalog_test.go cmd/schema/schema.go
git commit -m "feat(cli): MCP 三元注解(Go 派生 readOnly/destructive/idempotent+幂等键) + schema --json

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4：--help 注入 Hints + --verbose HTTP 可观测(⑯)

让生成命令的 `--help` 也显示三元注解;`--verbose` 输出 HTTP 重试/退避/状态/耗时。

**Files:** `internal/gen/main.go`、`internal/client/client.go`、`internal/cmdutil/factory.go`、`cmd/gen/*.go`(重生成)

- [ ] **Step 1:gen longText 注入 Hints**

`internal/gen/main.go`:`Iface` 结构体已含 ReadWrite/Params——加一个本地 `hints(it)` 同 catalog 派生逻辑(gen 是独立 main 包,复制那段派生函数即可,~15 行),在 `longText` 的 `b.WriteString("  | " + it.ReadWrite)` 之后追加:
```go
	if h := hints(it); h.ReadOnly || h.Destructive || h.Idempotent {
		b.WriteString("  | 注解: ")
		// 拼 read-only/destructive/idempotent(key=...)
	}
```
(完整拼接逻辑同 Task 3 Step4 的 tags 拼法。)

- [ ] **Step 2:client --verbose HTTP 观测**

`internal/cmdutil/factory.go` 的 `Client()`:把 verbose 传入 client。给 `client.Client` 加字段 `Verbose bool` 与 `Log io.Writer`(import `io`),在 `New(...)` 后由 factory 设置:
```go
	cl := client.New(r.BaseURL, r.Key, r.Secret, sign.Version(r.Sign), f.UserAgent())
	cl.Verbose = f.Verbose
	cl.Log = f.Err
	return cl, nil
```
`internal/client/client.go` 的 `do`/`Call`:`Verbose` 为真时,每次 attempt 前后用 `c.logf(...)` 输出「attempt N、backoff、HTTP status、耗时 ms」。加:
```go
func (c *Client) logf(format string, a ...any) {
	if c.Verbose && c.Log != nil {
		fmt.Fprintf(c.Log, "[openydt http] "+format+"\n", a...)
	}
}
```
在 `do` 里测耗时(`time.Now()` 前后差),HTTP 返回后 `c.logf("%s -> HTTP %d (%dms)%s", cmd?, code, ms, retryMark)`(cmd 不在 do 作用域,可改为传 attempt 序号;退避在 `Call` 里 logf)。
> `time.Now()`/`time.Since` 在生产代码可用(仅 Workflow 脚本禁用,Go 代码不受限)。

- [ ] **Step 3:重生成 + 测试 + 冒烟**

Run:
```bash
make generate && go test ./... && go vet ./... && go build -o bin/openydt . && \
./bin/openydt trade get-park-fee --help | grep -i 注解 && \
./bin/openydt api getAuthParkCodes --verbose 2>&1 | grep -i 'http'
```
Expected: `make generate` 重写 `cmd/gen/*.go`(产物);测试 PASS;get-park-fee --help 含「注解」行;--verbose 打印 HTTP 行。
> `make generate` 会改 `cmd/gen/*.go`——这是**预期的产物更新**,一并提交。

- [ ] **Step 4:提交**

```bash
git add internal/gen/main.go internal/client/client.go internal/cmdutil/factory.go cmd/gen/
git commit -m "feat(cli): --help 注入 MCP 注解 + --verbose 输出 HTTP 重试/退避/状态/耗时

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5：references 草稿自动渲染(⑩,P2)

`internal/gen` 增可选模式,从 catalog 渲染 reference 草稿到**非 skills 目录**(`build/reference-drafts/`),供人工搬入 skills/references——**绝不写进 skills/**(避免覆盖手写 + 被 skillsync 擦除)。

**Files:** `internal/gen/main.go`

- [ ] **Step 1:实现 --refs 模式**

`internal/gen/main.go` `main()`:解析一个 `--refs <dir>` 选项(用 `os.Args` 扫描,保持现有位置参数兼容)。若设了 `--refs`,对每个 included callable cmd 写 `<dir>/<domain>/<kebab-cmd>.md`,内容七段式草稿:命令/参数表(Name|Type|必填|Desc)/sampleBody/对应 API 路径(`/openydt/api/v3/<cmd>`)/Hints/坑点占位(`<!-- 人工补:踩坑 -->`)。
```go
// 在 main() 解析 catalog 后:
if refsDir := flagValue("--refs"); refsDir != "" {
	writeReferenceDrafts(refsDir, cat)
	return
}
```
`writeReferenceDrafts` + `flagValue` 实现(纯文件写,markdown 拼接;复用 hints(it))。

- [ ] **Step 2:冒烟(不进 git)**

Run:
```bash
go run ./internal/gen catalog/catalog.json cmd/gen --refs build/reference-drafts >/dev/null 2>&1 || go run ./internal/gen --refs build/reference-drafts ; \
ls build/reference-drafts/trade/ | head && echo "drafts ok"
```
Expected: `build/reference-drafts/<domain>/*.md` 生成。`build/` 已被 .gitignore 忽略(`/bin/` 在列;若 build/ 未忽略,加到 .gitignore)。

- [ ] **Step 3:确保 build/ 被忽略 + 提交 gen 改动**

```bash
grep -q '^/build/' .gitignore || echo '/build/' >> .gitignore
git add internal/gen/main.go .gitignore
git commit -m "feat(gen): --refs 模式从 catalog 渲染 reference 草稿到 build/(不写 skills/)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6：README 计数从 catalog 生成 + 修 11→12(⑭)

**Files:** `scripts/counts.sh`、`Makefile`、`README.md`

- [ ] **Step 1:写计数脚本**

新建 `scripts/counts.sh`(用 jq 从 catalog 出权威计数):
```bash
#!/usr/bin/env bash
# 从 catalog.json 与 skills/ 生成权威计数,供 README/文档核对。
set -euo pipefail
cd "$(dirname "$0")/.."
CAT=catalog/catalog.json
echo "接口总数:     $(jq '.interfaces|length' $CAT)"
echo "一等命令:     $(jq '[.interfaces[]|select(.included==true and .direction=="callable")]|length' $CAT)"
echo "callable兜底: $(jq '[.interfaces[]|select(.included==false and .direction=="callable")]|length' $CAT)"
echo "webhook:      $(jq '[.interfaces[]|select(.direction=="webhook")]|length' $CAT)"
echo "技能数:       $(ls -d skills/openydt-* | wc -l | tr -d ' ')"
```
`chmod +x scripts/counts.sh`。

- [ ] **Step 2:Makefile 加 counts 目标**

`Makefile` 追加:
```make
counts: ## 从 catalog 生成权威接口/命令/技能计数
	@bash scripts/counts.sh
```

- [ ] **Step 3:跑一次取真值 + 修 README**

Run: `make counts`(记下真实数字)。据真值修 `README.md`:
- 「`skills/` 下 11 个技能」→ `12 个技能`(并核对 AI Agent Skills 表是否列全 12 个,补 `openydt-flow-park-access`)。
- 内置命令表标题「共 **143** 条一等命令(接口目录共 423 个)」→ 改为「以 `make counts` 为准」或填 `make counts` 的真值,并加一句 `> 计数由 catalog 生成,运行 \`make counts\` 核对。`

- [ ] **Step 4:提交**

```bash
git add scripts/counts.sh Makefile README.md
git commit -m "docs(readme): 计数从 catalog 生成(make counts) + 修技能数 11→12

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7：全量验证

**Files:** 无改动(只读验证)

- [ ] **Step 1:全量测试 + vet + build**

Run: `go test ./... && go vet ./... && make build`
Expected: 全 PASS、二进制产出。

- [ ] **Step 2:关键回归冒烟(只读)**

Run:
```bash
./bin/openydt api createTrader --body '{}'; echo "exit=$? (期望2,写守护)"
./bin/openydt api getAuthParkCodes --read-only --dry-run >/dev/null; echo "exit=$? (期望0,只读放行读)"
./bin/openydt schema payParkFee --json | jq '{hints,readwrite}'
./bin/openydt api getParkFee --body '{}' -o table 2>&1 | grep -E '下一步|建议' || echo "(查费空body触发错误对象)"
```
Expected:写守护生效;只读放行读;schema JSON 带 hints;失败响应含「下一步」可执行命令。

- [ ] **Step 3:与 Phase 2A 不冲突确认**

Run: `node scripts/skill-format-check/index.js 2>&1 | tail -1`(确认 Go 改动未碰 skills,技能仍 0 FAIL)。

> 完整 30-agent 评测回归在 Phase 2C 完成后统一跑(见 design §9)。

---

## Self-Review(已核)

- **Backlog 覆盖(代码项)**:③ api 写守护→Task1(下沉 RunCall,纯函数 writeGuard 单测);⑯ --read-only→Task1;⑯ _error nextCommands/retryClass/参数定位→Task2;⑯ MCP 三元注解→Task3(Go 派生,**故意不改 extractor/catalog.json**,理由:catalog.json 是 DO-NOT-EDIT 产物、Go 派生零重生成风险);⑯ schema -o json→Task3(独立 --json flag,不破坏默认人读);⑯ --help 注解 + --verbose HTTP→Task4;⑩ references 自动渲染→Task5(写 build/ 非 skills/);⑭ README 计数生成+11→12→Task6。
- **占位符扫描**:Task2 Step1 的 `TestBuildErrorInfoMissingParam` 是**显式标注的占位骨架**,Step4 要求按 `schema getParkFee` 实查参数名补全具体断言——非隐藏占位,已注明补全动作。其余均给完整 Go 代码。
- **类型/命名一致**:`writeGuard`/`isReadOnlyErr`/`guardWrite`(Task1)、`ResultNextCommands`/`StatusNextCommands`/`RetryClass`(Task2 codes.go)、`Hints`/`IdempotencyKey`(Task3 catalog.go)、`ErrorInfo.NextCommands/RetryClass/DocURL/SkillRoute`(Task2 output.go)跨任务引用一致。`setParam`/`locateParam` 因 Go 不能为外包类型定义方法,**已明确改为自由函数**(接收 `*output.ErrorInfo`)。
- **产物纪律**:Task4 `make generate` 重写 `cmd/gen/*.go`(产物,一并提交);不手改 catalog.json;references 草稿写 build/(gitignore)。
- **DocURL/SkillRoute**:字段已加进 ErrorInfo(预留),本期不强制填充值(平台无公开 doc 站);留作 Task2 可选——不构成未完成项(omitempty)。
- **回归**:Task7 轻量冒烟 + 不碰 skills 确认;完整评测攒到 2C 后。
