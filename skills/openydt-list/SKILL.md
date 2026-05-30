---
name: openydt-list
version: 1.0.1
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
- 「创建特殊车辆类型 / VIP 分组」→ `openydt ticket add-special-car-type`（前置步骤，见业务流程）

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

## 示例

> 下列 parkCode/时间为占位示例；实际替换为你的授权车场与当前/未来时间（测试环境车场见 shared）。写操作建议先 `--dry-run` 预览、确认后再 `--yes`。

1. 添加一条黑名单车辆（写，需 `--yes`；`special-car-type-id` 来自 `add-special-car-type` 的响应）：

```bash
openydt blacklist add-black-list-car --yes \
  --park-code 2KNTYVWC \
  --car-code 粤YKK123 \
  --car-owner 车主 \
  --reason 原因 \
  --special-car-type-id 253
```

2. 查询某停车场黑名单列表（读，免 `--yes`，用 `--body` 传 parkCodeList 数组）：

```bash
openydt blacklist get-park-black-list \
  --body '{"parkCodeList":["2KNTYVWC"],"carCode":"粤YKK123","pageNum":1,"pageSize":10}'
```

3. 新增白名单放行规则（写，需 `--yes`，park-code-list 为数组用 `--body`）：

```bash
openydt redlist red-list-add --yes \
  --body '{"parkCodeList":["2KNTYVWC","2KNTYVCC"],"redlistParam":"粤YKK123"}'
```

4. 登记访客车辆（写，需 `--yes`；`special-car-type-id` 来自访客类型的 `add-special-car-type` 响应）：

```bash
openydt visitor add-visitor-car-new --yes \
  --park-code 2KKN885S \
  --car-no 粤YGW982 \
  --owner 李四 \
  --phone 13596156884 \
  --reason 访友 \
  --visit-from 20161214163930 \
  --visit-to 20161215163930 \
  --special-car-type-id 154
```
