---
name: openydt-flow-park-access
version: 1.0.2
description: "艾科智泊开放平台「车辆进出场」作业流程 SOP（进场 / 出场端到端编排）。当用户想让一辆车完整进场或出场、模拟真实进出场物理流程、跑通『补录/抓拍进场 → 校正』或『查费缴清 → 出口抓拍校正 → 自动放行/开闸』整条链路，或问『车怎么进场 / 怎么出场 / 进出场流程 / 进出场 SOP / 怎么把车弄进(出)车场 / 模拟一辆车进出 / 校正了车为什么没出场 / 出口怎么开闸放行』, 或在进出场过程中某一步卡壳/排错(如抓拍或校正后车却没进/出场、抓拍报 908、correct 成功但车不离场、不知下一步如何接续)时使用。本技能是跨域编排层，串联 parking(补录/校正/盘点) + device(抓拍/开闸) + trade(查费/缴费) 多条命令，并讲清跨命令的硬约束(配对出口、抓拍设备、通道放行模式自动/人工、缴费先于放行、908 偶发、令牌时效)与失败处理。边界：只查单条记录 / 在场车 / 锁车 / 调用单个命令，请直接用对应域技能 openydt-record / openydt-device / openydt-billing，不要走本流程技能。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt --help"
---

# openydt-flow-park-access — 车辆进出场作业流程 (SOP)

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**，了解认证 / profile / 签名 / 状态码 / 限速 / 安全等通用约定。本技能只讲“怎么把一辆车按真实物理流程弄进 / 弄出车场”的编排，**不复述各命令的参数表**——具体入参出参见各域技能。

## 何时用本技能（与域技能的分工）

- **用本技能**：要把一辆车**端到端**进场或出场、模拟真实进出场、跑通整条链路、不确定先调哪个命令、或某一步报错不知如何接续。
- **用域技能**（不要走本流程）：只查一条记录 / 在场车 / 锁车 / 只想调用某个单命令 → `[[openydt-record]]`（停车记录域）、`[[openydt-device]]`（设备域）、`[[openydt-billing]]`（缴费域）。

> 安全前提：进出场涉及**写操作**（补录、抓拍、校正、缴费、盘点），**仅在 test 环境演练**；每个写命令先 `--dry-run` 预览签名请求，确认后再加 `--yes`。详见 openydt-shared 的安全约定。

## 两个必须先理解的概念

**1. 进场的两条路本质不同：**

| | (a) 补录进场 | (b) 抓拍进场（模拟真实物理流程） |
| --- | --- | --- |
| 命令 | `supplement-parking-record-in` | `channel-snap` →（车牌不对则）`correct-car-on-channel` |
| 性质 | **强制进场**，车一定进 | **不保证进场** |
| 抓拍流水 | 不产生抓拍记录 | 产生（本就是模拟抓拍） |
| 适用 | 车已实际进场但平台漏报，需要补一条进场记录 | 演练真实进场物理过程：抓拍 → 识别 → 校正 |

**2. 通道放行模式（最反直觉的一点）：**
抓拍进场里 `correct-car-on-channel` **校正成功 ≠ 车一定进场**。是否真正进场取决于**该通道的放行模式**——若通道设了「严禁临时车进场」，校正后车也**不会**进车场。这是设计如此：抓拍链路严格按真实物理放行逻辑模拟。所以校正后**务必复核是否真的在场**（见下），不要假定校正完就进场了。

## 进场流程

先判断用哪条路：**只是要让记录里有这辆车（强制） → (a) 补录；要演练真实抓拍识别校正过程 → (b) 抓拍。**

### (a) 补录进场（强制，车必进）

1. （可选）`openydt parking check-channel-exist-car` 确认通道当前是否已有车，避免重复补录。
2. `openydt parking supplement-parking-record-in --yes`：传 `parkCode`、`carCode`、`enterTime`、`channelCode`、`carCodeType`、`carCodeColor`、`parkOrArea`。响应返回新生成的 `parkingCode`。
   > `recordType` / `enterCarType` 未传时 CLI 会按平台文档默认值自动补 `1`，无需手填。
3. 完成。补录是强制进场，记录立即生效。参数细节见 `[[openydt-record]]`。

### (b) 抓拍进场（模拟真实物理流程，不保证进场）

1. **抓拍**：`openydt device channel-snap --yes`（传 `parkCode`、`channelCode`），在进场通道触发一次抓拍。
   > 前提：该通道**有抓拍设备**；否则返回 `resultCode=908 找不到设备`，换有设备的通道。
2. **判断车牌**：看抓拍出的车牌是否就是目标车牌。
   - 是目标车牌 → 进入第 3 步前的放行判断。
   - 不是 → **校正**：`openydt parking correct-car-on-channel --yes`（传 `parkCode`、`channelCode`、`newCarNo`=目标车牌、`correctTime`）。
     > 若报「会话已过期」，说明该通道当前没有可校正的抓拍会话，需**先成功 `channel-snap`** 再校正。
3. **复核是否真的进场**（关键）：用 `openydt parking get-park-on-site-car`（传 `parkCodeList` + `enterTimeFrom`/`enterTimeTo`，**时间范围必填**）查在场车。
   - 在场 → 进场成功。
   - 不在场 → 多半是该通道**放行模式禁止临时车进场**所致，校正不会让它进场；改用 (a) 补录，或换放行模式允许的通道。

