# 设计:openydt 记忆/沉淀层(车场经验记忆 B + Profile 默认值 A)

> 日期:2026-05-30 · 状态:已确认形态,待写实现计划
> 调研依据:本会话两轮研究(飞书CLI/lark-skills/Agent生态/CLI自动学习/Agent自动记忆框架/格式选型/可行性裁决)+ 研读 `github.com/eze-is/web-access` 的"站点经验积累"机制。

## 1. 背景与问题

用户(人和 AI agent)每次用 openydt 都要重复一堆背景:**测哪个车场**(parkCode,如 PTD2YBBZ)、常用车牌、某车场的坑/有效调用方式、用哪个环境。目标:让这些被**自动沉淀**、下次**自动回忆/带上**,不必每次重述。

### 调研定论(指导本设计)
- **"自动套用进真实请求"有风险**(gh 自动默认 repo 致 agent 误发上游 PR、Terraform 隐式 workspace 致生产事故),尤其 openydt 有缴费/开闸写操作 + test/prod 红线 + agent 用户 → 碰请求的层要**保守**。
- **"自动沉淀知识给 agent 读"是安全且成熟的**(web-access 站点经验、feishu-release-approval 历史回填、Claude Code auto-memory)。护栏共识:**只记验证过的事实、当提示不当真、失败即回退并更新、按实体隔离、家目录存储**。
- **格式**:机器自动写的结构数据 → **JSON**;LLM 读写的经验知识 → **Markdown(带 YAML frontmatter)**。
- **多环境切换+记忆**:openydt **现已支持**(`Profile{Env,Sign}` + 持久化 `CurrentProfile`,`Resolve()` 优先级 flag>env>profile>默认)——本设计**不重做**,只复用。

## 2. 目标 / 非目标

### 目标
- **B 层(主):车场经验记忆** —— 按 parkCode 存一份经验文件(车场特征/有效调用/已知陷阱/常用车牌)。AI agent 启动时自动列出+读取匹配车场,**操作成功后自动追加验证过的事实**。让"测某车场"时背景自动到位。
- **A 层(辅):Profile 默认值** —— 给 Profile 增加 `DefaultPark`/`DefaultCarNo`,命令缺参时 fallback 到当前 profile 的默认值,**按环境隔离**(prod/test 各自的默认车场)。

### 非目标
- 不做"自动套用历史值进真实请求"的隐式学习(调研裁决:弊大于利)。A 层只用**显式 set 的**默认值。
- 不重做多环境/凭据/切换记忆(Profile 已支持)。
- B 层不引入向量库/语义检索(纯文件 + agent 读,对标 web-access)。
- 不改 `cmd/gen/*.go`(生成产物)。

## 3. 三层全景

| 层 | 内容 | 落点 | 格式 | 谁写 | 谁读 |
|---|---|---|---|---|---|
| **Profile(已有,不动)** | 多环境 key/secret/env/sign + 当前选择 | `~/.config/openydt-cli/config.json` | JSON | 用户 `config set/use` | Go `Resolve()` |
| **A 辅(新,极小)** | 默认 parkCode/carCode(并入 Profile) | 同 `config.json` 的 profile 字段 | JSON | 用户 `config set --default-park` | Go RunCall 缺参 fallback |
| **B 主(新)** | 车场经验:特征/有效调用/陷阱/常用车牌 | `~/.config/openydt-cli/park-notes/{parkCode}.md` | YAML frontmatter + Markdown | **AI agent 验证成功后自动写** | AI agent 启动自动回忆 |

关键:**A 层碰真实请求 → 保守(只显式默认值);B 层只是给 agent 读的知识 → 可全自动沉淀。** 二者互不耦合,可独立实现/测试。

---

## 4. B 层(主):车场经验记忆 —— 对标 web-access 站点经验

### 4.1 存储位置(已由技术现实定死)
`~/.config/openydt-cli/park-notes/{parkCode}.md`,一车场一文件。

⚠️ **为什么不放进 skill 目录**(web-access 是放 `references/site-patterns/`):openydt 的技能由 `npx skills` 分发,且**本仓库已实现技能自动同步**(install/update 时 `npx skills add` 重装、二进制漂移后台补同步)。技能目录会在再同步时被**重新拉取覆盖**,放里面的运行时经验会被擦掉。故必须放 `config.Dir()`(`~/.config/openydt-cli/`)下,**独立于技能包,跨同步存活**。技能里只放**指令**(指向该路径),数据在 config 目录。

