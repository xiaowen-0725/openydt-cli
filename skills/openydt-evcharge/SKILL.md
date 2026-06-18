---
name: openydt-evcharge
version: 1.0.0
description: "电动车充电域(evcharge)：充电站点列表/详情、充电桩(含枪口)列表、站点经营数据统计、充电订单列表/详情，全部只读。当用户要查充电站、充电桩、枪口状态、充电订单、运营商(代理商)经营数据/充电营收/电量电费、按运营商账号或手机号或站点编号查充电资产时使用。触发词：充电站、充电桩、充电枪/枪口、充电订单、运营商账号、代理商、充电营收、充电电量、电费/服务费、evcharge、plotCode、useraccount。注：充电是独立平台模块、复用同一签名鉴权；订单状态实测为 chargeStatus(充电状态)+orderState(订单状态)双维度，与对接文档的 orderStatus 英文枚举不一致(以平台实测为准)。停车(非充电)的查费/记录/月票/券请改用其它 openydt 域技能。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt evcharge --help"
---

# openydt-evcharge — 电动车充电域 (evcharge)

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**（认证 / profile / 签名 / 状态码 / 限速 / 安全规则）。未读共享基座不要执行任何命令。

## 何时用本技能

当用户需要对**电动车充电平台**做只读查询/统计时使用本技能，例如：列某运营商(代理商)名下的充电站、看某充电站详情、查充电桩及枪口实时状态、统计站点/桩的经营数据(订单数、充电电量、电费/服务费)、查充电订单列表或单笔订单详情。

意图路由：
- 要**站点**：按运营商账号/手机号列站 → `station-list`；按站点编号看单站(含 operatorName、pileCount) → `station-detail`。
- 要**充电桩/枪口** → `pile-list`(返回 `guns[]` 枪口状态)。
- 要**经营数据/营收/电量** → `station-statistics`(按站点聚合到每根桩)。
- 要**订单** → 列表用 `order-list`(按时间段，可加 plotCodes/pileId/useraccount 过滤)；单笔用 `order-detail`(按 orderNo)。
- 若要做**停车**(非充电)的查费/缴费、停车记录/在场、月票/券/设备/黑白名单等，请改用对应停车域技能 [[openydt-billing]] / [[openydt-record]] / [[openydt-park]] / [[openydt-monthticket]] / [[openydt-coupon]] / [[openydt-device]] / [[openydt-list]]。
- 充电平台若有本技能未覆盖的 cmd，用 [[openydt-api-explorer]] 以 `openydt api <cmd>` 兜底。

## 可用命令

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| 获取充电站点列表 | `openydt evcharge station-list` | 读 | useraccount / mobile(二选一)、pageNum、pageSize |
| 获取充电站点详情 | `openydt evcharge station-detail` | 读 | plotCode(必填) |
| 获取充电桩设备列表 | `openydt evcharge pile-list` | 读 | useraccount(必填)、plotCodes(数组，仅 --body)、pageNum、pageSize |
| 获取站点经营数据 | `openydt evcharge station-statistics` | 读 | useraccount(必填)、plotCodes(数组)、startDate、endDate(必填，≤30天)、分页 |
| 获取充电订单列表 | `openydt evcharge order-list` | 读 | startDate、endDate(必填，≤30天)、useraccount/plotCodes/pileId(可选过滤)、分页 |
| 获取订单详情 | `openydt evcharge order-detail` | 读 | orderNo(必填) |

> 全部为只读命令，无需 `--yes`。命令名已去掉冗余域前缀(`station-list` 而非 `evcharge-station-list`)；原始业务码 `evchargeStationList` 等仍可作 alias 使用。

### 入参约束（来自平台规则）

- `station-list`：`useraccount` 与 `mobile` **至少传一个**。
- `plotCodes` 是**站点编号数组**：只能经 `--body` 传(不是单值 flag)；不传或空数组表示不按站点过滤，默认最多 50 个。
- `startDate`/`endDate`：格式 `yyyy-MM-dd HH:mm:ss`，**间隔不能超过 30 天**(超过回 status=2 / resultCode=909「查询时间范围不能超过 30 天」)。
- 分页：`pageSize` **最大 50**(超过回业务失败「pageSize 不能大于 50」)；命令会按 catalog 文档默认值在缺省时注入 `pageNum=1` / `pageSize=10`。

## 业务流程

充电域均为只读查询，典型链路是**先定位运营商/站点，再下钻桩与订单**，强调用前序响应字段作为后续入参：

1. **定位站点**：用 `station-list`(传 `--useraccount` 或 `--mobile`)拿到站点的 `plotCode`/`plotName`；或已知 `plotCode` 直接 `station-detail` 看详情(返回 `operatorName`、`pileCount`、经纬度、运行状态)。
2. **下钻充电桩**：把第 1 步的 `plotCode` 放进 `pile-list` 的 `--body` 的 `plotCodes` 数组(配 `--useraccount`)，得到各桩 `pileId`、`enabled`、`runningStatus` 与 `guns[]` 枪口状态。
3. **经营统计**：`station-statistics`(useraccount + plotCodes + 时间段)按站点聚合到每根桩，给出 `orderCount`/`chargeTotalTimes`/`totalPower`/`totalAmount`/`totalElectricFee`/`totalServiceFee`。
4. **订单分析**：`order-list`(时间段必填，可用第 1/2 步的 `plotCodes`、`pileId`、`useraccount` 过滤)列订单并拿到 `orderNo`；需要单笔细节再用 `order-detail --order-no <orderNo>`。

