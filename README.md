# openydt-cli

广东艾科智泊 **智慧停车开放平台** 的命令行工具 —— 为人和 AI Agent 而生。

把开放平台的对外接口(查费 / 缴费 / 车场 / 停车记录 / 月票 / 电子券 / 设备控制等)封装成命令行,自动处理签名鉴权(v2/v3)、多授权商 profile、多环境(test/dev/prod),并附带一套开箱即用的 AI Agent **Skills**,让智能体零额外配置即可操作平台。

**元数据驱动**:接口目录 `catalog.json` 同时生成域命令与技能(`catalog.json → cmd/gen/*.go + skills/`)。改接口先改源头再重新生成,命令、参数提示、技能始终与平台对齐。

## 安装

需 Node ≥ 18:

```bash
npm i -g @openydt/openydt-cli      # 全局安装
openydt --version
# 免安装试用: npx @openydt/openydt-cli --help
```

安装 / 更新时会**自动把技能同步到本机已装的各 AI agent**(Claude Code / Codex / Cursor / Gemini CLI / OpenCode 等,经 `npx skills`,用户级)。手动同步:`openydt skill sync`。关闭自动同步:设环境变量 `OPENYDT_NO_SKILLS_SYNC=1`。

### 从源码构建(开发者)

```bash
make build          # 产出 bin/openydt
make catalog        # (可选)解析接口文档 Doc/*.vue → catalog/catalog.json
make generate       # (可选)catalog.json → cmd/gen/*.go(各域命令,同步内嵌副本)
```

> 发布见 [RELEASING.md](./RELEASING.md)。

## 快速开始

```bash
# 1. 配置授权商凭据(测试环境)
openydt config set --profile demo --key test --secret 123456 --env test

# 2. 验证凭据 / 签名链路
openydt auth test

# 3. 查费(生成的一等命令,带类型化参数)
openydt trade get-park-fee --car-code 粤EJW962 --park-code 1ZS7H5PQH9 -o table

# 4. 通用兜底:调用任意业务编码 cmd
openydt api getParkOnSiteCar --body '{"parkCodeList":["PTD2YBBZ"],"pageNum":1,"pageSize":10}'
```

## 三层命令体系

1. **域一等命令** `openydt <域> <命令>` —— 由 `catalog.json` 自动生成,带类型化 flag(含 `[可选: ...]` 枚举提示)、`--help`、写操作 `--yes` 守护。
2. **通用调用** `openydt api <cmd> --body '{...}'` —— 兜底任意可调用接口(含未做成一等命令的)。
3. **参数发现** `openydt schema [cmd]` —— 查看接口的必填 / 选填 / 类型 / **枚举可选值** / 示例 body(人和 AI Agent 自助发现入参)。
4. **配置 / 鉴权 / 技能** `openydt config | auth | skill`。

### 错误输出(AI Agent 友好)

失败响应不只回传原始 message,还附结构化提示:`status` / `resultCode` 的中文含义 + **可执行建议(hint)**;当 message 指明某参数时,自动从目录补出该字段的**类型 / 必填 / 说明 / 枚举可选值**;并标注是否 `retriable`。JSON 模式输出 `_error` 对象供 Agent 解析自纠,table 模式输出可读多行。

### 内置命令

10 个业务域,共 **143** 条一等命令(接口目录共 423 个):

| 域 | 命令 | 数量 |
|---|---|---|
| `trade` | 停车缴费(查费 / 缴费 / 补缴 / 预存) | 7 |
| `park` | 车场信息 / 计费测算 | 18 |
| `parking` | 停车记录 / 在场 / 进出 / 锁车 | 31 |
| `device` | 设备控制(开关闸 / 显示屏 / 扫码) | 11 |
| `ticket` | 月票 / VIP / 特殊车辆类型 | 29 |
| `blacklist` | 黑名单 | 3 |
| `redlist` | 白名单 | 3 |
| `visitor` | 访客 | 2 |
| `data` | 数据分析 | 9 |
| `coupon` | 电子券 / 商户 | 30 |

