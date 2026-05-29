# 发布流程

openydt-cli 通过 **Git tag + GitHub Actions(goreleaser)** 自动出多平台二进制,并提供 `curl|sh`、npm、源码三种安装方式。一次发布的标准步骤:

## 1. 准备
确保 `main` 上构建/测试通过,且生成产物是最新的:
```bash
make catalog generate   # 若接口目录有更新(会同步 catalog/ 与 internal/catalog/)
go test ./... && go vet ./... && make build
git commit -am "..."
```

## 2. 打标签触发发布(产出 GitHub Release)
```bash
git tag v0.1.0
git push origin v0.1.0
```
`.github/workflows/release.yml` 会跑 `go test` → `goreleaser release`,在 GitHub Releases 发布
`darwin/linux/windows × amd64/arm64` 的归档(`openydt-cli_0.1.0_<os>_<arch>.tar.gz` / windows 为 zip)+ `checksums.txt`。
版本号通过 ldflags 注入(`openydt --version`)。

## 3. 发布 npm 包(可选,提供 `npx`/`npm i -g`)
npm 包仅是个壳,postinstall 时按平台从上面的 Release 下载二进制,**版本必须与 tag 一致**:
```bash
cd npm
npm version 0.1.0 --no-git-tag-version   # 与 v0.1.0 对齐
npm publish --access public              # 需 npm 账号(npm login)
```

## 4. (可选)Homebrew
如需 `brew install`,新建一个 tap 仓库 `xiaowen-0725/homebrew-tap`,在 `.goreleaser.yml`
增加 `brews:` 块指向它即可;goreleaser 发布时会自动更新 formula。

## 安装方式速查(给使用者)
- 脚本:`curl -fsSL https://raw.githubusercontent.com/xiaowen-0725/openydt-cli/main/scripts/install.sh | sh`
- npm:`npm i -g openydt-cli` 或 `npx openydt-cli`
- 直接下载:GitHub Releases 里对应平台的归档,解压后把 `openydt` 放进 PATH
- 源码:`git clone … && make build`(产物 `bin/openydt`)

## 前置条件
- GitHub Actions 默认 `GITHUB_TOKEN` 即可发布 Release(无需额外 secret)。
- npm 发布需在本机 `npm login`(或 CI 配 `NPM_TOKEN`)。
- 包名 `openydt-cli` 若被占用,改用作用域名如 `@akeparking/openydt-cli`(同步改 npm/package.json 与 install.js 提示)。
