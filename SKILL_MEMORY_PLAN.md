# openydt 记忆/沉淀层 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 给 openydt 加两层"记忆":B 层(主)= AI agent 按车场自动沉淀/回忆的经验文件(纯 SKILL.md 约定);A 层(辅)= Profile 上的默认 parkCode/carNo,命令缺参时自动补全。

**Architecture:** A 层是 Go 改动,集中在 `internal/config`(Profile 加字段 + Active 选择器)、`internal/cmdutil`(RunCall 缺参注入,查内嵌 catalog 判断该命令是否收该字段)、`cmd/config`(新 `set-default` 子命令)。B 层零 Go 代码,只在 `skills/openydt-shared/SKILL.md` 加"车场经验"约定,数据存 `~/.config/openydt-cli/park-notes/{parkCode}.md`(独立于技能包,避免被技能同步擦除)。

**Tech Stack:** Go 1.26 + spf13/cobra;内嵌 catalog(`internal/catalog`);config 走 XDG(`~/.config/openydt-cli/`)。

**与 spec 的一处细化**:A 层写默认值用**独立子命令 `openydt config set-default`**(只更新现有/当前 profile 的默认字段,不必重输 key/secret),而非给 `config set` 加 flag(后者 Upsert 整条替换、且强制 key/secret)。详见 `SKILL_MEMORY_DESIGN.md`。

---

## File Structure

| 文件 | 改动 | 责任 |
|---|---|---|
| `internal/config/config.go` | 改:Profile 加 `DefaultPark`/`DefaultCarNo`;新增 `Active(profileFlag)` | A 层数据模型 + 当前 profile 选择 |
| `internal/config/defaults_test.go` | 建 | A 层 config 单测 |
| `internal/cmdutil/defaults.go` | 建 | 纯 `injectDefaults` + `Factory.applyDefaults` |
| `internal/cmdutil/defaults_test.go` | 建 | A 层注入单测(纯函数 + 真 catalog 集成) |
| `internal/cmdutil/run.go` | 改:RunCall 内挂 `body = f.applyDefaults(cmd, body)` | A 层注入挂钩点 |
| `cmd/config/config.go` | 改:新增 `set-default` 子命令;`list` 展示默认值 | A 层写入入口 |
| `cmd/config/setdefault_test.go` | 建 | `set-default` 命令测试 |
| `skills/openydt-shared/SKILL.md` | 改:新增 "## 车场经验" 节 | B 层(主)约定 |
| `README.md` / `CLAUDE.md` | 改:记忆层说明 | 文档 |

---

## Task 1: Profile 默认值字段 + Active 选择器

**Files:**
- Modify: `internal/config/config.go`(`Profile` 结构体 30-37 行;在 `Find` 之后加 `Active`)
- Test: `internal/config/defaults_test.go`

- [ ] **Step 1: 写失败测试** — 创建 `internal/config/defaults_test.go`:
```go
package config

import "testing"

func TestProfileDefaultsRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := &Config{}
	cfg.Upsert(Profile{Name: "test", Key: "k", Secret: "s", Env: "test", DefaultPark: "PTD2YBBZ", DefaultCarNo: "桂566666"})
	cfg.CurrentProfile = "test"
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	p, ok := got.Find("test")
	if !ok || p.DefaultPark != "PTD2YBBZ" || p.DefaultCarNo != "桂566666" {
		t.Fatalf("defaults round trip mismatch: %+v ok=%v", p, ok)
	}
}

func TestActiveByPrecedence(t *testing.T) {
	t.Setenv("OPENYDT_PROFILE", "")
	cfg := &Config{
		CurrentProfile: "test",
		Profiles: []Profile{
			{Name: "test", DefaultPark: "P_CUR"},
			{Name: "prod", DefaultPark: "P_PROD"},
		},
	}
	if p, ok := cfg.Active("prod"); !ok || p.DefaultPark != "P_PROD" {
		t.Fatalf("flag should win, got %+v", p)
	}
	if p, ok := cfg.Active(""); !ok || p.DefaultPark != "P_CUR" {
		t.Fatalf("should fall back to CurrentProfile, got %+v", p)
	}
	t.Setenv("OPENYDT_PROFILE", "prod")
	if p, ok := cfg.Active(""); !ok || p.DefaultPark != "P_PROD" {
		t.Fatalf("OPENYDT_PROFILE should win over CurrentProfile, got %+v", p)
	}
}

func TestActiveNoneFound(t *testing.T) {
	t.Setenv("OPENYDT_PROFILE", "")
	cfg := &Config{} // no current, no profiles
	if _, ok := cfg.Active(""); ok {
		t.Fatalf("expected no active profile")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/config/ -run "TestProfileDefaults|TestActive" -v`
