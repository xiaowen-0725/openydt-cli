---
name: openydt-skill-maker
version: 1.0.0
description: "为艾科智泊停车开放平台 CLI(openydt) 创建/编写自定义 Skill 的元技能(对标飞书 lark-skill-maker)。当用户要把 openydt 的某个接口或某条业务流程封装成可复用的 Skill、新建一个 openydt 域技能、规范 SKILL.md 的目录结构与 frontmatter、抽取 catalog 命令清单、给写操作加 --yes、或想知道 openydt skill 怎么写时使用。触发词：新建 openydt skill、写一个 openydt 技能、封装 openydt 接口、做一个停车域技能、openydt skill maker、技能模板、SKILL.md 规范、frontmatter 怎么写、命令清单怎么列、把这个接口做成技能、对标 lark-skill-maker、停车开放平台技能、skill 脚手架、技能目录结构。"
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

- `description` 是技能能否被正确召回的关键，**必须富含中文触发词**：先用一句话说清楚职责（覆盖哪个域、做读还是写），再罗列用户可能的口语说法（如「查车场、空车位、查费、发券、加黑名单、开月票」等同义/近义词），把读写动词和业务名词都铺开。参考 `openydt-park` / `openydt-coupon` 的 description 密度。
- `cliHelp` 填该域真实的帮助命令，便于人类核对子命令是否存在。

## 正文结构约定

正文按固定骨架写，从上到下：

1. **CRITICAL 先读 shared**：正文第一句必须是醒目提示，要求开始前先 Read `../openydt-shared/SKILL.md`，不在本技能重复签名/状态码/安全规则。
2. **何时用本技能 + 意图路由**：说清楚本域负责什么，并给出「这类诉求请改用 X 域」的路由（避免技能越界）。
3. **可用命令表**：用表格列出本域命令，列为「中文名 | 命令 | 读/写 | 关键参数」。
   - `命令` 列写**真实可执行**的 `openydt <域> <use>`。
   - `读/写` 列标明读还是写；**所有写命令在表内与示例里都要标注「需 `--yes`」**。
   - `关键参数` 标必填项（用 `*` 或「必填」），数组/对象型字段说明须用 `--body` JSON 传入。
4. **业务流程**（仅当有多步链路时）：描述需要回填上一步响应的链路（如「查费 → 计费测算」「建券 → 售券 → 发券」），强调字段必须取自上一步响应、不可臆造。无强依赖链则写明「各命令为独立查询」。
5. **示例**：给 2-4 个可直接复制运行的命令，至少包含一个读示例和（若有写命令）一个带 `--yes` 的写示例；参数尽量取自 catalog 的 `sampleBody` 或共享基座的测试车场。

## references/ 按需加载约定

- SKILL.md 主体要**短**（常驻上下文，控制 token）。把大块、低频内容下沉到 `references/<topic>.md`，并在主体里用相对链接指明「需要 X 时再 Read」。
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
3. 命令表逐条核对真实性与读写标注，写命令标 `--yes`。
4. 大块内容下沉 `references/`，主体留按需加载指引。
5. 自检：命令是否都真实存在、写操作是否都标 `--yes`、是否在开头要求先读 shared、description 触发词是否充分。

## 最小模板

把下面整段复制为新技能的 `SKILL.md` 起点，替换尖括号占位后逐项核对：

```markdown
---
name: openydt-<域>
version: 1.0.0
description: "<一句话职责：本域负责 X，含读/写>。触发词：<列尽用户可能说的中文同义词，如 查X、看X、新建X、修改X、删除X、X列表、X编码……>。"
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
