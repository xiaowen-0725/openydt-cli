---
name: openydt-coupon
version: 1.0.0
description: "电子券与商家域：创建/编辑/冻结/删除商家，建券模板、售卖、发券、查券、回收闭环。覆盖电子券/优惠券/抵扣券/停车券/商家券/商户/创建商家/建券/券模板/售券/卖券/发券/发放优惠券/扫码发券/固定券/打印券/回收券/退券/锁券/查券/已发放券/可用券/券二维码/券适用车场/售卖记录/发放记录等高频说法。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt coupon --help"
---

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**（认证 / profile / 签名 / 状态码 / 限速 / 安全规则）。未读共享基座不要执行任何命令。

## 何时用本技能

本技能覆盖**电子券与商家域（coupon）**：商家（trader）的创建与维护、电子券模板的创建、把券售卖给商家、给车辆发放电子券、查询券与商家信息、以及券的回收。车辆在停车场用电子券可抵扣部分或全部停车费。

意图路由：

- "新建一个商家 / 加个商户 / 录入商家账号" → `openydt coupon create-trader`（写，需 `--yes`）。
- "改商家信息 / 冻结商家 / 解冻 / 删除商家" → `openydt coupon edit-trader` / `frozen-trader` / `delete-trader`（写，需 `--yes`）。
- "查商家列表 / 看某个商家信息 / 校验商家账号密码" → `openydt coupon get-trader-list` / `get-trader-info-by-trader-code` / `validate-trader-account-and-password`。
- "建一个券模板 / 创建优惠券 / 建固定券" → `openydt coupon create-coupon-template` / `create-coupon` / `create-fixed-coupon`（写，需 `--yes`）。
- "把券卖给商家 / 售卖电子券" → `openydt coupon sell-coupon`（写，需 `--yes`）。
- "给这辆车发券 / 发放优惠券 / 按券编码发券 / 扫码发券" → `openydt coupon send-coupon` / `send-coupon-by-coupon-code` / `sync-scan-coupon-qr-code`（写，需 `--yes`）。
- "查这辆车有哪些券 / 已发放的券 / 可用券 / 券信息" → `openydt coupon query-car-code-valid-coupon` / `query-usable-coupon` / `query-coupon`。
- "回收这张券 / 退券 / 锁券 / 打印券" → `openydt coupon cancel-coupon` / `lock-coupon` / `print-coupon`（写，需 `--yes`）。

> 用券抵扣后的实际查费 / 缴费在 trade 域（`openydt trade --help`）；在场车确认在 parking 域（`openydt parking --help`）。

