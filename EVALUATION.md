# openydt-cli 技能 Agent 易用性评测报告

> 评测对象:`skills/openydt-*`(12 个技能)+ CLI 面 + 全集 meta。
> 真相源:`catalog/catalog.json`(included:true)、`cmd/gen/*.go`、`./bin/openydt <域> --help`、`internal/client/codes.go`、`internal/output/output.go`、`internal/cmdutil/run.go`、各 `skills/openydt-*/SKILL.md` 及 `references/`、`evals/routing-evals.json`。
> 评测口径:16 维 rubric(A1..E2);静态审查 + 4 个对标产品(lark-cli / Stripe / MCP 服务端 / Anthropic Skills+AGENTS.md·llms.txt)+ 5 个端到端实证用例。

## 1. 执行摘要

**总体健康度:良好,偏强。** 内容正确性(A 类)与跨技能路由/召回(B1/B3)是明显强项——命令表与真相源逐条核对几乎零漂移(多数域 A1=5),12 技能 description 全是 WHAT+WHEN 三段式且兄弟域边界双向互指(实证 E-route 5/5 意图零误路);结果解读(C2)在 shared 的 `result-reading-sop.md` 与 billing/flow 域做到了"金额单位=元、status vs data.code、0条≠无"的高细度教学,超出 Anthropic 公开示例。5 个实证用例全部 pass(E-pay 的 C4 为 partial),无 fail、无 blocked。

**但最严重的问题集中在正确性安全与渐进披露两块:**

1. **【P0·A1/C1/D1】`openydt-api-explorer` 核心安全断言与代码相反(幻觉)。** SKILL.md:68-73,127 整段加粗 + 速查表第 3 行宣称"`openydt api` 不判断读写、不拦截写、漏 `--yes` 可能直接把写请求发出去";但 `api.go:34→run.go:51` RunCall 无条件调 `guardWrite`,实测 `openydt api createCityOperationCouponTemplate`(无 `--yes`)被拦"是写操作,需加 --yes 确认"。这给 Agent 灌输了错误的安全心智模型(既制造无谓恐慌,又误导其相信平台无任何兜底)。
2. **【P0·A1/C2】`openydt-coupon` 券种判别字段归因错误。** SKILL.md:122 把金额券/时间券区分写成"由 balanceType 区分",实际 `balanceType`=结算类型(0销售/1发放/2使用),真正判别券种的是 `couponType`(0免费/1金额扣减/2折扣/3固定/4时间券)。会误导 Agent 读响应与建模板。
3. **【P1·A1】`openydt-monthticket` 命令表必填参数漏列。** `month-ticket-config-edit` 漏顶层必填 `price`(catalog group=null),漏列会触发 status=7;且把组内条件必填与顶层必填混同。
4. **【P2·B4】跨技能一致性漂移。** 写操作"读/写"列出现 4 种写法(park 全角合规 / data 半角 / monthticket `写(--yes)` / billing+coupon+device+list+record 裸"写"+表下备注),违反 `skill-maker:129` 自定硬规约;必填标记与关键参数列两套体系 4:4 分裂(flag 式 `--park-code` vs 裸 `parkCode`)。
5. **【P1·C5】全集缺"行为 eval"。** 所有域只有触发/路由 eval(`routing-evals.json`),无 Anthropic EDD 式 `expected_behavior` 端到端断言;`skill-maker` 指引作者跑的 `run_loop.py` 路径未锚定、所指"[[openydt-shared]] 的评测约定"在 shared 全文不存在(实证断裂)。

**一句话结论:** openydt-cli 技能集在内容真实性、路由去冲突、结果解读三项已达到甚至超过对标产品水平,但需立刻修掉 api-explorer 的反向安全幻觉与 coupon 券种归因两处会误导 Agent 的正确性缺陷,并补齐渐进披露(per-cmd references)、跨技能格式一致性与行为 eval 三块结构性短板。

## 2. 打分热力表

> 行=12 技能 + "CLI 面" + "全集 meta";列=16 维(A1..E2)。CLI 面只填 E1/E2,全集 meta 只填 B4,其余该行 NA。`*` 标注:E-route/E-pay 等实证用例对 park/record/monthticket 的部分 A2/A3/B3 维分值取了与"示例未用文档化测试值"群组同列的偏保守口径,详评中已注明其实际为强项。

| 技能 / 面 | A1 | A2 | A3 | B1 | B2 | B3 | B4 | C1 | C2 | C3 | C4 | C5 | D1 | D2 | E1 | E2 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| openydt-shared | 4 | 4 | 5 | 5 | 5 | 5 | NA | 5 | 5 | 3 | 3 | 4 | 4 | 5 | NA | NA |
| openydt-skill-maker | 5 | 4 | 5 | 4 | 3 | 4 | NA | 4 | 4 | 4 | 3 | 2 | 4 | 4 | NA | NA |
| openydt-api-explorer | 2 | 4 | 5 | 4 | 4 | 4 | NA | 2 | 3 | 4 | 3 | 3 | 3 | 3 | NA | NA |
| openydt-flow-park-access | 5 | 4 | 2 | 4 | 4 | 4 | NA | 4 | 5 | 4 | 5 | 4 | 4 | 4 | NA | NA |
| openydt-billing | 5 | 5 | 5 | 4 | 3 | 5 | NA | 4 | 5 | 5 | 4 | 4 | 5 | 5 | NA | NA |
| openydt-coupon | 4 | 5 | 5 | 5 | 3 | 5 | NA | 4 | 3 | 3 | 4 | 4 | 4 | 5 | NA | NA |
| openydt-data | 5 | 5 | 5 | 5 | 4 | 5 | NA | 4 | 4 | 4 | 4 | 4 | 4 | NA | NA | NA |
| openydt-device | 5 | 4 | 5 | 5 | 4 | 5 | NA | 4 | 5 | 4 | 4 | 4 | 4 | 5 | NA | NA |
| openydt-list | 5 | 5 | 4 | 4 | 4 | 5 | NA | 4 | 3 | 5 | 4 | 3 | 4 | 4 | NA | NA |
| openydt-monthticket | 4 | 5 | 5 | 5 | 4 | 5 | NA | 4 | 4 | 3 | 4 | 4 | 4 | 5 | NA | NA |
| openydt-park | 5 | 5 | 5 | 4 | 3 | 5 | NA | 4 | 5 | 3 | 4 | 4 | 4 | NA | NA | NA |
| openydt-record | 5 | 5 | 5 | 4 | 4 | 5 | NA | 4 | 4 | 3 | 3 | 4 | 4 | 5 | NA | NA |
| **CLI 面** | NA | NA | NA | NA | NA | NA | NA | NA | NA | NA | NA | NA | NA | NA | 5 | 5 |
| **全集 meta** | NA | NA | NA | NA | NA | NA | 3 | NA | NA | NA | NA | NA | NA | NA | NA | NA |

> 注:上表分值以详评(第 3 节)的逐维 evidence/gap 为准。`tools/eval/eval-output.json` 为机器可读副本。NA 含义:维不适用于该行 grain(如 data/park 的 D2 因本域无资金写操作而 NA;CLI 面只评 E1/E2;meta 只评 B4)。

## 3. 逐技能详评

### 3.1 openydt-shared(基座)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 4 | status 1-9 与 resultCode 901-912/1801 逐条与 `internal/client/codes.go` 一致(status-codes.md:7-34);域清单 SKILL.md:112 经 `./bin/openydt <域> --help` 全部 OK;park get-auth-park-codes / trade pay-park-fee / payback-batch / set-points / set-prestore-for-c-park 均真实存在;无幻觉命令。 | 全局 flag 表 SKILL.md:69-77 已与二进制漂移——`openydt --help` 多出已发布的全局 `--read-only`(也认 `OPENYDT_READ_ONLY=1`)与 `config set-default` 子命令,shared 全文无任何提及(grep 命中 NONE)。 |
| A2 | 4 | 三层命令模型 SKILL.md:108-121:域一等命令→`openydt api <cmd>`(line114)→schema 探索;作为基座不逐 cmd 列举,callable 盲区交由 api 兜底合理。 | 未指向 `openydt schema <cmd>` 作为"调 api 前先查 params/readwrite"的强制步;`--read-only` 会话级能力裁剪未纳入命令模型。 |
| A3 | 5 | SKILL.md:23 硬约束 "MUST 用文档化测试 parkCode(1ZS7H5PQH9/PTD2YBBZ)+当前/相对时间" 并给 why;示例 SKILL.md:118/159-160 用文档化测试值。 | none |
| B1 | 5 | description(SKILL.md:4)WHAT+WHEN 且声明单 owner;evals/routing-evals.json 1-14 全归 shared、15-19 正确分流。 | none |
| B2 | 5 | SKILL.md 仅 161 行;status-codes/result-reading-sop/write-idempotency/park-notes 四大块下沉 references/(各 22-44 行,<100 行无需 Contents);一行指针按需加载。 | none |
| B3 | 5 | SKILL.md:113 一行给齐全域 [[links]];读vs写边界(line138)、域vs编排边界清晰;sibling billing:13 反向断言先读 shared。 | none |
| C1 | 5 | SKILL.md:17-28 "⚠️ Agent 硬约束(MUST/NEVER·先读)"独立醒目区,每条附 why(对齐 Anthropic 规则+理由)。 | 未含 named anti-pattern "测试 key NEVER --sign v3→用 v2" 于硬约束区(仅 line104 散文);未列 ✅/⚠️/🚫 三级清单。 |
| C2 | 5 | references/result-reading-sop.md 教三层判读、金额一律元(时间券分钟例外)、0条≠不存在、分页非全量,有 Final Answer Check 五问。 | none |
| C3 | 3 | status-codes.md 列全状态/业务码/退出码;codes.go StatusHint/ResultHint 给 nextCommands;result-reading-sop.md 把 status≠1 引向各域错误表。 | shared 自身未内联"错误→诊断→下一步"速查表,只指向 codes.go;hint 的 retriable/nextCommands 未在 shared 文档面暴露,Agent 仅读技能拿不到可执行修复指引。 |
| C4 | 3 | 写操作 dry-run→yes、切 prod 前确认已前置;park-notes 回忆要求先确认环境。 | 缺 parkCode 缺失/支付方式歧义的"先问再做"决策树;无三级 ask-first 分层;未教"意图不确定时默认 --read-only 探查"。 |
| C5 | 4 | sibling 域技能硬性复述先读 shared;SKILL.md:13/21 双向声明形成遵循闭环;evals 提供触发实证锚。 | 无行为 eval(端到端断言),仅触发-召回 eval;遵循度靠约定而非实证度量。 |
| D1 | 4 | 写操作 MUST --yes、先 dry-run、prod 写前确认、NEVER 打印 key、prod 不记 PII;限速 300次/分+重试退避;test/dev/prod 物理隔离。 | 已发布的会话级硬写过滤 `--read-only` 未写入安全规则段(一个一等写安全控制点完全缺席);批量节流给数值未给姿势(如 ~4/s)。 |
| D2 | 5 | references/write-idempotency.md MUST 复用首次键/NEVER 换新键/907=幂等命中;各命令幂等键速查表逐条与 catalog 核对一致;无显式键的写给"重试前先读确认"兜底。 | none |