## 出场流程

> **关键前提（最易踩坑）：`correct-car-on-channel` 在出口 ≠ 放车出场。** 它只是「模拟出口设备抓到车、触发出场处理流程」。车是否真离场取决于 **放行模式 × 缴费状态**：
> - **自动放行 + 已缴清** → 系统自动放行，车出场（后端约数秒内闭环到 `leaveType=1`）。
> - **人工放行**（或仍欠费）→ 车被拦住，`correct` 仍回 `status=1` 但**车不出场**，需另外开闸（见第 4 步）。
> 所以**先把费缴清，再抓拍校正**；correct 返回 `status=1` 不代表车走了，**务必用 `get-car-out-list`/`get-park-on-site-car` 复核**。

1. **查费**：`openydt trade get-park-fee`（传 `carCode` + `parkCode`）。
   - 看 `data.shouldPayValue`（**单位：元**，`1` 即 1.00 元，不是 1 分）；取 `parkingCode`、`chargeDate`、`otherAttr` 供缴费回传。
   > 查费后 **10 分钟内**须完成缴费，令牌/账单否则失效。
2. **缴清费（先问后做）**：`shouldPayValue>0` 时**必须先缴清**，否则自动放行也不会放车。
   - **先询问用户「是否需要缴费？用什么支付方式？」**——缴费是真实写操作，不要默默执行。
   - `openydt trade pay-park-fee --yes`，回传 `parkingCode`、`chargeDate`、`actPayCharge`（缴清则 = `shouldPayValue`，单位元）、`payOrigin`、`paymentMode`、唯一 `billCode`。带券 0 元结算（actPayCharge=0 + 原样回传 otherAttr）见 `[[openydt-billing]]`。
   - 缴清后复核 `get-park-fee` 的 `shouldPayValue=0`。
3. **出口抓拍 + 校正**：
   - `openydt device channel-snap --yes`（`parkCode`、`channelCode`=出口通道）→ 抓出「未识别NNNNN」。
   - `openydt parking correct-car-on-channel --yes`（`newCarNo`=目标车 + `correctTime`）。
   > 前提：该出口**有抓拍设备**，且通常需是**进场通道配对的出口**；否则 `channel-snap` 报 `908 找不到设备`。
   > ⚠️ correct 偶发 `resultCode=908 会话已过期`：这是**间歇性服务端故障**（同一出口同一命令时成时败，疑似网关/后端会话一过性），**不是 TTL、与快慢无关**。**遇到就重试或换时机**，别据此断言"出场走不通"。
4. **复核 + 按需开闸**：
   - `openydt parking get-car-out-list`（`carNo` + `leaveStartTime`/`leaveEndTime`）查到 `leaveType=1` 且 `get-park-on-site-car` count→0 → 自动放行场，**已闭环，结束**。
   - 若 correct 成功但车仍在场 → 多半是**人工放行**：调 `openydt device op-gate --yes`（`parkCode`、`channelCode`、`opType=0` 开、`operator`、`operateTime`）或云车场 `openydt device cloud-open-gate --yes`（`channelId` 数字ID、`opType=0`）放车，再复核出场记录。

> **关于盘点离场 `inventory-car`**：⚠️ **对开放平台抓拍/补录进场的车是 no-op**（返回 status=1 但车不离场、getInventoryRecord 无记录），**不是可靠的出场兜底**。物理出场优先走上面的「缴清 → 出口 snap+correct →（人工放行则开闸）」。

## 跨命令硬约束与失败速查

| 现象 / 约束 | 含义 | 处理 |
| --- | --- | --- |
| `channel-snap` 报 `resultCode=908 找不到设备` | 该通道没有抓拍设备 | 换有抓拍设备的通道 |
| `correct-car-on-channel` 报 `908 会话已过期` | **间歇性服务端故障**（非 TTL、与快慢无关，时成时败） | **重试 / 换时机**；snap 成功后才有会话可校正 |
| 抓拍进场校正后车不在场 | 通道放行模式禁止临时车进场 | 改补录进场，或换放行模式允许的通道 |
| 出口 correct 回 `status=1` 但车不出场 | **设计如此**：仍欠费、或出口人工放行 | 先缴清费；人工放行场用 `op-gate`/`cloud-open-gate` 开闸 |
| `inventory-car` 返回成功但车不离场 | 对抓拍/补录进场车 **no-op**（匹配不上） | 别拿它当出场兜底；走「缴清+出口snap+correct」 |
| 金额理解 | `shouldPayValue`/`actPayCharge` 等单位是**元** | 别把 `1` 当 1 分 |
| 查费令牌 | 查费后 10 分钟内须缴费 | 超时重新查费 |

## 命令归属（参数见各域技能）

- 进场补录 / 校正 / 在场复核 / 盘点 → `[[openydt-record]]`
- 抓拍 `channel-snap` → `[[openydt-device]]`
- 查费 `get-park-fee` / 缴费 `pay-park-fee` → `[[openydt-billing]]`

> 进出场是跨域**流程**；单条命令的入参、出参、枚举值一律以上述域技能为准，本技能只负责把它们按正确顺序和约束串起来。