### 4.2 文件格式(frontmatter 机读 + 正文 LLM 读)
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
- 月票开通:需先确认车位号 + monthCfgId
## 已知陷阱
- dataAnalysis/* 在 test 多为 nodata
- deleteTrader 破坏性,跳过
## 常用车牌
- 桂566666(普通)、粤HH7772(VIP)  # 仅 test 环境记录
```
- frontmatter:`parkCode`(主键)、`aliases`(俗称,便于"朝阳大悦城"这类自然语言匹配)、可选 `traderCode`、`env`、`updated`(发现/更新日期)。
- 正文四节(车场特征/有效调用/已知陷阱/常用车牌),自由 Markdown,LLM 读。

### 4.3 自动回忆(读)—— 写进 `openydt-shared/SKILL.md`
所有 openydt 域技能本就先 Read `openydt-shared`。在其新增 "## 车场经验" 节,约定:
1. 任务开始前,`ls ~/.config/openydt-cli/park-notes/` 列出已有车场经验(文件名=parkCode;可读 frontmatter 的 aliases 做名称匹配)。
2. 用户指明目标车场(parkCode 或别名)后,**若有匹配文件,必须先 Read** 它,获取先验(特征/有效调用/陷阱/车牌)。
3. 经验标注 `updated` 日期,**当"可能有效的提示"而非保证**;若按经验操作失败 → 回退通用流程 + 更新该文件对应条目。

### 4.4 自动沉淀(写)—— 写进 `openydt-shared/SKILL.md`
4. openydt 命令**成功返回(status=1)后**,若发现该车场值得记录的**已验证**新事实(有效的签名版本/必需 body 字段、某接口在该环境 nodata、计费模式、稳定的关联 ID、常用车牌),**主动追加/更新**到 `park-notes/{parkCode}.md` 对应小节,并刷新 frontmatter `updated`。
5. **只写经过验证的事实,不写猜测**(对标 web-access、调研护栏"宁可漏记不可错记")。
6. 文件不存在则 `mkdir -p` 后按 4.2 模板创建。

### 4.5 隐私 / prod 红线护栏(写进 SKILL.md 约定)
- **车牌是 PII**:`常用车牌` 小节**只在 test/dev 环境**记录;**prod 环境不自动记真实车牌**(可记"该车场有 VIP 车类型"这类非 PII 事实)。
- parkCode/traderCode 是场地标识,非 PII,可记。
- frontmatter `env` 标明经验来自哪个环境,避免把 test 经验误用到 prod。
- 提供清理方式:`rm ~/.config/openydt-cli/park-notes/{parkCode}.md`(README 说明;可选后续加 `openydt config forget-park <code>`,本期不做)。

### 4.6 种子(可选,降低冷启动)
把现有 `openydt-shared/SKILL.md` "## 测试车场" 里静态的 PTD2YBBZ / 1ZS7H5PQH9 信息,作为 README/文档示例给出对应 `park-notes` 样例,用户/agent 首次可参照生成。**不预置文件**(保持"运行时积累"语义)。

### 4.7 B 层不需要 Go 代码
B 层 = `openydt-shared/SKILL.md` 的指令约定 + `~/.config/openydt-cli/park-notes/` 目录约定。agent 用 Read/Write/ls 操作。**零 Go 改动**(对标 web-access 以 SKILL.md 驱动)。

---

## 5. A 层(辅):Profile 默认值

### 5.1 数据模型(扩展现有 Profile)
`internal/config/config.go` 的 `Profile` 增加两个可选字段:
```go
type Profile struct {
    Name   string `json:"name"`
    Key    string `json:"key"`
    Secret string `json:"secret"`
    Env    string `json:"env,omitempty"`
    Sign   string `json:"sign,omitempty"`
    DefaultPark  string `json:"defaultPark,omitempty"`  // A 层新增
    DefaultCarNo string `json:"defaultCarNo,omitempty"` // A 层新增
}
```
默认值随 profile 走 → **天然按环境隔离**(prod/test 各自默认车场)。

### 5.2 写入(扩展 `config set`)
`cmd/config/config.go` 的 `set` 增加 flag `--default-park`、`--default-car-no`,写入对应 profile;`config list` 展示时附带显示默认值。示例:
```bash
openydt config set --profile test --key test --secret 123456 --env test --default-park PTD2YBBZ --default-car-no 桂566666
```

### 5.3 缺参 fallback(在统一调用链注入)
在 `internal/cmdutil/run.go` 的 `RunCall(cmd, body)` 内、签名/发送之前:
- 解析 compact body;若**缺 `parkCode`** 且 (a) 该 cmd 的内嵌 catalog schema 声明了 `parkCode` 参数,(b) 当前 profile 有 `DefaultPark` → 注入 `DefaultPark`。
- `DefaultCarNo` 同理注入车牌字段(字段名从 catalog schema 解析,常见 `carCode`;best-effort)。
- **优先级**:body 中已存在的字段值(无论来自 `--body`、`--body-file` 还是命令自带的字段 flag,到 RunCall 时都已组装进 body)> profile 默认值 > 无(原样)。即只补**缺失**字段,绝不覆盖任何已显式给出的值。
- **dry-run**:在注入**之后**再打印预览,使预览与实际发送一致。
- **可观测**:注入了默认值时,`--verbose` 在 stderr 标注 "已用 profile 默认 parkCode=PTD2YBBZ"。
- 注入逻辑放 `internal/cmdutil`(factory/run 层),**不改 generated 命令**。

### 5.4 A 层安全护栏
- prod 写操作仍由生成命令的 `--yes` 守护(不变)。默认值只补**缺失**字段,不覆盖显式值,降低误操作面。
- 默认值是**显式 set 的**(非自动学习),可预测、可 `config set` 改、可清。

---

## 6. 组件 / 改动清单

| 文件 | 改动 | 层 |
|---|---|---|
| `skills/openydt-shared/SKILL.md` | 新增 "## 车场经验" 节(回忆+沉淀+隐私约定),路径指向 `~/.config/openydt-cli/park-notes/` | B(主) |
| `internal/config/config.go` | Profile 加 `DefaultPark`/`DefaultCarNo`;`Resolved` 透传;(可加 `config.Dir()` 已存在) | A |
| `cmd/config/config.go` | `set` 加 `--default-park`/`--default-car-no`;`list` 展示默认值 | A |
| `internal/cmdutil/run.go`(+ 可能 `body.go`/`factory.go`) | RunCall 缺参 fallback 注入(查 catalog schema 是否有该字段) | A |
| `internal/cmdutil/*_test.go`、`internal/config/*_test.go` | A 层单测 | A |
| `README.md` / `CLAUDE.md` | 记忆层说明(B 路径/格式/隐私、A 默认值用法) | 文档 |
| `scripts/skill-format-check` | 不变(SKILL.md 仍合规) | — |

B 层零 Go 代码;A 层改动集中在 config + cmdutil,**不碰 generated**。

## 7. 数据流

**人用 CLI(A 层)**:`config set --default-park PTD2YBBZ` → 之后 `openydt api getParkFee --body '{"carCode":"粤A"}'`(没带 parkCode)→ RunCall 注入当前 profile 的 defaultPark → 签名发送。`--verbose` 可见注入。

**AI agent 对话(B 层)**:用户"测一下 PTD2YBBZ" → agent Read `openydt-shared` → 按约定 `ls park-notes/` → 命中 `PTD2YBBZ.md` 先读(知道 v2 签名、必需字段、坑、车牌)→ 直接发起正确调用,无需用户重述 → 成功后把新验证事实追加回 `PTD2YBBZ.md`。

## 8. 错误处理 / 护栏小结
- B 层:经验当提示不当真;失败回退通用流程并更新经验;只写已验证事实;prod 不记 PII 车牌;数据在 config 目录、不随技能同步丢失。
- A 层:只补缺失字段不覆盖显式值;dry-run 预览反映注入后 body;prod 写操作仍需 `--yes`;catalog 无该字段则不注入(避免给不收 parkCode 的接口塞参)。

## 9. 测试策略
- **A 层单测**(对标 CLAUDE.md 锚定风格):
  - config:Profile 含 DefaultPark/DefaultCarNo 的读写往返;`config set --default-park` 写入正确 profile。
  - RunCall fallback:body 缺 parkCode + cmd schema 有 parkCode + profile 有默认 → 注入;body 已有 parkCode → 不覆盖;cmd schema 无 parkCode → 不注入;无默认 → 不变。用注入式假 client / 断言最终 body(不真实发网)。
- **B 层**:`node scripts/skill-format-check/index.js` 仍 PASS;人工/对话验证回忆+写入闭环(真实文件读写,不进 Go 单测)。
- `go test ./... && go vet ./...` 全绿;跨平台 build。

## 10. 风险 / 已知点
- **catalog 字段名解析**:车牌字段命名不一(carCode/carNo/plateNo)。A 层对 parkCode(命名一致、价值最高)做可靠注入;carNo 注入按 catalog schema best-effort,若解析不到则跳过(不误注)。
- **B 层是 agent 行为约定,非强约束**:依赖 agent 遵循 SKILL.md(概率遵从)。可接受——失败只是回退到"像现在一样手动给背景",无副作用。
- **经验陈旧**:靠 `updated` 日期 + "失败即更新" 自纠;不做 TTL(文件少、人可查可删)。
- **跨 agent 平台**:park-notes 在 `~/.config/openydt-cli/`,与具体 agent 无关,所有读了 `openydt-shared` 的 agent 共享同一份经验。
