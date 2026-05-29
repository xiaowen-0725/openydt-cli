---
name: openydt-billing
version: 1.0.0
description: "停车缴费交易域：查停车费、缴费回传、批量补缴欠费、预存款/运营积分预置。覆盖查费/缴费/停车费/查停车费/算费/在线缴费/缴费回传/付款/补缴/欠费补缴/批量缴费/预存款/预付费/先付费后离场/免密支付/运营积分/积分抵扣/按时间算费等高频说法。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt trade --help"
---

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
   - `data.otherAttr.chargeBillToken` / `data.otherAttr.chargeBillNumber` → 缴费令牌 / 账单号（缴费时回传 `--body` 的 `otherAtrr`）；
   - `data.shouldPayValue` → 应缴金额（= 实付 `actPayCharge` + 券抵扣 `couponValue`）；
   - 响应里的 `parkingCode`、`chargeDate` → 下一步缴费的 `--parking-code`、`--charge-date`。
   > 查费后 **10 分钟内**须完成缴费，否则令牌/账单可能失效。
4. **缴费回传**（写，需 `--yes`） — 把第 3 步取到的令牌、账单、应缴/实付金额回传：
   ```
   openydt trade pay-park-fee --yes \
     --parking-code <来自查费 parkingCode> \
     --charge-date <来自查费 chargeDate> \
     --pay-date <yyyyMMddHHmmss> \
     --act-pay-charge <实付，<= shouldPayValue> \
     --pay-origin 9 --payment-mode 4 \
     --bill-code <第三方唯一订单号> \
     --body '{"otherAtrr":{"chargeBillToken":"<来自查费>","chargeBillNumber":"<来自查费>"}}'
   ```
   > 注意：带券缴费时 `couponList` 里的 `couponValue` + `actPayCharge` 必须等于查费返回的 `shouldPayValue`；`billCode` 须全局唯一，重试缴费时与首次保持一致以便去重对账。
5. **查订单记录** — 缴费后核对订单与明细：
   ```
   openydt parking get-pay-bill ...
   openydt parking get-park-detail ...
   ```

## 示例

实时查费（按车牌在指定车场查停车费，参数取自 catalog sampleBody）：

```
openydt trade get-park-fee --park-code 2KNTYVWC --car-code 粤EXX123
```

按时间段估算未来停车费用（小车，停 10 小时）：

```
openydt trade common-get-park-fee \
  --park-code 2KNTYVWC --car-type 1 \
  --start-time "2018-01-01 00:00:00" --end-time "2018-01-01 10:00:00"
```

缴费回传（写操作，必须加 `--yes`；金额/账单字段取自查费响应）：

```
openydt trade pay-park-fee --yes \
  --parking-code 180410135558832886170666 \
  --charge-date 20180411135558 --pay-date 20180411135658 \
  --act-pay-charge 3.2 \
  --pay-origin 9 --pay-origin-remark 微信 \
  --payment-mode 4 --payment-mode-remark 微信支付 \
  --bill-code 0
```
