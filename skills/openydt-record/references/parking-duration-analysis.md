# 基于离场明细复算停车时长

仅在用户要求基于 `get-car-out-list` 明细重算、核验停车时长，或生成时长分布/分位数时执行本流程。直接读取平台聚合统计走 [[openydt-data]]。

## 1. 取得一次性快照

按离场时间查询时，`get-car-out-list` 单次时间窗最多 1 天。月度分析按天执行 `--all-pages --out`，同一车场、接口和时间窗只导出一次；分页、重试与 NDJSON 规则以 [[openydt-shared]] 为准。

完成标准：每个日期都有一份 CLI 显示 `complete` 的 NDJSON，且最终记录数已保存。

```bash
openydt parking get-car-out-list \
  --park-code PTD2YBBZ \
  --leave-start-time 20260601000000 \
  --leave-end-time 20260601235959 \
  --all-pages --out 2026-06-01-out.ndjson
```

## 2. 形成经营离场样本并计算

从当前 `SKILL.md` 的绝对路径定位同目录下的 `scripts/parking_duration.py`；不要假设工作目录。脚本读取 CLI 领域语义，先按 `parkingCode` 保留 `leaveTime` 最晚的事件，再排除逻辑闭环、遗留记录和人工盘点，最后计算经营离场样本；可一次读取多份 NDJSON。

```bash
python3 <openydt-record-skill-dir>/scripts/parking_duration.py \
  2026-06-*-out.ndjson \
  --anomalies-out duration-anomalies.ndjson \
  > duration-summary.json
```

完成标准：脚本退出码为 0，汇总与异常文件均已生成，并满足三组守恒：

```text
rawDepartureEvents = deduplicatedDepartureEvents + duplicateDepartureEvents
deduplicatedDepartureEvents = operatingDepartureRecords + excludedBusinessRecords
durationCandidateRecords = validDurationRecords + excludedDurationRecords
```

## 3. 使用样本覆盖率

只有 `enterTime`、`leaveTime` 同时存在且合法的记录进入停车时长样本，时长为 `leaveTime - enterTime`。

- 缺 `enterTime`：归入 `missingEnterTime`，保持为缺失样本；`stoppingTime` 不作为自动替代值。
- 缺 `leaveTime`、时间无法解析、离场早于进场：分别归入对应异常。
- 原始闭环规模使用 `rawDepartureEvents`；经营车流使用 `operatingDepartureRecords`；时长均值、分布和分位数只使用 `validDurationRecords`。

完成标准：报告同时列出 `durationCoverageRate` 与每类异常数，并满足：

```text
excludedDurationRecords = missingEnterTime + missingLeaveTime
                        + invalidEnterTime + invalidLeaveTime
                        + negativeDuration
```

## 4. 写入报告

报告必须区分“离场记录事件”“经营离场样本”和“有效时长样本”，并披露每类业务排除数。有效样本为 0 时，结论写“停车时长不可计算”，而不是 0 小时。脚本以小时为单位，默认分布区间为左闭右开 `0-1、1-4、4-7、7-10、10-12、12+`，p50/p75/p90/p95 使用 nearest-rank；字段值以 `duration-summary.json` 为准，命令参数以脚本 `--help` 为准。

完成标准：任何时长结论都能追溯到输入文件、样本覆盖率和异常明细。
