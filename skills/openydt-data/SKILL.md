---
name: openydt-data
version: 1.0.3
description: "数据分析域(data)：缴费账单与账单汇总、车流量曲线、车牌 top 分布、实时在场统计、当天出车/交易次数、车位使用(含 echart 热力图)、停车时长分布等只读统计报表。当用户要做车场经营报表/趋势分析/数据体检(聚合统计而非单车明细)时使用。注：车位/时长/echart 三个统计接口平台契约标 write、调用需 --yes；单车明细/在场车辆见 parking 域(openydt-record)。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt data --help"
---

# openydt-data — 数据分析域 (data)

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**（认证 / profile / 签名 / 状态码 / 限速 / 安全规则）。未读共享基座不要执行任何命令。

## 何时用本技能

当用户需要对停车场做**只读数据分析与报表统计**时使用本技能，例如：查某车场的缴费账单 / 账单汇总、看车流量与车牌 top 分布、读停车场实时在场数据、统计当天出车次数与交易次数、按分钟看车流曲线、分析车位使用情况（含 echart 热力图）、做停车时长分布分析。

意图路由：
- 要**账单/缴费明细** → `get-bill-summary`；要**账单汇总指标** → `get-park-bill`。
- 要**车流量曲线/趋势** → 整段每分钟用 `get-traffic-flow`；车牌 top 分布用 `get-car-traffic-flow-analysis`。
- 要**实时/当天**数据 → 实时在场用 `get-real-time-park-info`；当天出车与交易次数用 `get-realtime-leave-and-charge-num`。
- 要**车位使用情况** → 列表数据用 `parking-place-used`；绘 echart 热力图用 `parking-place-used-for-echart`。
- 要**停车时长分布** → `parking-time-analyse`。
- 若要发起缴费、查车场配置，请改用 [[openydt-billing]]（trade 缴费）或 [[openydt-park]]（车场信息）。
- 若要单车明细/在场车辆/进出记录，请改用 [[openydt-record]]（parking 记录/在场）。
- 月票/券/设备/黑白名单/访客等，请改用 [[openydt-monthticket]] / [[openydt-coupon]] / [[openydt-device]] / [[openydt-list]]。

## 可用命令

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| 获取某停车场缴费账单信息 | `openydt data get-bill-summary` | 读 | parkCode, dimension(0天/1月/2年/3自定义), startTime, endTime, pageNum, pageSize |
| 获取车流量车牌top分布 | `openydt data get-car-traffic-flow-analysis` | 读 | parkCodeList(停车场编码列表 JSONArray) |
| 查询某车场账单汇总信息 | `openydt data get-park-bill` | 读 | parkCode, dimension(0天/1月/2年), startTime, endTime, pageNum, pageSize |
| 获取停车场实时数据 | `openydt data get-real-time-park-info` | 读 | parkCode |
| 获取车场当天的出车次数、交易次数 | `openydt data get-realtime-leave-and-charge-num` | 读 | parkCode |
| 获取单个停车场每分钟车流量情况 | `openydt data get-traffic-flow` | 读 | parkCode, startTime, endTime(间隔≤1天，格式 yyyy-MM-dd HH:mm) |
| 获取车位使用情况数据 | `openydt data parking-place-used` | 写(需 `--yes`) | parkCode, startDate, endDate(yyyyMMddHHmmss), minuteInterval(10 或 240) |
| 获取车位使用情况数据(echart热力图) | `openydt data parking-place-used-for-echart` | 写(需 `--yes`) | parkCode, startDate, endDate(yyyyMMddHHmmss), minuteInterval(10 或 240) |
| 获取车场停车时长分析数据 | `openydt data parking-time-analyse` | 写(需 `--yes`) | parkCode, startDate, endDate(yyyyMMddHHmmss), vipType(1临时车/2月租车), hourArea |

> 说明：`parking-place-used`、`parking-place-used-for-echart`、`parking-time-analyse` 在平台契约里标记为 write，调用时**必须带 `--yes`** 确认，否则会被写操作确认拦截。其余命令均为只读，无需 `--yes`。

## 业务流程

数据分析域均为**车场维度的只读/统计查询**，核心入参是 `parkCode`（停车场编号/编码）。典型分析链路如下，强调**用前序命令响应里的字段作为后续命令入参**：

