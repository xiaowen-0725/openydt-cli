# openydt-cli 技能 · Agent 易用性评测与改进 — 设计 (spec)

- **日期**: 2026-05-31
- **状态**: 已批准设计,待 writing-plans 出实现计划
- **分支**: `feat/skills-agent-eval`
- **作者**: brainstorming 协作产出(用户 + Claude)

---

## 1. 背景与目标

`openydt-cli`(广东艾科智泊智慧停车开放平台 CLI,Go + Cobra)第一版已交付,附带一套 AI Agent Skills(当前 `skills/openydt-*/` 共 **12 个**:1 基座 `openydt-shared`、1 元技能 `openydt-skill-maker`、1 兜底 `openydt-api-explorer`、1 编排 SOP `openydt-flow-park-access`、8 个域技能 billing/coupon/data/device/list/monthticket/park/record)。技能已经过一轮 skill-creator 标准整改(提交 `91c735d`):清晰 frontmatter、"先读 shared"、意图路由、读/写命令表、`--dry-run→--yes` 示例、`[[跨链接]]`、触发去冲突表、flow 技能的路由 evals。

**问题不是"修坏技能",而是 layer-2 增强**:`references/` 渐进披露约定已在 `skill-maker` 中**规定但几乎未使用**(仅 flow 有子目录,且是 evals 不是 references);"指令式规则"散落在各技能而非成体系;尚无与同类优秀产品的横向对标;尤其缺少对**行为属性**(规则是否真被遵守、响应是否被正确解读、报错能否自愈)的实证度量。

**目标**:对 openydt-cli 的 Agent 侧易用性做一次**研究驱动的评测**,产出一份打分化、带优先级的 `EVALUATION.md`,并据此出**改进设计**;用户批准后再由 writing-plans 出实现计划落地。

### 受众与范围

- **第一受众 = AI Agent**;人类可读性作为副产物,不单独做人类入口文档(README/quickstart 不在本轮范围,除非某改动顺带利于人类)。
- **覆盖四个面**:① Skills 正文与结构;② `references/` 引用文档层;③ CLI 自身 Agent 友好度(`--help`/错误 hint/`schema`/JSON 输出契约);④ **行为有效性**(遵循度 / 结果解读 / 错误自愈 / 意图澄清 / 任务效率)。

### 终态与审批门

**评测先行 → 改进设计 → 用户批准 → 落地**。本 spec 描述评测方法与产出;评测本身**不落地任何改动**。所有对技能/文档/CLI 的实际修改,等 `EVALUATION.md` + 改进设计经用户批准后,由后续实现计划执行。

---

## 2. 评分 rubric(16 维 / 5 类)

每维打分 **1–5**(1=缺失/有害,3=合格,5=最佳实践级)或 **N/A**(该维不适用于该技能)。`粒度` 列说明该维按什么单位打分:**逐技能** / **CLI 级** / **全集级(meta)**。标 ★ 为本轮新增维。