其余模块(城市运营、第三方车场接入、积分等)及 webhook 回调接口未生成一等命令;可调用类用 `openydt api <cmd>` 调用,webhook(平台主动推送)需自建接收端,详见 `openydt-api-explorer` 技能。

## 记忆 / 默认值

让你和 Agent 不必每次重复车场 / 车牌等背景:

- **默认值(随 profile,按环境隔离)**:`openydt config set-default --park PTD2YBBZ --car-no 桂566666`;之后命令缺 `parkCode` 时自动补全(显式值永远优先,`--verbose` 可见补全;prod 写操作仍需 `--yes`)。
- **车场经验(AI agent 自动沉淀 / 回忆)**:存 `~/.config/openydt-cli/park-notes/{parkCode}.{env}.md`(如 `PTD2YBBZ.test.md`),一车场一环境一文件、物理隔离 test/dev/prod。Agent 操作成功后自动追加该车场的有效调用 / 陷阱 / 常用车牌,下次自动回忆;prod 不记真实车牌。清理:删对应文件。

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

`skills/` 下 11 个技能,经 `npx skills` 分发到本机各 AI agent:

| Skill | 说明 |
|---|---|
| `openydt-shared` | 配置 / profile / 签名 / 状态码 / 限速 / 安全 / 车场经验(所有域技能先读它) |
| `openydt-billing` | 停车缴费闭环(进车 → 查费 → 缴费 → 查单) |
| `openydt-park` | 车场信息 / 计费测算 |
| `openydt-record` | 停车记录 / 在场 / 锁车 |
| `openydt-device` | 设备控制(开关闸 / 显示屏) |
| `openydt-monthticket` | 月票 / VIP 闭环(建类型 → 开通 → 续费 / 退费) |
| `openydt-list` | 黑名单 / 白名单 / 访客 |
| `openydt-data` | 数据分析 |
| `openydt-coupon` | 电子券闭环(建商家 + 模板 → 售卖 → 发放 → 回收) |
| `openydt-api-explorer` | 用 `api` 调用未封装接口 |
| `openydt-skill-maker` | 创建自定义 skill 的框架 |

格式校验:`node scripts/skill-format-check/index.js`。

## 测试与验证

```bash
make test           # 单元测试(含签名锚点向量 v2/v3/Base64)
make smoke          # 对测试环境查费冒烟
make e2e            # 端到端全量验证 → TEST_REPORT.md
```

端到端验证按业务流程链(缴费 / 月票 / 黑名单访客 / 电子券)驱动每个纳入接口,产出 `TEST_REPORT.md`。

## 状态码

响应包络 `{data, message, resultCode, status}`。`status`:1 成功 / 2 业务失败 / 3 系统异常 / 4 签名错误 / 5 key 错误 / 6 未授权 / 7 参数不完整 / 9 接口不存在。`status=2` 时看 `resultCode`(901-912、1801)。退出码:0 成功 / 1 业务失败 / 2 参数 / 4 鉴权 / 5 网络。

## 目录结构

```
cmd/            # root + api/auth/config/schema/skill + gen/(生成的各域命令)
internal/       # sign 签名 / client HTTP / config 配置 / output 输出 / catalog 目录 / cmdutil / skillsync
catalog/        # catalog.json(接口目录,由 tools/extractor 生成)
tools/extractor # Node: 解析 Doc/*.vue → catalog.json
internal/gen    # Go codegen: catalog.json → cmd/gen
skills/         # AI Agent 技能
tests/e2e       # 端到端验证 harness
```

## 致敬

openydt-cli 的形态与设计深受 **飞书 CLI** 启发 —— Go + Cobra、接口元数据驱动、为人和 AI Agent 而生、以 `npx skills` 分发技能。在此致敬并感谢这个优秀的开源项目:

- **飞书 CLI(larksuite/cli)** · <https://github.com/larksuite/cli>