## 可用命令

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| 获取商家信息 | `openydt coupon get-trader-info-by-trader-code` | 读 | `--trader-code` 必填 |
| 校验商家用户账户密码 | `openydt coupon validate-trader-account-and-password` | 读 | `--trader-user-account` 必填、`--trader-password` 必填 |
| 创建商家 | `openydt coupon create-trader` | 写 | `--trader-name` 必填、`--contact` 必填、`--phone` 必填、`--park-code` 必填、`--login-account` 必填、`--password` 必填、`--trader-type`(0美食/1酒店/2休闲娱乐/3生活服务/4购物/5其他) |
| 编辑商家 | `openydt coupon edit-trader` | 写 | `--trader-code`/`--trader-id`(二选一)、`--trader-name` 必填、`--contact` 必填、`--phone` 必填、`--park-code` 必填、`--login-account` 必填、`--password` 必填、`--trader-type` |
| 冻结或解冻商家 | `openydt coupon frozen-trader` | 写 | `--trader-code`/`--trader-id`(二选一)、`--park-code` 必填、`--status` 必填(0解冻/1冻结) |
| 删除商家（不可恢复） | `openydt coupon delete-trader` | 写 | `--park-code` 必填、`--trader-code`/`--trader-id`(二选一) |
| 查询商家列表 | `openydt coupon get-trader-list` | 读 | `--park-code` 必填、`--create-time-from` 必填、`--create-time-to` 必填、`--page-num` 必填、`--page-size` 必填(最多1000) |
| 创建电子券模板 | `openydt coupon create-coupon-template` | 写 | `--name` 必填、`--sell-from`/`--sell-to`/`--grant-from`/`--grant-to` 必填、`--valid-minute` 必填、`--balance-type` 必填、`--coupon-type` 必填、`--face-value` 必填、`--original-price`/`--real-price`/`--use-rule-from`/`--use-rule-to` 必填 |
| 创建电子券并售卖给商户 | `openydt coupon create-coupon` | 写 | `--trader-code` 必填、`--total-count` 必填(≤1万)、`--body`(parkCodeList、couponTemplate 子对象) |
| 创建固定电子券（一码多券） | `openydt coupon create-fixed-coupon` | 写 | `--trader-code` 必填、`--group-name` 必填、`--valid-from`/`--valid-to` 必填、`--max-num` 必填、`--sell-bill-id` 必填、`--uniq-no` 必填 |
| 售卖电子券给商家 | `openydt coupon sell-coupon` | 写 | `--trader-coupon-template-code` 必填、`--trader-code` 必填、`--sell-num` 必填、`--sell-money` 必填、`--sell-time` 必填、`--sell-remark`、`--transation-num` |
| 给指定车辆发放电子券 | `openydt coupon send-coupon` | 写 | `--trader-code` 必填、`--sell-bill-id` 必填、`--car-code` 必填、`--car-code-color`、`--parking-code` |
| 根据券编码发放商家券 | `openydt coupon send-coupon-by-coupon-code` | 写 | `--coupon-code` 必填、`--car-no`/`--card-no`(不能同时为空)、`--park-code`、`--grant-user-id`、`--car-code-color`、`--parking-code`、`--is-fixed-qr-code`、`--origin` |
| 同步电子券二维码扫码 | `openydt coupon sync-scan-coupon-qr-code` | 写 | `--coupon-code` 必填、`--grant-user-id` 必填 |
| 检查电子券是否可发放 | `openydt coupon check-coupon-whether-send-available` | 读 | `--coupon-code` 必填、`--fixed-status`(0非固定/1固定) |
| 检查电子券二维码有效性 | `openydt coupon check-coupon-qr-code-valid-status` | 读 | `--coupon-code` 必填、`--origin` 必填 |
| 打印电子券（标记已发放） | `openydt coupon print-coupon` | 写 | `--trader-code` 必填、`--coupon-sn` 必填 |
| 查询电子券打印记录 | `openydt coupon query-coupon-print-record` | 读 | `--coupon-code` 必填 |
| 锁定电子券 | `openydt coupon lock-coupon` | 写 | `--url` 必填 |
| 回收电子券 | `openydt coupon cancel-coupon` | 写 | `--trader-code` 必填、`--coupon-sn` 必填 |
| 查询已发放电子券（按车牌/卡号） | `openydt coupon query-car-code-valid-coupon` | 读 | `--park-code` 必填、`--car-code`、`--card-code`、`--coupon-sn` |
| 查询电子券信息 | `openydt coupon query-coupon` | 读 | `--trader-code` 必填、`--query-type` 必填(0指定/1全部/2可发放)、`--page` 必填、`--page-size` 必填、`--coupon-code-list` |
| 查询可用的电子券 | `openydt coupon query-usable-coupon` | 读 | `--trader-code` 必填、`--sell-bill-id` 必填 |
| 根据券编码查询适用车场 | `openydt coupon query-coupon-available-park-by-coupon-code` | 读 | `--coupon-code` 必填、`--fixed-status` 必填(0非固定/1固定) |
| 根据券模板代码查询券模板 | `openydt coupon query-coupon-template-by-coupon-code` | 读 | `--code` 必填(商家券模板代码) |
| 根据券编码查询券模板 | `openydt coupon query-coupon-template-by-coupon-sn` | 读 | `--coupon-sn` 必填 |
| 根据券编码查询商家信息 | `openydt coupon query-trader-info-by-coupon-code` | 读 | `--coupon-sn` 必填 |
| 查询电子券售卖记录 | `openydt coupon query-trader-coupon-sell-record` | 读 | `--sell-begin-time` 必填、`--sell-end-time` 必填、`--page-num`、`--page-size`、`--body`(parkCodeList) |
| 查询商家发券记录 | `openydt coupon get-trader-coupon-grant-record-list` | 读 | `--trader-id` 必填、`--begin-time` 必填、`--end-time` 必填、`--page-num` 必填、`--page-size` 必填、`--body`(parkCodeList)、`--need-coupon-park`、`--need-coupon-status` |
| 查询优惠券列表（按车牌） | `openydt coupon get-trader-coupon-list` | 读 | `--page-size`、`--page-num`、`--body`(parkCodes、carNos/carPlateList 二选一) |

> 所有**写**命令（`create-trader` / `edit-trader` / `frozen-trader` / `delete-trader` / `create-coupon-template` / `create-coupon` / `create-fixed-coupon` / `sell-coupon` / `send-coupon` / `send-coupon-by-coupon-code` / `sync-scan-coupon-qr-code` / `print-coupon` / `lock-coupon` / `cancel-coupon`）执行时**必须加 `--yes`** 确认，否则会被拦截。

