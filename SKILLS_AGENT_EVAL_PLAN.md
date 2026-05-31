# openydt-cli 技能 Agent 易用性评测 — 实现计划(Phase 1:评测)

> **For agentic workers:** REQUIRED SUB-SKILL: 用 `superpowers:executing-plans`(本计划核心是用 Workflow 工具编排一次评测,**建议 inline 执行**——主会话持有 Workflow 工具与全局上下文;subagent-per-task 不适合"任务本身就是 fan-out"的场景)。步骤用 checkbox(`- [ ]`)跟踪。

**Goal:** 跑一次研究驱动的评测(对标→静态审查→实证),产出根目录 `EVALUATION.md`(16 维热力打分表 + 对标差距 + 实证发现 + P0/P1/P2 backlog + 改进设计蓝图),交付到用户审批门。

**Architecture:** 单个 Workflow 脚本四相(Benchmark → StaticAudit → Empirical → Synthesize):Benchmark 5 个 agent 对标提炼;StaticAudit 12 技能 + CLI 面 + 全集一致性 按 16 维结构化打分(A1 走三方机械核对);Empirical 5 个用例各 actor→judge 两段(全程 read-only/dry-run/test-only);Synthesize 一个 agent 把结构化结果写成 `EVALUATION.md` + `tools/eval/eval-output.json`。评测阶段不落地任何技能改动。

**Tech Stack:** Workflow 工具(JS 脚本,`agent/parallel/pipeline/phase/log`,JSON Schema 结构化输出)、Go(`make build`/`go run`)、`jq`、openydt CLI(test 环境只读)。

**依据 spec:** `SKILLS_AGENT_EVAL_DESIGN.md`(同仓库根目录,已批准)。

---

## File Structure

| 文件 | 职责 | 动作 |
|---|---|---|
| `tools/eval/skills-agent-eval.workflow.mjs` | 评测 Workflow 脚本(四相编排 + 16 维 rubric + schema) | Create |
| `tools/eval/eval-output.json` | Workflow 产出的原始结构化聚合(供复核/复算) | Create(由 Workflow 写) |
| `EVALUATION.md`(仓库根) | 评测报告(7 节,见 spec §6) | Create(由 Workflow 写,Task 5 finalize) |
| `bin/openydt` | 评测用二进制(静态审查 A1 核对 + 实证只读调用) | Build(已 gitignore) |

> 不改:`cmd/gen/*.go`、`catalog/catalog.json`、`skills/**`(评测阶段只读它们)。

---

## Task 0：前置条件(构建 + 目录 + 凭据自检)

**Files:**
- Build: `bin/openydt`
- Create dir: `tools/eval/`

- [ ] **Step 1：构建二进制**

Run:
```bash
cd /Users/zhoujw/develop/tmp/openydt-cli && make build
```
Expected: 产出 `bin/openydt`;`ls -l bin/openydt` 可见。若 `make` 不可用:`go build -o bin/openydt .`。

- [ ] **Step 2：建评测目录**

Run:
```bash
mkdir -p /Users/zhoujw/develop/tmp/openydt-cli/tools/eval && echo ok
```
Expected: `ok`。

- [ ] **Step 3：凭据自检(决定实证哪些用例能跑活)**

Run:
```bash
cd /Users/zhoujw/develop/tmp/openydt-cli && ./bin/openydt config list 2>&1 | head -20; ./bin/openydt auth test 2>&1 | tail -5
```
Expected 之一:
- `✓ 认证通过 (status=1)` 并列出授权车场 → 实证 5 用例全部可跑活。
- 失败(无 profile / status=4/5/6)→ **不阻断计划**:实证里依赖实时只读调用的用例(E-pay 查费、E-onsite 在场车)会由 actor 自报 `blocked=true`;不依赖实时调用的用例(E-route 路由、E-error 诊断、E-api 到 dry-run 为止)仍照常跑。把这一情况记进最终报告(no silent caps),并在收尾告知用户"补 test 凭据可补全实证"。

> 安全:本计划全程只读 / `--dry-run` / 仅 test;任何步骤都不得对 prod 操作、不得加 `--yes` 真发写操作、不得打印 secret。

---

## Task 1：写评测 Workflow 脚本

**Files:**
- Create: `tools/eval/skills-agent-eval.workflow.mjs`

- [ ] **Step 1：写入完整脚本**