**差距小结:** 最该补 `--read-only`(全局 flag 表 + 安全规则段 + C4 探查姿势)、C4 三级 ask-first 清单、C3 把 billing 的错误自愈速查表升格为 shared 模板。

### 3.2 openydt-skill-maker(元技能)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 5 | SKILL.md:70-76 触发裁决表与真实 sibling description 逐条一致;SKILL.md:220-224 管道图命令全部真实存在;catalog included 计数=143 与 README 一致。 | none |
| A2 | 4 | SKILL.md:103-108 命令必须真实存在;SKILL.md:121 step5 要求扫 included:false 且 direction:callable 盲区加 api 指路。 | 只在散文里讲 api 兜底,未给可照抄的"callable盲区指路"模板句。 |
| A3 | 5 | SKILL.md:89 硬约束用 1ZS7H5PQH9/PTD2YBBZ+当前/相对时间、不照抄历史 sampleBody;SKILL.md:196 写示例带 --yes。 | none |
| B1 | 4 | SKILL.md:4 description 174 字符 WHAT+WHEN;routing-evals.json 19 条含 12 正例+7 兄弟域反例。 | 未对元技能(易 under-trigger)加 catch-all 触发语;evals 已造但无证据跑过。 |
| B2 | 3 | SKILL.md:96-100 完整阐述 references/ 按需加载约定;自身 224 行在预算内。 | 自身无 references/ 目录,统一渲染规约/Checklist/最小模板/workflow模板全内联;未把 lark"每命令一 reference+统一骨架"落成可复用模板,也未给"一级深""\>100行加目录"硬规则。 |
| B3 | 4 | SKILL.md:13 CRITICAL 先 Read shared;:132 要求跨技能引用用 [[openydt-<域>]];:200-215 清晰划界原子/域技能 vs workflow。 | 自身正文混用相对路径 prose 与它要求的 [[wiki-link]]规约,示范不自洽。 |
| C1 | 4 | SKILL.md:94 "why 约定"每条 MUST/NEVER/CRITICAL 配一句 why;:136-145 Pre-ship Checklist 逐项可勾。 | 未引入 Stripe 式 named anti-patterns 小节;未给三层边界分级。 |
| C2 | 4 | SKILL.md:92 要求新技能必含"结果解读要点"块并指向 result-reading-sop.md。 | 未把 lark"查询执行契约/has_more 非全量禁给全量结论"升格为制作规约硬条目。 |
| C3 | 4 | SKILL.md:90 要求每域必含"错误自愈速查表"三列,对标 billing。 | 未要求被造技能错误表逐条挂 nextCommands+retriable+0条/满页 hint。 |
| C4 | 3 | SKILL.md:113 写命令示例必带 --yes、建议先 --dry-run;:193-197 两步序列。 | 缺"意图澄清前置"制作要求;无 ⚠️Ask-first 层;未教 dry-run 输出向人复述 CONFIRM;未提 --read-only/--scope-park。 |
| C5 | 2 | SKILL.md:59/122/145 反复要求用 skill-creator 触发 eval、跑 run_loop.py;evals/routing-evals.json 确实存在(19 条)。 | **实证断裂:** grep shared 全文无任何 eval/评测约定;run_loop.py 路径未锚定;routing-evals.json 无 expected_behavior 行为断言、无证据被跑、无基线;MEMORY 记嵌套会话跑不出信号;EDD"先无技能量基线"未写进步骤。 |
| D1 | 4 | SKILL.md:110-113 写操作一律标 --yes、读/写列标"写(需 --yes)";:214 批量写注意限速。 | 未在制作规约复述"prod 不记车牌/批量验证仅 test"为必含项;未要求引入全局 --read-only 或结构化拦截。 |
| D2 | 4 | SKILL.md:91 "写操作幂等"节:要求点名幂等键、复述"重试复用首次键、907=幂等命中"、指向 write-idempotency.md。 | 未吸收 Stripe"改任一参数必须 mint 新 key"的 mismatch 守护;未教 thirdBillCode 逐条唯一性在批量 plan-validate 阶段查重。 |

**差距小结:** C5 实证断裂(指引作者跑的 eval 路径与约定不存在)最严重;其次 B2 自身未实践渐进披露 + 缺 per-cmd reference 模板;C4/C1 安全规约欠分级与具体禁项。

### 3.3 openydt-api-explorer(兜底)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 2 | SKILL.md:68-73,127 断言"openydt api 不判断读写、不拦截写、漏 --yes 可能直发"。机械核对:`api.go:34` 调 `f.RunCall`;`run.go:51` RunCall 无条件调 `f.guardWrite(cmd)`;`catalog.go:82` IsWrite 按 readwrite=write 要求 --yes。实测 `openydt api createCityOperationCouponTemplate`(无 --yes)被拦。其余命令表/示例 cmd 逐条核对 catalog 正确。 | **最醒目安全断言(整段加粗+速查表第3行)与真相源相反:** api 走 RunCall 同享 guardWrite,写 cmd 漏 --yes 会被拦而非直发。该幻觉教给 Agent 错误安全心智模型。须改为"api 与一等命令同享写守护;并存在 --read-only 硬过滤"。 |
| A2 | 4 | SKILL.md:17-32,110-117 系统覆盖 included=false 且 callable 兜底面,点名 cityOperationCoupon/thirdParkForBolian 等并给 jq 全量枚举法;webhook 单列说明改自建接收端。 | 存在更优 in-CLI 入口 `openydt schema` 但全文只导向 jq,未提 schema 命令(对标 lark"调原生前先 schema")。 |
| A3 | 5 | 示例均用 1ZS7H5PQH9/PTD2YBBZ,validFrom/To 用未来日期(2026-06-01~2027-06-01),显式警告照抄历史 sampleBody 会撞无效车场;实测 dry-run 输出 ts=20260601 正常。 | none |
| B1 | 4 | SKILL.md:4 description WHAT+WHEN,已是显式 catch-all 兜底 owner,与各域边界清晰互补。 | description 约 180 字偏长;触发清单与正文 §15 略重复。 |
| B2 | 4 | SKILL.md 共 165 行,信息密度高无冗余;无 references/ 目录。 | 未把"api 兜底单命令模板/schema 优先"下沉为 reference;可选补 references/api-fallback.md(首行先读 shared)。 |
| B3 | 4 | SKILL.md:13 顶部 CRITICAL 直链 shared;明确"先找一等命令再 api 兜底";回链 shared 的 write-idempotency.md;说明 monthticket/record 会指向本技能。 | 无 references/ 故无逐文档复述;到 schema 命令的路由缺失。 |
| C1 | 2 | SKILL.md:68-73 用加粗+⚠️给出"写操作必须 --yes"规则并带 why。 | 规则方向对,但事实前提(api 不拦截写)是错的,可执行规则建立在错误机制上,削弱可信度。应改写为"api 与一等命令同享写守护;漏 --yes 会被 CLI 拦下;需要绝对禁写用 --read-only"。 |
| C2 | 3 | SKILL.md:55 教写错字段名返回 909;:99 sampleResponse 帮预判;速查表区分 status 含义。 | 自身未教金额单位/status vs data.code/0条≠无;缺一句"返回字段含义见对应域,金额按元"。 |
| C3 | 4 | SKILL.md:121-129 错误自愈速查表覆盖 status=9/909/status=7→给出 jq 核对、按 params 逐项核对的恢复动作;回指 shared。 | **速查表第3行"写 cmd 漏 --yes 却真的发出去了"描述一个不存在的现象(实测会被拦),需删改;** 缺 0条/满页截断类 hint。 |
| C4 | 3 | SKILL.md:34 选择顺序消歧;:59,73 写操作先 --dry-run 再 --yes。 | 缺 parkCode/env/支付方式入参消歧前置;未引入 --read-only"不确定意图默认只读探查";写前复述未明确。 |
| C5 | 3 | SKILL.md:13 CRITICAL 强制先 Read shared;evals/routing-evals.json 20 条触发/去冲突 eval。 | 现有 eval 只测路由召回,无行为 eval;因 A1 机制描述错误,即便严格遵循文档也被灌输错误安全认知。 |
| D1 | 3 | SKILL.md:59,73,78-82 写操作教 --dry-run+--yes;点名 prod 尤其危险;默认 test。 | 核心写守护机制描述与代码相反(A1);完全未提 --read-only/OPENYDT_READ_ONLY(实测对 api 写 cmd 即使带 --yes 也拒绝);批量节流/prod 不记 PII 未提。 |
| D2 | 3 | SKILL.md:74,129 把写操作幂等/重试指向 shared 的 write-idempotency.md。 | 直发任意写 cmd(grantCityOperationCoupon/asynSuccess 回执)却未就近点明幂等键字段名;缺"写 cmd 重试前确认幂等键,变更任一参数须换新 key"。 |

**差距小结(本技能为全集最紧急修复对象):** A1/C1/C3/D1 集中受 SKILL.md:68-73,127 的反向安全幻觉拖累——须改写为"api 与一等命令同享写守护(实测漏 --yes 被拦)"并删除速查表第3行不存在的现象;同时补 `--read-only` 硬过滤宣传与 `openydt schema <cmd>` 调参指路。

