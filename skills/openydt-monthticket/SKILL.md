---
name: openydt-monthticket
version: 1.0.1
description: "月票/VIP 域(ticket)：月票类型与月票闭环(创建类型/开通/续费/退费/取消/冻结/解冻)、特殊车辆类型(访客/黑名单 VIP 组)创建与查询、车辆身份/车主/VIP 查询。当用户要开通续费月票、查月票将过期/已售、或建访客/黑名单 VIP 组时使用。月票的缴费/扣费记录属本域；临停实时算费请用 trade 域(openydt-billing)。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt ticket --help"
---

# openydt-monthticket — 月票 / VIP 域 (ticket)

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**（认证 / profile / 签名 / 状态码 / 限速 / 安全规则）。未读共享基座不要执行任何命令。

## 何时用本技能

当用户的诉求落在 **月票 / VIP / 会员包月 / 特殊车辆(访客、黑名单)/ 车辆身份** 业务时使用本技能,典型如:

- "给某车牌开通月票 / 续费 / 退费 / 取消月票"
- "新建一个月票类型 / 改月票类型 / 查月票类型详情"
- "查某车牌的月票记录 / 月票将过期 / 月票已售数量"
- "查车辆是不是 VIP / 查车主"、"月票冻结 / 解冻"
- "新增访客 VIP 组 / 黑名单 VIP 组的特殊车辆类型"

意图路由:
- 仅做**查询**(读) → 直接用对应 `get-*` 命令,无需 `--yes`。
- 涉及**创建 / 修改 / 开通 / 续费 / 退费 / 取消 / 冻结**(写) → 命令必须加 `--yes`(见下文每条标注)。
- 访客/黑名单的「名单成员」管理不在本域:访客用 `openydt visitor`,黑名单用 `openydt blacklist`,红名单用 `openydt redlist`（见 [[openydt-list]]）;但他们引用的「特殊车辆类型ID」由本域 `add-special-car-type` 创建并通过 `get-special-car-type-list` 查询。
- **临停实时算费**不在本域，请用 [[openydt-billing]]（`openydt parking get-park-fee`）。

## 可用命令

