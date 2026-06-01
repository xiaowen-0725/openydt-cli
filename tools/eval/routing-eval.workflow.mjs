export const meta = {
  name: 'openydt-routing-eval',
  description: 'openydt 技能路由触发评测:读 12 技能 evals → router subagent 分类 → 打分(非 nested claude -p)',
  phases: [
    { title: 'Read', detail: '12 reader 读各技能 description + evals' },
    { title: 'Route', detail: 'router subagent 批量分类 query → 预测召回技能' },
  ],
}

if (!args || !args.repo) throw new Error('需传 args.repo,例如 Workflow({scriptPath, args:{repo:"/abs/path/openydt-cli"}})')
const REPO = args.repo
const SKILLS = ['openydt-billing','openydt-record','openydt-park','openydt-device','openydt-monthticket','openydt-coupon','openydt-data','openydt-list','openydt-api-explorer','openydt-skill-maker','openydt-shared','openydt-flow-park-access']
const CANDS = [...SKILLS, 'none']

const READER_SCHEMA = { type:'object', additionalProperties:false, properties:{
  skill:{type:'string'}, desc:{type:'string'},
  evals:{ type:'array', items:{ type:'object', additionalProperties:false, properties:{
    query:{type:'string'}, expected:{type:'string'},
  }, required:['query','expected'] } },
}, required:['skill','desc','evals'] }

const ROUTER_SCHEMA = { type:'object', additionalProperties:false, properties:{
  predictions:{ type:'array', items:{ type:'object', additionalProperties:false, properties:{
    gid:{type:'integer'}, predicted:{type:'string'},
  }, required:['gid','predicted'] } },
}, required:['predictions'] }

function chunk(a, n) { const o = []; for (let i = 0; i < a.length; i += n) o.push(a.slice(i, i + n)); return o }

// ---- Phase 1: read descriptions + evals (12 parallel readers) ----
phase('Read')
const reads = (await parallel(SKILLS.map(s => () =>
  agent(`Read ${REPO}/skills/${s}/SKILL.md 的 frontmatter \`description\` 字段(整段),和 ${REPO}/skills/${s}/evals/routing-evals.json 的 evals 数组。
返回:skill="${s}";desc=该技能 frontmatter description 原文;evals=该 json 里每条的 {query, expected}(原样,别改)。`,
    { label:`read:${s}`, phase:'Read', agentType:'general-purpose', model:'sonnet', schema:READER_SCHEMA })
))).filter(Boolean)

const descFmt = reads.map(r => `- ${r.skill}: ${r.desc}`).join('\n')
// flatten + assign global ids; keep expected out of router input
let gid = 0
const all = []
for (const r of reads) for (const e of (r.evals || [])) { all.push({ gid: ++gid, query: e.query, expected: e.expected }) }
log(`读到 ${reads.length} 技能 / ${all.length} 条 query`)

// ---- Phase 2: route (batched router agents; query-only, expected hidden) ----
phase('Route')
const batches = chunk(all, 18)
const routed = (await parallel(batches.map((b, i) => () =>
  agent(`你是 openydt CLI 的「技能路由器」。下面是全部候选技能及其职责(description):
${descFmt}

规则:对每条用户诉求,判断**最该召回的那一个**技能;只能从这些 name 里选一个:${CANDS.join(' / ')}。若与本平台无关或无明确归属,选 none。只按职责判断,不要被措辞带偏。

待分类查询(按 [gid] 列出):
${b.map(e => `[${e.gid}] ${e.query}`).join('\n')}

对每个 gid 返回 {gid, predicted}(predicted 必须是上面候选之一)。`,
    { label:`route:${i + 1}/${batches.length}`, phase:'Route', agentType:'general-purpose', model:'sonnet', schema:ROUTER_SCHEMA })
))).filter(Boolean)

// ---- score (pure JS) ----
const expById = new Map(all.map(e => [e.gid, e.expected]))
const qById = new Map(all.map(e => [e.gid, e.query]))
const preds = new Map()
for (const r of routed) for (const p of (r.predictions || [])) preds.set(p.gid, p.predicted)

let correct = 0, scored = 0
const misroutes = []
const byExp = {}
for (const [id, exp] of expById) {
  const pred = preds.get(id)
  byExp[exp] = byExp[exp] || { total: 0, hit: 0 }
  byExp[exp].total++
  if (pred == null) { misroutes.push({ gid: id, query: qById.get(id), expected: exp, predicted: '(missing)' }); continue }
  scored++
  if (pred === exp) { correct++; byExp[exp].hit++ }
  else misroutes.push({ gid: id, query: qById.get(id), expected: exp, predicted: pred })
}
// 分母用 all.length(缺失预测计为未命中),与 ROUTING-BASELINE 口径一致。
const hitRate = all.length ? (correct / all.length) : 0
log(`命中 ${correct}/${all.length} (${(hitRate * 100).toFixed(1)}%);误路由 ${misroutes.length}`)

return {
  total: all.length, scored, correct,
  hitRate: Number((hitRate * 100).toFixed(1)),
  byExpected: byExp,
  misroutes,
}
