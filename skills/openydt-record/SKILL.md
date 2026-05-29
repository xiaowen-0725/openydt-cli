---
name: openydt-record
version: 1.0.0
description: "智慧停车开放平台-停车记录域(parking)：查询在场车/进出场记录、缴费账单与欠费记录、进车补录、车牌校正、锁车/解锁、拦截策略、自助进出场。触发词:停车记录/在场车/在场车辆/进场记录/出场记录/进出记录/查车/查在场/进车补录/补录进场/补录图片/锁车/锁定车辆/解锁/解锁车辆/查锁车状态/查费/查账单/缴费记录/支付账单/欠费/欠费记录/欠费条数/取消欠费/车辆欠费/异常开闸/非系统开闸/异常离场/车牌校正/校正车牌/通道权限/检查通道有车/扫码进出/扫通道码/路边车登记/拦截策略/代扣订单"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt parking --help"
---

# openydt-record — 停车记录域 (parking)

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**，了解认证 / profile / 签名 / 状态码 / 限速 / 安全等通用约定，再执行本域命令。

## 何时用本技能

当用户的需求落在“停车记录”相关业务上时使用本技能，包括：

- **在场 / 进出查询**：查在场车辆、进场记录、出场记录、停车记录详情、通道是否有车。
- **缴费 / 账单 / 欠费**：查缴费记录、支付账单（明细 / 文件）、车场账单、车辆欠费、运营商欠费记录与条数、异常开闸 / 异常离场、取消欠费、代扣订单更新。
- **补录与校正**：进车补录（`supplement-parking-record-in`）、进场图片补录、在场 / 通道 / 进出确认后的车牌校正。
- **锁车控制**：锁车 / 解锁 / 查锁车状态。
- **自助进出与策略**：扫通道码自助进出、路边车自助登记、车场拦截策略创建 / 删除、通道权限查询。

意图路由：
- “查在场车 / 在场车辆 / 现在场内有哪些车” → `get-park-on-site-car`
- “进场记录 / 进车记录” → `get-car-in-list`；“出场记录” → `get-car-out-list`
- “某条停车记录详情” → `get-park-detail`（或忽略状态 `get-park-detail-ignore-status`）
- “查费 / 缴费记录 / 账单” → `get-pay-bill` / `get-payment-record-detail-list` / `get-park-pay-bill-by-car-nos-and-pay-time`
- “欠费 / 欠费记录” → `get-car-arrearage-list` / `get-arrears-list-by-operator` / `get-arrears-count`
- “进车补录 / 补录进场” → `supplement-parking-record-in`
- “锁车 / 解锁 / 锁车状态” → `lock-car` / `unlock-car` / `get-car-lock-status`
- 跨域提示：月票/电子券/访客/黑名单等不在本域，分别使用 `openydt ticket` / `openydt coupon` / `openydt visitor` / `openydt blacklist`。

## 可用命令