命令格式:`openydt ticket <use>`。下表仅列 catalog 中 included 的命令(共 29 条)。

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| 新增线上月票类型 | `openydt ticket add-online-month-ticket-type` | 写(`--yes`) | parkCodes*, ticketName*, price*, timePeriod(startTime/endTime), channelList |
| 修改线上月票类型 | `openydt ticket month-ticket-config-edit` | 写(`--yes`) | monthTicketConfigId*, ticketName*, parkCodeList* |
| 查询线上月票类型详情 | `openydt ticket get-month-ticket-config-detail` | 读 | monthTicketConfigId* |
| 查询线上月票类型详情列表 | `openydt ticket get-month-ticket-config-detail-list` | 读 | parkCodeList*, monthTicketConfigId, ticketStatus |
| 查看月票已购买数量 | `openydt ticket get-month-ticket-sell-num` | 读 | monthId* |
| 月票类型名额扣减 | `openydt ticket deduct-month-ticket-config` | 写(`--yes`) | monthTicketConfigId*, thirdpartyBillCode*, thirdpartyIdentify*, parkCodes*, carNo*, num*, startTime*, endTime* |
| 开通线上月票 | `openydt ticket add-online-month-ticket` | 写(`--yes`) | monthTicketConfigId*, parkCodes*, carNo*, billCode*, userName*, userPhone*, originPrice*, favorPrice*, payOrigin*, payMode*, timePeriodList*, thirdpartyIdentify(预约开通必填) |
| 续费线上月票 | `openydt ticket renew-online-vip-ticket` | 写(`--yes`) | billCode*, monthTicketId*, originPrice*, favorPrice*, payOrigin*, payMode*, renewBy*, renewTime*, timePeriodList* |
| 取消线上月票 | `openydt ticket cancel-online-vip-ticket` | 写(`--yes`) | parkCode*, billCode*, refundPrice*, payOrigin*, payMode* |
| 根据月票类型取消线上月票 | `openydt ticket cancel-online-month-ticket-by-month-ticket-type` | 写(`--yes`) | monthTicketConfigId*, parkCode* |
| 修改线上月票订单信息 | `openydt ticket edit-online-vip-ticket` | 写(`--yes`) | monthTicketId*, carNo*, userName*, userPhone*, remark1*~remark3*, editBy*, editTime* |
| 查询线上月票 | `openydt ticket get-online-vip-ticket` | 读 | monthTicketId* |
| 查询线上月票记录 | `openydt ticket get-online-month-ticket-list` | 读 | parkCodeList*, pageNum*, pageSize* |
| 查询车牌线上月票 | `openydt ticket get-online-month-ticket-by-car-card` | 读 | carCode* |
| 查询线上月票支付信息 | `openydt ticket get-online-month-ticket-payment` | 读 | parkCode*, operateTimeFrom*, operateTimeTo*, pageNum*, pageSize* |
| 通过车牌查看月票预约信息详情 | `openydt ticket get-month-ticket-bill-detail` | 读 | thirdBillCode* |
| 查询将要过期的月票 | `openydt ticket get-will-expire-month-ticket-bill` | 读 | validFrom, validTo, pageNum, pageSize |
| 查询月票账号交易记录 | `openydt ticket get-month-ticket-account-transation-record` | 读 | monthTicketBillId*, transationTimeStart*, transationTimeEnd*, pageNum*, pageSize* |
| 查询月票账号扣费记录 | `openydt ticket get-month-ticket-account-use-record` | 读 | monthTicketBillId*, pageNum*, pageSize* |
| 申请月票冻结 | `openydt ticket apply-month-ticket-freeze` | 写(`--yes`) | monthTicketBillId*, frozenStartTime*, frozenEndTime*, reason* |
| 月票冻结 | `openydt ticket freeze-month-ticket` | 写(`--yes`) | monthTicketBillId*, frozenStartTime*, frozenEndTime* |
| 月票解冻 | `openydt ticket un-freeze-month-ticket` | 写(`--yes`) | monthTicketBillId* |
| 添加特殊车辆类型 | `openydt ticket add-special-car-type` | 写(`--yes`) | parkCode*, specialCarTypeName*, vipGroupType*(1访客/2黑名单) |
| 获取特殊车辆类型列表 | `openydt ticket get-special-car-type-list` | 读 | parkCodeList*, vipGroupType*(1访客/2黑名单) |
| 查询车辆的车主及VIP | `openydt ticket get-car-owner-and-vip-type` | 读 | parkCode*, carNo* |
| 获取车辆身份 | `openydt ticket get-vip-by-car-no` | 读 | carCode*, enterTime*, leaveTime*, parkCode*, parkingCode* |
| 通过车牌和时间获取VIP信息 | `openydt ticket get-vip-by-car-no-and-time` | 读 | carCode*, enterTime*, parkCode* |
| 获取车场协议 | `openydt ticket get-park-agreement` | 读 | parkCodeList* |
| 同步电子券二维码扫码(车场协议) | `openydt ticket park-agreement-save` | 写(`--yes`) | parkCode*, agreementTitle*, agreementContent* |

> `*` 表示必填。写操作均通过 `f.ConfirmWrite` 拦截,执行时**必须**带 `--yes`,否则会被拒绝。
> 参数以 `openydt ticket <cmd> --help` 为准,有出入信 --help。

## 业务流程

### 月票闭环(创建类型 → 开通 → 续费/退费 → 查询)

务必把**前序命令响应里的字段作为后续命令的入参**,不要凭空编造 ID。

1. **创建月票类型** — `openydt ticket add-online-month-ticket-type --yes`(body 见下方示例)。
   从响应取 **`data.monthTicketConfigId`** 作为「月票类型ID」。
2. **开通线上月票** — `openydt ticket add-online-month-ticket --yes`,用上一步的 `monthTicketConfigId` 填入 `--month-ticket-config-id`(或 body 的 `monthTicketConfigId`)。
   从响应取 **`data.monthTicketId`(月票订单id)** 用于续费 / 退费 / 修改。
