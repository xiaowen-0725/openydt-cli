# openydt-cli

广东艾科智泊 **智慧停车开放平台** 的命令行工具 —— 为人和 AI Agent 而生。

把开放平台的对外接口(查费/缴费/车场/停车记录/月票/电子券/设备控制等)封装成命令行,
自动处理签名鉴权(v2/v3)、多授权商 profile、多环境(test/dev/prod),并附带一套
AI Agent **Skills**,让智能体零额外配置即可操作平台。

> 形态参照飞书 CLI:Go + Cobra,接口元数据驱动(`catalog.json` → 自动生成命令与技能)。

## 安装 / 构建

```bash
make build          # 产出 bin/openydt
# 或
go build -o bin/openydt .
```

从开放平台文档前端重新生成接口目录与命令:

```bash
make catalog        # 解析 open-api-front 的 Doc/*.vue → catalog/catalog.json
make generate       # catalog.json → cmd/gen/*.go(各域命令)
```

## 快速开始

```bash
# 1. 配置授权商凭据(测试环境)
openydt config set --profile demo --key test --secret 123456 --env test

# 2. 验证凭据/签名链路
openydt auth test

# 3. 查费(生成的一等命令,带类型化参数)
openydt trade get-park-fee --car-code 粤EJW962 --park-code 1ZS7H5PQH9 -o table

# 4. 通用兜底:调用任意业务编码 cmd
openydt api getParkOnSiteCar --body '{"parkCodeList":["PTD2YBBZ"],"pageNum":1,"pageSize":10}'
```

## 三层命令体系

1. **域一等命令** `openydt <域> <命令>` —— 由 `catalog.json` 自动生成,带参数校验、`--help`、写操作 `--yes` 守护。
2. **通用调用** `openydt api <cmd> --body '{...}'` —— 兜底任意可调用接口(含未做成一等命令的)。
3. **配置/鉴权** `openydt config|auth`。

内置 10 个业务域,共 **143** 条一等命令:

| 域 | 命令 | 数量 |
|---|---|---|
| `trade` | 停车缴费(查费/缴费/补缴/预存) | 7 |
| `park` | 车场信息 / 计费测算 | 18 |
| `parking` | 停车记录 / 在场 / 进出 / 锁车 | 31 |
| `device` | 设备控制(开关闸/显示屏/扫码) | 11 |
| `ticket` | 月票 / VIP / 特殊车辆类型 | 29 |
| `blacklist` | 黑名单 | 3 |
| `redlist` | 白名单 | 3 |
| `visitor` | 访客 | 2 |
| `data` | 数据分析 | 9 |
| `coupon` | 电子券 / 商户 | 30 |

其余模块(城市运营、第三方车场接入、积分等)及 webhook 回调接口未生成一等命令;
可调用类用 `openydt api <cmd>` 调用,webhook(平台主动推送)需自建接收端,详见 `openydt-api-explorer` 技能。

## 全局参数

| flag | 说明 |
|---|---|
| `--profile` | 授权商 profile(默认当前) |
| `--env` | 环境 test\|dev\|prod(默认 test) |
| `-o, --output` | 输出 json\|table |
| `--sign` | 签名版本 v2\|v3(默认 v2) |
| `-y, --yes` | 确认执行写操作 |
| `--dry-run` | 只打印将发送的签名请求,不实际发送 |
| `-v, --verbose` | 调试信息 |

环境变量可覆盖 profile:`OPENYDT_PROFILE / OPENYDT_KEY / OPENYDT_SECRET / OPENYDT_ENV / OPENYDT_SIGN`。

## 签名

- 请求:`POST {base}/openydt/api/v3/{cmd}?sign=...`,头 `Authorization: Base64(key:时间)`、`Content-Type: application/json;charset=utf-8`。
- **v2**(默认):`sign = md5(key + ":" + 时间 + ":" + secret)`,不含 body。
- **v3**:`sign = md5(key + ":" + 时间 + ":" + 紧凑body + ":" + secret)`,含 body。
- 时间格式 `yyyyMMddHHmmss`(本地时间),有效期 10 分钟。
- 客户端内置自定义 `User-Agent`、重试 + 指数退避(吸收网关间歇 404 / 连接重置)、限速友好。

> ⚠️ 实测:测试环境的 `test` key **仅接受 v2 签名**;`--sign v3` 会返回 `status=4 签名错误`(除非平台对该 key 开通 v3)。v3 实现遵循官方算法,生产环境如开通即可用。

## AI Agent Skills

`skills/` 下 11 个技能(对标飞书 CLI):

| Skill | 说明 |
|---|---|
| `openydt-shared` | 配置/profile/签名/状态码/限速/安全(所有域技能先读它) |
| `openydt-billing` | 停车缴费闭环(进车→查费→缴费→查单) |
| `openydt-park` | 车场信息 / 计费测算 |
| `openydt-record` | 停车记录 / 在场 / 锁车 |
| `openydt-device` | 设备控制(开关闸/显示屏) |
| `openydt-monthticket` | 月票 / VIP 闭环(建类型→开通→续费/退费) |
| `openydt-list` | 黑名单 / 白名单 / 访客 |
| `openydt-data` | 数据分析 |
| `openydt-coupon` | 电子券闭环(建商家+模板→售卖→发放→回收) |
| `openydt-api-explorer` | 用 `api` 调用未封装接口 |
| `openydt-skill-maker` | 创建自定义 skill 的框架 |

格式校验:`node scripts/skill-format-check/index.js`。

## 测试与验证

```bash
make test           # 单元测试(含签名锚点向量 v2/v3/Base64)
make smoke          # 对测试环境查费冒烟
make e2e            # 端到端全量验证 → TEST_REPORT.md
```

端到端验证按业务流程链(缴费/月票/黑名单访客/电子券)驱动每个纳入接口,产出 `TEST_REPORT.md`。

## 状态码

响应包络 `{data, message, resultCode, status}`。`status`:1成功 / 2业务失败 / 3系统异常 / 4签名错误 / 5key错误 / 6未授权 / 7参数不完整 / 9接口不存在。`status=2` 时看 `resultCode`(901-912、1801)。退出码:0成功 / 1业务失败 / 2参数 / 4鉴权 / 5网络。

## 目录结构

```
cmd/            # root + api/auth/config + gen/(生成的各域命令)
internal/       # sign 签名 / client HTTP / config 配置 / output 输出 / catalog 目录 / cmdutil
catalog/        # catalog.json(接口目录,由 tools/extractor 生成)
tools/extractor # Node: 解析 Doc/*.vue → catalog.json
internal/gen    # Go codegen: catalog.json → cmd/gen
skills/         # AI Agent 技能
tests/e2e       # 端到端验证 harness
```