### 3.4 openydt-flow-park-access(编排)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 5 | SKILL.md:42-75 全部命令逐条核对真实存在;字段 shouldPayValue/parkingCode/chargeDate/otherAttr.chargeBillNumber 经 getParkFee sampleResponse 证实;pay 步骤 otherAtrr 精确匹配平台真实拼写(非 otherAttr)。无幻觉/无过期计数。 | 轻微:pay-park-fee 的 payDate 为必填,SKILL.md:73 缴费字段清单未列——但完整参数表下沉到 [[openydt-billing]],属可接受的编排层简化。 |
| A2 | 4 | 编排层(SKILL.md:13,88-94),触及的进出场 callable 命令均给一等命令,明确把单命令参数归属回指域技能;出场链路涵盖盘点兜底。 | scan-channel-code-in-out/roadside-car-check-in/supplement-parking-record-image/correcting-car-code-after-...-confirm-phone 未编排也未指路;枚举附录仅指向在线 /Api/appendixData,未指 api-explorer。 |
| A3 | 2 | 占位示例存在(故非 NA);SKILL.md:42-75 均为占位抽象(parkCode/carCode/channelCode/newCarNo…)。 | **全文 grep 无任何文档化测试 parkCode 或当前时间,无一条 flag-complete 可跑串。** 至少应在进场(b)与出场主路各给一条用 PTD2YBBZ/1ZS7H5PQH9+当前时间的端到端可跑串,并示范一次 -o table/json(对比 billing:103-128 已做)。 |
| B1 | 4 | description WHAT+WHEN 俱全,末句显式去冲突"只查单条/在场车/锁车请用对应域";routing-evals 18 例覆盖正反向。 | description ~439 字符远超目标;正文未设独立"不应使用本技能"反向小节。 |
| B2 | 4 | SKILL.md 仅 94 行;开篇声明"不复述各命令参数表",参数/出参/枚举下沉到域技能;结构清晰。 | 无自有 references/;缺"枚举码→中文映射表"与 ASCII 数据流图。 |
| B3 | 4 | [[links]] 全部解析到真实技能目录;命令归属表逐命令标 owner 域;首行 CRITICAL 强制先 Read shared。 | 缺每步"从上一步响应取哪个字段当本步入参"的显式数据流(进场链路未画),以及 ASCII 进/出场流程图。 |
| C1 | 4 | CRITICAL 先读 shared;写操作"仅 test、先 --dry-run 后 --yes";"校正成功≠进场,务必复核在场"带 why;跨命令硬约束速查表每行现象→含义→处理。 | 未采用三级标签;缺 Stripe 式具名反模式;写操作未用低自由度措辞固定顺序。 |
| C2 | 5 | 金额强约束"shouldPayValue 单位:元,1 即 1.00 元不是 1 分";教从查费响应取下游字段;在场复核教读 get-park-on-site-car 判断(0条≠车场无此车而是可能未进场);字段名经核对准确。 | none |
| C3 | 4 | 错误自愈成表:channel-snap 908→换通道;会话已过期→先 snap 再校正;查费令牌超时→重新查费;每条现象→诊断→下一步完整。 | 未把通用业务码/退出码 hint/retriable 显式回指 shared;缴费 907 幂等命中场景未进本技能速查表。 |
| C4 | 5 | 进场开篇先判断补录(强制)还是抓拍(模拟)消除路径歧义;缴费步骤强制"先询问是否需要缴费?用什么支付方式?——缴费是真实写操作不要默默执行";全程写操作先 --dry-run 后 --yes。 | none |
| C5 | 4 | 首行 CRITICAL MUST 先 Read shared、每个写命令带 --yes、安全前提点明先 --dry-run、路由按域归属明确;routing-evals 18 例。 | 仅触发 eval,无行为 eval;遵循度未经实证度量。 |
| D1 | 4 | 开篇明确进出场写操作"仅在 test 环境演练",每写命令 --dry-run 后 --yes;prod 门/PII/限速下沉到 shared,首行强制先读。 | 正文未自带 prod 隔离/PII/限速一句话提醒;批量盘点离场未提节流;未提示"不确定意图先 --read-only/--dry-run 探查"。 |
| D2 | 4 | 缴费写"唯一 billCode",完整幂等机制显式下沉到 [[openydt-billing]]。 | 就地未教"重试复用首次 billCode、绝不新生成、907 视为成功";盘点离场批量场景无幂等键/重复盘点防护提示。 |

**差距小结:** A3 可跑性最该补(全文无 flag-complete 可跑串);A2 同链路 callable 变体未指路;B2/B3 缺枚举码→中文映射表与 ASCII 数据流图。

### 3.5 openydt-billing(trade 域)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 5 | 命令表7条与 `trade --help`+`cmd/gen/trade.go` Use 名逐条一致;flags 与 trade.go:108-145 BuildBody/Flags 吻合;catalog included=true 的 trade cmd 恰为这7个;无过期数字。 | none |
| A2 | 5 | catalog trade callable 共15条:included=true 的7条全有一等命令;included=false 的8条 SKILL.md:44 明列并指向 api 兜底+[[api-explorer]];webhook reducePrestore/refundParkFee 正确未列。 | none |
| A3 | 5 | SKILL.md:108/116 用 1ZS7H5PQH9+未来时间;显式警告照抄历史 sampleBody;缴费示例用占位符强制取自上一步响应。 | get-park-fee 示例带真实车牌 粤EJW962(test 可接受),可加"prod 用脱敏车牌"提示。轻微。 |
| B1 | 4 | description WHAT+WHEN 约130字,末句去冲突"历史账单用 parking 域";evals 13/14 锚定边界。 | 无"不应使用本技能"反向小节;billing↔record↔coupon 三角归属未在三方文件互指闭环。 |
| B2 | 3 | SKILL.md 仅130行;evals/ 已建。 | 无 references/;pay-park-fee/payback-batch/set-points/set-prestore 4个高参写命令字段全内联,无"必读 reference"列。 |
| B3 | 5 | SKILL.md:13 直链 shared;[[api-explorer]]/[[shared]] 指到 references;读/写归属清晰,28行明确在场车/订单/进车补录属 parking。 | none |
| C1 | 4 | CRITICAL MUST 先读 shared;所有写命令"必须加 --yes";"务必把前序响应字段作后续入参";金额单位规则带 why(别把1当1分付0.01)。 | 未列 named anti-patterns;写操作未给低自由度"严格5步不增删 flag"措辞。 |
| C2 | 5 | 金额单位=元逐字段点名;shouldPayValue=actPayCharge+couponValue 等式;教从 data.otherAttr.chargeBillNumber 等取值(字段经 catalog 核对存在);10分钟时效解读。 | 0条≠无 仅指向 shared 未在本域内联一句。轻微。 |
| C3 | 5 | SKILL.md:90-99 "错误自愈速查"表:912→重查、907→幂等命中改 get-pay-bill、连接超时→同 billCode 重发、909→schema 核必填;每行现象→含义→恢复动作含可执行命令;被 MEMORY 标为可推广模板。 | none |
| C4 | 4 | 写操作先 --dry-run 再 --yes;缴费示例默认 --dry-run 起手;parkCode 缺失分流说明。 | 无三级分级;支付方式/env=prod/批量写"先问再做"未做成一等可扫规则;--dry-run 输出未要求渲染 CONFIRM 复述行。 |
| C5 | 4 | 入口 MUST 先 Read shared;写命令在 trade.go:104/159/183 经 `f.ConfirmWrite()` 硬拦截;evals 16条触发/路由断言。 | evals 仅触发/路由,缺行为 eval(expected_behavior)。 |
| D1 | 5 | 写命令必须 --yes(代码 ConfirmWrite);先 dry-run 后 yes;示例"仅 test"标注;全局 --read-only 守护已存在;env 隔离;批量节流由 shared/client 重试退避承载。 | 示例车牌 粤EJW962 为真实 PII(test 可接受);未在本域复述"prod 不记 PII 车牌"。轻微。 |
| D2 | 5 | "缴费幂等:billCode 全局唯一,重试必须复用首次 billCode(绝不新生成);907=幂等命中改 get-pay-bill;payback-batch 每条 thirdBillCode 同理";与代码注解(trade.go:101/156/180/225)一致。 | none |

**差距小结:** B2 渐进披露落点缺失(4个高参写命令无 per-cmd reference);C4 三层化 + CONFIRM 复述行;C5 行为 eval。本域 A/D 类近满分,是范例域之一。

### 3.6 openydt-coupon(券与商家域)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 4 | SKILL.md:34-65 命令表 30 条与 `coupon --help`(30)逐条比对零幻觉零遗漏;参数/枚举与 catalog/help verbatim 一致;`--total-count ≤1万` 正确。 | **line122 把金额券/时间券区分写成"由 balanceType 区分",实属错误:** balanceType=结算类型,真正区分券种的是 couponType(0免费/1金额扣减/2折扣/3固定/4时间券);全文未在散文点明 couponType。 |
| A2 | 5 | coupon 域 30 个 included callable 全有一等命令;唯一 excluded couponFlow 及跨域 thirdCoupon*/cityOperationCoupon* 在 SKILL.md:69 显式点名并 [[api-explorer]] 兜底。 | none |
| A3 | 5 | create-trader 用 1ZS7H5PQH9、query 用 PTD2YBBZ;销售时间 2026-06-01;明确警告照抄历史 sampleBody;实跑 delete-trader --dry-run 成功(test/v2)。 | none |
| B1 | 5 | description WHAT+WHEN+去冲突"用券抵扣后实际查费/缴费在 trade 域"约 110 字;evals 20 条显式区分 coupon vs billing/park/record。 | none |
| B2 | 3 | SKILL.md 共 165 行;无 references/;create-coupon-template 13必填、create-coupon 嵌套 couponTemplate 全内联。 | 写命令参数复杂,宜下沉 references/openydt-coupon-<cmd>.md(统一骨架+首行先读 shared),命令表加必读 reference 列。 |
| B3 | 5 | 边界声明(查费/缴费→trade,在场车→parking);跨域券接口→api-explorer;幂等/0条解读→shared;链接目标均存在。 | none |
| C1 | 4 | CRITICAL 先读 shared;写命令清单+"必须加 --yes 否则被拦截";不可逆操作先 --dry-run;实跑 delete-trader 无 --yes 被拦截。 | 硬约束以 MUST 为主缺"规则+why"散文;未设三层;可补具名禁项"时间券 faceValue 单位是分钟非元"。 |
| C2 | 3 | 金额量纲(faceValue 金额券=元/时间券=分钟)、"查券0条≠该车无券"、教从前序响应取 traderCode/sellBillId/couponSn。 | **line122 关键归因错误(value/time 区分写成 balanceType,实为 couponType)会误导 Agent 读响应判别券种;** 未教 status vs data.code 区分。 |
| C3 | 3 | "错误自愈速查(券域)"两行(发券失败/status=2;售券重发疑似重复→复用 transationNum);通用码回指 shared。 | 速查表仅 2 行覆盖窄;未给 hint/retriable 字段化提示,未覆盖 status=7/6/0条与满页分页。 |
| C4 | 4 | 意图路由先按动词分流并标读/写;写操作(尤其 delete-trader/cancel-coupon 不可逆)先 --dry-run 后 --yes。 | 未在写前显式要求消歧 parkCode/env;缺 Stripe 式 CONFIRM 一行复述。 |
| C5 | 4 | 强制先 Read shared;所有写命令统一带 --yes,示例先 --dry-run;实证:delete-trader 无 --yes 被拦、--dry-run 走 test v2。 | 无行为 eval。 |
| D1 | 4 | 全部 14 个写命令清单+--yes 守护;不可逆操作先 --dry-run;coupon --help 暴露全局 --read-only;test 隔离。 | 未提批量发券/售券限速节流;create-trader 示例含 --password 123456 明文,未提醒 prod 勿在命令行留密码/不记 PII。 |
| D2 | 5 | 幂等键 sell-coupon.transationNum、create-fixed-coupon.uniqNo"重试复用同值绝不新生成";错误表"售券重发后疑似重复→复用首次 transationNum";CLI help 自带注解 idempotent。 | none |

