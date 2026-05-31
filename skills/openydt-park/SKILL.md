---
name: openydt-park
version: 1.0.2
description: "车场信息域(park)：车场列表与编码、基本/系统/区域/通道/云车场设备信息、空车位与剩余车位、车场收费标准查询与其他车型计费测算、车辆免费停车信息、查某车应显示的屏显语音内容、车辆优惠券记录(只读)。当用户要按 parkCode 查车场属性、车位余位、收费标准时使用。边界：实时算费/缴费见 trade(openydt-billing)，向设备下发屏显/播报见 device，券的发放/回收见 coupon。"
metadata:
  requires:
    bins: ["openydt"]
  cliHelp: "openydt park --help"
---

# openydt-park — 车场信息域

> **CRITICAL：开始前 MUST 先用 Read 工具读取 [`../openydt-shared/SKILL.md`](../openydt-shared/SKILL.md)**，掌握认证 / profile / 签名 / 状态码 / 限速 / 安全规则后再执行任何 `openydt park` 命令。

## 何时用本技能

当用户的诉求落在「停车场静态/动态信息」上时使用本域，例如：

- 查停车场：列出授权商所有车场、看单个车场的基本信息 / 系统信息 / 区域 / 通道 / 云车场设备。
- 查车位：单个车场空车位、区域内（按经纬度半径）空车位、单个车场实时剩余车位、剩余车位 + 免费时长。
- 查费 / 计费：根据车场编码取收费信息（查费）、对某收费标准做其他车型计费测算。
- 车辆相关：车辆免费停车信息、车辆屏显及语音、车辆优惠券（电子券）记录。
- 授权 / 维护：取该账号下所有授权车场编码；设置（上报）车场实时车位（写操作）。

意图路由：
- 涉及「订单/账单/缴费/月票」→ 用 `openydt trade`。
- 涉及「在场车/进出场/锁车/放行」→ 用 `openydt parking`。
- 涉及「黑/白名单、访客、电子券核销、设备」→ 分别用 `openydt blacklist` / `openydt redlist` / `openydt visitor` / `openydt coupon` / `openydt device`。
- 「车场本身的属性、车位、车场维度查费」→ 留在本域 `openydt park`。

本域以**查询为主、无强依赖链**：各查询命令大多只需要 `parkCode`（停车场编号），可独立调用。仅 `get-park-ydt-other-car-type-charge-info` 需要先调用查费命令拿到结果对象后回填（见业务流程）。

## 可用命令

`<use>` 为命令真实 kebab 名，调用形如 `openydt park <use>`。所有命令支持 `--body '<json>'` 直接传完整请求体；列出的字段也提供同名 flag（flag 会合并覆盖 `--body`）。

| 中文名 | 命令 | 读/写 | 关键参数 |
| --- | --- | --- | --- |
| 获取授权商所有停车场信息 | `openydt park get-park-list` | 读 | （无） |
| 获取该账号下所有授权车场编码 | `openydt park get-auth-park-codes` | 读 | （无） |
| 获取停车场全部信息 | `openydt park get-all-park-info` | 读 | `--park-code`、`--update-time-from`、`--update-time-to`（均可选） |
| 获取单个停车场基本信息 | `openydt park get-park-info` | 读 | `--park-code`（可选） |
| 获取停车场系统信息(VEMS/云车场) | `openydt park get-park-system-info` | 读 | `--park-code`（可选） |
| 获取车场区域信息 | `openydt park get-park-area-info` | 读 | `--park-code`* |
| 获取云停车场通道信息 | `openydt park get-channel-info` | 读 | `--park-code`* |
| 获取云停车场设备信息 | `openydt park get-cloud-park-device-info` | 读 | `--park-code`* |
| 获取单个停车场空车位信息 | `openydt park get-ept` | 读 | `--park-code`* |
| 获取区域内所有停车场空车位信息 | `openydt park get-area-ept` | 读 | `--longitude`*、`--latitude`*、`--radius`* |
| 获取单个停车场实时车位信息 | `openydt park get-park-remain-carport` | 读 | `--park-code`*、`--area-code`、`--area-id` |
| 获取停车场剩余车位和免费时长 | `openydt park get-park-remain-carport-and-free-time` | 读 | `parkCodes`*（车场编号数组，用 `--body`） |
| 根据车场编码获取收费信息(查费) | `openydt park get-park-ydt-charge` | 读 | `--park-code`*、`--start-date-time` |
| 其他车型计费测算 | `openydt park get-park-ydt-other-car-type-charge-info` | 读 | `--park-code`*、`--standard-seq`*、`--car-type`*、`--start-time`*、`parkYdtChargeVo`*（用 `--body`） |
| 获取车辆免费停车信息 | `openydt park get-car-free-parking-info` | 读 | `--park-code`*、`--parking-code`*、`--car-code`、`--vip-type`、`--is-support-special-charge-rule`、`--is-support-trigger-event` |
| 获取车辆屏显及语音 | `openydt park get-display-voice-by-car-code` | 读 | `--park-code`*、`--car-code`*、`--vip-type`*、`channelCodeList`*（用 `--body`）、`--feature-pass-time`、`--whether-charge` |
| 获取车辆优惠券信息列表 | `openydt park get-car-coupon-record` | 读 | `--page-num`*、`--page-size`*、`carPlateList`*（车牌列表，用 `--body`） |
| 设置停车场实时车位信息 | `openydt park set-park-remain-carport` | 写（需 `--yes`） | `--park-code`*、`remainCarportList`*（区域车位数组，用 `--body`） |