> 字段传递要点：`station-list`/`station-detail` 的 `plotCode` → `pile-list`/`station-statistics`/`order-list` 的 `plotCodes` 数组(经 --body)；`pile-list` 的 `pileId` → `order-list` 的 `--pile-id` 过滤；`order-list` 的 `orderNo` → `order-detail` 的 `--order-no`。

## 响应字段/枚举速查（prod 实测，与对接文档有出入）

> ⚠️ 以下为正式环境实测结果；对接文档写的单一 `orderStatus` 英文枚举(PAID/CHARGING…)在 prod **未出现**。以平台实测为准；若平台后续澄清，以平台为准。

- **订单**(`order-list`/`order-detail`)有**两个状态维度**：
  - `chargeStatus` / `chargeStatusName` —— 充电状态：`9002`=充电中，`9003`=已完成。
  - `orderState` / `orderStatusName` —— 订单状态：`1`=下单，`3`=已支付。
  - 金额字段(`totalAmount`/`electricFee`/`serviceFee`)单位「元」，电量 `chargingPower` 单位「kWh」，`chargingDuration` 形如「12分」「1分12秒」。
- **充电桩**(`pile-list`)：`enabled` 0停用/1启用；桩 `runningStatus`(原值)0运行/1离线；`plotCode` 为所属站点；`guns[].status` 实测 **0=离线 / 2=空闲 / 4=已插枪**(对接文档仅给 2=空闲)。
- **站点经营**(`station-statistics`)：`piles[]` 含 `chargeTotalTimes`(充电次数，文档未列)及 orderCount/totalDuration/totalPower/totalAmount/totalElectricFee/totalServiceFee；recordList 项含 `stationName`。
- **站点**(`station-list`/`station-detail`)：`runningStatus` 1运营中/0已停业；`station-detail` 另含 `operatorName`、`pileCount`、`lng`/`lat`、`region`。

## 错误自愈速查

| 现象 | 含义 | 恢复动作 |
| --- | --- | --- |
| resultCode 909「查询时间范围不能超过 30 天」 | `order-list`/`station-statistics` 时间窗 >30 天 | 收窄 `startDate`/`endDate` 到 ≤30 天 |
| 业务失败「pageSize 不能大于 50」 | 分页超限 | `--page-size` ≤ 50 |
| `station-list` 报参数错误 | `useraccount` 与 `mobile` 都没传 | 至少给一个 |
| `plotCodes` 没生效 | 当成单值 flag 传了 | 用 `--body '{"plotCodes":["C3CMHJX"]}'`(数组只能经 body) |
| 返回空 / nodata | 多为时间窗无业务，或在 **test 环境查充电**(充电数据当前在 prod 验证) | 换时间窗，或 `--env prod` + 对应 prod 凭据核实，非接口故障 |

> 通用 status/resultCode、退出码、限速重试、签名版本见 [[openydt-shared]]。充电复用同一签名鉴权(默认 v2)。

## 示例

> 充电业务数据目前在 **prod** 验证(test 环境可能无充电数据)；如需查 prod，加 `--env prod` 并配 prod 凭据。下列为只读命令，安全。样例站点 `C3CMHJX`、运营商 `DLS001` 为实测值。

```bash
# 1) 按运营商账号列充电站(二选一：--useraccount 或 --mobile)
openydt evcharge station-list --useraccount DLS001 --page-size 10

# 2) 看单站详情(含运营商、桩数、运行状态)
openydt evcharge station-detail --plot-code C3CMHJX

# 3) 列站内充电桩与枪口(plotCodes 是数组，必须经 --body)
openydt evcharge pile-list --body '{"useraccount":"DLS001","plotCodes":["C3CMHJX"],"pageNum":1,"pageSize":10}'

# 4) 站点经营数据(时间段 ≤30 天)
openydt evcharge station-statistics --body '{"useraccount":"DLS001","plotCodes":["C3CMHJX"],"startDate":"2026-06-01 00:00:00","endDate":"2026-06-18 23:59:59","pageNum":1,"pageSize":10}'

# 5) 充电订单列表(时间段必填，可按站点/桩/运营商过滤)，拿到 orderNo
openydt evcharge order-list --body '{"plotCodes":["C3CMHJX"],"startDate":"2026-06-01 00:00:00","endDate":"2026-06-18 23:59:59","pageNum":1,"pageSize":10}'

# 6) 单笔订单详情(用第 5 步响应里的 orderNo)
openydt evcharge order-detail --order-no 26061820675194016111704626880659
```