**差距小结:** A1/C2 券种归因错误(balanceType→应为 couponType)最该修;B2 缺 per-cmd reference;C3 错误自愈仅 2 行 + C5 无行为 eval。

### 3.7 openydt-data(数据分析域)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 5 | 命令表 9 条与 `data --help`(9)及 catalog 逐条一致;读/写分类(6读+3写)与 catalog rw 完全吻合;参数枚举校对无误(dimension/minuteInterval 10-240/vipType/hourArea/pageSize≤1000)。 | none |
| A2 | 5 | data 域 9 个接口全部 callable,命令表逐条给一等命令,无遗漏;get-car-traffic-flow-analysis 仅 --body 通道已明确教法。 | none |
| A3 | 5 | 显式警示照抄历史 sampleBody;示例用 1ZS7H5PQH9/PTD2YBBZ 与 2026-06;dry-run→yes 安全环完整。 | 示例#1 get-park-bill --dimension 2(年)却给同一天 start/end,演示连贯性瑕疵,不影响可跑性。 |
| B1 | 5 | description WHAT+WHEN 齐备,末句去冲突"单车明细见 record";routing-evals 18 例覆盖正向+兄弟域。 | description 偏长(>150字),无独立"不应使用本技能"反向小节。 |
| B2 | 4 | SKILL.md 仅 96 行组织清晰;无 references/。 | 3 个 write 统计命令参数细节未下沉 references,命令表无"必读 reference"列;若后续参数增多易膨胀。 |
| B3 | 5 | 意图路由段全域 [[links]] 到位;读vs写、域vs域边界双处声明清晰。 | none |
| C1 | 4 | CRITICAL 先读 shared;3 写命令"必须带 --yes"(实测拦截);提醒两种时间格式坑。 | 未采用三层或具名反模式;写操作段未用低自由度措辞。 |
| C2 | 4 | 金额单位"元";"test 多数统计接口返回 nodata 属正常≠无数据"(0条≠无);指向 result-reading-sop.md。 | 未教 status vs data.code;聚合统计返回字段无字段级解读。 |
| C3 | 4 | "错误自愈速查(统计)"表:间隔>1天/格式错→缩窗改格式;minuteInterval 非10/240→改;nodata→换时间窗;通用码回指 shared。 | 未给可执行 nextCommands;高频业务码未在本域复述;分页可信边界未在 get-bill-summary/get-park-bill 回指。 |
| C4 | 4 | 业务流程第1步"先确定目标车场 parkCode";写命令示例强制先 --dry-run 再 --yes。 | 未要求向用户复述"将对 {parkCode} 在 {时间窗} 跑写统计"再确认;env 未作澄清前置项。注:3个 write 是统计接口非真扣费,风险低。 |
| C5 | 4 | 强制 Read shared;routing-evals 18 例;写命令示例演示 dry-run→yes;实测写守护拦截。 | evals 仅路由,无行为 eval。 |
| D1 | 4 | 3 write 命令挂 --yes 守护(实测拦截);示例先 --dry-run;全局 --read-only/--env;本域纯统计无 PII 写入风险;env 隔离。 | prod 门禁/prod 不记车牌未在本域复述;未说明"统计 write 接口其实只读不改数据"。 |
| D2 | NA | 本域 9 接口均统计/报表读取性质:6 read,3 虽标 write 但语义是聚合统计查询,不涉金额扣费、无幂等键、重复调用无资金副作用。 | 不适用;若严格按 MCP idempotentHint,可在 SKILL 点明"重复调用安全"以消解顾虑,但非 D2 本意。 |

**差距小结:** B2 为 3 个 write 统计命令补 reference;C2/C3 补字段级解读 + 可执行 nextCommands + 分页契约回指;C5/C4 行为 eval + 点明 write 实为只读幂等统计。

### 3.8 openydt-device(设备域)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 5 | 命令表 11 条全部与 `device --help` 逐条一致;catalog included=true 的 11 个全覆盖无幻觉;关键参数与 catalog params 完全吻合(opGate/changeChannelMode/cloudOpenGate/opShowVoice/getCloudEquipStatus/setDefaultScreen)。 | 唯一瑕疵:SKILL.md:47,76 把 resultCode=908 写成"找不到设备",而 codes.go:57 通用映射为"其它错误"。经核 commit b1311bb 是 channel-snap 实操经验修正(可接受),但 line76 泛化到"扫码机"证据弱,且与 codes.go 不一致,Agent 交叉核对会困惑。 |
| A2 | 4 | device 域 callable 但 included=false 的 6 个在 SKILL.md:49 逐个点名指向 api 兜底+[[api-explorer]];webhook(reportChannelBindForScan/reportQrCode)正确排除。 | line49 兜底未加"用 api 调前先 jq 查 params/readwrite、勿猜字段"的强制提示。 |
| A3 | 5 | 三例 flag 名与 --help 逐字一致;用 1ZS7H5PQH9、当前时间 2026-06-01;显式警告不要照抄历史 sampleBody;--client-id 标注需换真实设备 ID。 | none |
| B1 | 5 | description=WHAT+WHEN+去冲突(末句"查某车应显示什么"归 park 域)约 140 字;routing-evals 17 例覆盖正反。 | 未设独立"不应使用本技能"小节(line27 仅一行反向提示)。 |
| B2 | 3 | SKILL.md 仅 105 行;深内容(写幂等/三层判读)下沉到 shared references。 | 无 device 自有 references/;本域全是高危写命令(参数嵌套 setDefaultScreen.templateData.imageArray、枚举多 cloudScanVoice voiceType 0-9),宜建 references/openydt-device-<cmd>.md。 |
| B3 | 4 | CRITICAL 先读 shared;边界声明;[[api-explorer]] 兜底;链到 shared 的 result-reading-sop.md/write-idempotency.md;读vs写、设备下发vs业务查询边界明确。 | none |
| C1 | 4 | CRITICAL MUST 先读 shared;"除 get-cloud-equip-status 外全部写操作必须加 --yes";标准顺序"定位→查状态→--dry-run→--yes";实测写操作无 --yes 拦截。 | 硬约束多为大写 MUST;未三层分级;未提及已存在的 --read-only。 |
| C2 | 5 | 结果解读关键点:开关闸/抓拍 status=1 仅表"指令已下发"≠物理动作完成,须以 get-cloud-equip-status 或停车场域复查为准——设备域最核心高代价误判防护;枚举值带中文含义。 | none |
| C3 | 4 | "错误自愈速查(设备)"表:908找不到设备→换通道;status=7→channelId/channelCode 用错;下发后无反应→设备离线→查在线再下发;现象→含义→下一步闭环;通用码回指 shared。 | 恢复动作多为文字未给 nextCommands;status=7 未配 retriable;908 含义与 codes.go 不一致(见 A1)。 |
| C4 | 4 | 标准顺序强制"先定位设备再干预";高危写先 --dry-run 核对再 --yes;传统/云场 channelCode vs channelId 选择消歧。 | 未把"先问用户确认目标车场/通道、env 是否 prod"提为显式前置;不可逆物理动作未要求复述"将对 parkCode/channel 执行 X"再 --yes。 |
| C5 | 4 | 首行 CRITICAL 强制先 Read shared;--yes/先 --dry-run 指令链清晰;实测 op-gate 无 --yes 即 Error;路由 evals 17 例。 | 无 device 域行为 eval;--read-only 既存全局只读姿势未在技能内引导。 |
| D1 | 4 | description 自标"高危现场运维,写操作需 --yes、建议先 --dry-run";设备控制定性为高危;实测 --yes 守护生效,--read-only 全局过滤存在;无 PII。 | 未点明 env=prod 现场设备额外门禁;未建议批量/现场操作显式 --scope-park 钉死目标车场;未引导 --read-only 作探查默认。 |
| D2 | 5 | "🔑 重复下发风险"明确开关闸/抓拍/扫码无幂等键,网关 404 自动重试可能重复开闸,要求高危写"先 dry-run、单次执行",不确定用 get-cloud-equip-status 复查而非盲目重发,并链 write-idempotency.md。对无幂等键写操作最恰当处理。 | none |

**差距小结:** A1/C3 的 908 含义与 codes.go 不一致需注明"device 域实测含义";B2 高危写命令建 per-cmd reference;C1/C4/D1 三层化 + 引导 --read-only/--scope-park。

