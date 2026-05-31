# 停车记录域 业务流程

## 业务流程

> 通用原则：**先用读命令定位记录，拿到响应里的字段（如 `parkingCode`、`parkCode`、`channelCode`/`channelId`、`carCode`、欠费记录 `recordId` 等）作为后续写命令的入参，不要凭空填写。**

### 1. 在场 / 进出记录查询 → 详情下钻

1. 查在场车：`openydt parking get-park-on-site-car`，传 `parkCodeList`、`enterTimeFrom`、`enterTimeTo`、分页。
   响应里每条车记录会带 `parkingCode`（停车记录编号）、`carCode`（车牌）、`channelCode`（进出通道）。
2. 需要按进 / 出时段筛：进场用 `get-car-in-list`（`isPresence` 区分是否在场），出场用 `get-car-out-list`。
3. 下钻单条详情：把上一步响应里的 `parkCode` + `parkingCode`（或 `carCode`）传给 `openydt parking get-park-detail`；
   若记录状态异常导致查不到，改用 `openydt parking get-park-detail-ignore-status`。

### 2. 进车补录（写）

1.（可选）先 `openydt parking check-channel-exist-car`（传 `parkCode`、`channelCode`）确认通道当前是否已有车，避免重复补录。
2. 补录进场记录：`openydt parking supplement-parking-record-in --yes`，
   传 `parkCode`、`carCode`、`enterTime`、`channelCode`、`carCodeType`、`carCodeColor`、`parkOrArea`。
   响应会返回新生成的 `parkingCode`。
3.（可选）补图片：用上一步返回的 `parkingCode` 作为入参，调用
   `openydt parking supplement-parking-record-image --yes`，传 `parkCode`、`parkingCode`、`parkOrArea`、`carCodeImage`、`carImage`、`parkingType`。
4. 如发现进场车牌识别有误，用 `parkCode` + `parkingCode` 调 `openydt parking correct-car-no --yes` 做在场车牌校正；
   若车还卡在通道未确认，用 `parkCode` + `channelCode` 调 `openydt parking correct-car-on-channel --yes`。

### 3. 锁车 / 解锁（写）

1. 先查锁车状态：`openydt parking get-car-lock-status`（传 `carNo` 或 `cardNumber`），确认当前是否已锁。
2. 锁车：`openydt parking lock-car --yes`，传 `carNo`（或 `cardNumber`）与 `lockReason`。锁定后车辆离场会在出入口提示“车辆已锁定”。
3. 解锁：`openydt parking unlock-car --yes`，传同一 `carNo`（或 `cardNumber`）与 `unlockReason`。
4. 操作后可再次 `get-car-lock-status` 校验状态是否变更。

### 4. 欠费查询 → 取消欠费（写）

1. 查车辆欠费：`openydt parking get-car-arrearage-list`，或查运营商欠费：`openydt parking get-arrears-list-by-operator`（先 `get-arrears-count` 看总数）。
   响应里每条欠费记录带 `recordId`。
2. 取消欠费：把上一步的 `recordId` 传给 `openydt parking cancellation-of-arrears --yes`，并填 `status`、`remark`、`operator`。