把以下内容**逐字**写入 `tools/eval/skills-agent-eval.workflow.mjs`:

```javascript
export const meta = {
  name: 'openydt-skills-agent-eval',
  description: 'openydt-cli 技能 Agent 易用性评测:对标→静态审查→实证→合成 EVALUATION.md',
  phases: [
    { title: 'Benchmark', detail: '对标 5 类产品,提炼可借鉴模式' },
    { title: 'StaticAudit', detail: '12 技能 + CLI 面 + 全集一致性 按 16 维打分' },
    { title: 'Empirical', detail: '5 个实证用例:actor 执行 → judge 评分(read-only/dry-run)' },
    { title: 'Synthesize', detail: '合成 EVALUATION.md + 改进设计 + eval-output.json' },
  ],
}

const REPO = (args && args.repo) ? args.repo : '/Users/zhoujw/develop/tmp/openydt-cli'

// ---- 16 维 rubric(与 spec §2 一致)----
const DIMENSIONS = [
  { id:'A1', cat:'A 内容正确性', name:'真实性/与真相源漂移', grain:'skill', desc:'命令/参数/枚举与 catalog.json(included:true)+cmd/gen/*.go+`openydt <域> --help` 逐条一致;有无幻觉/陈旧命令/过期计数' },
  { id:'A2', cat:'A 内容正确性', name:'callable 盲区覆盖', grain:'skill', desc:'域内 direction=callable 接口是否都有一等命令或 api 兜底指路,无遗漏' },
  { id:'A3', cat:'A 内容正确性', name:'示例可直接跑', grain:'skill', desc:'示例用文档化测试 parkCode(1ZS7H5PQH9/PTD2YBBZ)/当前时间;不照抄 catalog 历史 sampleBody' },
  { id:'B1', cat:'B 结构', name:'召回/触发去冲突', grain:'skill', desc:'description WHAT+WHEN、约100-150字、不堆砌;裸触发短语单 owner、无兄弟域冲突' },
  { id:'B2', cat:'B 结构', name:'渐进披露 references/', grain:'skill', desc:'SKILL.md 精简(≤500行)、大块下沉 references/ 并写按需加载指引' },
  { id:'B3', cat:'B 结构', name:'跨技能路由', grain:'skill', desc:'[[links]] 到位、边界清晰(域vs编排、读vs写归属)' },
  { id:'B4', cat:'B 结构', name:'跨技能一致性', grain:'set', desc:'结构骨架/术语/格式统一,Agent 可迁移学习(命令表列名、安全措辞、CRITICAL 头)' },
  { id:'C1', cat:'C 指令与行为', name:'规则清晰/可执行', grain:'skill', desc:'MUST/NEVER/决策树明确、可执行、无法忽略;硬约束醒目' },
  { id:'C2', cat:'C 指令与行为', name:'结果解读支撑', grain:'skill', desc:'教 Agent 读响应:金额单位(元vs分)、status vs data.code、字段含义、0条≠无' },
  { id:'C3', cat:'C 指令与行为', name:'错误自愈闭环', grain:'skill', desc:'错误→诊断→下一步闭环;用 hint/retriable;常见失败速查' },
  { id:'C4', cat:'C 指令与行为', name:'意图澄清/确认前置', grain:'skill+emp', desc:'先消歧(parkCode/env/支付方式),写操作先问再 dry-run 后 yes' },
  { id:'C5', cat:'C 指令与行为', name:'遵循度(实证)', grain:'skill+emp', desc:'实跑是否真先读 shared/加 --yes/按路由选域(由实证度量)' },
  { id:'D1', cat:'D 安全稳健', name:'写操作安全(含PII/限速/env)', grain:'skill', desc:'--yes/--dry-run/prod门;prod不记车牌;批量节流;test/dev/prod隔离' },
  { id:'D2', cat:'D 安全稳健', name:'幂等/重试安全', grain:'skill', desc:'写操作(缴费 billCode 唯一性/补缴)是否教幂等键,重试不重复扣费' },
  { id:'E1', cat:'E CLI 可供性', name:'CLI 可供性', grain:'cli', desc:'--help/schema/JSON契约(_error)/机器可读错误便于自解析自纠' },
  { id:'E2', cat:'E CLI 可供性', name:'可诊断/可观测', grain:'cli', desc:'--verbose/--dry-run/请求回显作为调试支点' },
]
const SKILLS = ['openydt-shared','openydt-skill-maker','openydt-api-explorer','openydt-flow-park-access','openydt-billing','openydt-coupon','openydt-data','openydt-device','openydt-list','openydt-monthticket','openydt-park','openydt-record']
const PER_SKILL_DIMS = DIMENSIONS.filter(d => d.grain === 'skill' || d.grain === 'skill+emp')
const CLI_DIMS = DIMENSIONS.filter(d => d.grain === 'cli')
const SET_DIMS = DIMENSIONS.filter(d => d.grain === 'set')
const rubricText = (dims) => dims.map(d => `- ${d.id} ${d.name}(${d.cat}):${d.desc}`).join('\n')

// ---- schema ----
const SCORE_SHAPE = { type:'object', additionalProperties:false, properties:{
  value:{ type:'string', enum:['1','2','3','4','5','NA'] },
  evidence:{ type:'string', description:'file:line 或具体引用' },
  gap:{ type:'string', description:'差距/问题;无则写 none' },
}, required:['value','evidence','gap'] }
const scorecardSchema = (dimIds) => ({ type:'object', additionalProperties:false, properties:{
  target:{ type:'string' },
  scores:{ type:'object', additionalProperties:false,
    properties: Object.fromEntries(dimIds.map(id => [id, SCORE_SHAPE])), required: dimIds },
  topGaps:{ type:'array', items:{ type:'string' } },
}, required:['target','scores','topGaps'] })
const BENCHMARK_SCHEMA = { type:'object', additionalProperties:false, properties:{
  product:{ type:'string' },
  patterns:{ type:'array', items:{ type:'object', additionalProperties:false, properties:{
    pattern:{ type:'string' }, dims:{ type:'array', items:{ type:'string' } },
    oursToday:{ type:'string' }, gap:{ type:'string' }, recommendation:{ type:'string' },
  }, required:['pattern','dims','oursToday','gap','recommendation'] } },
}, required:['product','patterns'] }
const TRACE_SCHEMA = { type:'object', additionalProperties:false, properties:{
  case:{ type:'string' }, blocked:{ type:'boolean' }, blockedReason:{ type:'string' },
  readSharedFirst:{ type:'boolean' }, commands:{ type:'array', items:{ type:'string' } },
  usedYesOnReadonly:{ type:'boolean' }, usedDryRunBeforeYes:{ type:'boolean' }, askedBeforeWrite:{ type:'boolean' },
  interpretation:{ type:'string' }, steps:{ type:'integer' }, narrative:{ type:'string' },
}, required:['case','blocked','blockedReason','commands','narrative'] }
const VERDICT_SCHEMA = { type:'object', additionalProperties:false, properties:{
  case:{ type:'string' },
  checkpoints:{ type:'array', items:{ type:'object', additionalProperties:false, properties:{
    dim:{ type:'string' }, name:{ type:'string' }, verdict:{ type:'string', enum:['pass','partial','fail','blocked'] }, note:{ type:'string' },
  }, required:['dim','name','verdict','note'] } },
  rootCause:{ type:'string' },
}, required:['case','checkpoints','rootCause'] }

// ================= Phase 0: Benchmark =================
phase('Benchmark')
const BENCH = [
  { product:'飞书 lark CLI + 本机已装 lark-* 技能', focus:'同架构(Go+Cobra+npx skills)技能结构/references用法/frontmatter/shared基座/命令表风格/workflow技能范式。lark-* 技能就在本机:可用 Bash `ls ~/.claude/plugins/cache/**/skills/ 2>/dev/null` 或在已加载技能列表里找 lark-* 的 SKILL.md 直接 Read。' },
  { product:'Stripe(Agent Toolkit + LLM 文档 + idempotency)', focus:'写操作幂等键/确认/可预览范式;llms.txt 式 LLM 友好文档结构。' },
  { product:'MCP 服务端(GitHub / Supabase / Cloudflare)', focus:'工具描述写法/只读模式/结构化错误/project-scoping/destructive&idempotent 提示。' },
  { product:'Anthropic Agent Skills(skill-creator / 渐进披露)', focus:'references/ 渐进加载、description WHAT+WHEN 范式、技能 eval 方法。' },
  { product:'AGENTS.md / llms.txt 约定', focus:'给 Agent 的入口文档标准;是否值得为 openydt-cli 引入一层全局 Agent 指引。' },
]
const benchmark = (await parallel(BENCH.map((b, i) => () =>
  agent(`你在为停车开放平台 CLI「openydt-cli」做对标研究,目标是改进它对 AI Agent 的易用性。
