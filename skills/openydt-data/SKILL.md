---
name: openydt-data
version: 1.0.0
description: "宇视智慧停车开放平台数据分析域 openydt data：车场缴费账单、账单汇总、车流量、车牌top分布、停车场实时数据、当天出车/交易次数、每分钟车流量、车位使用情况、停车时长分析等只读数据统计与报表。触发词：数据分析/统计/报表/查账单/缴费账单/账单汇总/车流量/车流统计/车牌top/车牌分布/实时数据/在场车/出车次数/交易次数/每分钟车流/车位使用/车位使用率/车位占用/热力图/echart/停车时长/时长分析/停车时长分布"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt data --help"
---

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**（认证 / profile / 签名 / 状态码 / 限速 / 安全规则），否则命令会因缺少凭证或签名而失败。

## 何时用本技能

当用户需要对停车场做**只读数据分析与报表统计**时使用本技能，例如：查某车场的缴费账单 / 账单汇总、看车流量与车牌 top 分布、读停车场实时在场数据、统计当天出车次数与交易次数、按分钟看车流曲线、分析车位使用情况（含 echart 热力图）、做停车时长分布分析。

意图路由：
- 要**账单/缴费明细** → `get-bill-summary`；要**账单汇总指标** → `get-park-bill`。
- 要**车流量曲线/趋势** → 整段每分钟用 `get-traffic-flow`；车牌 top 分布用 `get-car-traffic-flow-analysis`。
- 要**实时/当天**数据 → 实时在场用 `get-real-time-park-info`；当天出车与交易次数用 `get-realtime-leave-and-charge-num`。
- 要**车位使用情况** → 列表数据用 `parking-place-used`；绘 echart 热力图用 `parking-place-used-for-echart`。
- 要**停车时长分布** → `parking-time-analyse`。
- 若要发起缴费、查车场配置、月票/券、设备、黑白名单、访客等，不属于本域，请改用对应技能（trade / parking / ticket / coupon / device / blacklist / redlist / visitor 等）。

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

> 字段传递要点：上游车场域响应中的 `parkCode` → 本域全部命令的 `--park-code`；分页查询响应里的总数/分页信息 → 调整下一次的 `--page-num` / `--page-size`；多场分析时把各车场 `parkCode` 收集进 `--body` 的 `parkCodeList` 数组。

## 示例

```bash
# 1) 查询某车场账单汇总（按年维度，只读，无需 --yes）
openydt data get-park-bill \
  --park-code 2KKN6112 --dimension 2 \
  --start-time "2019-08-17 00:00:00" --end-time "2019-08-17 23:59:59" \
  --page-num 1 --page-size 10

# 2) 获取车流量车牌 top 分布（多场，用 --body 传 parkCodeList 数组，只读）
openydt data get-car-traffic-flow-analysis \
  --body '{"parkCodeList":["765OB49GJ","765MQK2TX"]}'

# 3) 获取车场停车时长分析数据（write 操作，必须带 --yes）
openydt data parking-time-analyse \
  --park-code 2KKN885S \
  --start-date 20190910000000 --end-date 20190910235959 \
  --vip-type 1 --hour-area "0-1,1-4,4-7,7-10,10-12,12-0" \
  --yes
```
