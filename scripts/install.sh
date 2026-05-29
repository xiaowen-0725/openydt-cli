#!/bin/sh
# openydt-cli 安装脚本
#   curl -fsSL https://raw.githubusercontent.com/xiaowen-0725/openydt-cli/main/scripts/install.sh | sh
# 可选环境变量: VERSION=v0.1.0 (默认最新)  BIN_DIR=/usr/local/bin
set -e

REPO="xiaowen-0725/openydt-cli"
BIN="openydt"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$os" in
  darwin) os="darwin" ;;
  linux)  os="linux" ;;
  *) echo "不支持的系统: $os (Windows 请从 Releases 下载 zip)"; exit 1 ;;
esac
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "不支持的架构: $arch"; exit 1 ;;
esac

ver="${VERSION:-}"
if [ -z "$ver" ]; then
  ver="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep -m1 '"tag_name"' | cut -d'"' -f4)"
fi
[ -n "$ver" ] || { echo "无法确定版本, 请设置 VERSION=v0.1.0"; exit 1; }
num="${ver#v}" # 去掉前导 v, 匹配归档名

url="https://github.com/$REPO/releases/download/$ver/openydt-cli_${num}_${os}_${arch}.tar.gz"
echo "下载 $url"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
curl -fsSL "$url" -o "$tmp/pkg.tar.gz"
tar -xzf "$tmp/pkg.tar.gz" -C "$tmp"

if [ -w "$BIN_DIR" ]; then
  install -m 0755 "$tmp/$BIN" "$BIN_DIR/$BIN"
else
  echo "需要权限写入 $BIN_DIR, 使用 sudo"
  sudo install -m 0755 "$tmp/$BIN" "$BIN_DIR/$BIN"
fi

echo "✓ 已安装到 $BIN_DIR/$BIN ($ver)"
"$BIN_DIR/$BIN" --version || true
echo "接着: openydt config set --profile demo --key <key> --secret <secret> --env test"