| 维度 | 含义(打分看什么) | 粒度 |
|---|---|---|
| **A 内容正确性** | | |
| A1 ★ 真实性 / 与真相源漂移 | 技能里的命令/参数/枚举是否与 `catalog.json`(included:true)+ `cmd/gen/*.go` + `openydt <域> --help` **逐条一致**;有无幻觉命令、陈旧参数、过期计数(如 README 说"11 技能"实为 12) | 逐技能 |
| A2 ★ callable 盲区覆盖 | 域内 `direction=callable` 接口是否都有一等命令或"用 `openydt api` 兜底"的指路,无遗漏 | 逐技能 |
| A3 示例可直接跑 | 示例用文档化测试 parkCode(`1ZS7H5PQH9`/`PTD2YBBZ`)、当前/相对时间;不照抄 catalog 2016–2019 历史 sampleBody(否则复制即撞 904/911/空结果) | 逐技能 |
| **B 结构与可读性** | | |
| B1 召回准确 / 触发去冲突 | description 是否 WHAT+WHEN、约 100–150 字、不堆砌同义词;裸触发短语是否单 owner,与兄弟域无冲突 | 逐技能 |
| B2 结构精简 / 渐进披露 | SKILL.md 是否精简(≤500 行约定)、大块内容是否下沉 `references/` 并写明按需加载触发条件 | 逐技能 |
| B3 跨技能路由 | `[[links]]` 是否到位、边界是否清晰(域 vs 编排、读 vs 写归属) | 逐技能 |
| B4 ★ 跨技能一致性 | 结构骨架/术语/格式是否统一,使 Agent 跨技能**迁移学习**(命令表列名、安全提示措辞、CRITICAL 头是否一致) | 全集(meta) |
| **C 指令与 Agent 行为** | | |
| C1 规则清晰 / 可执行 | MUST/NEVER/决策树是否明确、可执行、"无法被忽略";硬约束是否醒目 | 逐技能 |
| C2 结果解读支撑 | 是否教 Agent 正确读响应:金额单位(元 vs 分)、外层 `status` vs `data.code`、关键字段含义、"0 条不等于无" | 逐技能 |
| C3 错误自愈闭环 | 是否给出 错误→诊断→下一步 的闭环;是否利用 CLI 的 hint/retriable 信号;常见失败有无速查 | 逐技能 |
| C4 ★ 意图澄清 / 确认前置 | 是否教 Agent 先消歧(parkCode/env/支付方式),写操作前先问、再 `--dry-run`、后 `--yes`,而非猜了就发 | 逐技能 + 实证 |
| C5 遵循度(实证) | Agent 实跑时**是否真的**先读 shared、加 `--yes`、按路由选域(由 §4 实证评测度量) | 逐技能 + 实证 |
| **D 安全与稳健** | | |
| D1 写操作安全 | `--yes` 守护、`--dry-run` 预览、prod 门;**并入** PII 红线(prod 不记车牌)、限速/批量礼仪(~4–5/s)、test/dev/prod 环境隔离正确性 | 逐技能 |
| D2 ★ 幂等 / 重试安全 | 写操作(尤其缴费 `billCode` 唯一性、补缴)是否教幂等键,使重试不重复扣费;与 C3 自愈耦合 | 逐技能 |
| **E CLI Agent 可供性** | | |
| E1 CLI 可供性 | `--help` 文案、`schema` 发现、JSON 输出契约(`_error` 结构等)、机器可读错误是否便于 Agent 自解析自纠 | CLI 级 |
| E2 ★ 可诊断 / 可观测 | `--verbose`/`--dry-run`/请求回显是否构成有效的调试支点 | CLI 级 |

> 未单列为独立轴的 PII / 限速 / 环境隔离,统一收编进 **D1** 一起评。

---

## 3. 对标产品集(5 类)

Phase0 先对标提炼可借鉴模式,逐条映射到上面 16 维,作为静态审查的打分参照。每类的**提炼目标**:

| # | 对标对象 | 提炼什么 |
|---|---|---|
| 1 | 飞书 lark CLI + 本机已装 lark-* 技能 | 同架构(Go+Cobra+`npx skills`)的技能结构、`references/` 用法、frontmatter、shared 基座模式、命令表风格、workflow 技能范式 —— 就在本机,可直接 diff |
| 2 | Stripe(Agent Toolkit + LLM 文档 + idempotency/确认) | 金融级写操作的幂等键、确认、可预览范式;`llms.txt` 式 LLM 友好文档结构 —— 对标 D1/D2 |
| 3 | MCP 服务端(GitHub / Supabase / Cloudflare) | 工具描述写法、**只读模式**、结构化错误、project-scoping、destructive/idempotent 提示 —— 对标 C2/C3/E1 |
| 4 | Anthropic Agent Skills 官方指引(skill-creator / 渐进式披露) | `references/` 渐进加载、description WHAT+WHEN 范式、技能 eval 方法 —— 核对 B1/B2 差距 |
| 5 | AGENTS.md / llms.txt 约定 | 新兴"给 Agent 的入口文档"标准,评估是否值得为 openydt-cli 引入一层全局 Agent 指引 |

