---
name: openydt-shared
version: 1.0.0
description: "艾科智泊停车开放平台(openydt) CLI 共享基座：首次使用 openydt、配置 profile、签名(v2/v3)与多环境(test/dev/prod)、响应包络与状态码/业务码处理、退出码、限速与重试、写操作安全规则。当用户第一次使用 openydt、配置/切换 profile、处理签名或环境问题、解读 status/resultCode、遇到限速或鉴权失败、或执行任何停车场写操作前触发。所有 openydt 域技能都应先 Read 本基座。"
metadata:
  requires:
    bins: ["openydt"]
---

# openydt CLI 共享基座

本技能是艾科智泊停车开放平台 CLI(`openydt`)的共享基础规则。所有 openydt 域技能(park / parking / trade / coupon / ticket / device / blacklist / visitor / data 等)在执行具体任务前，都应先 Read 本文件，以统一处理配置、签名、状态码、限速与安全。

`openydt` 把开放平台接口封装成命令行：自动处理签名鉴权(v2/v3)、多授权商 profile、多环境(test/dev/prod)，并内置重试与退避。

## 配置 profile 与凭据

凭据按「授权商 profile」管理，每个 profile 含 key/secret/env/sign。配置文件位于 `~/.config/openydt-cli/config.json`(尊重 `XDG_CONFIG_HOME`)，权限 0600。

```bash
# 新增或更新一个授权商 profile（首次使用从这里开始）
openydt config set --profile demo --key test --secret 123456 --env test --sign v2

# 列出所有 profile（secret 已脱敏），带 * 的是当前 profile
openydt config list

# 切换当前 profile
openydt config use demo

# 打印配置文件路径
openydt config path
```

- `config set` 的 `--profile / --key / --secret` 必填；`--env` 默认 `test`，`--sign` 默认 `v2`。
- 第一次 `config set` 时，若尚无当前 profile，会自动把它设为当前 profile。

### 环境变量覆盖（适合 CI）

下列环境变量优先级高于 profile 中的值，可在不写配置文件的情况下临时覆盖：

| 变量 | 含义 |
|------|------|
| `OPENYDT_PROFILE` | 选择 profile 名 |
| `OPENYDT_KEY` | 覆盖 key |
| `OPENYDT_SECRET` | 覆盖 secret |
| `OPENYDT_ENV` | 覆盖环境 test\|dev\|prod |
| `OPENYDT_SIGN` | 覆盖签名版本 v2\|v3 |

优先级(从低到高)：内置默认 < profile < 环境变量 < 命令行显式 flag。空值会被忽略。只要设置了 `OPENYDT_KEY`+`OPENYDT_SECRET`，即使没有同名 profile 也能直接调用。

## 全局 flag

所有命令通用：

| Flag | 说明 |
|------|------|
| `--profile <名>` | 指定授权商 profile（默认当前 profile） |
| `--env test\|dev\|prod` | 指定环境（默认 test） |
| `--output`, `-o json\|table` | 输出格式（默认 json） |
| `--sign v2\|v3` | 签名版本（默认按 profile，否则 v2） |
| `--yes`, `-y` | 确认执行写操作 |
| `--dry-run` | 只打印将发送的签名请求，不实际发送 |
| `--verbose`, `-v` | 输出调试信息到 stderr |

各环境 base URL：

- test → `https://openapi-test.yidianting.com.cn`
- dev → `https://openapi-dev.yidianting.xin`
- prod → `https://open.yidianting.xin`

## 认证验证

配置好后，先做一次冒烟验证(内部调用 `getAuthParkCodes` 确认凭据/签名链路可用)：

```bash
openydt auth test
```

成功输出 `✓ 认证通过 (status=1)` 并列出授权车场；失败会打印 status/message/resultCode 并以对应退出码返回。

## 签名

请求路径形如 `POST {base}/openydt/api/v3/{cmd}?sign={sign}`，并带 `Authorization: base64(key:ts)` 头。时间戳 `ts` 为本地时间 `yyyyMMddHHmmss`，有效期 10 分钟。

| 版本 | 算法 | 说明 |
|------|------|------|
| v2(默认) | `lower(md5(key:ts:secret))` | 不含 body；测试环境默认可用 |
| v3 | `lower(md5(key:ts:body:secret))` | 含 compact 后的 body |

