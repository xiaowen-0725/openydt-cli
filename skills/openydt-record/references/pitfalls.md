# 停车记录域 字段易错点

### 字段易错点（实测踩坑）

- **`get-car-out-list` 不同于 `get-car-in-list`**：出场查询用 **`carNo`（单数 String）**，时间用 **`leaveStartTime`/`leaveEndTime`**（出场时段，≤1 天）或 `enterTimeFrom`/`enterTimeTo`（进场时段，≤1 个月）——**两组时间至少传一组，否则报「…不能同时为空」**；`pageSize` 上限 100。**不要照抄 `get-car-in-list` 的 `carNoArray`/`startTime`/`endTime`**（那是进场查询专用，会报参数错误）。
- **`get-park-on-site-car` 的 `enterTimeFrom`/`enterTimeTo` 必填**：不传时间范围会返回 0 条（易误判“无在场车”），起止间隔有上限。
- **判断“是否离场”不要只看 `get-park-detail`**：车离场后 detail 有时仍回 `status=1`、而 `get-park-detail-ignore-status` 又查不到，两者可能不一致。判断在场/离场以 `get-car-out-list`（已出场）或 `get-park-on-site-car`（仍在场）为准。
- **`scan-channel-code-in-out` 外层 `status=1` 不等于真的进出成功**：当通道实际无车（如补录车）时，外层回 `status=1`，但 `data.code` 为非 0（如 `8 当前通道没有车辆`）、并未真正进/出场。务必**检查 `data.code` 是否为 0**，非 0 视为业务失败并看 `data.msg`。
- **`correct-car-on-channel` 报「会话已过期」**：说明该通道当前没有可校正的抓拍会话，需先成功 `openydt device channel-snap` 生成待出/待进车会话，再调用本命令校正。