> 备注：标 `*` 为必填。`carPlateList` / `channelCodeList` / `parkCodes` / `parkYdtChargeVo` / `remainCarportList` 等数组/对象型字段没有独立 flag，需通过 `--body` JSON 传入。适用场景：标注「云停车场」的命令（区域、通道、云车场设备、查费、计费测算、设置车位）仅对云车场有效，其余 VEMS 传统车场与云车场通用。

## 业务流程

### 查费 → 其他车型计费测算（唯一需要回填上一步响应的链路）

1. 先查费：`openydt park get-park-ydt-charge --park-code <parkCode>`，从响应中拿到查费信息对象（即 `parkYdtChargeVo`，含 `chargeTypeSeq` / `parkSysType` / `parkYdtChargeStandardVoList` 等，其中每个收费标准带有计费规则 ID `standardSeq` 与车辆类型 `carType`）。
2. 再测算：把上一步整个查费返回结果作为 `parkYdtChargeVo` 字段、并指定要测算的 `standardSeq`（收费标准 ID）、`carType`（计费规则的车辆类型，即「特殊车辆类型 ID」）、`startTime`（计费开始时间），调用：
   `openydt park get-park-ydt-other-car-type-charge-info --body '{...}'`
   - 强调：`standardSeq`、`carType`、`parkYdtChargeVo` 必须取自上一步 `get-park-ydt-charge` 的响应字段，不可臆造。

> **解读 `get-park-ydt-charge` 返回（实测，避免误读）**：它返回的是车场**计费组**（`parkYdtChargeStandardVoList`，一个车型可挂多套标准），**不是计费规则原文**。每套标准里：
> - `standardSeq` = **计费规则 ID**（不是普通自增序号，测算时回填的就是它）；`type`：0 自定义 / 1 免费 / 2 循环递增 / 3 按次固定。
> - `chargeMap` 是**费用预览估算，不是规则本身**：key 为停车**时长档位（1/2/3/4/8 小时**这 5 个离散点），`value.fee` = 停该时长的**应缴总额（单位：元）**。它以**当前时刻**为进场时间试算，跨昼夜 / 工作日（顶层 `workDay`）结果会变；看不到时段边界、递增步长、封顶明细。`stoppingTimeStr` 恒为空。
> - 想看“停 1/2/3/4/8 小时各多少钱”用 `chargeMap` 足够；要精确算某车某时段费用，请用 trade 域实时查费（`[[openydt-billing]]` 的 `get-park-fee`），别拿 `chargeMap` 当精确账单。
>
> **呈现给用户时，别甩原始 JSON，按这个风格转成人话**（每套标准 = 车型 + `type` 通俗说法 + 用 `chargeMap` 讲“停多久多少钱” + 封顶/免费，末尾带估算提醒）：
>
> > 车场 PTD2YBBZ 的收费（工作日）：
> > - **小型车**（挂了 2 套）：① 按次固定 **1 元/次**、封顶 25 元；② 自定义阶梯 **约 60 元/小时**（1h≈60、4h≈240、8h≈480）。
> > - **大型车**：按次 **0.03 元/次**（像测试值）。
> > - 新能源小车/大车同理（见各套 `type` 与 `chargeMap`）。
> >
> > ⚠️ 以上是按“现在进场”试算的预览，只覆盖 1/2/3/4/8 小时这几档；某辆车某时段的准确金额请用实时查费 `get-park-fee`。
>
> 要点：`type` 翻成「按次固定 / 每小时 X 元 / 免费 / 自定义」这种用户能懂的话；金额带「元」；多套标准分条列；务必保留末尾的估算提醒。

其余命令均为独立查询：拿到 `parkCode`（可先用 `get-park-list` / `get-auth-park-codes` 获取授权车场编码）后即可直接调用对应查询。

## 示例

1) 列出账号下所有授权车场（无参数）：

```bash
openydt park get-park-list
```

2) 查某车场的实时剩余车位（参数取自 catalog sampleBody）：

```bash
openydt park get-park-remain-carport --park-code 2KNTYVWC
```

3) 按经纬度 + 半径查区域内所有车场空车位：

```bash
openydt park get-area-ept --longitude 113.158978 --latitude 23.046084 --radius 200
```

4) 写操作示例——设置车场实时车位（**写操作必须加 `--yes` 确认**）：

```bash
openydt park set-park-remain-carport --yes --body '{
  "parkCode": "2KNTYVWC",
  "remainCarportList": [
    {"areaId": 1, "totalRemain": 1, "tempRemain": 1, "fixedRemain": 1}
  ]
}'
```