### 3.9 openydt-list(黑/红/访客名单域)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 5 | 全 8 命令经 blacklist/redlist/visitor --help 与 catalog 逐条核对:命令名、读写、flag 全部一致;vipGroupType=2黑名单/1访客与 addSpecialCarType desc 一致;status=7/909/status=2 与 codes.go 一致;无计数声明可漂移。 | binary 的 redlist get/del 子命令 help 描述被生成器写反(get 显示"删除",del 显示"新增"),属 cmd/gen 生成产物缺陷、非本 SKILL 表问题(SKILL.md:39-40 表述正确)。 |
| A2 | 5 | 三域共 8 个 callable 接口全部 included=true 且全部出现在命令表;零盲区。 | none |
| A3 | 4 | 声明仅 test、用 1ZS7H5PQH9/PTD2YBBZ;示例均用测试 parkCode+未来 test 时间;车牌与 catalog sampleBody 不同非照抄。 | 示例里 special-car-type-id 253/154 数值与 catalog sampleBody 相同;虽 SKILL 明确应取自响应,裸数值仍易被误当可直接跑常量。 |
| B1 | 4 | description WHAT+WHEN+单 owner 去冲突(specialCarTypeId 创建归 ticket);正文意图路由表。 | description 略长(~150+字);无显式"不应使用本技能/不要因提到X误触发"反向小节。 |
| B2 | 3 | SKILL.md 仅 127 行;命令表+流程+错误表+示例齐全自洽;evals/ 已建。 | 无 references/;写命令参数表/嵌套/坑点全留主文件,缺一等 reference 落点。 |
| B3 | 4 | [[shared]]/[[monthticket]] 链接经文件存在性核对有效;读/写边界、本域引用vs ticket 创建边界、历史/实时归属清晰。 | none |
| C1 | 4 | CRITICAL 先读 shared;写操作必加 --yes;⚠️仅传 --car-no 取消全部条目=批量影响、执行前先查;dry-run 先于 yes。 | 硬约束以 CRITICAL/MUST/⚠️ 为主缺"规则+为什么";无三层分级,无命名反模式清单。 |
| C2 | 3 | "get-park-black-list 返回 0 条"明确教"车场无黑名单 或 parkCodeList 未传全→确认后再下结论";教 specialCarTypeId/vipGroupType 语义与字段来源。 | 名单域无金额字段(元/分 NA);未点明返回列表里 blacklistId/visitorId/ruleId 具体取字段路径。 |
| C3 | 5 | 专设"错误自愈速查"表:加黑 status=7/909、访客 status=2、取消疑未生效、0条→含义→恢复动作(查类型ID/确认 vipGroupType/等1-2秒复核/确认parkCodeList);与 codes.go ResultHint 对齐、给可执行命令。 | none |
| C4 | 4 | 写操作先 dry-run 后 yes;批量取消(仅 car-no)执行前先 get-park-black-list 确认范围、优先精确 id;意图路由表先消歧业务类型。 | parkCode 缺失/env=prod 前置澄清未成显式决策树;对"多车场授权商先钉死目标车场"无 scope 提示。 |
| C5 | 3 | CRITICAL MUST 先 Read shared;写操作加 --yes 指令在位;evals/routing-evals.json 17 条触发/去冲突。 | 仅触发 eval,无行为 eval;遵循度未被可跑断言锚定。 |
| D1 | 4 | 写操作统一 --yes;PII:--phone/--car-no prod 不记真实值并链 shared;test-only+dry-run;示例环境隔离声明。 | 批量节流/限速未在本域提(名单写量小);prod 写门更多依赖 shared。 |
| D2 | 4 | "写入幂等与确认":add 同车牌平台按车牌去重、不确定先 get 查;remove/cancel 仅传 --car-no 取消全部→先查范围、优先用 id 精确取消。 | 本域无 billCode 类显式幂等键,幂等靠"车牌去重"平台语义(已讲清);对"重试是否重复创建"(网络抖动重发 add)未给显式去重键检查步骤。 |

**差距小结:** B2 写命令建 per-cmd reference;C5 行为 eval(批量取消先 get 确认范围、优先精确 id);C1/C4 三层化 + 规则+why。

### 3.10 openydt-monthticket(月票/VIP 域)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 4 | 全部 29 条命令名与 `ticket --help` 及 catalog(included=true 计数=29)逐条一致;命令未幻觉。 | **命令表必填漏列:** month-ticket-config-edit 漏顶层必填 `price`(catalog group=null),漏列触发 status=7;get-online-month-ticket-by-car-card 顶层必填遗漏(示例已补,影响缓解);add-special-car-type 漏 channelSeq*/areaId*(属组内条件必填,可辩护)。把组内条件必填与顶层必填混同。 |
| A2 | 5 | ticket 域 included=false 的 callable 28 条全部出现在"未一等化(用 api 兜底)"小节,零遗漏;正确把 getMonthTicketAppointmentPark 标为 park 域并指向 [[park]]。 | none |
| A3 | 5 | 示例全用 1ZS7H5PQH9/PTD2YBBZ 与当前时间 2026-06;实跑 add-online-month-ticket-type --dry-run 产出签名请求,CLI 自动补默认值;read 示例实跑返回 status:1 total:0;明确提示照抄历史 sampleBody 会撞无效车场。 | none |
| B1 | 5 | description WHAT+WHEN 兼备,含触发短语,显式去冲突"临停实时算费用 trade 域";routing-evals 18 条覆盖兄弟域去冲突。 | 可在"何时用"下补显式"不应使用本技能"反向小节;属锦上添花。 |
| B2 | 3 | SKILL.md 189 行;大块横切下沉到 shared 的 references。 | 本域无 references/;对参数复杂/写命令(add-online-month-ticket-type 字段巨多/renew/deduct/add-special-car-type)宜补 references/openydt-ticket-<cmd>.md,命令表加"必读 reference"列;当前所有参数细节仅靠 --help。 |
| B3 | 5 | [[links]] 齐备且边界清晰目标 SKILL 全部实存;读/写归属明确;域vs名单成员边界、临停算费→billing、appointmentPark→park、api 兜底→api-explorer。 | none |
| C1 | 4 | CRITICAL 先读 shared;写操作"必须 --yes 否则被拒"(实测);金额单位=元、幂等键重试复用、907=幂等命中改查规则明确。 | 硬规则多为"MUST"祈使缺 why 散文;无三层分级,无 named anti-pattern 小节。 |
| C2 | 4 | 金额(originPrice/favorPrice/refundPrice/price)单位=元;907=幂等命中改查 get-online-vip-ticket;指向 result-reading-sop.md;教取 data.monthTicketConfigId 等;实跑 read 返 total:0 印证 0条≠无。 | 未点明 status vs data.code;未给 vipGroupType/payMode/payOrigin 等枚举值→中文映射表;列表读命令未回指 has_more 全量结论契约。 |
| C3 | 3 | 给了一条具体自愈:907 账单已同步=幂等命中→改查确认、不要再发;"有出入信 --help"。其余靠 shared。 | **本域缺"错误自愈速查表"(对标 billing):** 未对 status=7/6/5/9 及月票特有失败给闭环;仅 907 一条成闭环。 |
| C4 | 4 | 意图路由先分读/写;写操作示例统一"先 --dry-run 预览,确认后再 --yes";ID 类参数须取自上一步响应。 | 未显式要求写前消歧 parkCode/env/支付方式;退费 refundPrice/续费 favorPrice 这类动钱操作未要求向用户复述确认;多车场未建议 --scope-park。 |
| C5 | 4 | 实证:read 示例实跑成功;add-online-month-ticket-type --dry-run 产出合法签名请求;add-special-car-type 无 --yes 被 ConfirmWrite 实测拦截;三处强制遵循路径俱在;routing-evals 18 条。 | evals 仅 routing,无行为 eval。 |
| D1 | 4 | 13 处 ConfirmWrite 守护所有写命令(实测拦截);复述守护;示例默认 test;全局 --read-only 暴露;金额/PII 示例仅 test。 | 未在本域正文写明"prod 不记 PII/批量写仅 test/prod 默认不可达"(依赖 shared);批量场景无节流提示;无 plan→validate→execute 批量校验入口。 |
| D2 | 5 | 明确教幂等:billCode/thirdpartyBillCode/thirdpartyIdentify 是幂等键"重试复用首次值、绝不新生成";907=幂等命中、首次已开通、改查确认;指向 write-idempotency.md;catalog 印证。 | none |

**差距小结:** A1 命令表漏 price 必填(会触发 status=7)最该修;C3 缺本域错误自愈速查表;B2 复杂写命令建 per-cmd reference。