3. **续费 / 退费**:
   - 续费 — `openydt ticket renew-online-vip-ticket --yes`,用第 2 步的 `monthTicketId` 填 `--month-ticket-id`,并给新的 `timePeriodList`。
   - 退费(取消)— `openydt ticket cancel-online-vip-ticket --yes`,用 `monthTicketId` + `parkCode` + `refundPrice`;或按类型批量取消 `openydt ticket cancel-online-month-ticket-by-month-ticket-type --yes`(用 `monthTicketConfigId`)。
4. **查询**:
   - 按类型/车场列表查 — `openydt ticket get-online-month-ticket-list`。
   - 按车牌查 — `openydt ticket get-online-month-ticket-by-car-card`(用 `carCode`)。
   - 单笔订单查 — `openydt ticket get-online-vip-ticket`(用 `monthTicketId`)。

> 💰 金额（`originPrice`/`favorPrice`/`refundPrice`/`price`）单位是**元**；开通/续费/扣减的 `billCode`/`thirdpartyBillCode`/`thirdpartyIdentify` 是幂等键——**重试复用首次值、绝不新生成**；收到 `907 账单已同步` = 幂等命中、首次已开通，改查 `get-online-vip-ticket` 确认，不要再发。详见 [[openydt-shared]] 的 `references/write-idempotency.md` 与 `references/result-reading-sop.md`。

### 月票冻结 / 解冻

- 冻结需要**月票订单id**:先用 `get-online-month-ticket-by-car-card` 或 `get-online-month-ticket-list` 查到 **`monthTicketBillId`(月票订单id)**。
- 冻结 — `openydt ticket apply-month-ticket-freeze --yes` 或 `openydt ticket freeze-month-ticket --yes`(传 `monthTicketBillId` + `frozenStartTime` + `frozenEndTime`)。
- 解冻 — `openydt ticket un-freeze-month-ticket --yes`(传 `monthTicketBillId`)。

### 特殊车辆类型(供访客 / 黑名单复用)

- 用 `openydt ticket add-special-car-type --yes` 创建特殊车辆类型,`vipGroupType=1` 为访客VIP组、`vipGroupType=2` 为黑名单VIP组。
- 从响应取**特殊车辆类型ID**(也可用 `openydt ticket get-special-car-type-list --vip-group-type 1` 并传 `--body '{"parkCodeList":["1ZS7H5PQH9"]}'` 查询列表拿到ID;parkCodeList 和 vipGroupType 均为必填)。
- 该「特殊车辆类型ID」供 `openydt visitor`(访客)与 `openydt blacklist`(黑名单)在添加名单成员时作为入参复用,实现「类型在 ticket 域创建、名单成员在各自域管理」的闭环。

### 未一等化的 ticket 子特性（用 `openydt api` 兜底）

以下 ticket 域接口在 catalog 中可调用(callable)但未做成一等命令，无专属 `openydt ticket` 子命令；需要时用通用兜底 `openydt api <cmd>`（写操作必须 `--yes`，建议先 `--dry-run` 预览），body 照抄 `catalog/catalog.json` 的 sampleBody，详见 [[openydt-api-explorer]]：

- **月票预约/排队**：`bookOrCancelReservation`(写)、`checkMonthTicketAppointment`、`getMonthTicketAppointmentByCarCode` / `getMonthTicketAppointmentDetail` / `getMonthTicketAppointmentLineUp`。
  > ⚠️ `getMonthTicketAppointmentPark`(查月票预约可开通的车场)属 **park 域**，请用 `openydt api getMonthTicketAppointmentPark` 调用，见 [[openydt-park]]。