研究允许用 WebSearch/WebFetch + context7 取最新文档;不可得时以本机 lark-* 技能与 Anthropic skill-creator 标准为主锚。

---

## 4. 评测方法(三轨并行 → 打分报告)

用 **Workflow 编排**(契合 ultracode,fan-out 后汇总):

```
Phase0 北极星
  对标 5 类产品 → 提炼可借鉴模式,逐条映射到 16 维 rubric  → BENCHMARK-NOTES(中间产物)

Phase1 三轨并行
  ├─ 轨1 静态审查   每技能 1 agent(12)+ 1 个 CLI 面 agent;按 16 维打分(1–5/N-A)+ 证据(文件:行)+ 差距;A1 用三方机械核对
  ├─ 轨2 对标提炼   每类 1 agent(5);产出"模式 → 我们现状 → 差距 → 借鉴建议"
  └─ 轨3 实证评测   每用例 1 agent;按 §5 跑真实任务,行为维打 pass/partial/fail,多次取方差

Phase2 汇总
  合并三轨 → EVALUATION.md(见 §6)+ 改进设计(backlog → 具体改动蓝图)
```

- **A1 真实性轴(机械核对)**:对每个技能命令表里的 `openydt <域> <use>`,核对其存在于 `cmd/gen/<域>.go` 或 `catalog.json` 的 `included:true` 接口;对 `openydt api <cmd>` 兜底,核对该 `cmd` 在 catalog 存在且 `direction=callable`。对不上即记缺陷(幻觉/漂移)。
- **静态审查参照对标**:轨1 打分时引用轨2 的模式,故 Phase0 在前;但轨1 与轨2 可并行启动,轨1 在 Phase2 汇总时再吸收轨2 结论做"差距"标注。
- **变异控制**:实证用例每个跑 3–5 次,报多数判定 + 方差标记(行为不稳定本身是缺陷信号)。

---

## 5. 实证 Agent 评测设计(拿真信号,零 prod 风险)

给装好 openydt 技能的 subagent 派真实任务,观测行为维。**安全铁律:全程 read-only + `--dry-run`,仅 test 环境,绝不真实写、绝不碰 prod。** 缴费等写操作只验证到 `--dry-run` 预览签名请求为止。

| 用例 | 任务 | 观测点(维) |
|---|---|---|
| E-pay | "给车 `粤EJW962` 在 `1ZS7H5PQH9` 查费,并演练缴费(到 dry-run 为止)" | 先读 shared?(C5) 缴费前先问支付方式?(C4) 写操作先 `--dry-run` 后才提 `--yes`?(D1) 金额解成元?(C2) 注意到 `billCode` 唯一性?(D2) |
| E-onsite | "查 `PTD2YBBZ` 现在有哪些在场车" | `enterTimeFrom/To` 时间范围必填、0 条不误判"无车"?(C2) 完成步数/turn 数(可选非评分观测,不计入 rubric) |
| E-error | 注入一次 `resultCode=909` 或"会话已过期" | 读 hint → 诊断 → 给出正确下一步自纠?还是空转/瞎试?(C3) |
| E-api | "调用一个未一等化的接口(如城市运营券模板)" | 用 `openydt api` + 查 catalog 取 sampleBody,而非臆造命令/参数?(A2) |
| E-route | 触发正例/反例(复用既有 32/32 思路,覆盖易撞域:查费/在场车/账单/屏显/电子券/特殊车辆类型) | 召回正确 owner 域、兄弟域不误触发?(B1) |

**实证评分**:每用例对每个观测点判 pass / partial / fail,跨多次运行取多数 + 标注方差;根因归到具体技能与维度,回填 rubric 的 C4/C5(及 A2/B1 的实证佐证)。

---

## 6. 产出物

### 6.1 `EVALUATION.md`(评测报告,落根目录 `EVALUATION.md`,对齐仓库 `TEST_REPORT.md` 约定)

结构:

