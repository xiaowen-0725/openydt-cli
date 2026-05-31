---
name: openydt-list
version: 1.0.2
description: "车牌名单管理域：黑名单(blacklist 禁入/高收费)、白名单(redlist 免费放行规则)、访客(visitor 限时来访登记)的增删查。当用户要拉黑/解黑、配置警车等放行规则、登记或取消访客车时使用。名单引用的「特殊车辆类型ID(specialCarTypeId)」由 ticket 域创建(openydt-monthticket)，本域仅作入参引用、不负责创建。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt blacklist --help / openydt visitor --help / openydt redlist --help"
---

# openydt-list — 车牌名单管理域 (blacklist / redlist / visitor)

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**（认证 / profile / 签名 / 状态码 / 限速 / 安全规则）。未读共享基座不要执行任何命令。

## 何时用本技能

当用户要在停车场维护**车牌名单**时使用本技能，包括三类业务：

- **黑名单（blacklist）**：禁止某车辆进场 / 收高额费用。说法如「拉黑这台车」「加黑名单」「查黑名单列表」「解除黑名单」。
- **白名单（redlist）**：放行规则，免费 / 特权通行。说法如「加白名单」「警车放行」「白名单规则」「删除放行规则」。
- **访客（visitor）**：临时来访登记，限时通行。说法如「登记访客车」「访客放行」「取消访客预约」。

意图路由：
- 「加黑 / 拉黑 / 查黑名单 / 解除黑名单」→ `openydt blacklist ...`
- 「加白 / 白名单规则 / 删除白名单」→ `openydt redlist ...`
- 「访客登记 / 取消访客」→ `openydt visitor ...`
- 「创建特殊车辆类型 / VIP 分组」→ `openydt ticket add-special-car-type`（前置步骤，见业务流程；详见 [[openydt-monthticket]]）

> 注意：本技能命令分布在 **blacklist / redlist / visitor** 三个子命令域，调用前缀各不相同。

## 可用命令

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| 添加黑名单车辆 | `openydt blacklist add-black-list-car` | 写 | `--park-code` `--special-car-type-id` `--car-code` `--car-owner` `--reason`（均必填） |
| 查询黑名单车辆列表 | `openydt blacklist get-park-black-list` | 读 | `--body`(parkCodeList 必填) / `--car-code` `--owner` `--page-size` `--page-num` |
| 取消黑名单车辆 | `openydt blacklist remove-black-list-car` | 写 | `--park-code`(必填) `--blacklist-id` / `--car-no`（二选一） |
| 新增白名单规则 | `openydt redlist red-list-add` | 写 | `--redlist-param` `--park-code-list`（必填）`--plate-color` `--operator` `--remark` |
| 查询白名单规则 | `openydt redlist get-red-list` | 读 | `--park-code-list`(必填) |
| 删除白名单规则 | `openydt redlist del-red-list` | 写 | `--rule-id`(必填，来自查询结果) |
| 添加访客车辆 | `openydt visitor add-visitor-car-new` | 写 | `--park-code` `--car-no` `--owner` `--special-car-type-id` `--visit-from` `--visit-to`（均必填） |
| 取消访客车辆 | `openydt visitor cancel-visitor-car-new` | 写 | `--park-code`(必填) `--visitor-id` / `--car-no`（二选一） |

> 所有**写操作**（add/remove/del/cancel）均需追加 `--yes` 确认。读操作（get-*）无需 `--yes`。

## 业务流程

黑名单 / 访客登记必须先有「特殊车辆类型」，其 `specialCarTypeId` 是后续命令的必填入参。完整闭环如下，**务必用前序命令响应里的字段作为后续命令入参**：

1. **创建特殊车辆类型**（前置，属 ticket 域）
   `openydt ticket add-special-car-type --yes`
   - 黑名单用 `vipGroupType=2`，访客用 `vipGroupType=1`。
   - 从响应中取回的 **`specialCarTypeId`** 是步骤 2 的必填入参（黑名单 `--special-car-type-id` / 访客 `--special-car-type-id`）。

2. **登记车辆**（用步骤 1 的 `specialCarTypeId` + 车牌 + parkCode）
   - 黑名单：`openydt blacklist add-black-list-car --yes`
   - 访客：`openydt visitor add-visitor-car-new --yes`

3. **查询确认**
   - 黑名单：`openydt blacklist get-park-black-list`，从返回列表中取回 **`blacklistId`**，供步骤 4 精确取消使用。
   - 白名单：`openydt redlist get-red-list`，从返回中取回 **`ruleId`**，供 `del-red-list` 使用。