1. **确定目标车场 parkCode**：`parkCode` 来自车场域命令（如 `openydt park`/`openydt parking` 列表/查询接口）的响应字段，或来自用户/上游系统。**先拿到 parkCode，再把它填入本域所有命令的 `--park-code`。**
2. **快速体检（实时面）**：用 `get-real-time-park-info`（实时在场/车位）+ `get-realtime-leave-and-charge-num`（当天出车次数、交易次数）拿到当天概览，确认车场有数据、编号正确。
3. **车流分析**：
   - 单场分钟级曲线：`get-traffic-flow`，注意 `startTime`/`endTime` 间隔不允许超过一天（格式 `yyyy-MM-dd HH:mm`）。
   - 多场车牌 top 分布：`get-car-traffic-flow-analysis`，把多个车场的 `parkCode` 组成 `parkCodeList` 数组传入。
4. **账单分析**：先用 `get-park-bill` 看账单汇总指标（按 dimension 选天/月/年），需要逐条明细时再用 `get-bill-summary`（dimension 支持 3 自定义时间段），两者均分页（`pageNum`/`pageSize`，pageSize 最多 1000 条）。
5. **车位与时长分析（写确认，需 `--yes`）**：
   - 车位使用情况：`parking-place-used`（列表）/ `parking-place-used-for-echart`（热力图坐标数据），`minuteInterval` 仅支持 10 或 240。
   - 停车时长分布：`parking-time-analyse`，按 `vipType`（1 临时车 / 2 月租车）和 `hourArea` 时长区间（形如 `0-1,1-4,4-7,7-10,10-12,12-0`）切分。
   - 上述命令的时间参数用 `yyyyMMddHHmmss` 紧凑格式，与第 3/4 步的 `yyyy-MM-dd` 格式不同，注意区分。

> 字段传递要点：上游车场域响应中的 `parkCode` → 本域全部命令的 `--park-code`；分页查询需要全量时使用 `--all-pages --out <file>.ndjson`；多场分析时把各车场 `parkCode` 收集进 `--body` 的 `parkCodeList` 数组。

> 📖 **结果解读**：`get-park-bill`/`get-bill-summary` 返回的金额字段单位「元」；**test 环境多数统计接口返回 nodata 属正常，不等于车场无数据**（换 prod 或确认时间窗有业务）。三层判读/空结果见 [[openydt-shared]] 的 references/result-reading-sop.md。

### 数据分析硬规则

- 需要账单等分页明细时用 `--all-pages --out <file>.ndjson` 一次性导出；同一车场、接口、时间窗的明细只获取一次，后续指标复用本地文件。
- 涉及时长、分位数、中位数、分布、跨表关联、去重、金额情景模拟等计算，必须编写并实际运行代码，不得依靠自然语言心算。
- 计算产物至少保留：脚本、输入文件或文件哈希、输入条数、过滤/异常条数、有效条数和关键指标校验结果，以便复跑和审计。

## 错误自愈速查（统计）

| 现象 | 含义 | 恢复动作 |
| --- | --- | --- |
| `get-traffic-flow` 报错/空 | startTime/endTime 间隔 >1 天 或格式非 `yyyy-MM-dd HH:mm` | 缩到 ≤1 天、改对格式 |
| `parking-place-used*` 报错 | `minuteInterval` 非 10/240 | 改为 10 或 240 |
| 返回 nodata | test 环境常态 / 时间窗无业务 | 换时间窗或 prod 核实，非接口故障 |

> 通用码/退出码/重试见 [[openydt-shared]]；金额单位=元、0 条≠无见其 references/result-reading-sop.md。

## 示例

> 示例 parkCode/时间为文档化测试值（仅 test）；照抄 catalog 历史 sampleBody 会撞无效车场/过期时间窗。

```bash
# 1) 查询某车场账单汇总（按年维度，只读，无需 --yes）
openydt data get-park-bill \
  --park-code 1ZS7H5PQH9 --dimension 2 \
  --start-time "2026-06-01 00:00:00" --end-time "2026-06-01 23:59:59" \
  --page-num 1 --page-size 10

# 2) 获取车流量车牌 top 分布（多场，用 --body 传 parkCodeList 数组，只读）
openydt data get-car-traffic-flow-analysis \
  --body '{"parkCodeList":["1ZS7H5PQH9","PTD2YBBZ"]}'

# 3) 获取车场停车时长分析数据（write 操作，先 --dry-run 预览再 --yes）
openydt data parking-time-analyse \
  --park-code PTD2YBBZ \
  --start-date 20260601000000 --end-date 20260601235959 \
  --vip-type 1 --hour-area "0-1,1-4,4-7,7-10,10-12,12-0" \
  --dry-run
# 确认无误后去掉 --dry-run 加 --yes 实发
```