1. **执行摘要** — 总体健康度、最严重 3–5 个问题、一句话结论。
2. **打分热力表** — 行 = 12 技能 + "CLI 面" + "全集 meta";列 = 16 维;格 = 1–5 / N/A。一眼看出弱项聚集处。
3. **逐技能详评** — 每技能:各维得分 + 证据(`文件:行`)+ 具体差距。
4. **对标差距表** — `模式(来自哪个产品)→ 我们现状 → 差距 → 借鉴建议`。
5. **实证发现** — 每用例:pass/partial/fail + 运行方差 + 根因(归到技能×维度)。
6. **优先级 backlog** — 每条:`缺陷 → 影响维 → 建议改动 → 预估工作量 → P0/P1/P2`。P0=正确性/安全(如 A1 幻觉命令、D2 重复扣费风险);P1=高频行为有效性(C2/C3/C4);P2=结构/一致性打磨(B2/B4)。

### 6.2 改进设计(并入本初始化产出,作为"批准后落地"依据)

把 backlog 映射成具体改动蓝图:

- **`references/` 目录蓝图** — 哪些技能要建 `references/`、放什么(字段字典/枚举表/错误码处置手册/长业务流程)、主体如何写按需加载指引。
- **指令规则强化点** — 哪些散落规则要升格为 MUST/NEVER、加决策树、加结果解读小节(单位/`status` vs `data.code`)、加错误自愈速查。
- **CLI 友好度补丁清单** — `--help`/`schema`/JSON 错误契约/`--verbose` 的可改进项(注意:命令本身是生成产物,改的是抽取器/codegen/输出层,不手改 `cmd/gen`)。
- **跨技能一致性整改** — 统一骨架/术语/措辞的清单。
- 是否引入一层全局 **AGENTS.md / Agent 指引**(由对标 5 决定)。

---

## 7. 边界 / 非目标(YAGNI)

- **不改生成产物**:`cmd/gen/*.go`、`catalog/catalog.json` 是生成物。若评测发现命令/参数问题,改的是抽取器(`tools/extractor/extract.mjs`)/codegen(`internal/gen`)/技能文档,**不手改生成物**。
- **不重写人类 README / quickstart**(本轮聚焦 Agent 面);仅当某改动顺带利于人类可读性才捎带。
- **实证评测全程 read-only / dry-run / test-only**,不碰 prod、不做真实写、不打印密钥。
- **评测阶段不落地任何改动** —— 所有实际修改等 `EVALUATION.md` + 改进设计经用户批准后,由 writing-plans 实现计划执行。
- 不新增领域技能(一等命令覆盖已满,见既有决策);workflow 型技能(经营日报/缴费对账)是否新增由评测结论决定,不在本 spec 预设。

---

## 8. 成功标准

本初始化(评测阶段)完成的判据:

1. 12 个技能 + CLI 面 + 全集 meta 在 16 维上**全部有打分与证据**(无空格,N/A 须注明理由)。
2. 5 类对标各产出"可借鉴模式 → 差距 → 建议"。
3. §5 的 5 个实证用例**全部跑过**,有 pass/partial/fail 判定与方差。
4. 产出**带优先级(P0/P1/P2)的 backlog**,每条可追溯到具体维度与证据。
5. 产出**改进设计蓝图**(references/ + 规则 + CLI + 一致性),可直接喂给 writing-plans。
6. A1 真实性轴**零未核查命令**(每条命令表项都过了三方机械核对)。

---

## 9. 后续(交接 writing-plans)

本 spec 经用户审阅批准后,调用 **writing-plans** 出实现计划。计划预期形态:

- **Plan Phase 1 — 跑评测**:执行 §4 三轨 Workflow → 产出根目录 `EVALUATION.md` + 改进设计。
- **审批门(用户)**:用户审阅 `EVALUATION.md` + 改进设计,圈定要落地的条目(P0 必做,P1/P2 由用户取舍)。
- **Plan Phase 2 — 落地批准项**:按改进设计改技能/references/抽取器/输出层;每改一处原子提交;改后用既有 subagent 评测思路回归验证行为维。
- **收尾**:技能 `version` 按改动 bump(skillsync 按版本下发,需 release 生效);更新受影响文档与 README 计数漂移(如 11→12)。
