---
name: openydt-api-explorer
version: 1.0.1
description: "通用 API 兜底与接口探索：当某接口没有专属一等命令时，用 openydt api <cmd> --body 直发任意 callable 接口，并从 catalog.json 查 cmd/参数、区分能主动调的 callable 与只能被动接收的 webhook。当某 cmd 在域技能里找不到专属子命令、要调未一等化的接口(如城市运营券/第三方车场/小区门禁/月票预约/授权访客/证件规则/车辆标签等)、或问『这个 cmd 怎么调 / 平台会不会回调我』时使用。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt api --help"
---

# openydt-api-explorer — 通用 API 兜底 / 接口探索

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**（认证 / profile / 签名 / 状态码 / 限速 / 安全规则）。`openydt api` 与一等命令走完全相同的签名、包络、退出码与限速逻辑，未读共享基座不要执行任何命令。

## 何时用本技能

平台共 423 个接口，只有约 143 个被做成了**域一等命令**（`openydt <域> <命令>`，参数已结构化为 flag）。当你要调的接口**没有专属子命令**时，用本技能的通用兜底：

```bash
openydt api <cmd> --body '{...}'
```

`api` 对**任意可调用（direction=callable）的业务编码 cmd** 自动签名并 `POST`，覆盖一等命令没点名的所有域，例如：

- `cityOperationCoupon` 城市运营券（创建/发放城市运营券模板）
- `thirdParkForBolian` 第三方车场缴费接入回执
- `upward` 上行数据上报回执（如 `asynSuccess`）
- `community` 小区门禁（如 `getAuthCommunities`）
- `ad` 广告统计、`preferential`/`score` 积分兑换、`invoice` 发票、`ydtUser` 用户认证等
- **有专属域、但子特性未一等化**的接口：月票预约/排队(`ticket·appointment`)、月票授权访客(`ticket·authorize`)、月票证件规则(`ticket·certificate`)、车辆标签(`parking·tag`：`addCarTags`/`delCarTags`)——monthticket / record 域技能会把这些指向本技能，同样用 `api` 调（写操作记得 `--yes`）

这些域在 `catalog.json` 里 `included=false`（没生成一等命令），但只要 `direction=callable` 就能用 `api` 直接调。

> 选择顺序：**先找一等命令**（`openydt <域> --help` 或对应域技能），找不到再用 `api` 兜底。一等命令把参数拆成了 flag、自动判定读写、更不易出错；`api` 是“原始 JSON 直发”，更通用但要你自己保证 body 正确。

## 用法：openydt api

```bash
# 1) 行内 JSON body（最常用）
openydt api getParkFee --body '{"parkCode":"1ZS7H5PQH9","carCode":"粤EJW962"}'

# 2) 无参接口可省略 body
openydt api getAuthParkCodes

# 3) 从文件读 body
openydt api getParkOnSiteCar --body-file ./body.json

# 4) 从 stdin 读 body（- 表示 stdin），适合管道 / 大 body
echo '{"parkCode":"PTD2YBBZ"}' | openydt api getParkOnSiteCar --body-file -
```

- `<cmd>`：业务编码，**就是 catalog 里的 `cmd` 字段**（如 `getParkFee`、`createCityOperationCouponTemplate`）。注意是 cmd，不是 `dir` 路径。
- `--body` 与 `--body-file` **互斥**；二者都不给则发送空 body（`{}`），仅适合无参接口。
- body 是**原始 JSON**：CLI 会先做 JSON compact 再用于签名与发送（与一等命令一致），字符串内部空格如 `"2019-04-16 00:11:25"` 会保留。
- 参数命名与嵌套结构**完全照搬接口定义**（见下文从 catalog 查参数）；写错字段名通常返回 `status=2 / resultCode=909 请求参数错误` 或 `status=7 请求参数不完整`。

### --dry-run 预览

不确定 body 或在 prod 环境前，先用 `--dry-run` 只打印**将发送的签名请求**（URL / sign / ts / compact 后的 body），不实际发送：

```bash
openydt api createCityOperationCouponTemplate --dry-run \
  --body '{"parkCodeList":["PRJ9YJ19"],"couponTemplate":{"name":"抵扣1元券","faceValue":1}}'
```

### 写操作必须 --yes（重要：api 不自动判定读写）

**`api` 命令本身不会判断 cmd 是读还是写**，是否需要确认完全交给你负责：

- 凡是会改变平台状态的 cmd（建/改/删、发券、缴费、开闸、上报回执等），**你必须显式加 `--yes`**，否则被安全拦截不执行。
- 一等命令会自己识别写操作并要求 `--yes`；`api` 不会替你识别，所以用 `api` 调写接口时**务必自己确认 readwrite 并加 `--yes`**。
- 判断某 cmd 读写：查 catalog 的 `readwrite` 字段（`read` / `write`），见下文。

```bash
# 写操作（catalog readwrite=write），必须 --yes
openydt api createCityOperationCouponTemplate --yes \
  --body '{"parkCodeList":["PRJ9YJ19"],"couponTemplate":{"name":"抵扣1元券","totalNum":2,"couponType":1,"faceValue":1,"validFrom":"2019-04-28 00:00:00","validTo":"2020-04-28 00:00:00"}}'
```

