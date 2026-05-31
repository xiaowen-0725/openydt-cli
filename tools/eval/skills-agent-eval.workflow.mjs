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
