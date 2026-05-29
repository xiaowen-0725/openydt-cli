// postinstall: 按平台从 GitHub Releases 下载对应的 openydt 二进制(版本与本包一致)。
const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");

const pkg = require("./package.json");
const REPO = "xiaowen-0725/openydt-cli";

const plat = { darwin: "darwin", linux: "linux", win32: "windows" }[process.platform];
const arch = { x64: "amd64", arm64: "arm64" }[process.arch];
if (!plat || !arch) {
  console.error(`[openydt] 不支持的平台: ${process.platform}/${process.arch}`);
  process.exit(1);
}

const ext = plat === "windows" ? "zip" : "tar.gz";
const asset = `openydt-cli_${pkg.version}_${plat}_${arch}.${ext}`;
const url = `https://github.com/${REPO}/releases/download/v${pkg.version}/${asset}`;
const binDir = path.join(__dirname, "bin");
fs.mkdirSync(binDir, { recursive: true });
const archive = path.join(binDir, asset);

(async () => {
  console.log("[openydt] 下载", url);
  const res = await fetch(url, { redirect: "follow" });
  if (!res.ok) throw new Error(`下载失败 HTTP ${res.status} — 确认已发布 v${pkg.version}`);
  fs.writeFileSync(archive, Buffer.from(await res.arrayBuffer()));

  if (ext === "zip") execFileSync("unzip", ["-o", archive, "-d", binDir], { stdio: "inherit" });
  else execFileSync("tar", ["-xzf", archive, "-C", binDir], { stdio: "inherit" });

  const src = path.join(binDir, plat === "windows" ? "openydt.exe" : "openydt");
  const dst = path.join(binDir, plat === "windows" ? "openydt-bin.exe" : "openydt-bin");
  fs.renameSync(src, dst);
  fs.chmodSync(dst, 0o755);
  fs.unlinkSync(archive);
  console.log("[openydt] 安装完成");
})().catch((e) => {
  console.error("[openydt]", e.message);
  process.exit(1);
});
