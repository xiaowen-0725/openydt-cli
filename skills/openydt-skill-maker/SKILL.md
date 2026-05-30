---
name: openydt-skill-maker
version: 1.0.1
description: "创建或规范化 openydt(艾科智泊停车开放平台 CLI)自定义 Skill 的元技能。当用户要新建一个 openydt 域技能、把某接口或多步业务流程固化成可复用 Skill、或规范化已有 SKILL.md(frontmatter / 命令表 / 触发词 / 写操作 --yes 守护)时使用。对标飞书 lark-skill-maker。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt --help"
---

# openydt-skill-maker — openydt 自定义 Skill 制作器

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**，掌握 openydt 的配置 / profile / 签名(v2/v3) / 响应包络与状态码 / 限速 / 写操作安全规则。所有你新建的 skill 都要复用这套基座，不要在新 skill 里重复这些内容。

本技能是「制作其它 openydt skill」的元技能，对标飞书 `lark-skill-maker`：当用户想把 openydt 的某个接口（原子封装）或某条多步业务流程（编排）固化成可复用 Skill 时，按本指南产出一个符合规范的 `SKILL.md`。

## 何时用本技能

- 用户要新建一个 openydt 域技能（如「在场车 / 月票 / 黑名单」域）。
- 用户要把某个 `openydt api <cmd>` 或某条「查费 → 缴费」式业务链路封装成可复用 Skill。
- 用户要把已有 SKILL.md 规范化（补 frontmatter、整理命令表、给写操作加 `--yes`）。

不属于本技能：实际调用业务接口（用对应域技能或 `openydt api`）、改 Go 命令代码（改 `cmd/gen` 须先改抽取器/codegen，而非手写）。

## 目录结构

每个 skill 是 `skills/<skill-name>/` 下的一个目录，至少含 `SKILL.md`；正文较长或有大块参考资料时，把按需内容拆到 `references/`：

```
skills/
  openydt-shared/SKILL.md        # 共享基座（所有 skill 都先 Read 它）
  openydt-<域>/
    SKILL.md                     # 主入口：frontmatter + 正文（必须精简，常驻上下文）
    references/                   # 可选：按需加载的大块内容
      <topic>.md                 # 如字段字典、完整枚举表、长业务流程
```

命名约定：技能目录名用 `openydt-<域或场景>`（kebab-case），与 frontmatter 的 `name` 一致。

## frontmatter 字段说明（同规范）

frontmatter 必须与现有 openydt 技能一致：

```yaml
---
name: openydt-<域>                 # 必填，与目录名一致，kebab-case
version: 1.0.0                      # 必填，语义化版本
description: "<一句话职责>。<触发词富集>。"  # 必填，见下方要求
metadata:
  requires:
    bins: ["openydt"]              # 必填，声明依赖 openydt 二进制
  cliHelp: "openydt <域> --help"   # 建议：让使用者快速看到真实子命令
---
```

- `description` 是技能能否被正确召回的关键，写法对标 Anthropic skill-creator：**一句 WHAT（本域职责 + 读/写边界）+ 一句 WHEN（典型触发场景），整体精简到约 100–150 字**，给 3–6 个有区分度的代表场景即可，**不要堆砌同义词**。
  - 为什么不堆砌：把「读写动词 × 业务名词」笛卡尔积式铺开，会与兄弟域抢同一个词（如「查费/缴费」曾被 5 个技能同时认领）、稀释判别信号、撑大常驻 token，反而降低召回准确度。判别力来自**边界**而非数量。
  - 同名跨域场景要加一句边界，例如 record 注明「历史账单/缴费记录（实时算费请用 trade 域）」。具体归属见下方「触发词去冲突」。
  - 触发是否准确，按本技能「制作步骤」用 skill-creator 的触发 eval **客观验收**，而非靠「词够不够多」的主观判断。
- `cliHelp` 填该域真实的帮助命令，便于人类核对子命令是否存在。

## 触发词去冲突（每个触发短语只归一个 owner 域）

skill-creator 的首要扣分项是「与兄弟技能触发词冲突」。openydt 是单一平台下的多域技能库，最容易撞车。规则：**每个裸触发短语只归一个 owner 域，其余域改用限定词 + 在 description 写一句边界**；读 vs 写同名场景按「主操作」归属。新技能上线前，必须与既有域的 description 做一次触发词去重。

已知裁决表（务必遵守，新增域据此对齐）：

