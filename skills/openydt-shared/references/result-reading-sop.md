# 结果解读契约 (result-reading SOP)

> openydt 命令统一返回包络 `{data, message, resultCode, status}`。读懂结果是 Agent 正确决策的前提。本契约是硬规则;各域技能遇结果解读疑问回到这里。

## 三层判读(必须按顺序看)
1. **`status`(传输/鉴权/业务总判)**:1=成功;2=业务失败(看 `resultCode`);3=系统异常(可重试);4/5/6=签名/key/未授权;7=参数不完整;9=接口不存在。**status≠1 一律不是成功。**
2. **`resultCode`(status=2 时的业务码)**:901-912 / 1801,含义见 [status-codes.md](status-codes.md) 的「响应包络与状态码」表。如 904 车场不存在、907 账单已同步(=幂等命中,见 [`write-idempotency.md`](write-idempotency.md))、912 查费超时需重新查费。
3. **`data` 内层("伪成功"陷阱)**:**`status=1` ≠ 业务动作成功**——某些接口外层 status=1,但 `data.code`/`data.msg` 才是真实结果。已知:`scanChannelCodeInOut` 通道无车时外层 status=1 而 `data.code≠0`(如 8 当前通道没有车辆);`get-park-fee` status=1 但 data 可能为空(无在场/已离场)。**对这类接口必须查 data 内层,别只看外层 status。**

## 金额单位:一律是「元」(小数)
- `shouldPayValue`/`paidValue`/`actPayCharge`/`couponValue`(trade)、`originPrice`/`favorPrice`/`refundPrice`/`price`(ticket)、金额型 `faceValue`(coupon)、`value.fee`(park chargeMap)、data 域账单金额——**单位都是元**。`shouldPayValue:1` = 1.00 元,**不是 1 分**。别把 1 当 1 分去付 0.01(只缴一分、仍欠 0.99)。
- 例外量纲:coupon **时间券** 的 faceValue 单位是「分钟」(由 `balanceType` 区分金额券/时间券);data 域 `minuteInterval` 单位是分钟。读金额型字段先确认量纲。

## 空结果 ≠ 不存在(防误判)
- **0 条 / `data` 空 / nodata ≠「业务对象不存在」**。常见诱因:必填筛选项没给全(如 `get-park-on-site-car` 不传 `enterTimeFrom/To` 返 0 条)、test 环境多数 dataAnalysis 接口 nodata、车已离场。**先确认必填筛选项给全,再下「无」的结论。**
- 判「是否在场」:以 `get-park-on-site-car`(在场)/`get-car-out-list`(已离场)为准,不要只看 `get-park-detail`(两者可能不一致)。

## 分页:别拿一页当全量
- 带 `pageNum`/`pageSize` 的查询,**单页 ≠ 全量**。需要全量时优先用 `--all-pages --out <file>.ndjson`，CLI 会按最大页尺寸顺序翻到尽；同一查询只导出一次，后续计算复用本地文件。给「共 N 条 / 全部」结论前，核对导出完成进度与最终记录数。

## Final Answer Check(回话前自检)
1. `status==1`?(否→按三层判读处置,见 [`write-idempotency.md`](write-idempotency.md) 与各域错误自愈表) 2. 该接口要不要看 `data` 内层? 3. 金额量纲对了吗(元)? 4. 「0 条」是真没有,还是筛选没给全? 5. 分页是否已覆盖全量?