Expected: 编译失败(`DefaultPark`/`DefaultCarNo`/`Active` 未定义)。

- [ ] **Step 3: 实现** — 在 `internal/config/config.go` 把 `Profile` 结构体改为(加最后两字段):
```go
// Profile is one authorized-merchant credential set.
type Profile struct {
	Name         string `json:"name"`
	Key          string `json:"key"`
	Secret       string `json:"secret"`
	Env          string `json:"env,omitempty"`  // test|dev|prod
	Sign         string `json:"sign,omitempty"` // v2|v3
	DefaultPark  string `json:"defaultPark,omitempty"`  // 缺参时自动补的 parkCode
	DefaultCarNo string `json:"defaultCarNo,omitempty"` // 缺参时自动补的车牌
}
```
并在 `Find` 方法之后新增 `Active`:
```go
// Active returns the profile selected by the same precedence as Resolve
// (flag > OPENYDT_PROFILE > CurrentProfile). Returns false when none resolves.
func (c *Config) Active(profileFlag string) (Profile, bool) {
	name := firstNonEmpty(profileFlag, os.Getenv("OPENYDT_PROFILE"), c.CurrentProfile)
	if name == "" {
		return Profile{}, false
	}
	return c.Find(name)
}
```
(`os` 与 `firstNonEmpty` 在本文件已有,无需新增 import。)

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/config/ -run "TestProfileDefaults|TestActive" -v`
Expected: PASS。

- [ ] **Step 5: 提交**
```bash
git add internal/config/config.go internal/config/defaults_test.go
git commit -m "feat(config): Profile 加 DefaultPark/DefaultCarNo + Active 选择器"
```

---

## Task 2: 缺参默认值注入(cmdutil)

**Files:**
- Create: `internal/cmdutil/defaults.go`
- Test: `internal/cmdutil/defaults_test.go`
- Modify: `internal/cmdutil/run.go`(`RunCall` 内,40-46 行的 JSON 校验之后)

依赖 Task 1(`Profile.DefaultPark`、`Config.Active`)。catalog API(本仓库已有,见 `run.go` 的 `buildErrorInfo`):`catalog.Embedded() (*Catalog, error)`、`cat.Find(cmd) (Iface, bool)`、`it.FindParam(name) (Param, bool)`。

- [ ] **Step 1: 写失败测试** — 创建 `internal/cmdutil/defaults_test.go`:
```go
package cmdutil

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

// 纯函数:用假 hasParam,不碰 catalog/config/网络
func TestInjectDefaultsPure(t *testing.T) {
	hasPark := func(n string) bool { return n == "parkCode" }

	// 缺 parkCode + cmd 收 parkCode + 有默认 → 注入
	out, inj := injectDefaults(`{}`, "PTD2YBBZ", "", hasPark)
	if !strings.Contains(out, `"parkCode":"PTD2YBBZ"`) || len(inj) != 1 {
		t.Fatalf("expected parkCode injected, got %s %v", out, inj)
	}

	// body 已有 parkCode → 不覆盖、原样返回
	out, inj = injectDefaults(`{"parkCode":"X"}`, "PTD2YBBZ", "", hasPark)
	if out != `{"parkCode":"X"}` || len(inj) != 0 {
		t.Fatalf("must not overwrite existing, got %s %v", out, inj)
	}

	// cmd 不收 parkCode → 不注入
	out, inj = injectDefaults(`{}`, "PTD2YBBZ", "", func(string) bool { return false })
	if out != `{}` || len(inj) != 0 {
		t.Fatalf("must not inject when param absent, got %s %v", out, inj)
	}

	// 无默认 → 原样
	out, _ = injectDefaults(`{}`, "", "", hasPark)
	if out != `{}` {
		t.Fatalf("no defaults => unchanged, got %s", out)
	}

	// carNo:命中第一个该 cmd 声明的车牌字段名
	hasCar := func(n string) bool { return n == "carCode" }
	out, inj = injectDefaults(`{}`, "", "桂566666", hasCar)
	if !strings.Contains(out, `"carCode":"桂566666"`) || len(inj) != 1 {
		t.Fatalf("expected carCode injected, got %s %v", out, inj)
	}
}

