---
name: openydt-park
version: 1.0.0
description: "智慧停车开放平台·车场域(park)：查车场/停车场列表、车场基本信息、车场系统信息、车场区域、通道信息、云车场设备、空车位/剩余车位、区域内空车位、剩余车位与免费时长、车场收费信息(查费)、其他车型计费测算、车辆免费停车信息、车辆屏显及语音、车辆优惠券/电子券记录、授权车场编码、设置实时车位。触发词：查车场、车场信息、停车场列表、车场编码、parkCode、空车位、剩余车位、车位余位、免费时长、查费、收费标准、计费测算、车型计费、免费停车、屏显语音、优惠券、电子券、授权车场、通道信息、云车场设备、区域信息、设置车位。"
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

1. 先查费：`openydt park get-park-ydt-charge --park-code <parkCode>`，从响应中拿到查费信息对象（即 `parkYdtChargeVo`，含 `chargeTypeSeq` / `parkSysType` / `parkYdtChargeStandardVoList` 等，其中每个收费标准带有自增 ID `standardSeq` 与车辆类型 `carType`）。
2. 再测算：把上一步整个查费返回结果作为 `parkYdtChargeVo` 字段、并指定要测算的 `standardSeq`（收费标准 ID）、`carType`（计费规则的车辆类型，即「特殊车辆类型 ID」）、`startTime`（计费开始时间），调用：
   `openydt park get-park-ydt-other-car-type-charge-info --body '{...}'`
   - 强调：`standardSeq`、`carType`、`parkYdtChargeVo` 必须取自上一步 `get-park-ydt-charge` 的响应字段，不可臆造。

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