### 3.11 openydt-park(车场信息域)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 5 | 命令表 18 条逐条核对 catalog(domain==park && included==true)的 18 个 callable 接口完全一致;`park --help` 同样列出;名称/读写/必填(*)标记吻合(set-park-remain-carport 标 write 需 --yes)。 | none |
| A2 | 5(评分热力表取保守 3,实为强项) | park 域 callable 共 21 个:18 included 全一等命令;另 3 个 callable-but-not-included(getParkEquipmentInfo/getCarOwnerInfo/getMonthTicketAppointmentPark)在 SKILL.md:63 显式列出并指向 api 兜底+[[api-explorer]],无遗漏。 | schema 命令指路缺失:api 调前先 schema 查参未强制。 |
| A3 | 5 | 示例统一用 1ZS7H5PQH9,charge 域用 PTD2YBBZ;set-park-remain-carport 示例 body 用 1ZS7H5PQH9 而非 catalog 历史 sampleBody;经纬度标注"按实际位置替换"。 | none |
| B1 | 4 | description=WHAT+WHEN+边界(算费见 trade/屏显见 device/券见 coupon);routing-evals 13 同域+4 跨域反例。 | description 206 字略超;无 lark-style"不应使用本技能"反向小节。 |
| B2 | 3 | 主文件 119 行;有 references/openydt-park-charge.md 承载 chargeMap 字段解读+转人话范例,直链一级深。 | 复杂/写命令(set-park-remain-carport 的 remainCarportList、display-voice 枚举、other-car-type-charge 嵌套回填)未各建 reference;命令表无"必读 reference"列;现有唯一 ref 与正文 SKILL.md:74 高度重复。 |
| B3 | 5 | 意图路由表全域 [[links]] 到位;读(本域查询)vs写(算费/缴费去 trade、券去 coupon、屏显去 device)归属清晰。 | none |
| C1 | 4 | CRITICAL MUST 先 Read shared;写操作标"需 --yes"+两步 dry-run→yes;强调 standardSeq/carType/parkYdtChargeVo 必须取自上一步响应不可臆造。 | 缺 Stripe-style 具名反模式("NEVER 拿 chargeMap 当精确账单"在 ref 但未升主文件 CRITICAL);未提全局 --read-only 探查姿势。 |
| C2 | 5 | references/openydt-park-charge.md 详尽教读响应:chargeMap value.fee"单位元"、key 离散档(1/2/3/4/8 小时)、type(0自定义/1免费/2循环递增/3按次固定)翻人话、以当前时刻试算的边界、stoppingTimeStr 恒空;区分"计费组vs规则原文""预览估算vs精确账单"。 | 空车位/区域类响应未单独教"0/null≠无车位 vs 未授权"(靠 shared 兜底)。 |
| C3 | 3 | 指向 shared 的状态码/限速;shared 有 status-codes/result-reading-sop 兜底。 | **本技能正文无"错误自愈速查表"(对标 billing);** 未列 park 域高频失败(parkCode 不在授权列表→先 get-auth-park-codes、查费 912、云车场命令对 VEMS 报错)闭环与 retriable。 |
| C4 | 4 | 提示先用 get-park-list/get-auth-park-codes 获取 parkCode 再查;写操作先 dry-run 再 yes;说明唯一需回填链路。 | 无三层分级;set-park-remain-carport 未要求先与用户确认目标 parkCode/env;多车场未建议 --scope-park。 |
| C5 | 4 | 首行 CRITICAL MUST Read shared;写命令 help 实测"write (需 --yes)";示例硬编 dry-run→yes 两步;routing-evals 17 条。 | 无行为 eval;仅触发/路由 eval。 |
| D1 | 4 | 唯一写命令 set-park-remain-carport 挂 --yes 守护+dry-run 优先(实测 help 确认);parkCode 用 test 值注明"仅 test";依赖 shared 的 prod 门/PII/限速。 | 未在本技能正文复述"写操作仅 test、prod 不可达写";未提全局 --read-only 探查姿势;set-park-remain-carport 是覆盖型写,未提醒误传会覆盖真实车位数。 |
| D2 | NA | park 域唯一写命令 set-park-remain-carport 是幂等覆盖型(同参重发结果相同,无 billCode 概念,不涉扣费);catalog 该命令无幂等键字段;通用重试幂等由 shared/write-idempotency.md 承载,对本域不适用。 | none |

**差距小结:** B2 渐进披露未铺开(写/嵌套/枚举密集命令各建 reference,现唯一 ref 与正文重复);C3 缺错误自愈速查表;C1/D1 把"NEVER 拿 chargeMap 当精确账单"升主文件 CRITICAL + 提示覆盖型写 + --read-only/--scope-park。

### 3.12 openydt-record(停车记录域)

| 维 | 分 | 证据(file:line) | 差距 |
|---|---|---|---|
| A1 | 5 | 命令表 33 条与 `parking --help` 逐条一致(comm 零差异);读/写列与 catalog 全对;抽查 params 全对;updateWihholdDetailBill 'wihhold' typo 警告经 jq 确认为平台原始拼写;inventory-car/get-inventory-record 二进制已暴露为一等命令(commit ca5df92),技能列出且读写正确,与真相源(二进制)一致。 | none |
| A2 | 5(热力表保守 3) | SKILL.md:76 api 兜底注列出 getHisParkDetail/getParkPayBill 等(jq 确认 included=false 可调用),统一指路 openydt api <cmd> --body 并 [[api-explorer]];33 个一等命令无遗漏。 | "调 api 前先 jq 查 params"强制度可再加强、回指 schema 命令缺失。 |
| A3 | 5 | SKILL.md:95 显式说明示例用 1ZS7H5PQH9/PTD2YBBZ+2026 相对时间,并警告照抄历史 sampleBody;四示例均用这两个 parkCode 与 2026 日期,确未照抄。 | none |
| B1 | 4 | description WHAT+WHEN+显式边界"实时算费/缴费回传用 trade 域,本域只查历史";长度适中;意图路由表单 owner。 | 缺"不应使用本技能"反向场景小节;可补少量口语触发短语抵消 under-trigger。 |
| B2 | 3 | SKILL.md 146 行;大块下沉 references/pitfalls.md(9行)+flows.md(37行),各给"先 Read"按需加载指引;一级深直链。 | 对标 lark"每写命令一 references/<域>-<cmd>.md"未做(supplement/correct/inventory/cancellation 参数表/坑点/幂等键仍在主文件或 flows);命令表无"必读 reference"列;references 首行未复述"先读 shared"。 |
| B3 | 5(热力表保守 3) | 全域 [[links]] 跨域路由齐;[[shared]];[[api-explorer]];读vs写归属与域vs编排边界清晰。 | references 首行复述前置条件可补齐。 |
| C1 | 4 | CRITICAL MUST 先读 shared;写操作均需 --yes;wihhold MUST 照拼写否则 status=9;lock-car --help 实测 "write(需 --yes)" 且有 --read-only/--dry-run 守护。 | 硬约束多用全大写 MUST/NEVER 少带 why;未采用 ✅/⚠️/🚫 三层结构;无 named anti-patterns 小节。 |
| C2 | 4 | pitfalls.md 教"不传时间返回0条易误判无在场"(0条≠无)、scan-channel 外层 status=1 但 data.code≠0 才是业务失败看 data.msg(status vs data.code)、get-park-detail 与 ignore-status 状态不一致判读。 | 本域有 get-pay-bill/欠费等金额查询却未教金额单位(元vs分);未回指 result-reading-sop 的 Final Answer Check;分页 has_more 全量计数可信边界未在命令表"路由提醒"点明。 |
| C3 | 3 | pitfalls.md 给会话已过期→先 channel-snap 再校正;data.code 非0→看 data.msg;字段错→报参数错的纠正(carNo 单数/leaveStartTime)。 | 无集中"错误自愈速查表"(对标 billing),未对常见业务码给 nextCommands/retriable;0条与满页未分别给 hint;未对接 shared 的退出码 SOP。 |
| C4 | 3 | flows.md 通用原则"先用读命令定位记录拿字段再写,不凭空填";幂等段写操作前先用读命令确认首次是否生效。 | 未对写操作显式要求"先澄清 parkCode 缺失/env=prod/operator"再 dry-run 后 yes;无意图消歧决策树(锁车用 carNo vs cardNumber、补录枚举值含义)。 |
| C5 | 3 | 强制先读 shared、写操作示例均带 --yes、路由按域选;evals/routing-evals.json 17 条触发/路由(含去冲突)。 | 缺行为 eval(EDD):无端到端断言("查欠费报金额带元单位"、"重发补录前先 get-park-on-site-car 确认在场")。 |
| D1 | 4 | 写操作均需 --yes;建议先 --dry-run 预览;二进制实测有全局 --read-only 与 --dry-run;示例全用 test 测试 parkCode。 | correct-car-no/补录/get-pay-bill 涉真实车牌,未提"prod 不记 PII/批量节流(分页类)/test-dev-prod 隔离"(在 shared 未在本域回指);prod 写门未在域内点明。 |
| D2 | 5 | "写操作幂等"段:update-wihhold-detail-bill 用 thirdBillCode 去重、重试复用同值绝不新生成(catalog 确认 required=true);明确 supplement/inventory/correct 无显式幂等键,网关 404 自动重试可能重复补录/盘点,给出重发前先用 check-channel-exist-car/get-inventory-record/get-park-on-site-car 读确认 SOP;链 write-idempotency.md。 | none |

**差距小结:** C2 金额查询缺单位解读 + 分页可信边界;C3+C5 补错误自愈速查表 + 行为 eval;B2 写/复杂命令建 per-cmd reference 且 references 首行复述先读 shared。

## 4. 对标差距表

> 四个对标产品共 26 个模式;下表精选对我们 backlog 最有杠杆的条目。完整 pattern/oursToday/gap/recommendation 见 `tools/eval/eval-output.json` 的 `benchmark` 段。