// 集成:走真 catalog(getBillSummary 真实含 parkCode 参数)+ 真 config(临时 XDG)
func TestApplyDefaultsWithRealCatalog(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("OPENYDT_PROFILE", "")
	cfg := &config.Config{CurrentProfile: "test"}
	cfg.Upsert(config.Profile{Name: "test", Key: "k", Secret: "s", Env: "test", DefaultPark: "PTD2YBBZ"})
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	f := &Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

	// getBillSummary 收 parkCode → 注入
	if got := f.applyDefaults("getBillSummary", "{}"); !strings.Contains(got, `"parkCode":"PTD2YBBZ"`) {
		t.Fatalf("expected parkCode injected for getBillSummary, got %s", got)
	}
	// 已显式给 parkCode → 不覆盖
	if got := f.applyDefaults("getBillSummary", `{"parkCode":"OTHER"}`); !strings.Contains(got, `"OTHER"`) || strings.Contains(got, "PTD2YBBZ") {
		t.Fatalf("must not overwrite explicit parkCode, got %s", got)
	}
	// 未知命令 → catalog 找不到 → 原样
	if got := f.applyDefaults("no_such_cmd_xyz", "{}"); got != "{}" {
		t.Fatalf("unknown cmd => unchanged, got %s", got)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/cmdutil/ -run "TestInjectDefaults|TestApplyDefaults" -v`
Expected: 编译失败(`injectDefaults`/`applyDefaults` 未定义)。

- [ ] **Step 3: 实现** — 创建 `internal/cmdutil/defaults.go`:
```go
package cmdutil

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xiaowen-0725/openydt-cli/internal/catalog"
	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

// carFields are the candidate body field names for a license plate, in priority
// order. The command's catalog schema decides which (if any) actually applies.
var carFields = []string{"carCode", "carNo"}

// injectDefaults fills missing body fields from profile defaults. It only adds a
// field when (a) the body omits it and (b) hasParam reports the command declares
// it. Existing values are never overwritten. Returns the (possibly rewritten)
// compact body and the list of injected field names. Pure + unit-testable.
func injectDefaults(body, defaultPark, defaultCarNo string, hasParam func(string) bool) (string, []string) {
	if defaultPark == "" && defaultCarNo == "" {
		return body, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		return body, nil
	}
	if m == nil {
		m = map[string]any{}
	}
	var injected []string
	if defaultPark != "" {
		if _, has := m["parkCode"]; !has && hasParam("parkCode") {
			m["parkCode"] = defaultPark
			injected = append(injected, "parkCode")
		}
	}
	if defaultCarNo != "" {
		for _, fld := range carFields {
			if hasParam(fld) {
				if _, has := m[fld]; !has {
					m[fld] = defaultCarNo
					injected = append(injected, fld)
				}
				break // command uses at most one plate field name
			}
		}
	}
	if len(injected) == 0 {
		return body, nil
	}
	out, err := json.Marshal(m)
	if err != nil {
		return body, nil
	}
	return string(out), injected
}

// applyDefaults resolves the active profile's defaults, consults the embedded
// catalog for which fields this command accepts, and injects missing ones. Any
// error (no config / no profile / no catalog) is non-fatal — the body is
// returned unchanged.
func (f *Factory) applyDefaults(cmd, body string) string {
	cfg, err := config.Load()
	if err != nil {
		return body
	}
	p, ok := cfg.Active(f.Profile)
	if !ok || (p.DefaultPark == "" && p.DefaultCarNo == "") {
		return body
	}
	cat, err := catalog.Embedded()
	if err != nil {
		return body
	}
	it, ok := cat.Find(cmd)
	if !ok {
		return body
	}
	has := func(name string) bool { _, ok := it.FindParam(name); return ok }
	out, injected := injectDefaults(body, p.DefaultPark, p.DefaultCarNo, has)
	if len(injected) > 0 && f.Verbose {
		fmt.Fprintf(f.Err, "[openydt] 已用 profile %q 默认值补全: %s\n", p.Name, strings.Join(injected, ", "))
	}
	return out
}
```

- [ ] **Step 4: 挂钩 RunCall** — 在 `internal/cmdutil/run.go` 的 `RunCall` 中,把 JSON 校验那段:
```go
	if !json.Valid([]byte(body)) {
		return usageErr(fmt.Errorf("--body 不是合法 JSON: %s", body))
	}
	c, err := f.Client()
```
改为(在校验之后、`f.Client()` 之前插入一行):
```go
	if !json.Valid([]byte(body)) {
		return usageErr(fmt.Errorf("--body 不是合法 JSON: %s", body))
	}
	body = f.applyDefaults(cmd, body) // 缺参 → 补 profile 默认值(dry-run 预览也反映)
	c, err := f.Client()
```

- [ ] **Step 5: 跑测试确认通过 + 全包构建**

Run: `go test ./internal/cmdutil/ -run "TestInjectDefaults|TestApplyDefaults" -v && go build ./...`
Expected: PASS,build 无输出。

- [ ] **Step 6: 提交**
```bash
git add internal/cmdutil/defaults.go internal/cmdutil/defaults_test.go internal/cmdutil/run.go
git commit -m "feat(cmdutil): RunCall 缺参注入 profile 默认 parkCode/carNo(查 catalog)"
```

---

## Task 3: `openydt config set-default` 命令 + list 展示默认值

**Files:**
- Modify: `cmd/config/config.go`(`New` 注册;`newList` 内追加默认值展示;新增 `newSetDefault`)
- Test: `cmd/config/setdefault_test.go`

依赖 Task 1。

- [ ] **Step 1: 写失败测试** — 创建 `cmd/config/setdefault_test.go`:
```go
package config

import (
	"bytes"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

func TestSetDefaultUpdatesCurrentProfile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("OPENYDT_PROFILE", "")
	seed := &config.Config{CurrentProfile: "test"}
	seed.Upsert(config.Profile{Name: "test", Key: "k", Secret: "s", Env: "test"})
	if err := seed.Save(); err != nil {
		t.Fatal(err)
	}

	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	cmd := New(f)
	cmd.SetArgs([]string{"set-default", "--park", "PTD2YBBZ", "--car-no", "桂566666"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("set-default: %v", err)
	}

	got, _ := config.Load()
	p, _ := got.Find("test")
	if p.DefaultPark != "PTD2YBBZ" || p.DefaultCarNo != "桂566666" {
		t.Fatalf("defaults not saved: %+v", p)
	}
	// 凭据未被破坏
	if p.Key != "k" || p.Secret != "s" {
		t.Fatalf("creds clobbered: %+v", p)
	}
}

func TestSetDefaultRequiresAtLeastOne(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	seed := &config.Config{CurrentProfile: "test"}
	seed.Upsert(config.Profile{Name: "test", Key: "k", Secret: "s"})
	_ = seed.Save()
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	cmd := New(f)
	cmd.SetArgs([]string{"set-default"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error when neither --park nor --car-no given")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd/config/ -run TestSetDefault -v`
Expected: FAIL(`set-default` 未知命令 → Execute 返回 error,或编译期 newSetDefault 未定义)。

- [ ] **Step 3: 实现** — 在 `cmd/config/config.go`:

(a) `New` 注册新子命令:
```go
	cmd.AddCommand(newSet(f), newList(f), newUse(f), newPath(f), newSetDefault(f))
```

(b) `newList` 的循环里,在原 `fmt.Fprintf(... env=%s sign=%s\n ...)` 那条之后追加默认值行:
```go
			fmt.Fprintf(f.Out, "%s%-16s key=%s secret=%s env=%s sign=%s\n",
				marker, p.Name, p.Key, mask(p.Secret), orDefault(p.Env, config.DefaultEnv), orDefault(p.Sign, config.DefaultSign))
			if p.DefaultPark != "" || p.DefaultCarNo != "" {
				fmt.Fprintf(f.Out, "    默认: park=%s carNo=%s\n", orDefault(p.DefaultPark, "-"), orDefault(p.DefaultCarNo, "-"))
			}
```

(c) 新增 `newSetDefault`(放在 `newPath` 之后):
```go
func newSetDefault(f *cmdutil.Factory) *cobra.Command {
	var profile, park, carNo string
	cmd := &cobra.Command{
		Use:   "set-default",
		Short: "为某 profile 设置默认 parkCode / 车牌(命令缺参时自动补全)",
		Example: `  openydt config set-default --park PTD2YBBZ --car-no 桂566666
  openydt config set-default --profile prod --park 1ZS7H5PQH9`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if park == "" && carNo == "" {
				return cmdutil.ExitError{Code: output.ExitUsage, Err: fmt.Errorf("至少提供 --park 或 --car-no")}
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			name := profile
			if name == "" {
				name = cfg.CurrentProfile
			}
			if name == "" {
				return cmdutil.ExitError{Code: output.ExitUsage, Err: fmt.Errorf("无当前 profile;先 openydt config set 或加 --profile")}
			}
			p, ok := cfg.Find(name)
			if !ok {
				return cmdutil.ExitError{Code: output.ExitUsage, Err: fmt.Errorf("profile %q 不存在", name)}
			}
			if park != "" {
				p.DefaultPark = park
			}
			if carNo != "" {
				p.DefaultCarNo = carNo
			}
			cfg.Upsert(p)
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(f.Out, "已更新 profile %q 默认值: park=%s carNo=%s\n", name, orDefault(p.DefaultPark, "-"), orDefault(p.DefaultCarNo, "-"))
			return nil
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&profile, "profile", "", "目标 profile(默认当前 profile)")
	fl.StringVar(&park, "park", "", "默认 parkCode")
	fl.StringVar(&carNo, "car-no", "", "默认车牌")
	return cmd
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./cmd/config/ -v`
Expected: PASS。

- [ ] **Step 5: 提交**
```bash
git add cmd/config/config.go cmd/config/setdefault_test.go
git commit -m "feat(cmd/config): set-default 子命令 + list 展示默认值"
```

---

## Task 4: B 层(主)—— openydt-shared 车场经验约定

**Files:**
- Modify: `skills/openydt-shared/SKILL.md`(在 "## 测试车场" 节之前插入 "## 车场经验" 节)

无 Go 代码;用 `node scripts/skill-format-check/index.js` 验证仍合规。

- [ ] **Step 1: 插入新节** — 在 `skills/openydt-shared/SKILL.md` 的 "## 测试车场(仅测试环境)" 行之前,插入:
```markdown
## 车场经验(自动沉淀,跨 session 复用)

按车场积累的经验存在 **openydt 配置目录**下的 `park-notes/{parkCode}.md`(默认 `~/.config/openydt-cli/park-notes/`,或 `$XDG_CONFIG_HOME/openydt-cli/park-notes/`;父目录即 `openydt config path` 所示文件的所在目录)。**不要**放进技能目录——技能会被 `npx skills` 同步覆盖,经验会丢。

**任务开始前(回忆)**:
1. `ls ~/.config/openydt-cli/park-notes/`(目录不存在或为空属正常)。文件名即 parkCode;必要时读各文件 frontmatter 的 `aliases` 做车场名/俗称匹配。
2. 用户指明目标车场(parkCode 或别名)后,若有匹配文件**必须先 Read** 它,据此选择签名版本、必填字段、避开已知陷阱、复用常用车牌。
3. 经验标注 `updated` 日期,**当"可能有效的提示"而非保证**;按经验操作若失败 → 回退通用流程,并**更新**该文件对应条目。

**openydt 命令成功(status=1)后(沉淀)**:
4. 若发现该车场值得记录的**已验证**新事实(可用签名版本、必填 body 字段、某接口在该环境 nodata、计费模式、稳定的关联 ID、常用车牌),主动**追加/更新**到 `park-notes/{parkCode}.md` 对应小节并刷新 frontmatter `updated`。文件不存在则按下方模板创建(先 `mkdir -p` 该目录)。
5. **只写经过验证的事实,不写猜测。**
6. **隐私红线**:车牌是 PII —— `常用车牌` 仅在 **test/dev** 记录;**prod 环境不要写真实车牌**(可写"该车场有 VIP 车类型"这类非 PII 事实)。frontmatter 用 `env` 标明经验来自哪个环境,避免把 test 经验误用到 prod。

文件格式:
```markdown
---
parkCode: PTD2YBBZ
aliases: [智汇云测试车场]
traderCode: O563CNQTNZVY
env: test
updated: 2026-05-30
---
## 车场特征
- 云车场;test 环境授权;按时段计费
## 有效调用
- getParkFee:v2 签名调通,body 仅需 carCode
## 已知陷阱
- dataAnalysis/* 在 test 多为 nodata
## 常用车牌
- 桂566666(普通)、粤HH7772(VIP)   # 仅 test/dev
```
```

- [ ] **Step 2: 校验技能格式仍合规**

Run: `node scripts/skill-format-check/index.js 2>&1 | tail -3`
Expected: `openydt-shared` 仍 PASS(或维持原 WARN,不得新增 FAIL)。

- [ ] **Step 3: 提交**
```bash
git add skills/openydt-shared/SKILL.md
git commit -m "feat(skills): openydt-shared 新增车场经验(自动沉淀/回忆)约定"
```

---

## Task 5: 文档

**Files:**
- Modify: `README.md`(AI Agent Skills 段附近)
- Modify: `CLAUDE.md`(新增"记忆/默认值"小节)

- [ ] **Step 1: README 追加** — 在 README "AI Agent Skills" 那段之后追加一段:
```markdown
**记忆 / 默认值**:
- 默认车场/车牌(随 profile,按环境隔离):`openydt config set-default --park PTD2YBBZ --car-no 桂566666`;之后命令缺 `parkCode` 自动补全(显式值永远优先,`--verbose` 可见补全)。
- 车场经验(AI agent 自动沉淀/回忆):存 `~/.config/openydt-cli/park-notes/{parkCode}.md`,agent 操作成功后自动追加该车场的有效调用/陷阱/常用车牌,下次自动回忆。清理:删对应 `.md`。prod 不记真实车牌。
```

- [ ] **Step 2: CLAUDE.md 追加** — 在 `## 技能同步(skillsync)` 节之后插入:
```markdown
## 记忆 / 默认值
- **A 层(Profile 默认值)**:`Profile.DefaultPark/DefaultCarNo`(`internal/config`);`openydt config set-default` 写;`RunCall` 经 `cmdutil/applyDefaults` 在缺参时注入(查内嵌 catalog 是否收该字段,只补缺失、不覆盖显式值,dry-run 预览反映注入)。按环境隔离(随 profile)。
- **B 层(车场经验,主)**:`~/.config/openydt-cli/park-notes/{parkCode}.md`(frontmatter+Markdown)。**纯 `openydt-shared/SKILL.md` 约定,无 Go 代码**。agent 启动回忆 + 成功后沉淀已验证事实;存 config 目录避免被技能同步擦除;prod 不记 PII 车牌。
```

- [ ] **Step 3: 提交**
```bash
git add README.md CLAUDE.md
git commit -m "docs: 记忆/默认值(A profile 默认值 + B 车场经验)说明"
```

---

## Task 6: 全量验证

- [ ] **Step 1: 测试 + vet**

Run: `go test ./... 2>&1 | grep -vE "no test files" && go vet ./...`
Expected: 相关包(config / cmdutil / cmd/config 等)全 `ok`,vet 无输出。

- [ ] **Step 2: 跨平台编译**

Run:
```bash
GOOS=windows GOARCH=amd64 go build ./... && echo win-ok
GOOS=linux GOARCH=arm64 go build ./... && echo linux-ok
```
Expected: 两行 ok。

- [ ] **Step 3: 技能格式校验**

Run: `node scripts/skill-format-check/index.js 2>&1 | tail -2`
Expected: 无新增 FAIL。

- [ ] **Step 4: 真机冒烟(默认值注入,用 --dry-run 不发网)**

Run:
```bash
make build
export XDG_CONFIG_HOME=$(mktemp -d)
./bin/openydt config set --profile test --key test --secret 123456 --env test
./bin/openydt config set-default --park PTD2YBBZ
./bin/openydt config list
./bin/openydt --dry-run --verbose api getBillSummary 2>&1 | grep -i parkCode
```
Expected:`config list` 显示 `默认: park=PTD2YBBZ`;dry-run 预览的请求体里含 `"parkCode":"PTD2YBBZ"`(自动补全)。

---

## Self-Review(已执行)

- **Spec 覆盖**:B 层(车场经验:存储位置/格式/回忆/沉淀/隐私)→ Task 4;A 层(Profile 字段)→ Task 1、(set-default 写)→ Task 3、(缺参 fallback)→ Task 2;文档 → Task 5;验证 → Task 6。spec §5.2 的"set 加 flag"细化为独立 `set-default` 子命令(已在 plan 顶部标注)。
- **占位符**:无 TBD/TODO;每个代码步骤含完整代码与精确命令。
- **类型一致**:`Profile.DefaultPark/DefaultCarNo`、`Config.Active(profileFlag)(Profile,bool)`、`injectDefaults(body,defaultPark,defaultCarNo string, hasParam func(string)bool)(string,[]string)`、`Factory.applyDefaults(cmd,body string) string`、catalog `Embedded()/Find/FindParam` 在各任务间一致;`getBillSummary`(真实含 parkCode 参数)用于集成测试。
- **不碰 generated**:无 `cmd/gen/*.go` 改动。
