---
name: openydt-billing
version: 1.0.3
description: "停车缴费交易域(trade)：临停车辆实时/估算查费、缴费信息回传、欠费批量补缴、预存款与运营积分预置。当用户要算停车费、回传/同步缴费订单、批量补缴欠费、给车充值预存款做自动扣费时使用。本域是「实时查费/缴费」的归属域；历史账单/缴费记录查询请用 parking 域(openydt-record)。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt trade --help"
---

# openydt-billing — 停车缴费交易域 (trade)

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**（认证 / profile / 签名 / 状态码 / 限速 / 安全规则）。未读共享基座不要执行任何命令。

## 何时用本技能

本技能覆盖**停车缴费交易域（trade）**：临停车辆查费、缴费信息回传、欠费批量补缴，以及预存款 / 运营积分预置（用于车辆自动扣费）。

意图路由：

- "这辆车停车费多少 / 查一下停车费 / 算费 / 出场要交多少钱" → `openydt trade get-park-fee`（实时查费，10 分钟内须完成缴费）。
- "未来停 X 小时大概多少钱 / 按时间段估算费用" → `openydt trade common-get-park-fee`。
- "缴费成功了把订单回传 / 同步缴费 / 付款回写" → `openydt trade pay-park-fee`（写，需 `--yes`）。
- "把这几条欠费一起补缴 / 批量补缴 / 离场欠费补缴" → `openydt trade payback-batch`（写，需 `--yes`）。
- "给车预存款 / 充值后自动扣费 / 先付费后离场免密支付" → `openydt trade set-prestore-for-c-park` 或 `set-prestore-for-c-park-first-pay-before-leave`（写，需 `--yes`）。
- "预置运营积分 / 积分自动抵扣车费" → `openydt trade set-points`（写，需 `--yes`）。

> 在场车确认、订单查询、进车补录等属于 parking 域，见 `openydt parking --help`。

## 可用命令

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| 按时间获取停车费用（估算未来时段） | `openydt trade common-get-park-fee` | 读 | `--park-code` 必填、`--car-type` 必填、`--start-time` 必填、`--end-time` 必填、`--charge-group`（仅云车场） |
| 获取停车费用（实时查费，10 分钟内缴费） | `openydt trade get-park-fee` | 读 | `--car-code` 车牌、`--park-code`（不传则全系统按车牌/卡号搜）、`--card-code`、`--parking-code`、`--need-pay-qr-code`、`--body`(couponList) |
| 缴费信息回传 | `openydt trade pay-park-fee` | 写 | `--parking-code` 必填、`--charge-date` 必填、`--pay-date` 必填、`--act-pay-charge` 必填、`--pay-origin` 必填、`--payment-mode` 必填、`--bill-code` 必填、`--body`(otherAtrr) |
| 欠费批量补缴 | `openydt trade payback-batch` | 写 | `--body`(paybackList[]：parkCode/parkingCode/paybackTime/paybackValue/couponValue/payOrigin/paymentMode/thirdBillCode) |
| 预置运营积分（自动抵扣车费） | `openydt trade set-points` | 写 | `--park-code`、`--app-id` 必填、`--parking-code` 必填、`--third-bill-code` 必填、`--rule` 必填、`--max` 必填、`--points-value` 必填、`--pay-origin` 必填、`--payment-mode` 必填 |
| 预置预存款（云车场自动扣费） | `openydt trade set-prestore-for-c-park` | 写 | `--parking-code` 必填、`--third-bill-code` 必填、`--prestore-amount` 必填、`--pay-origin` 必填、`--payment-mode` 必填、`--park-code` |
| 预置预存款（先付费后离场） | `openydt trade set-prestore-for-c-park-first-pay-before-leave` | 写 | `--parking-code` 必填、`--prestore-amount` 必填、`--park-code` |

> 所有**写**命令（`pay-park-fee` / `payback-batch` / `set-points` / `set-prestore-for-c-park` / `set-prestore-for-c-park-first-pay-before-leave`）执行时**必须加 `--yes`** 确认，否则会被拦截。

> 本表未列、但属 trade 域的可调用接口（如 `getCloudParkChargeInfoMap`/`setChargeRule`/`getChargeRule`/`setPrestore`/`setPrestoreFlow`/`setPrestoreForFirstPayThenLeave`/`syncAutoCoupon`），用通用兜底 `openydt api <cmd> --body '{...}'` 调用（写操作记得 `--yes`、先 `--dry-run`）；cmd/字段/sampleBody 见 catalog，详见 [[openydt-api-explorer]]。

## 业务流程

### 停车缴费闭环（查费 → 缴费 → 对账）

逐步执行，**务必把前序命令响应里的字段作为后续命令入参**，不要凭空构造：

1. **进车补录（云车场）** — 若车辆入场未上报，先补录进车记录：
   ```
   openydt parking supplement-parking-record-in --yes ...
   ```