**重要**：实测测试 key 仅接受 v2；用 v3 调用测试 key 会返回「签名错误」(status=4)，除非平台对该 key 专门开通了 v3。默认保持 v2 即可，仅在平台明确为该 key 开通 v3 后再用 `--sign v3`。

签名用的 body 与实际发送的 body 必须字节一致：CLI 会先做一次 JSON compact 再同时用于签名与发送（字符串内部空格如 `"2019-04-16 00:11:25"` 会保留）。

## 三层命令模型

调用任意接口有三条路径，按优先级选择：

1. **域一等命令**(首选)：`openydt <域> <命令>`，参数已结构化为 flag，最易用。例如 `openydt park get-auth-park-codes`、`openydt parking <子命令>`。当前内置域：`blacklist coupon data device park parking redlist ticket trade visitor`。
2. **通用兜底**：`openydt api <cmd> --body '{...}'`，对任意业务编码 cmd 自动签名并 POST，覆盖任何可调用接口。
   ```bash
   openydt api getParkFee --body '{"carCode":"粤EJW962"}'
   openydt api getAuthParkCodes
   echo '{"parkCode":"PTD2YBBZ"}' | openydt api getParkOnSiteCar --body-file -
   ```
   `--body` 与 `--body-file` 互斥；`--body-file -` 从 stdin 读取。
3. **schema 探索**(若有)：用于发现接口与字段，再回到 ① 或 ②。

## 响应包络与状态码

平台统一包络：`{"data":..., "message":"...", "resultCode":N, "status":N}`。

`status` 含义：

| status | 含义 |
|--------|------|
| 1 | 业务成功 |
| 2 | 业务失败（看 resultCode） |
| 3 | 系统异常 |
| 4 | 签名错误 |
| 5 | key 错误 |
| 6 | 未授权 |
| 7 | 请求参数不完整 |

当 `status == 2` 时看 `resultCode`(常见业务码，源自 `internal/client/codes.go`)：

| resultCode | 含义 |
|------------|------|
| 901 | 系统发生异常 |
| 902 | 远程服务器未响应 |
| 903 | 运营商不存在 |
| 904 | 停车场不存在 |
| 905 | 未找到在场车辆 |
| 906 | 账单不存在 |
| 907 | 账单已同步 |
| 908 | 其它错误 |
| 909 | 请求参数错误 |
| 910 | 找不到授权商下面的停车场 |
| 911 | 无权限操作该停车场 |
| 912 | 查费已超时，请重新查费 |
| 1801 | 找不到指定车辆 |

### 进程退出码

| 退出码 | 含义 |
|--------|------|
| 0 | 成功(status=1) |
| 1 | 业务失败(status=2 或其它非成功) |
| 2 | 参数错误(用法错误) |
| 4 | 鉴权失败(status=4 签名 / 5 key / 6 未授权) |
| 5 | 网络/传输失败 |

## 限速与重试

- 授权车场数 < 60 的授权商：限速 **300 次/分**。批量调用时自行节流，避免触发 429。
- 客户端已内置重试 + 指数退避(约 400ms 起，带抖动，默认最多重试 3 次)。
- 遇网关偶发 404、连接重置、429/502/503/504 会自动重试；非包络的 HTML 错误页不重试。
- `查费超时(resultCode 912)` 是业务态，需按提示重新查费，不是网络重试范畴。

## 安全规则

- **写操作必须 `--yes`**：缴费、开闸、发券、开通月票、加/移黑名单等任何会改变平台状态的操作，必须显式带 `--yes` 才会执行，避免误操作。
- **先 `--dry-run` 预览**：危险或不确定的请求先用 `--dry-run` 查看将发送的签名请求(URL/sign/ts/body)，确认无误后再去掉。
- **不要明文输出密钥**：不要把 key/secret 打印到终端或日志；`config list` 已对 secret 脱敏。
- 默认在 `test` 环境验证；切到 `prod` 前务必与用户确认。

## 测试车场（仅测试环境）

| parkCode | 用途 |
|----------|------|
| `1ZS7H5PQH9` | 可查费，配套测试车牌 `粤EJW962` |
| `PTD2YBBZ` | 有存量数据，适合查记录 / 查在场车辆 |

示例：

```bash
openydt api getParkFee --body '{"parkCode":"1ZS7H5PQH9","carCode":"粤EJW962"}'
openydt api getParkOnSiteCar --body '{"parkCode":"PTD2YBBZ"}'
```