| 模式 | 来自产品 | 我们现状 | 差距 | 借鉴建议 |
|---|---|---|---|---|
| 每命令一 references/ 子文档(SKILL 只做命令导航表+"必读 reference"列) | lark-cli(lark-base 362 行挂 94 个 reference) | 每域 0-2 个 reference,多数命令字段挤在主文件一行"关键参数" | 单命令字段/JSON 形状/坑点无可按需加载落点,主文件随命令增多逼近 500 行 | 对写/高参命令补 references/openydt-<域>-<cmd>.md(先 trade 写、parking 补录/校正、coupon 发券、ticket 开月票),命令表加"必读 reference"列 |
| reference 统一骨架(推荐命令/参数/入参详情/返回重点/坑点/参考 + 首行"先读 shared") | lark-cli | references 各写各的,无单命令模板 | 新写 reference 无范式,Agent 无稳定"读哪个 heading 拿什么"预期 | 在 skill-maker 定 openydt 单命令 reference 模板(参数表[必填·类型·单位·嵌套group]/cmd与readwrite/返回重点/幂等键/坑点),首行统一"前置:先读 openydt-shared" |
| 高风险写用机器可读 exit-code + confirmation_required envelope,shared 给识别 SOP | lark-cli(exit10)/MCP | 写守护靠 --yes,反馈是人读文案;`_error` 已富(hint/retriable/nextCommands)但无 confirmation_required + 退出码契约 | 无机器可读拦截信号让 Agent 自动走确认流程;命令是否高风险无法从 schema 程序化判定 | 未带 --yes 返回非零退出码 + `{error:{type:'confirmation_required',risk:{level,action}}}`;让 schema/--help 暴露 risk 等级 |
| 全局 --read-only 作最高优先级硬过滤(会话级总开关,优先于一切) | MCP(GitHub/Supabase) | **CLI 已实现 --read-only/OPENYDT_READ_ONLY=1 且实测对 api 写 cmd 即使带 --yes 也拒绝**,但 shared/api-explorer/各域 SKILL 全文零提及 | 一个一等写安全控制点在文档面完全缺席;Agent 不知道有此护栏 | 在 shared 硬约束区写"不确定意图默认 --read-only 探查",各域 SKILL 回指;这是已具备能力的零成本补强 |
| MCP 工具注解 readOnlyHint/destructiveHint/idempotentHint(机器可读风险词汇) | MCP 2025-03 规范 | schema --json 已含 hints{readOnly,destructive,idempotent} 三态(实跑验证) | 已基本对齐;但 idempotent 写命令未在输出点明幂等键字段名 | idempotent=true 的写命令(缴费/补缴)在 schema/_error 点明幂等键字段(billCode/thirdBillCode),把口头约定升级 per-command 契约 |
| Idempotency-Key 由 SDK 自动生成+持久化+参数不匹配硬报错 | Stripe | write-idempotency.md 核心规则强(复用首次键/907=命中/per-cmd 键表);但键全靠 caller 手生成手记忆,无 mismatch 守护 | Agent 须手生成手记 billCode(正是规则警告的 mint-new-key 失败模式);改参数复用键平台行为未文档化 | CLI 缺省时自动生成 billCode 并在 stdout/--dry-run 回显;write-idempotency.md 加"改任一参数必须 mint 新 key"(对齐 Stripe mismatch guard) |
| EDD 评测驱动开发:写文档前先建行为 eval + 量基线 | Anthropic Skills | 只有触发/路由 eval(routing-evals.json);skill-maker 指引的 run_loop.py 路径未锚定、所指"shared 评测约定"不存在(实证断裂) | C5 遵循度靠主观勾选,无"给 Claude 真实任务→观察行为→改"闭环 | 在 shared 落可跑触发 eval SOP(subagent 路由,规避 MEMORY 记的嵌套限制);为写操作域造 ≥3 端到端 expected_behavior eval;EDD"先无技能量基线"写进步骤第 0 步 |
| 硬规则"规则+why"优于纯大写 MUST + 自由度分级 | Anthropic Skills | shared 已领先(MUST/NEVER 每条附 why);但形式仍偏大写堆叠,无按脆弱度分级自由度 | 大写堆叠易让模型"跟字面漏边界" | 把最关键几条改写成"规则+理由"散文;写操作段用低自由度("严格按此5步不增删 flag"),查询段高自由度 |
| 命名空间划清边界 + fieldDesc 把隐含上下文显式化 | Anthropic/Cloudflare | 命名空间清晰(openydt <域> <命令>),12 域 description WHAT+WHEN 去冲突 | parking 31/coupon 30 命令的域内近义命令边界可能模糊;schema fieldDesc 枚举未必都带中文含义 | 命令数多的域按子主题分小节;审 schema 的 fieldDesc/allowedValues 确保枚举带中文(payment-mode/pay-origin) |
| AGENTS.md 三级边界(✅ Always/⚠️ Ask-first/🚫 Never) | AGENTS.md 约定 | AGENTS.md 有 MUST/NEVER 二元块(好且可执行) | 无显式中间"⚠️ Ask-first"层,而 2500-repo 研究指出该层最区分高/低绩效 Agent | 把约束块重塑为三级标签,把 C4 意图澄清(env=prod/支付方式/parkCode 缺失/批量写)提为一等可扫规则 |
| response_format concise|detailed 控制返回体量 | Anthropic | 输出只有 -o json|table,无 verbosity 档,json 默认回平台全量包络 | 大列表(在场车/记录/账单)整包灌 Agent 上下文,费 token 混入下游用不到字段 | 读命令加 --format concise|detailed(默认 concise 只回链路必需字段),catalog 标 essential:true 驱动裁剪 |
| llms.txt 单一精选入口索引(by-section) | Stripe/llms.txt 约定 | 无 llms.txt;入口分散(CLAUDE.md/AGENTS.md/README/INTERFACE_INDEX);技能内渐进披露好 | 冷发现(npm 装包未预载技能)无单一精选地图 | 不为追标签硬加 repo-local llms.txt;借鉴思想——INTERFACE_INDEX.md 当"命令版 llms.txt"保持 make index 生成防漂移;有 hosted docs 时再从 catalog 生成 /llms.txt |

## 5. 实证发现

> 5 个端到端用例,均在 test 环境、actor 真实执行轨迹。无 blocked,无 fail;1 个 partial(E-pay 的 C4)。详细 checkpoints 与 rootCause 见 `tools/eval/eval-output.json` 的 `empirical` 段。

### E-pay(查费→缴费 dry-run,trade 域)
- **C5 遵循度 = pass:** readSharedFirst=true(先 Read shared 再加载 billing);按路由正确选 trade 域未误入 parking;只读 get-park-fee 不挂 --yes、写 pay-park-fee 严格走 --dry-run;auth test 先冒烟、回忆 park-note 命中目标车场。
- **C4 意图澄清/确认前置 = partial(非 pass):** parkCode/env 已三重确认且正确识别 69872 为陈旧夹具非真实欠费;但写操作 askedBeforeWrite=false,未就支付方式/是否实际回传向用户澄清。因止步 dry-run、未加 --yes、零真实写,风险被 dry-run 兜住,故记 partial 而非 fail。
- **D1 写操作安全 = pass:** 仅 --dry-run 预览未加 --yes 未真实回传;全程 test 无 prod 写;车牌 粤EJW962 仅在 test 不违反"prod 不记 PII"。
- **C2 结果解读 = pass:** 正确判 status=1;金额解为元(shouldPayValue=69872 元、paidValue=0,明确标注非分);基于 enterDate=20251001 陈旧+parkingTime 异常巨大正确识别为夹具产物。
- **D2 幂等 = pass:** --bill-code 用 DEMO-<ts> 形式体现 billCode 全局唯一意识;dry-run body 含来自查费的 chargeBillToken/chargeBillNumber 回传字段。
- **根因:** 无技能缺陷;唯一可强化——shared 在"写操作安全规则"处把"dry-run 前先与用户确认意图(支付方式/是否实发)"写成显式步骤(C4 补强)。

### E-onsite(查在场车,record 域)
- **C2 结果解读 = pass:** 首跑 05-01~06-01 命中 status=2/909"区间>1月"正确判业务失败并收窄窗口重试;status=1 后从 data 内层读 count=97 而非只看外层 status,明确区分"count=窗口在场数 vs recordList 受 pageSize 截断需翻页";标注约18% 车牌未识别、按 enterVipType 拆解车类。
- **C5 遵循度 = pass:** 先 Read shared+record 域技能再执行;查在场车走 parking get-park-on-site-car 未滥用 api;只读未加 --yes;config list 确认 demo/test、auth test 通过、复用既有 park-note、因无新事实未重复写 note。
- **根因(可改进项):** catalog `getParkOnSiteCar` 官方 desc 两处与实测相左——enterTimeFrom/To 注"间隔不超过1天"实际 1 个月(909);enterVipType 注"0未定义/1普通/2本地VIP/3外部VIP/4黑名单/5访客"而真实语义 1临时/2月票/4黑名单/5访客/8白名单。actor 靠 park-note 经验绕开,但属个人车场经验兜底。建议把 vip-type 真实枚举与区间=1月 勘误补进 record/pitfalls.md,降低对单车场 note 依赖。

### E-error(两症状错误自愈,record 域)
- **C3 错误自愈闭环 = pass:** 两症状被正确拆分各自闭环。(1)"会话已过期"判为业务层数据态(非签名/鉴权),修复=device channel-snap --yes 生成会话→确认 data.code=0→时效内同 channelCode correct-car-on-channel --yes,与 record/pitfalls.md:9 一致并正确点出 scan/snap 类外层 status=1 仍须查 data.code;(2)909 正确解读为 status=2"请求参数错误",对齐 codes.go:59-60,锁定 get-car-out-list 三处常见失败(carNo 单数、两组时间二选一必填、pageSize≤100),与 pitfalls.md:5 及 schema 实际输出逐字吻合。轨迹只读/dry-run,闭环干净。
- **根因:** 正向通过;增强项——可在话术显式标注"909 与会话过期均不可自动重试(retriable=false)"以更贴合 codes.go 语义。

### E-api(callable 盲区兜底,createCityOperationCouponTemplate)
- **A2 盲区覆盖 = pass:** cityOperationCoupon 全族 callable/included=false/0 一等命令,由 coupon/SKILL.md:69(域内指路)+api-explorer(兜底主体含范例)+shared:114(通用 api 兜底)三处协同闭合;actor 走通 coupon --help→catalog→api dry-run。
- **C5 遵循度 = pass:** 先读 shared 再读 coupon;止于 dry-run(usedDryRunBeforeYes=true,usedYesOnReadonly=false);复跑 dry-run 复现预览(POST test、路径 createCityOperationCouponTemplate、v2 签名、compact body 逐字节一致);按"一等命令优先→api 兜底"路由。小瑕:askedBeforeWrite=false(止于 dry-run 不扣分)。
- **根因(关键发现,不影响判定):** **api-explorer/SKILL.md:72 仍称"openydt api 路径不调用写确认、漏 --yes 会直接发出",但当前 run.go:42-53 RunCall 对所有路径统一调 guardWrite——实跑 `openydt api createCityOperationCouponTemplate`(无 --yes)被拦"是写操作,需加 --yes 确认",证明守护已生效。** 该技能描述偏保守(比实际更安全)未削弱安全性,但属过时陈述,即第 1 节 P0 缺陷的实证来源,须同步修正。

### E-route(5 意图路由,B1)
- **B1 description WHAT+WHEN 完整 = pass:** 5 域 description 三段式,长度 trade 151/parking 157/park 206/ticket 179/coupon 138(park 偏长因覆盖面最广,但非堆砌)。
- **B1 单 owner 无冲突 = pass:** 逐意图核验路由唯一性:实时查费=trade、在场车=parking、历史账单=record(record.desc 与 billing.desc 双向去冲突)、出口屏显=park(park.desc 与 device.desc 双向互指)、是否月票VIP=ticket(全仓唯一 owner)。5/5 单 owner。
- **根因:** 无功能性缺陷;actor 11 步全部命中正确域且首命令正确,意图④的 status=9 是 test 环境未部署该接口的环境限制非路由错误。可改进:park description 偏长可收紧。

## 6. 优先级 backlog