> 命令统一以 `openydt parking <use>` 调用。写操作（写）均需追加 `--yes` 确认。

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| 检查通道是否有车 | `openydt parking check-channel-exist-car` | 读 | parkCode, channelCode |
| 在场车辆查询 | `openydt parking get-park-on-site-car` | 读 | parkCodeList, enterTimeFrom, enterTimeTo, pageNum, pageSize |
| 进场记录查询 | `openydt parking get-car-in-list` | 读 | parkCode, isPresence, startTime, endTime, pageNum, pageSize |
| 出场记录查询 | `openydt parking get-car-out-list` | 读 | parkCode, pageNum, pageSize |
| 停车记录详情 | `openydt parking get-park-detail` | 读 | parkCode, parkingCode/carCode (任选定位) |
| 停车记录详情(忽略状态) | `openydt parking get-park-detail-ignore-status` | 读 | parkCode, parkingCode/carCode (任选定位) |
| 通道权限查询 | `openydt parking get-channel-permission` | 读 | parkCode, channelId, carCode, operatorTime, plateColor |
| 缴费记录查询 | `openydt parking get-pay-bill` | 读 | parkingCode (+parkCode) |
| 支付账单明细列表 | `openydt parking get-payment-record-detail-list` | 读 | parkCode, pageNum, pageSize |
| 支付账单文件 | `openydt parking get-payment-record-detail-file` | 读 | parkCode, payTime |
| 按车牌+支付时间查账单 | `openydt parking get-park-pay-bill-by-car-nos-and-pay-time` | 读 | carNoList, pageNum, pageSize |
| 车辆欠费记录 | `openydt parking get-car-arrearage-list` | 读 | carNo, parkCodeList (可选筛选) |
| 运营商欠费记录 | `openydt parking get-arrears-list-by-operator` | 读 | pageSize, pageNum |
| 运营商欠费条数 | `openydt parking get-arrears-count` | 读 | (body 可空) |
| 欠费图片详情 | `openydt parking get-arrears-detail` | 读 | (body 可空) |
| 非系统开闸记录 | `openydt parking get-abnormal-open-gate-list` | 读 | parkCodeList, openGateTimeFrom, openGateTimeTo |
| 异常离场记录 | `openydt parking get-abnormal-out-list` | 读 | parkCodeList, leaveTimeFrom, leaveTimeTo |
| 查锁车状态 | `openydt parking get-car-lock-status` | 读 | carNo / cardNumber |
| 进车补录 | `openydt parking supplement-parking-record-in` | 写 | parkCode, carCode, enterTime, channelCode, carCodeType, carCodeColor, parkOrArea |
| 进场图片补录 | `openydt parking supplement-parking-record-image` | 写 | parkCode, parkingCode, parkOrArea, carCodeImage, carImage, parkingType |
| 在场车牌校正 | `openydt parking correct-car-no` | 写 | parkCode, parkingCode, newCarNo, correctTime, correctName, operateType |
| 通道待进出车牌校正 | `openydt parking correct-car-on-channel` | 写 | parkCode, channelCode, newCarNo, correctTime |
| 进出确认拍照后车牌校正 | `openydt parking correcting-car-code-after-car-in-out-confirm-phone` | 写 | parkCode, channelId, newCarNo, correctTime |
| 锁车 | `openydt parking lock-car` | 写 | carNo / cardNumber, lockReason |
| 解锁 | `openydt parking unlock-car` | 写 | carNo / cardNumber, unlockReason |
| 扫通道码自助进出场 | `openydt parking scan-channel-code-in-out` | 写 | parkCode, channelSeq, userUniqCode |
| 路边车自助登记 | `openydt parking roadside-car-check-in` | 写 | carNo, positionNo, longitude, latitude |
| 创建车场拦截策略 | `openydt parking create-intercept-policy` | 写 | parkCode, policyName, tags |
| 删除车场拦截策略 | `openydt parking delete-intercept-policy` | 写 | parkCode |
| 取消欠费 | `openydt parking cancellation-of-arrears` | 写 | recordId, status, remark, operator |
| 更新代扣流程订单 | `openydt parking update-wihhold-detail-bill` | 写 | thirdBillCode, billStatus, billCallbackDate |

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

## 示例

> 以下参数取自 catalog 的 sampleBody，可直接替换为真实值运行。

读：查询某车场指定时段进场记录（含指定车牌过滤、分页）

```bash
openydt parking get-car-in-list --body '{
  "parkCode": "2KNTYVWC",
  "carNoArray": ["粤YAL876", "粤A66666"],
  "isPresence": "0",
  "startTime": "20171015000000",
  "endTime": "20171015235959",
  "pageNum": 1,
  "pageSize": 10
}'
```

写：进车补录（注意 `--yes`，响应返回 `parkingCode` 供后续补图/校正使用）

```bash
openydt parking supplement-parking-record-in --yes --body '{
  "parkCode": "2KNTYVWC",
  "carCode": "湘OQKKZA",
  "enterTime": "20171015000000",
  "channelCode": "AA123C",
  "carCodeType": 1,
  "carCodeColor": 1,
  "parkOrArea": 1
}'
```

写：锁定指定车辆（注意 `--yes`）

```bash
openydt parking lock-car --yes --body '{
  "cardNumber": "A12345",
  "carNo": "粤YZZ568",
  "lockReason": "reason"
}'
```