- **月票授权访客**：`addMonthTicketAuthorizeVisitor`(写)、`allowBuyMonthlyTicket`(写)、`cancelMonthTicketAuthorizeVisitorByCarNo` / `cancelMonthTicketAuthorizeVisitorById`(写)、`getMonthTicketAuthorizeVisitor` / `getMonthTicketAuthorizeVisitorHis`。
- **月票证件规则**：`monthTicketCertifiRuleSave`(写)、`saveOrUpdateMonthTicketCretifi`(写)、`delMonthTicketCertifiRule`(写)、`getMonthTicketCertifiRuleList` / `getMonthTicketCertificateInfoList` / `getCertificateByInfo`。
- **线下月卡（整套未一等化）**：`getMonthCard` / `pauseMonthCard`(写) / `recoverMonthCard`(写) / `refundMonthCard`(写) / `renewMonthCard`(写)——线下月卡的查询/暂停/恢复/退费/续费，均用 `openydt api <cmd>` 调，body 见 catalog。
- **VIP 类型扩展**：`addCusVipType`(写) / `editCusVipType`(写) / `getVipType` / `addVipTicket`(写) / `setVipChargeRule`(写) / `querySupportMonthTicketParkList`——自定义 VIP 类型与计费规则，同样用 `api` 兜底。

## 示例

> 示例 parkCode/时间为文档化测试值（仅 test 环境；照抄 catalog 历史 sampleBody 会撞无效车场/过期有效期）。写操作建议先用 `--dry-run` 预览签名请求，确认后再 `--yes`；ID 类参数（monthTicketConfigId 等）须取自上一步响应，不要照抄示例值。

1) 新增线上月票类型(写,需 `--yes`)— 成功后从响应 `data.monthTicketConfigId` 取类型ID:
```bash
# 先 dry-run 预览
openydt ticket add-online-month-ticket-type --dry-run --body '{
  "parkCodes": "1ZS7H5PQH9",
  "ticketName": "测试月票类型01",
  "price": 10,
  "timePeriod": { "startTime": "2026-06-01 00:00:00", "endTime": "2026-06-30 23:59:59" }
}'
# 确认后实发
openydt ticket add-online-month-ticket-type --yes --body '{
  "parkCodes": "1ZS7H5PQH9",
  "ticketName": "测试月票类型01",
  "price": 10,
  "timePeriod": { "startTime": "2026-06-01 00:00:00", "endTime": "2026-06-30 23:59:59" }
}'
```

2) 用上一步的 `monthTicketConfigId` 开通线上月票(写,需 `--yes`):
```bash
# 先 dry-run 预览
openydt ticket add-online-month-ticket --dry-run --body '{
  "carNo": "粤EJW962",
  "billCode": "<全局唯一订单号,如 your-app-20260601-0001>",
  "parkCodes": "1ZS7H5PQH9,PTD2YBBZ",
  "originPrice": 10,
  "favorPrice": 5,
  "payMode": 4,
  "payModeRemark": "微信支付",
  "payOrigin": 7,
  "payOriginRemark": "智慧停车",
  "userName": "王五",
  "userPhone": "18000000000",
  "thirdpartyIdentify": "<第三方唯一标识,预约开通必填>",
  "monthTicketConfigId": "<取自上一步响应的 data.monthTicketConfigId>",
  "timePeriodList": [
    { "startTime": "2026-06-01 00:00:00", "endTime": "2026-06-30 23:59:59" }
  ]
}'
# 确认后实发(--yes)
openydt ticket add-online-month-ticket --yes --body '{
  "carNo": "粤EJW962",
  "billCode": "<全局唯一订单号,如 your-app-20260601-0001>",
  "parkCodes": "1ZS7H5PQH9,PTD2YBBZ",
  "originPrice": 10,
  "favorPrice": 5,
  "payMode": 4,
  "payModeRemark": "微信支付",
  "payOrigin": 7,
  "payOriginRemark": "智慧停车",
  "userName": "王五",
  "userPhone": "18000000000",
  "thirdpartyIdentify": "<第三方唯一标识,预约开通必填>",
  "monthTicketConfigId": "<取自上一步响应的 data.monthTicketConfigId>",
  "timePeriodList": [
    { "startTime": "2026-06-01 00:00:00", "endTime": "2026-06-30 23:59:59" }
  ]
}'
```

3) 按车牌查询线上月票(读,无需 `--yes`):
```bash
openydt ticket get-online-month-ticket-by-car-card --body '{
  "carCode": "粤EJW962",
  "buyMethod": 2,
  "ticketType": 1,
  "validStatus": 0,
  "startTime": "20260601000000",
  "endTime": "20261231235959",
  "pageNum": 1,
  "pageSize": 10
}'
```
