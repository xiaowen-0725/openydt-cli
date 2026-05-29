# 发布流程(npm)

面向使用者的安装方式只有 **npm**(`npm i -g openydt-cli`)。npm 包是个壳,安装时按平台从
**GitHub Releases** 下载原生二进制——所以一次发布 = 先出 GitHub Release(二进制),再发 npm 包。
两步都由 **打 tag** 自动完成(GitHub Actions)。

## 一次性准备
在 GitHub 仓库 `Settings → Secrets and variables → Actions` 添加:
- **`NPM_TOKEN`**:npm 账号的发布 token(npmjs.com → Access Tokens → Generate,类型选 Automation)。
  没有它,GitHub Release 仍会发布,但 npm 不会自动发(会给 warning)。

## 每次发布
```bash
# 确保 main 上构建/测试通过、生成产物最新
make catalog generate    # 若接口目录有更新
go test ./... && make build && git commit -am "..."

# 打标签即触发发布
git tag v0.1.0
git push origin v0.1.0
```
`.github/workflows/release.yml` 会:
1. `go test` → `goreleaser release`:在 GitHub Releases 发布 darwin/linux/windows × amd64/arm64 归档 + checksums。
2. 把 `npm/package.json` 版本同步为 `0.1.0`,`npm publish --access public` 发布到公共 npm。

用户随后:`npm i -g openydt-cli`(postinstall 自动下载对应平台二进制)。

## 手动发 npm(不走 CI 时)
```bash
npm login                                   # 需 npm 账号
cd npm
npm version 0.1.0 --no-git-tag-version      # 与 GitHub Release 的 v0.1.0 对齐
npm publish --access public
```

## 备注
- 包名 `openydt-cli` 若被占用,改作用域名(如 `@akeparking/openydt-cli`),同步改 `npm/package.json` 的 `name`。
- 正式环境 base URL 已内置 `prod = https://open.yidianting.xin`(默认仍是 test)。