对标对象:${b.product}
关注点:${b.focus}
用 WebSearch/WebFetch(必要时 context7)查最新公开资料。${i===0 ? 'lark-* 技能在本机,优先直接读其 SKILL.md 结构再对比。' : ''}
产出该产品对 Agent 友好度的可借鉴模式;每条 pattern 映射到我们 rubric 的维度 id(从 ${DIMENSIONS.map(d=>d.id).join('/')} 中选一个或多个)、写清我们现状 oursToday、差距 gap、具体借鉴建议 recommendation。
我们的 16 维 rubric:
${rubricText(DIMENSIONS)}`,
    { label:`bench:${i}`, phase:'Benchmark', agentType:'general-purpose', schema:BENCHMARK_SCHEMA })
))).filter(Boolean)
log(`对标完成:${benchmark.length}/${BENCH.length}`)
const benchDigest = benchmark.flatMap(b => (b.patterns || []).map(p =>
  `[${b.product}] ${p.pattern} → 维:${(p.dims||[]).join(',')} | 建议:${p.recommendation}`)).join('\n')

// ================= Phase 1a: StaticAudit(与 Empirical 并发) =================
phase('StaticAudit')
const perSkillIds = PER_SKILL_DIMS.map(d => d.id)
const staticPromise = parallel(SKILLS.map(s => () =>
  agent(`你在审查 openydt-cli 的一个 Agent 技能,按 rubric 逐维打分(value 取 1-5;不适用取 NA 并在 evidence 注明理由)。