## 业务流程

### 电子券闭环（建商家+建模板 → 售卖 → 发券 → 查询 → 回收）

逐步执行，**务必把前序命令响应里的字段作为后续命令入参**（商家编码 traderCode、券模板编码 traderCouponTemplateCode、销售账单 sellBillId、券唯一编号 couponSn），不要凭空构造：

1. **创建商家 + 创建券模板**（写，均需 `--yes`）：
   ```
   openydt coupon create-trader --yes \
     --trader-name <商家名> --contact <联系人> --phone <手机号> \
     --park-code <车场编号> --login-account <账号> --password <密码>

   openydt coupon create-coupon-template --yes \
     --name <券名> --sell-from ... --sell-to ... --grant-from ... --grant-to ... \
     --valid-minute 60 --balance-type 0 --coupon-type 1 \
     --face-value 500 --original-price 600 --real-price 500 \
     --use-rule-from 0 --use-rule-to 1000
   ```
   - 从 `create-trader` 响应取**商家编码 `traderCode`**（也可用 `get-trader-list` 反查）→ 作为后续所有命令的 `--trader-code`；
   - 从 `create-coupon-template` 响应取**券模板编码**（商家券模板代码）→ 作为第 2 步 `sell-coupon` 的 `--trader-coupon-template-code`（也可用 `query-coupon-template-by-coupon-code --code <模板代码>` 核对）。

2. **售卖给商家**（写，需 `--yes`）— 用第 1 步拿到的商家编码 + 模板编码：
   ```
   openydt coupon sell-coupon --yes \
     --trader-coupon-template-code <来自 create-coupon-template> \
     --trader-code <来自 create-trader> \
     --sell-num 100 --sell-money 0.01 --sell-time "2018-04-16 09:00:00"
   ```
   - 从 `sell-coupon` 响应取**销售账单 `sellBillId`** → 作为第 3 步 `send-coupon` 的 `--sell-bill-id`（也可用 `query-usable-coupon --trader-code <traderCode> --sell-bill-id <id>` 核对该批次可发放余量）。

3. **发放给车辆**（写，需 `--yes`）— 用第 2 步拿到的 `sellBillId` + 商家编码：
   ```
   openydt coupon send-coupon --yes \
     --trader-code <来自 create-trader> \
     --sell-bill-id <来自 sell-coupon> \
     --car-code <车牌> --car-code-color 1
   ```
   > 也可按券编码发券：`send-coupon-by-coupon-code --yes --coupon-code <券编码> --car-no <车牌>`（`--car-no` 与 `--card-no` 不能同时为空）。发券前可先 `check-coupon-whether-send-available --coupon-code <券编码>` 确认是否可发。

4. **查询券**（读）— 确认券已落到车辆 / 商家：
   ```
   openydt coupon query-car-code-valid-coupon --park-code <车场> --car-code <车牌>
   openydt coupon query-coupon --trader-code <traderCode> --query-type 1 --page 1 --page-size 10
   ```
   - 从查询响应取**券唯一编号 `couponSn`** → 作为第 5 步 `cancel-coupon` 的 `--coupon-sn`（按券编码反查商家用 `query-trader-info-by-coupon-code --coupon-sn <couponSn>`）。

5. **回收**（写，需 `--yes`）— 用第 4 步取到的 `couponSn` + 商家编码：
   ```
   openydt coupon cancel-coupon --yes --trader-code <traderCode> --coupon-sn <来自 query>
   ```

## 示例

创建商家（写操作，必须加 `--yes`；参数取自 catalog sampleBody）：

```
openydt coupon create-trader --yes \
  --trader-name 测试商家 --trader-type 5 \
  --contact 联系人 --phone 13800000000 \
  --park-code 2KNTYVWC --login-account trader001 --password 123456
```

售卖电子券给商家（写操作，必须加 `--yes`；模板编码 / 商家编码取自前序响应）：

```
openydt coupon sell-coupon --yes \
  --trader-coupon-template-code GCSH3FI1YNDN \
  --trader-code NWTSZY49BH67 \
  --sell-num 100 --sell-money 0.01 --sell-remark 测试 \
  --sell-time "2018-04-16 09:00:00"
```

按车牌查询已发放电子券（读操作）：

```
openydt coupon query-car-code-valid-coupon --park-code 2KKN6111 --car-code 粤B88888
```