| 触发词 | owner（独占裸词） | 其余域改法 |
| --- | --- | --- |
| 查费 / 算费 / 缴费 / 在线缴费 | billing(trade) 实时算费 | record→「查缴费记录/账单明细」；monthticket→「月票缴费/扣费记录」；park 仅「车场收费标准查询」；coupon 不认领 |
| 屏显 / 语音播报 | device（下发：推屏显/喊话/播报） | park→「查屏显内容/应显示什么」 |
| 在场车 | record（明细） | data→「实时在场数量/在场统计」 |
| 查账单 | record =「缴费记录/账单明细」 | data =「账单汇总/经营报表」 |
| 电子券 / 优惠券 | coupon（查/发/收闭环） | park→「车辆优惠券记录(只读)」；monthticket→「车场协议同步扫码」 |
| 特殊车辆类型 / specialCarTypeId / VIP分组 | monthticket（创建/查询） | list 仅「作入参引用」，不认领「创建」 |
| 泛词「查车」 | —（禁用泛词，拆成限定短语） | record「查在场车/进出记录」、monthticket「查车主身份/是否 VIP」、park「查车场信息」 |

## 正文结构约定

正文按固定骨架写，从上到下：

1. **CRITICAL 先读 shared**：正文第一句必须是醒目提示，要求开始前先 Read `../openydt-shared/SKILL.md`，不在本技能重复签名/状态码/安全规则。
2. **何时用本技能 + 意图路由**：说清楚本域负责什么，并给出「这类诉求请改用 X 域」的路由（避免技能越界）。
3. **可用命令表**：用表格列出本域命令，列为「中文名 | 命令 | 读/写 | 关键参数」。
   - `命令` 列写**真实可执行**的 `openydt <域> <use>`。
   - `读/写` 列标明读还是写；**所有写命令在表内与示例里都要标注「需 `--yes`」**。
   - `关键参数` 标必填项（用 `*` 或「必填」），数组/对象型字段说明须用 `--body` JSON 传入。
4. **业务流程**（仅当有多步链路时）：描述需要回填上一步响应的链路（如「查费 → 计费测算」「建券 → 售券 → 发券」），强调字段必须取自上一步响应、不可臆造。无强依赖链则写明「各命令为独立查询」。
5. **示例**：给 2-4 个可直接复制运行的命令，至少含一个读示例；若有写命令，**必须先给一条 `--dry-run` 预览签名请求、再给一条 `--yes` 实发**两步序列，让「先预演后执行」可直接照抄。示例卫生（硬约束）：parkCode **必须用共享基座文档化的测试车场**（`1ZS7H5PQH9` / `PTD2YBBZ`），时间参数用当前/相对时间或中性占位——**不要照抄 catalog `sampleBody` 里 2016–2019 的历史值**，否则用户复制即撞 904/911 或空结果。

## references/ 按需加载约定

- SKILL.md 主体要**短**（常驻上下文，控制 token）：**正文控制在 500 行以内，命中即拆**。把大块、低频内容下沉到 `references/<topic>.md`，并在主体里用相对链接指明「需要 X 时再 Read」。
- 适合放 references 的内容：完整字段字典 / 长枚举表（如券类型、车辆类型全集）、超过两步的完整业务流程、错误码到处置动作的详表。
- 在主体里写清触发条件，例如：「**处理建券 → 售券 → 发券完整链路前，先 Read [`references/coupon-flow.md`](references/coupon-flow.md)**」，让模型按需加载而非默认全读。

## 命令必须真实存在

- 命令表里的每条 `openydt <域> <use>` **必须真实存在**，来源只有两类，二选一核对：
  - **域一等命令**：以 `cmd/gen/<域>.go` 的真实子命令为准（这些由 codegen 生成），或运行 `openydt <域> --help` 核对。
  - **catalog 接口**：以 `catalog/catalog.json` 中 `included: true` 的 `interfaces[]` 为准，每条含 `cmd`(业务编码) / `domain` / `readwrite` / `params` / `sampleBody`，可据此推断命令名、读写属性与示例参数。
- 不要臆造命令名或参数。`included:false` 的接口（标 `excludeReason`）属于越界范围，不要为它建技能。
- 若目标接口尚无域一等命令，可在技能里用通用兜底 `openydt api <cmd> --body '{...}'` 调用（见 shared 的三层命令模型），但仍要确认该 `cmd` 在 catalog 中存在且可调用。

## 写操作标 --yes

- catalog 中 `readwrite: "write"`（或任何会改变平台状态的操作：缴费、开闸、发券、开通月票、加/移黑名单、设置车位等）一律为**写命令**。
- 写命令在命令表的「读/写」列标「写（需 `--yes`）」，并在示例中实际带上 `--yes`，必要时建议先 `--dry-run` 预览签名请求。这条与 shared 的安全规则一致，新技能不得弱化。

## 制作步骤