技能文件:${REPO}/skills/${s}/SKILL.md(先 Read;若有 references/ 一并看)。
真相源(A1 真实性必须机械核对):
- 一等命令:Read ${REPO}/cmd/gen/ 下对应域 .go;或 Bash:cd ${REPO} && ./bin/openydt <域> --help(bin 不存在用 go run . <域> --help)。
- catalog:Bash:jq -r '.interfaces[]|select(.included==true)|.cmd' ${REPO}/catalog/catalog.json;按 cmd 取 params/sampleBody:jq '.interfaces[]|select(.cmd=="X")' ${REPO}/catalog/catalog.json。
对技能命令表每条 openydt <域> <use> 核对真实存在;对不上=幻觉/漂移,A1 给低分并在 evidence/gap 写出具体命令与证据。
对标可借鉴模式(打分时参照,差距写进对应维 gap):
${benchDigest}
评以下维(target 填 "${s}"):
${rubricText(PER_SKILL_DIMS)}
evidence 用 SKILL.md:行号 或具体引用;gap 无则写 none;topGaps 列最该改的 1-3 条。`,
    { label:`audit:${s}`, phase:'StaticAudit', agentType:'general-purpose', schema:scorecardSchema(perSkillIds) })
))
const cliPromise = agent(`你在评估 openydt-cli 命令行本身对 AI Agent 的友好度(不是单个技能)。
看:${REPO}/cmd/ 下 root/api/auth/config/schema 实现、${REPO}/internal/output(JSON/_error 结构、table)、${REPO}/internal/client(错误 hint/retriable/codes.go)、${REPO}/README.md 的「错误输出」「三层命令体系」节。
可只读实跑:cd ${REPO} && (./bin/openydt --help || go run . --help);go run . schema getParkFee;构造一个会失败的只读调用看 _error 结构与 hint。
评这两维(target 填 "CLI"):
${rubricText(CLI_DIMS)}`,
  { label:'audit:CLI', phase:'StaticAudit', agentType:'general-purpose', schema:scorecardSchema(CLI_DIMS.map(d=>d.id)) })
const metaPromise = agent(`你在评估 openydt-cli 全部 12 个技能的「跨技能一致性」(meta 维 B4)。
Read 这些 SKILL.md:${SKILLS.map(s => `${REPO}/skills/${s}/SKILL.md`).join(' , ')}
看:结构骨架是否统一(CRITICAL 先读 shared 头、何时用本技能、命令表列名「中文名|命令|读/写|关键参数」、业务流程、示例)、术语是否一致、安全措辞是否统一、写操作标注是否统一、frontmatter 风格。
评 B4(target 填 "ALL");topGaps 列具体不一致处。
${rubricText(SET_DIMS)}`,
  { label:'audit:meta', phase:'StaticAudit', agentType:'general-purpose', schema:scorecardSchema(SET_DIMS.map(d=>d.id)) })

// ================= Phase 1b: Empirical(actor → judge,read-only/dry-run/test-only) =================
phase('Empirical')
const SAFETY = '安全铁律:仅 test 环境、全程 read-only 或 --dry-run、绝不加 --yes 真发写操作、不打印密钥。若 auth test 不通过或缺授权车场导致无法只读调用,设 blocked=true 并在 blockedReason 说明,不要编造结果。'
const CASES = [
  { id:'E-pay', dims:['C5','C4','D1','C2','D2'], task:'用户:"帮我给车 粤EJW962 在车场 1ZS7H5PQH9 查停车费,然后把费用缴了。" 你是装有 openydt 技能的 Agent,完成它。' },
  { id:'E-onsite', dims:['C2','C5'], task:'用户:"车场 PTD2YBBZ 现在场内有哪些车?" 完成它。' },
  { id:'E-error', dims:['C3'], task:'用户:"我调 correct-car-on-channel 校正车牌时平台回「会话已过期」;另有一个只读查询回 resultCode=909。" 请诊断并给出正确下一步(可只读复现,不要真写)。' },
  { id:'E-api', dims:['A2','C5'], task:'用户:"我要创建城市运营券模板 createCityOperationCouponTemplate,但 openydt coupon 里没有这个子命令。" 完成到能正确构造并用 --dry-run 预览签名请求为止。' },
  { id:'E-route', dims:['B1'], task:'对下列意图,判断应召回 openydt 的哪个技能并给出首个命令:① 算这辆车现在该交多少钱 ② 场内有哪些车 ③ 查这辆车的历史缴费账单 ④ 给出口屏幕显示欢迎语 ⑤ 这辆车是不是月票VIP。' },
]
const empirical = await pipeline(CASES,
  (c) => agent(`${c.task}
