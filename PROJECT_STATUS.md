# 项目状态 / 进展记录(PROJECT_STATUS)

> 更新时间锚点:对应 commit `9ffad1e`、tag `v0.1.1`(npm 已发布)。本文件用于追溯项目进展与现状。

## 1. 这是什么
`openydt-cli` —— 广东艾科智泊「智慧停车开放平台」的命令行工具,形态参照飞书 CLI(Go + Cobra,接口元数据驱动),为人和 AI Agent 而生。自动处理签名鉴权、多授权商 profile、多环境。

- 仓库:https://github.com/xiaowen-0725/openydt-cli (分支 `main`)
- 依赖的外部仓库(本机,非本仓库):
  - 接口文档前端真相源:`/Users/zhoujw/develop/git/open-api-front`(`src/components/Doc/**/*.vue` = 423 接口)
  - 参照对象飞书 CLI:`/Users/zhoujw/develop/github/cli`
- SaaS 运营后台(测试环境,采真实数据用):`https://bemmgr-test.yidianting.com.cn`

## 2. 进展(提交里程碑)
| commit | 内容 |
|---|---|
| d309f48 | v1 框架:Go+Cobra、签名/客户端/配置/输出、catalog 抽取、143 命令 codegen、11 skill、首版 E2E |
| 68eed81 | E2E 真实数据驱动:非 NODEPLOY 失败 72→1;CLAUDE.md 记录签名/网关约定 |
| 99ec3e5 | `schema` 参数发现命令(含枚举可选值)+ AI Agent 友好的结构化错误 |
| 538c812 | 「停车服务3.0」接口支持对照表 `SUPPORTED_INTERFACES.html`(门户风格) |
| 6815490→765026a | 发布:GitHub Release 流水线 + npm 安装;收敛为 npm-only;改用作用域名 `@openydt/openydt-cli` |

## 3. 规模 / 架构
- 手写 Go ~2808 行 / 20 文件 + 生成命令 11 文件 / **143 条一等命令** / **11 个 skill**
- catalog:**423 接口**,其中 143 纳入为一等命令(停车缴费/车场/记录/设备/月票VIP/黑白名单/访客/数据分析/电子券)
- 关键目录:`cmd/`(root+api+auth+config+schema+gen/) · `internal/`(sign/client/config/output/catalog/cmdutil/gen) · `tools/extractor`(Doc→catalog.json) · `skills/` · `tests/e2e`
- 三层命令:① `openydt <域> <命令>` ② `openydt api <cmd>` 兜底 ③ `openydt schema <cmd>` 查参数
- catalog 既驱动命令 codegen,也内嵌进二进制(`//go:embed`)供 schema/错误富化

## 4. 关键事实(实施时直接照用)
- **签名**:`时间`=本地 yyyyMMddHHmmss(有效10分钟);v2=`md5(key:时间:secret)`(默认);v3=`md5(key:时间:紧凑body:secret)`;`Authorization=base64(key:时间)`;sign 入 `?sign=` 查询;POST。**实测测试 key 仅接受 v2**(v3 回 status=4)。
- **网关**:腾讯 TGW + APISIX,间歇性 404/连接重置 → 客户端内置重试+退避 + 自定义 User-Agent。
- **环境**:默认 `test`;`prod = https://open.yidianting.xin`(正式),`dev = openapi-dev.yidianting.xin`,`test = openapi-test.yidianting.com.cn`。E2E 有硬护栏只在 test 跑。
- **测试凭据**:key=`test` secret=`123456`;数据丰富云车场 `PTD2YBBZ`(智汇云测试专用车场123412);另 `1ZS7H5PQH9`。
- **生成产物边界**:`cmd/gen/*.go` 与 `catalog/catalog.json`(及 `internal/catalog/catalog.json` 内嵌副本)是生成物,**勿手改**;改命令请改 `tools/extractor/extract.mjs` 或 `internal/gen` 后 `make catalog generate`。详见 `CLAUDE.md`。

## 5. 测试现状(`TEST_REPORT.md`,对测试环境实跑)
- **PASS=106 / BIZFAIL=0 / ERROR=0 / SKIP=1 / NODATA=25 / NODEPLOY=11**
- **非 NODEPLOY 真失败 = 1**(仅 `deleteTrader`,故意跳过的破坏性操作)
- NODATA=环境限制(无物理设备/无欠费数据/无待校正会话/服务端异常/幂等终态等,接口本身已调通);NODEPLOY=测试环境未部署该接口
- 重跑:`make e2e`(依赖真实 fixtures + `tests/e2e/recipes.json`)

## 6. 接口支持对照(便于扩展)
`SUPPORTED_INTERFACES.html`(仅「停车服务3.0」platform/*,共 250):
- ✅ 已支持 143 · ⚪ 未做命令(api 兜底,**待扩展候选**)56 · 🔔 Webhook 回调 43 · ➖ 非接口 8
- 重生成:`node tools/gen-support-doc.mjs`

## 7. 发布现状 ✅(已发布)
- ✅ **npm 已公开发布:`@openydt/openydt-cli`(latest = `0.1.1`)** —— 对外安装:`npm i -g @openydt/openydt-cli`(得到 `openydt` 命令,postinstall 按平台下载二进制)。
- ✅ **GitHub Release `v0.1.0` / `v0.1.1` 已发布**(各 6 平台二进制 + checksums);npm 包安装时从对应 Release 下载。
- ✅ **CI 全自动**:`.github/workflows/release.yml` 打 `v*` tag → goreleaser 发 Release + `npm publish`;`NPM_TOKEN` secret 已配可用 token(Classic Automation,带 bypass-2FA)。
- npm 包页面(README + package.json)**已移除所有源码地址**(0.1.1);`install.js` 仍含 Release 下载 URL(功能必需,非对外文档)。
- 备注:0.1.0 是首版(含源码地址,已被 0.1.1 取代为 latest);组织首次发布曾走 npm "Staged Packages" 需手动确认,后续 CI 直发。

## 8. 待办 / 下一步
1. **安全(尽快)**:对话中暴露过几个 npm token,到 npmjs.com 删除/吊销,仅保留 CI 用的那个(已存为 `NPM_TOKEN` secret)。
2. **license 确认**:`npm/package.json` 暂填 MIT,如不开源改 `UNLICENSED`。
3. **正式环境实测**:`prod` 地址已内置(`https://open.yidianting.xin`)但未用正式凭据实发验证过;拿到正式 key 可 `openydt --env prod auth test`。
4. **扩展接口**:从 `SUPPORTED_INTERFACES.html` 的「⚪ 未做命令」挑需要的(如月票会员车类型 0/10、积分、发票),放开 extractor 纳入规则后 `make catalog generate`,再打 tag 发新版。
5. 可选:命令 shell 自动补全、schema/错误格式单测、把对照表也出一份 Markdown。

## 9. 常用命令速查
```bash
make build            # 构建 bin/openydt
make catalog generate # 重抽取接口 + 重生成命令(同步内嵌 catalog)
make e2e              # 端到端验证 → TEST_REPORT.md(仅 test)
go test ./... && go vet ./...
node tools/gen-support-doc.mjs   # 重生成接口支持对照表 HTML
# 发布(已打通): git tag v0.1.x && git push origin v0.1.x  → CI 自动发 GitHub Release + npm publish
```
