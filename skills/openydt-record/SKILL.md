---
name: openydt-record
version: 1.0.8
description: "停车记录域(parking)：查询在场/进出/停车详情/历史缴费/欠费，执行补录校正/锁车/自助进出，以及按离场事件语义分析经营车流、停车时长、逃费和放行类型。用户请求上述记录级任务时使用；实时查费缴费用 openydt-billing，平台聚合统计用 openydt-data。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt parking --help"
---

# openydt-record — 停车记录域 (parking)

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**，了解认证 / profile / 签名 / 状态码 / 限速 / 安全等通用约定，再执行本域命令。

## 意图路由

- “查在场车 / 在场车辆 / 现在场内有哪些车” → `get-park-on-site-car`
- “进场记录 / 进车记录” → `get-car-in-list`；“出场记录” → `get-car-out-list`
- “全量进出明细 / 经营分析原始数据” → 按 [[openydt-shared]] 的全量分页规则导出本地 NDJSON 快照
- “基于离场明细重算 / 核验停车时长、分布或分位数” → **先读 [`references/parking-duration-analysis.md`](references/parking-duration-analysis.md)，再运行随附脚本**
- “逃费 / 疑似跟车 / 按放行模式统计 / 异常放行 / 遥控开闸分析” → **先读 [`references/parking-record-enums.md`](references/parking-record-enums.md)，以 CLI schema 的枚举与领域语义为准**
- “直接调用平台停车时长分布等聚合统计” → [[openydt-data]]
- “某条停车记录详情” → `get-park-detail`（或忽略状态 `get-park-detail-ignore-status`）
- “缴费记录 / 历史账单” → `get-pay-bill` / `get-payment-record-detail-list` / `get-park-pay-bill-by-car-nos-and-pay-time`（实时应缴金额请用 trade 域 `get-park-fee`，本域只查历史）
- “欠费 / 欠费记录” → `get-car-arrearage-list` / `get-arrears-list-by-operator` / `get-arrears-count`
- “进车补录 / 补录进场” → `supplement-parking-record-in`
- “盘点离场 / 批量清场 / 盘点记录” → `inventory-car`（写）/ `get-inventory-record`（读）
- “锁车 / 解锁 / 锁车状态” → `lock-car` / `unlock-car` / `get-car-lock-status`
- 跨域提示：月票/电子券/访客/黑名单等不在本域，分别使用 `openydt ticket`（见 [[openydt-monthticket]]）/ `openydt coupon`（见 [[openydt-coupon]]）/ `openydt visitor` / `openydt blacklist`（见 [[openydt-list]]）；实时算费/缴费见 [[openydt-billing]]。

## 可用命令

> 命令统一以 `openydt parking <use>` 调用。写操作（写）均需追加 `--yes` 确认。

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| 检查通道是否有车 | `openydt parking check-channel-exist-car` | 读 | parkCode, channelCode |
| 在场车辆查询 | `openydt parking get-park-on-site-car` | 读 | parkCodeList, enterTimeFrom, enterTimeTo, pageNum, pageSize |
| 进场记录查询 | `openydt parking get-car-in-list` | 读 | parkCode, isPresence, startTime, endTime, pageNum, pageSize |
| 出场记录查询 | `openydt parking get-car-out-list` | 读 | parkCode, **carNo**(单数), leaveStartTime/leaveEndTime 或 enterTimeFrom/enterTimeTo(二选一必填), pageNum, pageSize(≤100) |
| 停车记录详情 | `openydt parking get-park-detail` | 读 | parkCode, parkingCode/carCode (任选定位) |
| 查盘点记录 | `openydt parking get-inventory-record` | 读 | parkCodeList, inventoryStartTime, inventoryEndTime, remark |
| 盘点离场 | `openydt parking inventory-car` | 写 | parkCode, enterTimeEnd, carNo/carNos/parkingCodes, remark |
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

> 本表未列、但属 parking 域的可调用接口（如 `getHisParkDetail`/`getParkPayBill`/`getParkingPosition`/`getParkingSpaceInfo`/`paymentRecordQuery*`/`selfInOutForCloudPark`/`typingRandomCodeInOut`/`scanChannelCodeInOutFlow`/`supplyCarIn-Out-Pic` 等），用 `openydt api <cmd> --body '{...}'` 调用（写操作 `--yes`、先 `--dry-run`）；车辆标签打标/取消：`openydt api addCarTags` / `openydt api delCarTags`（写，需 `--yes`），body 见 `catalog/catalog.json` 的 sampleBody；详见 [[openydt-api-explorer]]。

> ⚠️ `update-wihhold-detail-bill` 里的 `wihhold` 是平台接口编码 `updateWihholdDetailBill` 的**原始拼写**（平台侧 typo，本应是 withhold）。CLI 按平台编码逐字发送，**必须照此拼写、不要「纠正」为 withhold**，否则平台返回 `status=9 接口不存在`。

> 查询、校正进出记录或处理通道会话前，读 [`references/pitfalls.md`](references/pitfalls.md)；它收录出场字段、在场时间范围、状态判定与会话过期等分支。

### 写操作幂等（避免重试重复）

- `update-wihhold-detail-bill` 用 `thirdBillCode` 去重——重试复用同值，绝不新生成。
- `supplement-parking-record-in` / `inventory-car` / `correct-*` **无显式幂等键**：网络超时/网关 404 触发的自动重试可能重复补录或重复盘点离场。重发前**先用读命令确认首次是否已生效**：
  - 补录前：`check-channel-exist-car` 确认通道是否已有在场记录；
  - 盘点前：`get-inventory-record` 确认该时段是否已存在盘点操作；
  - 在场查验：`get-park-on-site-car` 确认车牌是否仍在场，避免重复补录重复入账。
- 详见 [[openydt-shared]] 的 `../openydt-shared/references/write-idempotency.md`。

> 多步业务流程(在场/进出查询→详情、进车补录、锁车/解锁、欠费→取消)见 [references/flows.md](references/flows.md),跑多步链路前先 Read。

## 示例

> 示例 parkCode/时间为文档化测试值（仅 test 环境：`1ZS7H5PQH9` / `PTD2YBBZ`，2026 相对时间）；照抄历史 sampleBody 中的旧 parkCode/时间会撞无效车场或过期时间段。写操作建议先把 `--yes` 换成 `--dry-run` 预览签名请求，确认后再 `--yes`。

读：查询某车场指定时段进场记录（含指定车牌过滤、分页）

```bash
openydt parking get-car-in-list --body '{
  "parkCode": "1ZS7H5PQH9",
  "carNoArray": ["粤EJW962"],
  "isPresence": "0",
  "startTime": "20260601000000",
  "endTime": "20260601235959",
  "pageNum": 1,
  "pageSize": 10
}'
```

读：查询某车场指定**出场**时段的出场记录（注意字段名与 in-list 不同：`carNo` 单数 + `leaveStartTime`/`leaveEndTime`）

```bash
openydt parking get-car-out-list --body '{
  "parkCode": "PTD2YBBZ",
  "carNo": "粤EJW987",
  "leaveStartTime": "20260531000000",
  "leaveEndTime": "20260531235959",
  "pageNum": 1,
  "pageSize": 10
}'
```

写：进车补录（注意 `--yes`，响应返回 `parkingCode` 供后续补图/校正使用）

```bash
openydt parking supplement-parking-record-in --yes --body '{
  "parkCode": "1ZS7H5PQH9",
  "carCode": "粤EJW962",
  "enterTime": "20260601080000",
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
  "carNo": "粤EJW962",
  "lockReason": "测试锁车"
}'
```