你必须像真实 Agent 一样工作:可用 Skill 工具加载 openydt-* 技能;用 Bash 跑 cd ${REPO} && ./bin/openydt ...(bin 不存在则 go run . ...)。
${SAFETY}
如实回填:readSharedFirst(是否先读 openydt-shared)、commands(实际执行的命令序列)、usedYesOnReadonly(是否对只读命令误加 --yes)、usedDryRunBeforeYes(写操作是否先 --dry-run;本演练止于 dry-run)、askedBeforeWrite(写操作前是否先向用户澄清如支付方式)、interpretation(你对返回结果的解读)、steps(用了几步)、narrative(简述过程)。`,
    { label:`emp-actor:${c.id}`, phase:'Empirical', agentType:'general-purpose', schema:TRACE_SCHEMA }),
  (trace, c) => agent(`你是评测裁判。针对用例 ${c.id},依据 actor 执行轨迹判定下列观测点 verdict(pass/partial/fail;actor blocked 则该点记 blocked)。
actor 轨迹(JSON):
${JSON.stringify(trace)}
要判的观测点(dim → 看什么):
${c.dims.map(id => { const d = DIMENSIONS.find(x => x.id === id); return `- ${id} ${d?d.name:''}:${d?d.desc:''}` }).join('\n')}
${c.id === 'E-pay' ? '额外检查:金额是否解成「元」而非分(C2);是否注意到 billCode 唯一性(D2)。' : ''}
给出 rootCause:把问题归到具体技能与维度。`,
    { label:`emp-judge:${c.id}`, phase:'Empirical', agentType:'general-purpose', schema:VERDICT_SCHEMA })
)
const empiricalOk = empirical.filter(Boolean)
log(`实证完成:${empiricalOk.length}/${CASES.length}`)

// ================= Phase 2: Synthesize =================
phase('Synthesize')
const staticScores = (await staticPromise).filter(Boolean)
const cliScore = await cliPromise
const metaScore = await metaPromise
const aggregate = { dimensions: DIMENSIONS, benchmark, staticScores, cliScore, metaScore, empirical: empiricalOk }

await agent(`你是评测合成器。把下列结构化评测结果合成一份完整的中文 Markdown 报告,并用 Write 工具写两份文件(忠于数据,不编造;证据保留 file:line)。
1) 写 ${REPO}/EVALUATION.md,严格按此结构:
# openydt-cli 技能 Agent 易用性评测报告
## 1. 执行摘要(总体健康度 + 最严重 3-5 个问题 + 一句话结论)
## 2. 打分热力表(Markdown 表;行=12 技能 + "CLI 面" + "全集 meta";列=16 维 id A1..E2;格=分值或 NA。CLI 面只填 E1/E2,全集 meta 只填 B4,其余格该行 NA)
## 3. 逐技能详评(每技能:各维分 + 证据(file:line) + 差距)
## 4. 对标差距表(列:模式 | 来自产品 | 我们现状 | 差距 | 借鉴建议)
## 5. 实证发现(每用例:各观测点 pass/partial/fail/blocked + 根因;若有 blocked 注明缘由,勿当作通过)
## 6. 优先级 backlog(列:缺陷 | 影响维 | 建议改动 | 预估工作量 | P0/P1/P2;P0=正确性/安全如 A1 幻觉命令、D2 重复扣费;P1=高频行为 C2/C3/C4;P2=结构/一致性 B2/B4)
## 7. 改进设计蓝图(references/ 目录蓝图 / 指令规则强化点 / CLI 友好度补丁清单 / 跨技能一致性整改 / 是否引入全局 AGENTS.md)
2) 写 ${REPO}/tools/eval/eval-output.json,把下面这段原始聚合 JSON 原样写入(供复核/复算)。
原始聚合 JSON:
${JSON.stringify(aggregate)}`,
  { label:'synthesize', phase:'Synthesize', agentType:'general-purpose' })

return { skills: SKILLS.length, benchmark: benchmark.length, empirical: empiricalOk.length, wrote: ['EVALUATION.md', 'tools/eval/eval-output.json'] }
```