## 从 catalog 查可用 cmd 及参数

接口清单在 [`../../catalog/catalog.json`](../../catalog/catalog.json)（绝对路径 `/Users/zhoujw/develop/tmp/openydt-cli/catalog/catalog.json`）。顶层结构 `{generatedFrom, count, interfaces:[...]}`，每个 interface 对象关键字段：

| 字段 | 含义 |
| --- | --- |
| `cmd` | 业务编码，**直接作为 `openydt api <cmd>` 的 cmd** |
| `domain` / `dir` | 所属域 / 文档路径（仅供归类，**不是** api 的入参） |
| `direction` | `callable`=可主动调 / `webhook`=平台主动推送（见下节，**不能调**） |
| `readwrite` | `read` / `write`——决定调用时是否要加 `--yes` |
| `included` | 是否已做成一等命令；`false` 表示要用 `api` 兜底 |
| `excludeReason` | 未做成一等命令的原因（`out-of-scope-domain` / `deprecated` / `no-endpoint` / `vems-only` 等） |
| `params` | 参数定义数组：`name` / `required` / `type` / `desc` / `group`（`group` 非空表示该参数嵌在某个子对象里） |
| `sampleBody` | 官方示例请求 body——**构造 `--body` 的最佳起点，照抄字段名再改值** |
| `sampleResponse` | 示例响应，帮你预判返回字段 |

用 `jq` / `python3` 检索（不要在终端打印密钥，这里只查清单不涉及凭据）：

```bash
# 按 cmd 看完整定义（params + sampleBody）
jq '.interfaces[] | select(.cmd=="createCityOperationCouponTemplate")' catalog/catalog.json

# 列出某个未点名域所有“可调用”的 cmd 及读写
jq -r '.interfaces[] | select(.domain=="cityOperationCoupon" and .direction=="callable") | "\(.cmd)\t\(.readwrite)\t\(.explain)"' catalog/catalog.json

# 全量“没做成一等命令但可调用”的接口（included=false 且 callable）
jq -r '.interfaces[] | select(.included==false and .direction=="callable") | "\(.domain)\t\(.cmd)\t\(.readwrite)"' catalog/catalog.json

# 只看某 cmd 的参数清单（含嵌套 group）
jq '.interfaces[] | select(.cmd=="createCityOperationCouponTemplate") | .params' catalog/catalog.json
```

**included=false 但 direction=callable 的接口照样能用 `api` 调**——常见于城市运营券、第三方车场接入（thirdParkForBolian）、上行回执（upward）、小区门禁（community）、广告/积分/发票/ydtUser 等“未点名域”。`included=false` 只是“没生成专属命令”，不代表“不能调”；判断能否调用看 `direction`，判断要不要 `--yes` 看 `readwrite`。

构造流程：① `jq` 取目标 cmd 的 `sampleBody` → ② 照抄字段名、按 `params` 校对必填/类型/嵌套 → ③ `--dry-run` 预览签名请求 → ④ 写操作加 `--yes` 正式发。

## 不能调用的 webhook（平台主动推送）

catalog 中 `direction=webhook` 的接口（共 61 个，如 `reportParkinglotChange` 车位变更上报）是**平台主动 POST 到你方接收端**的回调，方向与 callable 相反：

- **CLI 不能主动调这些 cmd**——它们没有“你方请求平台”的入口，`openydt api <webhook-cmd>` 不是正确用法（平台不提供该方向端点）。
- 要接收这类推送，需**你自建一个 HTTP 接收端（webhook receiver）**，向平台登记回调地址，由平台在事件发生时把 `sampleBody` 形态的数据 POST 给你；你的服务负责验签、处理并按约定返回（很多上行回执对应一个 `upward` 域的 callable 确认 cmd，如 `asynSuccess`）。
- 区分方法：调用前先 `jq '.interfaces[]|select(.cmd=="<cmd>")|.direction'`，是 `webhook` 就别用 `api` 调，改为自建接收端；是 `callable` 才用 `api`。

## 示例

查某未点名域有哪些可调 cmd（读，纯查 catalog，不发请求）：

```bash
jq -r '.interfaces[] | select(.domain=="thirdParkForBolian" and .direction=="callable") | "\(.cmd)\t\(.readwrite)"' catalog/catalog.json
```

用 sampleBody 起手、先预览再发（写操作示例，正式发要加 `--yes`）：

```bash
# 1) 取示例 body
jq -r '.interfaces[]|select(.cmd=="createCityOperationCouponTemplate")|.sampleBody' catalog/catalog.json

# 2) 预览签名请求，确认无误
openydt api createCityOperationCouponTemplate --dry-run --body-file ./body.json

# 3) 正式发（写操作，必须 --yes）
openydt api createCityOperationCouponTemplate --yes --body-file ./body.json
```

读接口直接调（无需 --yes）：

```bash
echo '{}' | openydt api getAuthCommunities --body-file -   # community 域，readwrite=read
```