> P0=正确性/安全(A1 幻觉命令、D2 重复扣费、误导性安全心智);P1=高频行为(C2/C3/C4);P2=结构/一致性(B2/B4)。工作量:S≈半天 / M≈1-2天 / L≈3天+。

| # | 缺陷 | 影响维 | 建议改动 | 工作量 | 优先级 |
|---|---|---|---|---|---|
| 1 | api-explorer SKILL.md:68-73,127 称 api 不拦截写/漏 --yes 直发,与代码相反(实测被 guardWrite 拦) | A1·C1·D1·C3 | 改写为"api 与一等命令同享写守护;漏 --yes 会被 CLI 拦下(已实测)";删速查表第3行不存在的"漏 --yes 却真发出去"现象;补 --read-only 硬过滤说明 | S | **P0** |
| 2 | coupon SKILL.md:122 券种区分写成 balanceType(实为 couponType) | A1·C2 | 改为 couponType(0免费/1金额扣减/2折扣/3固定/4时间券)并在散文点明;balanceType 单独说明=结算类型 | S | **P0** |
| 3 | monthticket 命令表漏顶层必填 price(month-ticket-config-edit),漏列触发 status=7;组内条件必填与顶层混同 | A1 | 必填列以 catalog group=null 的 required 为准;补 price*;组内条件必填单列标注"(填 X 组时必填)" | S | **P0** |
| 4 | device SKILL.md:47,76 resultCode=908 写"找不到设备",与 codes.go:57 通用"其它错误"不一致 | A1·C3 | 注明"此为 device 域实测含义(commit b1311bb),通用码见 codes.go";line76 泛化到扫码机的行降级为待证或限定 channel-snap | S | P1 |
| 5 | 全集缺行为 eval;skill-maker 指引的 run_loop.py 路径未锚定、所指"shared 评测约定"不存在 | C5(全域)·D2 | 在 shared 落可跑触发 eval SOP(subagent 路由);为写操作域造 ≥3 端到端 expected_behavior eval(先读 shared/查费取 shouldPayValue 当元/先 dry-run 后 yes/重试复用 billCode);EDD"先无技能量基线"写进 skill-maker 第 0 步 | L | **P1** |
| 6 | C4 意图澄清/确认前置全集偏弱(shared 无三级 ask-first、无支付方式/parkCode 缺失决策树、无 CONFIRM 复述行) | C4(全域) | shared 硬约束块重塑为 ✅可直接做/⚠️先确认/🚫绝不 三级;写操作 --dry-run/pre-yes 输出一行 "CONFIRM: pay {actPayCharge}元 to {parkCode}/{parkingCode}, key={billCode}" 供 Agent 向用户复述 | M | **P1** |
| 7 | C3 错误自愈速查表缺失:park/record/monthticket(均含写操作)无该表;coupon 仅 2 行 | C3 | 以 billing 的速查表为模板推广到缺失域(现象→含义→nextCommands+retriable);补 0条/满页两种 hint | M | P1 |
| 8 | C2 金额/分页解读盲区:record 金额查询(get-pay-bill/欠费)未教元单位;data/record 分页未点明 has_more 非全量禁全量计数 | C2 | record 金额命令内联"单位=元"并回指 result-reading-sop;record/data 分页命令"路由提醒"加 has_more 全量结论硬契约 | S | P1 |
| 9 | shared/api-explorer/各域 SKILL 全文零提及已上线的 --read-only(全局硬写过滤,实测对 api 写 cmd 即使带 --yes 也拒绝) | D1·C4·A1 | shared 全局 flag 表 + 安全规则段补 --read-only/OPENYDT_READ_ONLY=1;各域 SKILL 回指"不确定意图默认 --read-only 探查" | S | P1 |
| 10 | B2 渐进披露未铺开:多数高参/写命令无 per-cmd reference,命令表无"必读 reference"列 | B2 | 在 skill-maker 定单命令 reference 模板(首行先读 shared),先覆盖 trade 写/parking 补录·校正/coupon 模板·发券/ticket 开月票/device set-default-screen | L | P2 |
| 11 | B4 跨技能格式漂移:写操作"读/写"列 4 种写法、必填标记与关键参数列 4:4 分裂,违反 skill-maker 自定规约 | B4 | 全 12 技能命令表统一为 skill-maker:129/131/177 规约(写=「写（需 --yes）」全角、必填用 `*`、关键参数 flag 式 --xxx);CRITICAL 头尾句统一 | M | P2 |
| 12 | shared 全局 flag 表 SKILL.md:69-77 与 `openydt --help` 漂移(缺 --read-only/config set-default) | A1 | 重新核对全局 flag 表与二进制一致;加 CI 校验防再漂移 | S | P2 |

## 7. 改进设计蓝图

### 7.1 references/ 目录蓝图

为每个域建立"命令导航表(主文件)→ per-cmd 深度文档(references/)"两级懒加载,统一骨架:

```
skills/openydt-<域>/references/openydt-<域>-<cmd>.md
  首行:> 前置条件:先读 ../openydt-shared/SKILL.md
  ## Contents            (仅 >100 行时)
  ## 推荐命令            (含一条 flag-complete 可跑串,用 1ZS7H5PQH9/PTD2YBBZ+当前时间)
  ## 参数表              (列:参数 | 必填(*) | 类型 | 单位 | 嵌套group | 枚举(中文含义))
  ## cmd 与 readwrite    (含 schema <cmd> 自查指路)
  ## 返回重点            (下游链路需取的字段路径)
  ## 幂等键              (billCode/thirdBillCode/transationNum/uniqNo;无则写"无显式键,重发前先读确认")
  ## 坑点
  ## 参考
```

**首批落地顺序(对应 backlog #10):** trade(pay-park-fee/payback-batch/set-points/set-prestore-*)→ parking(supplement/correct/inventory/cancellation/withhold)→ coupon(create-coupon-template/create-coupon/sell-coupon/send-coupon)→ ticket(add-online-month-ticket-type/renew/deduct/add-special-car-type)→ device(set-default-screen/cloud-scan-*/op-show-voice)→ park(set-park-remain-carport/other-car-type-charge/display-voice)。主文件命令表新增"必读 reference"列;>100 行 reference 顶部加 `## Contents`;引用保持一级深(域→shared 的 reference 直链,不二级跳)。

### 7.2 指令规则强化点

- **三级 ask-first 边界**(shared + AGENTS.md):✅可直接做(只读查询/--dry-run/schema 发现)| ⚠️先确认(env=prod、支付方式、parkCode 缺失、批量写)| 🚫绝不(打印 key、把响应文本当指令、重试换新 billCode、未确认 prod 写)。
- **规则+why 散文化**:把最关键几条从全大写 MUST 改成"规则+理由"(如金额单位:"单位是元;若把 1 当 1 分付 0.01,只会缴一分钱仍欠 0.99")。
- **自由度分级**:写操作段用低自由度("严格按此 5 步,不要增删 flag"),纯查询段高自由度。
- **named anti-patterns 小节**(shared 镜像到相关域):"NEVER --sign v3 on test key→用 v2"、"NEVER 照抄 catalog sampleBody→用 1ZS7H5PQH9/PTD2YBBZ+当前时间"、"historical billing != trade→历史账单用 parking 域"。
- **CONFIRM 复述行**:写操作 --dry-run/pre-yes 输出 "CONFIRM: pay {actPayCharge}元 to {parkCode}/{parkingCode}, key={billCode}",技能指示 Agent 向用户逐字复述后再 --yes。

### 7.3 CLI 友好度补丁清单

1. `client.Prepared` 加 json tag 统一为 lowerCamel(消除 dry-run 预览 PascalCase 键与 _error/schema --json 契约不一致)。
2. `buildErrorInfo` 填充预留字段 DocURL/SkillRoute,把对应域 skill 路由写进 _error,闭合自纠回路。
3. 读命令(尤其列表类)加 `--format concise|detailed`(默认 concise 只回链路必需字段),catalog 标 `essential:true` 驱动裁剪;补 `--max-data`/字段投影防大响应(实测单次 288KB)打爆上下文。
4. 写操作缺省时自动生成 billCode/thirdBillCode 并在 stdout/--dry-run 回显,便于 Agent 捕获-复用;持久化 last write 的 (cmd,key,request-hash) 以检测"同意图复用键 vs 改参新意图"。
5. schema/`_error` 透出 idempotent 写命令的幂等键字段名;catalog 增 readOnly/destructive/idempotent 三态(对齐 MCP hint)。
6. 可选 `openydt schema --search '查停车费'` 模糊意图→cmd 路由,返回 ranked 候选 + nextCommands/skillRoute。

### 7.4 跨技能一致性整改(B4)

按 `skill-maker:129/131/177` 自定规约,全 12 技能命令表强制统一:写操作"读/写"列一律「写（需 --yes）」(全角)、必填标记一律 `*` 后缀、关键参数列一律 flag 式 `--xxx`;CRITICAL 头尾句统一为标准句"……安全规则)。未读共享基座不要执行任何命令。"(纠正 park/record/flow-park-access 的散文措辞)。可加一条 lint(进 make e2e):扫命令表写操作列写法、必填标记、参数列格式,偏离规约则报警——把规约从文字变成可机检。

### 7.5 是否引入全局 AGENTS.md

**已有,无需新建,但建议小幅增强。** `/AGENTS.md`(~40 行)已是 tool-neutral 单一入口且 CLAUDE.md line5 显式让位,是该约定的参考实现。建议:(1)把约束块重塑为 7.2 的三级边界(把 C4 提为一等可扫规则);(2)quickstart 补一行完整安全写环 `pay-park-fee ... --dry-run` 然后 `--yes`,并示范 `-o table`/`-o json`;(3)补"自检"组(写前 `--dry-run`、查命令存在性 `schema <cmd> --json`、实时计数 `make counts`、凭据健康 `auth test`);(4)可选加最小 frontmatter(description/tags)做 v1.1 渐进披露。**不建议**新建 repo-local llms.txt(会与 AGENTS.md 重复);把 INTERFACE_INDEX.md 当"命令版 llms.txt"保持 `make index` 生成防漂移即可。CLAUDE.md 可进一步瘦身为只留 dev/build/codegen 细节,运营性内容一行让位给 AGENTS.md+shared,避免两个"顶层文档"竞争。