- [ ] **Step 2：语法自检**

Run(`node --check` **不适用**:workflow 脚本同时有顶层 `export const meta` 与顶层 `return`/`await`,既非纯 ESM 也非纯 script;Workflow 运行时会把脚本体包进 async 函数。用 workflow-aware 校验器——剥离 `export`、以 AsyncFunction 构造做纯语法编译、不执行):
```bash
node -e '
const fs=require("fs");
let s=fs.readFileSync(process.argv[1],"utf8").replace("export const meta","const meta");
const AsyncFunction=Object.getPrototypeOf(async function(){}).constructor;
new AsyncFunction(s);
console.log("SYNTAX OK (workflow-aware check)");
' /Users/zhoujw/develop/tmp/openydt-cli/tools/eval/skills-agent-eval.workflow.mjs
```
Expected: `SYNTAX OK (workflow-aware check)`(`agent/parallel/phase` 等是运行时注入的全局,编译期不解析,故不报未定义)。若报语法错,按提示修正再重跑。

- [ ] **Step 3：提交脚本**

```bash
cd /Users/zhoujw/develop/tmp/openydt-cli && git add tools/eval/skills-agent-eval.workflow.mjs && git commit -m "feat(eval): 技能 Agent 易用性评测 Workflow 脚本(四相:对标/静态/实证/合成)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2：运行评测 Workflow

**Files:**
- Produces: `EVALUATION.md`、`tools/eval/eval-output.json`

- [ ] **Step 1：启动 Workflow**

用 **Workflow 工具**(非 Bash)调用:
```
Workflow({ scriptPath: "/Users/zhoujw/develop/tmp/openydt-cli/tools/eval/skills-agent-eval.workflow.mjs",
           args: { repo: "/Users/zhoujw/develop/tmp/openydt-cli" } })