2. **确认在场** — 查在场车，确认目标车牌在场并拿到所属车场：
   ```
   openydt parking get-park-on-site-car --park-code-list <park>
   ```
3. **实时查费** — 按车牌 + 车场查费：
   ```
   openydt trade get-park-fee --car-code <车牌> --park-code <park>
   ```
   从响应里取后续缴费所需字段：
   - `data.otherAttr.chargeBillNumber` → 账单号（缴费时回传 `--body` 的 `otherAtrr`）；
   - `data.shouldPayValue` → 应缴金额（= 实付 `actPayCharge` + 券抵扣 `couponValue`）；
   - 响应里的 `parkingCode`、`chargeDate` → 下一步缴费的 `--parking-code`、`--charge-date`。
   > 查费后 **10 分钟内**须完成缴费，否则令牌/账单可能失效。
   > 💰 **金额单位是「元」**（小数，如 `shouldPayValue: 1` 表示 1.00 元，不是 1 分）。`shouldPayValue` / `paidValue` / `actPayCharge` / `couponValue` 均为元。**别把 1 当成 1 分去付 0.01**，否则只缴了一分钱、仍欠 0.99。
4. **缴费回传**（写，需 `--yes`） — 把第 3 步取到的令牌、账单、应缴/实付金额回传：
   ```
   openydt trade pay-park-fee --yes \
     --parking-code <来自查费 parkingCode> \
     --charge-date <来自查费 chargeDate> \
     --pay-date <yyyyMMddHHmmss> \
     --act-pay-charge <实付，<= shouldPayValue> \
     --pay-origin 9 --payment-mode 4 \
     --bill-code <第三方唯一订单号> \
     --body '{"otherAtrr":{"chargeBillNumber":"<来自查费 otherAttr.chargeBillNumber>"}}'
   ```
   > 注意：带券缴费时 `couponList` 里的 `couponValue` + `actPayCharge` 必须等于查费返回的 `shouldPayValue`；`billCode` 须全局唯一，重试缴费时与首次保持一致以便去重对账。
5. **查订单记录** — 缴费后核对订单与明细：
   ```
   openydt parking get-pay-bill ...
   openydt parking get-park-detail ...
   ```

> 🔑 **缴费幂等**：`billCode` 全局唯一，**重试必须复用首次 billCode**（绝不新生成）；重发收到 `907 账单已同步` = 幂等命中、首次已成功，改用 `openydt parking get-pay-bill` 核对而非再缴。`payback-batch` 每条 `thirdBillCode` 同理。详见 [[openydt-shared]] 的 `../openydt-shared/references/write-idempotency.md`。

## 错误自愈速查（查费/缴费）

| 现象 | 含义 | 恢复动作 |
| --- | --- | --- |
| `resultCode=912 查费已超时` | 查费账单 10 分钟失效 | 重新 `get-park-fee` 取新账单，再缴 |
| 缴费回 `907 账单已同步` | 幂等命中，首次已成功 | 不重发；`openydt parking get-pay-bill` 核对金额一致 |
| 连接超时/网关 404（写发出后不确定是否成功） | 可能已成功 | **用同一 billCode** 重发（平台去重），或先查 `get-pay-bill` |
| `resultCode=909` / `status=7` | 缴费字段/必填错 | 用 `openydt schema payParkFee` 核对必填，字段取自查费响应 |

> 通用码/退出码/重试见 [[openydt-shared]]；金额单位=元见其 `../openydt-shared/references/result-reading-sop.md`。

## 示例

> 下列 parkCode/时间为文档化测试值（仅 test 环境）；照抄 catalog 历史 sampleBody 会撞无效车场/过期时间。写操作先 `--dry-run` 预览、确认后再 `--yes`。

实时查费（按车牌在指定车场查停车费）：

```
openydt trade get-park-fee --park-code 1ZS7H5PQH9 --car-code 粤EJW962
```

按时间段估算未来停车费用（小车，停 10 小时）：

```
openydt trade common-get-park-fee \
  --park-code 1ZS7H5PQH9 --car-type 1 \
  --start-time "2026-06-01 00:00:00" --end-time "2026-06-01 10:00:00"
```

缴费回传（写操作；金额/账单字段取自查费响应）。**先 `--dry-run` 预览签名请求，确认无误再把 `--dry-run` 换成 `--yes` 实发**；`--bill-code` 必须全局唯一（用你的订单号，重试与首次保持一致以便去重）：

```
openydt trade pay-park-fee --dry-run \
  --parking-code <来自查费响应 parkingCode> \
  --charge-date <来自查费响应 chargeDate> --pay-date <yyyyMMddHHmmss> \
  --act-pay-charge 3.2 \
  --pay-origin 9 --pay-origin-remark 微信 \
  --payment-mode 4 --payment-mode-remark 微信支付 \
  --bill-code <你的唯一订单号>
```
