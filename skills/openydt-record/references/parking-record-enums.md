# 离场事件与放行类型分析

用于经营车流、逃费、异常放行、遥控开闸或 `leaveType` 分布分析。枚举名称和事件语义由 CLI schema 维护，本文件只规定取值步骤与完成标准。

## 1. 读取字段绑定

先查询产生数据的原接口，不以旧开发平台页面中的 0-3 简表为完整字典。

```bash
openydt schema getCarOutList --json > getCarOutList-schema.json
```

从 `responseEnums[]` 读取名称，从 `domainSemantics` 读取记录集、去重和统计处理。完成标准：目标字段、原始 code、名称和语义来自同一次 schema 输出。

## 2. 先形成经营离场样本

对 `getCarOutList` 快照运行随附 `parking_duration.py`，由脚本执行 schema 中的规则：按 `parkingCode` 保留 `leaveTime` 最晚事件，再将每条记录归入经营样本或业务排除。完成标准：

```text
rawDepartureEvents = deduplicatedDepartureEvents + duplicateDepartureEvents
deduplicatedDepartureEvents = operatingDepartureRecords + excludedBusinessRecords
```

未知或缺失类型保留原值并进入业务排除，不归并到任何已知类型。

## 3. 按业务意图取数

- 经营车流、停车时长：使用 `operatingDepartureRecords` / `validDurationRecords`。
- 逃费：使用脚本的 `escapeRecords`，业务口径为 `leaveType=6（可疑跟车）`；当前没有独立逃费查询接口。
- 欠费放行：按 schema 对应类型单列，属于欠费业务，不并入逃费。
- 异常放行：从离场记录全集按对应类型统计；异常离场独立流水只补充原因和过程，不作为离场总量来源。
- 遥控开闸：离场记录中的对应类型表示车辆已离场且由工作人员使用实体遥控器放行；非系统开闸流水只表示操作发生，不保证车辆离场。
- 盘点、逻辑闭环和遗留类型：按 `domainSemantics` 单列业务排除。

## 4. 写入报告

报告按“code + 名称”展示经营样本分布，并同时披露原始事件数、重复事件数、业务排除数、未知类型数。逃费结论注明“可疑跟车”口径，开闸操作流水注明“不代表离场”。

完成标准：每个占比都能回算到经营离场样本，且所有原始事件都能通过守恒关系回算。