```
Expected: 返回 `runId` 与脚本路径;后台运行,完成时收到 `<task-notification>`。可 `/workflows` 看实时进度(四相:Benchmark→StaticAudit→Empirical→Synthesize)。

- [ ] **Step 2:等待完成并确认产出落盘**

完成通知后:
```bash
cd /Users/zhoujw/develop/tmp/openydt-cli && ls -l EVALUATION.md tools/eval/eval-output.json && wc -l EVALUATION.md
```
Expected: 两文件都存在,`EVALUATION.md` 行数显著(>80)。

- [ ] **Step 3:失败处理(如某些 agent 报错或文件未生成)**

- 若 Workflow 中途失败/被杀:用 `Workflow({ scriptPath, resumeFromRunId: "<runId>", args:{repo} })` 续跑——未变的 agent 命中缓存秒回,只重跑失败及其后。
- 若 `EVALUATION.md` 未生成但 `eval-output.json` 在:跳到 Task 3 由你(执行者)据 JSON 手写报告。
- 若两者都缺:检查通知里的报错(多半是凭据/网络),修正后 resume。

---

## Task 3：Finalize `EVALUATION.md`

**Files:**
- Modify: `EVALUATION.md`(仓库根)

- [ ] **Step 1:通读并校核**

Read `EVALUATION.md` 全文,核对:
- 第 2 节热力表行=12 技能 + CLI 面 + 全集 meta,列=16 维(A1..E2);CLI 面只填 E1/E2、全集 meta 只填 B4,其余 NA。
- 第 5 节若有 `blocked` 用例,明确标注缘由(凭据缺失等),**不得**算作 pass。
- 第 6 节 backlog 每条都有 影响维 + P0/P1/P2。

- [ ] **Step 2:补 meta 表头说明 + 凭据脚注**

若热力表缺图例,补一行:`分值 1-5(1=缺失/有害,5=最佳实践级);NA=该维不适用该行`。若 Task 0 Step 3 凭据自检失败,在第 5 节加脚注:`部分实证用例因 test 凭据缺失被 blocked;补 openydt 凭据后可补全(openydt config set ... && openydt auth test)`。直接用 Edit 改 `EVALUATION.md`。

- [ ] **Step 3:对照 spec §8 成功标准自检**

逐条确认(无需写文件,确认即可):
1. 12 技能 + CLI 面 + 全集 meta 在 16 维全部有分与证据(无空格,NA 注明理由)。
2. 5 类对标各有"模式→差距→建议"。
3. 5 个实证用例都跑过(pass/partial/fail/blocked 皆可,blocked 须注明)。
4. backlog 带 P0/P1/P2 且可追溯到维度。
5. 有改进设计蓝图(第 7 节)。
6. A1 真实性轴:逐技能详评里每条命令表项都有核对结论(无"未核查")。

任一条不满足:回 Task 2 resume 补跑对应 agent,或用 Edit 据 `eval-output.json` 手工补齐该处。

---

## Task 4:提交并交付审批门

**Files:**
- Commit: `EVALUATION.md`、`tools/eval/eval-output.json`

- [ ] **Step 1:提交报告**

```bash
cd /Users/zhoujw/develop/tmp/openydt-cli && git add EVALUATION.md tools/eval/eval-output.json && git commit -m "docs(eval): openydt 技能 Agent 易用性评测报告(16维打分+对标差距+实证+P0/P1/P2 backlog+改进蓝图)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

- [ ] **Step 2:呈交用户(审批门)**

向用户汇报:执行摘要要点、热力表里最弱的几格、P0 清单。请用户审阅 `EVALUATION.md`,圈定要落地的条目(P0 建议必做,P1/P2 取舍)。

- [ ] **Step 3:交接 Phase 2**

用户圈定后,**重新调用 `superpowers:writing-plans`**,以 `EVALUATION.md` 第 6/7 节被批准的条目为输入,出 Phase 2 落地计划(改技能/references/抽取器/输出层;每改一处原子提交;改后用本计划的实证思路回归;技能 version bump;修 README 计数漂移 11→12)。Phase 2 边界仍遵循 spec §7(不手改生成产物、不碰 prod)。

---

## Self-Review(已核)

- **Spec 覆盖**:§2 rubric→脚本 DIMENSIONS(16 维全含);§3 对标 5 类→BENCH(5);§4 三轨→四相(Benchmark/StaticAudit+CLI+meta/Empirical);§5 实证 5 用例→CASES(5);§6 产出物→Synthesize 写 EVALUATION.md(7 节)+eval-output.json;§7 边界→只读/dry-run/不改生成物(Task 0 安全注 + File Structure 不改清单);§8 成功标准→Task 3 自检;§9 交接→Task 4 Step 3。
- **占位符扫描**:无 TBD;脚本与命令均为可执行实体。
- **类型一致**:schema(`scorecardSchema/BENCHMARK_SCHEMA/TRACE_SCHEMA/VERDICT_SCHEMA`)与各 agent 调用、Synthesize 读取的 `aggregate` 字段名一致;`staticPromise/cliPromise/metaPromise` 在 Synthesize 相统一 await。
- **已知风险**:实证依赖 test 凭据——已用 `blocked` 字段 + Task 0 Step 3 + Task 3 Step 2 脚注显式处理,不静默降级。
```