1. 确定域/场景与目标接口，去 `catalog/catalog.json`（`included:true`）或 `openydt <域> --help` 核对真实命令、读写属性、必填参数、`sampleBody`。
2. 新建 `skills/openydt-<名>/SKILL.md`，按上面 frontmatter 与正文骨架填写；description 富集中文触发词。
3. 命令表逐条核对真实性与读写标注：读/写须与 catalog `readwrite` **逐条一致**，写命令标 `--yes`；**只读域不得混入 write 命令**（如确有平台契约标 write 的「伪写」统计接口，须在 description/正文显式说明「该接口契约标 write，调用需 --yes」）。
4. 大块内容下沉 `references/`，主体留按需加载指引。
5. 顺带核对盲区：去 catalog 看本域有没有「有 endpoint、`included:false` 但 `direction:callable`」的接口（常见排除理由 appointment/authorize/certificate/tag 等多是「功能未一等化」而非废弃）。这类接口无专属命令，**应在正文加一句「用 `openydt api <cmd>` 调用，详见 api-explorer」的指路**，而非漏掉。
6. 自检（客观优先于主观）：命令是否都真实存在、读写标注是否与 catalog 一致、写操作是否都标 `--yes` 且示例含 `--dry-run`、是否在开头要求先读 shared、**触发词是否与既有域去重（对照「触发词去冲突」表）**。触发是否准确，**用 skill-creator 的触发 eval 客观验收**：构造正例（应召回本域）与反例（易误召回的兄弟域场景），跑 `run_loop.py` 确认本域命中、冲突域不误召回，达标再上线。

## 最小模板

把下面整段复制为新技能的 `SKILL.md` 起点，替换尖括号占位后逐项核对：

```markdown
---
name: openydt-<域>
version: 1.0.0
description: "<一句 WHAT：本域负责 X，含读/写边界>。<一句 WHEN：典型触发场景，3–6 个有区分度的代表说法，勿堆砌同义词；与兄弟域同名场景加一句边界>。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt <域> --help"
---

# openydt-<域> — <中文域名>

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**（认证 / profile / 签名 / 状态码 / 限速 / 安全规则）。未读共享基座不要执行任何命令。

## 何时用本技能

<本域负责什么>。意图路由：
- <这类诉求> → 留在本域。
- <那类诉求> → 改用 `openydt <其它域> --help`。

## 可用命令

`<use>` 为命令真实 kebab 名，调用形如 `openydt <域> <use>`。数组/对象型字段用 `--body '<json>'` 传入。

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| <查询某资源> | `openydt <域> <use-read>` | 读 | `--xxx`*（必填） |
| <修改某资源> | `openydt <域> <use-write>` | 写（需 `--yes`） | `--xxx`*、`yyyList`*（用 `--body`） |

> 标 `*` 为必填。

## 业务流程

<若有需回填上一步响应的链路，在此描述，强调字段取自上一步响应、不可臆造；否则写「各命令为独立查询，拿到必填参数即可直接调用」。>

## 示例

1) 读示例：

```bash
openydt <域> <use-read> --xxx <值>
```

2) 写示例（**写操作必须加 `--yes`**，必要时先 `--dry-run`）：

```bash
openydt <域> <use-write> --yes --body '{"xxx":"...","yyyList":[{...}]}'
```
```

## workflow（编排）型技能模板（用于跨域 / 多步业务流程）

上面的「最小模板」是**原子 / 域技能**（包装一个域的命令）。当一个**业务结果**需要把多个命令（常跨域）串成固定管道、且步骤间要回填上一步响应字段时，做成 **workflow（编排）型技能**，对标飞书 `lark-workflow-*`。

| | 原子 / 域技能 | workflow 技能 |
| --- | --- | --- |
| 触发 | API/域级意图（「查在场车」） | 业务结果（「出一份车场经营日报」「缴费后对账」） |
| 正文 | 命令表 + 独立查询 | 固定多步管道 + 步骤间数据回填 |
| 组合 | 自包含 | 组合多个域命令 |
| 命名 | `openydt-<域>` | `openydt-workflow-<场景>` |

额外约定：
- 触发写「业务结果」而非「调某接口」，触发词用结果名（经营日报 / 对账 / 催缴），**不与被组合的域技能抢词**。
- 正文用 **ASCII 管道图**画清数据流与回填字段（哪个字段取自上一步响应）。
- 含写的步骤：**先 `--dry-run` 预览、确认后再 `--yes`**；批量写注意限速（见 shared，约 4/s）。
- 仅当「跨域 / 需多步回填 / 需汇总」才独立成 workflow——单域已能在正文写全的闭环不必重复成技能。

管道图示意（放进 workflow 技能的「## 流程」节）：

```
{parkCode} ─► openydt parking get-park-on-site-car         ──► 确认在场
          ─► openydt trade get-park-fee                    ──► 取 chargeBillToken / shouldPayValue
                └─► openydt trade pay-park-fee --dry-run → 确认 → --yes   （写）
                      └─► openydt parking get-pay-bill      ──► 反查金额是否一致
```