4. **清理 / 取消**
   - 黑名单：`openydt blacklist remove-black-list-car --yes`（用步骤 3 的 `--blacklist-id`，或仅传 `--car-no` 取消该车牌全部黑名单）
   - 访客：`openydt visitor cancel-visitor-car-new --yes`（用 `--visitor-id`，或仅传 `--car-no` 取消最新一次访客）
   - 白名单规则：`openydt redlist del-red-list --yes --rule-id <步骤3的ruleId>`

白名单规则相对独立，无需特殊车辆类型，直接用 `openydt redlist red-list-add` 新增（支持单车牌或 `*警` 这类通配规则）。

### 写入幂等与确认

- `add-black-list-car`/`add-visitor-car-new` 重复回传同车牌：平台按车牌去重（同车牌重复加黑不新建条目）；不确定首次是否生效先 `get-park-black-list` 查。
- ⚠️ `remove-black-list-car`/`cancel-visitor-car-new` **仅传 `--car-no`（不带 id）会取消该车牌全部条目**——批量影响，执行前先 `get-park-black-list` 确认范围，优先用查询拿到的 `blacklistId`/`visitorId` 精确取消。
> PII：`--phone`/`--car-no` 是 PII，prod 不记真实值（见 [[openydt-shared]]）；特殊车辆类型创建见 [[openydt-monthticket]]。

## 错误自愈速查（黑白名单/访客）

| 现象 | 含义 | 恢复动作 |
| --- | --- | --- |
| 加黑名单失败 `status=7`/`resultCode=909` | `specialCarTypeId` 未传/不属本车场/类型不匹配 | 先 `openydt ticket get-special-car-type-list` 查本车场的类型 ID，确认 `vipGroupType=2` 后再加黑 |
| 访客登记失败 `status=2` | `specialCarTypeId` 属黑名单类型（vipGroupType=2）或未授权 | 确认用访客类型（vipGroupType=1）的 `specialCarTypeId` |
| 取消黑名单/访客后疑似未生效 | 仅传 `--car-no` 但平台侧查询有延迟 | 等 1-2 秒后 `get-park-black-list` 复核；若存在多条同车牌条目改用 `--blacklist-id` 精确取消 |
| `get-park-black-list` 返回 0 条 | 车场无黑名单 **或** `parkCodeList` 未传全 | 确认 `parkCodeList` 正确后再下「无黑名单」结论；见 [[openydt-shared]] 结果解读 SOP |
> 通用状态码与重试规则见 [[openydt-shared]]；结果三层判读见其 `../openydt-shared/references/result-reading-sop.md`。

## 示例

> 下列 parkCode/时间为文档化测试值（仅 test；`1ZS7H5PQH9`/`PTD2YBBZ` 为测试车场，生产环境替换为授权车场）。写操作建议先 `--dry-run` 预览、确认后再 `--yes`。

1. 添加一条黑名单车辆（写，需 `--yes`；`special-car-type-id` 来自 `add-special-car-type` 的响应）：

```bash
openydt blacklist add-black-list-car --yes \
  --park-code 1ZS7H5PQH9 \
  --car-code 粤EJW962 \
  --car-owner 车主 \
  --reason 原因 \
  --special-car-type-id 253
```

2. 查询某停车场黑名单列表（读，免 `--yes`，用 `--body` 传 parkCodeList 数组）：

```bash
openydt blacklist get-park-black-list \
  --body '{"parkCodeList":["1ZS7H5PQH9"],"carCode":"粤EJW962","pageNum":1,"pageSize":10}'
```

3. 新增白名单放行规则（写，需 `--yes`，park-code-list 为数组用 `--body`）：

```bash
openydt redlist red-list-add --yes \
  --body '{"parkCodeList":["1ZS7H5PQH9","PTD2YBBZ"],"redlistParam":"粤EJW962"}'
```

4. 登记访客车辆（写，需 `--yes`；`special-car-type-id` 来自访客类型的 `add-special-car-type` 响应）：

```bash
openydt visitor add-visitor-car-new --yes \
  --park-code PTD2YBBZ \
  --car-no 粤EJW962 \
  --owner 李四 \
  --phone 13800000000 \
  --reason 访友 \
  --visit-from 20260601000000 \
  --visit-to 20260602000000 \
  --special-car-type-id 154
```
