# CLAUDE.md — openydt-cli

艾科智泊停车开放平台 CLI(Go + Cobra)。详见 `README.md`。

## 构建 / 命令
- `make build` → `bin/openydt`(单二进制);`go test ./...`;`go vet ./...`
- `make catalog` → Node 抽取器解析 `open-api-front` 的 `Doc/*.vue` → `catalog/catalog.json`
- `make generate` → `go run ./internal/gen` 读 catalog.json 生成 `cmd/gen/*.go`
- `make e2e` → `tests/e2e` 端到端验证 → `TEST_REPORT.md`(仅对 test 环境)

## 元数据驱动(改接口先改源头)
- 真相源 = `open-api-front/src/components/Doc/**/*.vue`(每文件一接口),不是 `state.js`。
- 链路:Doc → `catalog/catalog.json` → `cmd/gen/*.go` + skills。
- **`cmd/gen/*.go` 与 `catalog/catalog.json` 是生成产物,DO NOT EDIT**;要改命令请改抽取规则(`tools/extractor/extract.mjs`)或生成器(`internal/gen`)后重跑 `make catalog generate`。

## 签名(`internal/sign`,已实测+单测锚定)
- 请求:`POST {base}/openydt/api/v3/{cmd}?sign=...`;头 `Authorization: Base64(key + ":" + 时间)`、`Content-Type: application/json;charset=utf-8`。
- `时间` = **本地时间** `yyyyMMddHHmmss`,有效 10 分钟。
- **v2(默认)**:`sign = lower(md5(key + ":" + 时间 + ":" + secret))`,**不含 body**。
- **v3**:`sign = lower(md5(key + ":" + 时间 + ":" + 紧凑body + ":" + secret))`,含 body。
- **不变量**:用于签名的 body 必须与实际发出的 body **逐字节一致** → 先 `sign.CompactBody(body)` 一次,签名和发送都用它。
- ⚠️ 测试 key 只接受 **v2**;`--sign v3` 会回 `status=4 签名错误`(除非平台对该 key 开通 v3)。
- 单测向量(勿改,见 `internal/sign/sign_test.go`):

```
v2:  md5("test:20260529220831:123456")                          = a2710e9f093f22691ff24cf944aadd48
v3:  md5('abc:20170416142030:{"parkCode":"123123"}:123456')     = f1e1e7f8bc710dd6633bc0d9a9336207
auth: base64("test:20260529220831")                             = dGVzdDoyMDI2MDUyOTIyMDgzMQ==
```

## 技能同步(skillsync)
- 真相源 = 仓库 `skills/openydt-*/SKILL.md`(11 个);分发交给外部 `npx skills`(vercel-labs 包管理器,自动探测本机已装 agent)。
- **主路径**:`npm i/update -g @openydt/openydt-cli` 的 postinstall 跑 `npx skills add xiaowen-0725/openydt-cli -g -y`(best-effort,失败不阻断安装),并写 `~/.config/openydt-cli/skills-state.json`。
- **兜底**:Go 二进制每条普通命令前 `skillsync.MaybeTrigger` 本地比对版本;漂移则 detached 后台跑 `openydt skill sync --quiet`(输出进 `skills-sync.log`,**每版本最多每 6 小时 fork 一次,瞬时失败可自愈**,不碰 stdout)。
- **手动**:`openydt skill sync [--force]`。
- **opt-out**:`OPENYDT_NO_SKILLS_SYNC=1`;CI / DEV / 非 release 版本自动跳过。
- `internal/skillsync/*` 有单测锚定;真实 `npx` 不进单测(用 override 隔离)。

## 网关 / 客户端(`internal/client`)
- 网关为腾讯 TGW + APISIX,**会间歇性把合法路由 404、并重置连接** → 客户端必须重试 + 指数退避(对 404/429/5xx/连接错误重试,已实现)。
- 必须设自定义 `User-Agent`(`openydt-cli/<ver>`);默认 `Go-http-client`/空 UA 可能被网关挡。
- 限速:授权车场 < 60 → 300 次/分(5/s);E2E 已按 ~4/s 节流。
- 响应包络 `{data,message,resultCode,status}`;`status` 1成功/2业务失败/4签名/5key/6未授权/7参数/9接口不存在;业务码见 `internal/client/codes.go`。

## 安全
- 写操作(缴费/开闸/发券/开通月票/黑名单等)由生成命令自动挂 `--yes` 守护;先 `--dry-run` 预览签名请求。
- E2E/写操作仅对 test 环境;勿在 prod 跑批量验证。
