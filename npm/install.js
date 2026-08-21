// postinstall: 按平台从 GitHub Releases 下载对应的 openydt 二进制(版本与本包一致)。
const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");

const pkg = require("./package.json");
const REPO = "xiaowen-0725/openydt-cli";
const GITHUB_API = "https://api.github.com";
const DOWNLOAD_TIMEOUT_MS = 15_000;

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

async function install() {
  console.log("[openydt] 下载", url);
  const content = await downloadReleaseAsset({ version: pkg.version, asset });
  fs.writeFileSync(archive, content);

  if (ext === "zip") execFileSync("unzip", ["-o", archive, "-d", binDir], { stdio: "inherit" });
  else execFileSync("tar", ["-xzf", archive, "-C", binDir], { stdio: "inherit" });

  const src = path.join(binDir, plat === "windows" ? "openydt.exe" : "openydt");
  const dst = path.join(binDir, plat === "windows" ? "openydt-bin.exe" : "openydt-bin");
  fs.renameSync(src, dst);
  fs.chmodSync(dst, 0o755);
  fs.unlinkSync(archive);
  console.log("[openydt] 安装完成");

  // best-effort: 同步 AI Agent 技能到本机已装的各 agent(失败绝不影响安装)
  syncSkills();
}

async function downloadReleaseAsset({
  version,
  asset,
  fetchImpl = fetch,
  repo = REPO,
  apiRoot = GITHUB_API,
  timeoutMs = DOWNLOAD_TIMEOUT_MS,
}) {
  const directUrl = `https://github.com/${repo}/releases/download/v${version}/${asset}`;
  try {
    const response = await fetchWithTimeout(fetchImpl, directUrl, { redirect: "follow" }, timeoutMs);
    if (response.ok) return Buffer.from(await response.arrayBuffer());
    console.warn(`[openydt] GitHub 直连下载失败(HTTP ${response.status}),尝试备用通道`);
  } catch (error) {
    console.warn(`[openydt] GitHub 直连下载失败(${error.message}),尝试备用通道`);
  }

  const headers = {
    Accept: "application/vnd.github+json",
    "User-Agent": `openydt-installer/${version}`,
    "X-GitHub-Api-Version": "2022-11-28",
  };
  const releaseUrl = `${apiRoot}/repos/${repo}/releases/tags/v${version}`;
  const releaseResponse = await fetchWithTimeout(fetchImpl, releaseUrl, { headers }, timeoutMs);
  if (!releaseResponse.ok) {
    throw new Error(`备用通道查询版本失败 HTTP ${releaseResponse.status} — 确认已发布 v${version}`);
  }

  const release = await releaseResponse.json();
  const releaseAsset = Array.isArray(release.assets)
    ? release.assets.find((candidate) => candidate.name === asset)
    : undefined;
  if (!releaseAsset || typeof releaseAsset.url !== "string") {
    throw new Error(`备用通道未找到安装包 ${asset} — 确认 GitHub Release 资源已上传完整`);
  }

  console.log("[openydt] 使用 GitHub API 备用通道下载");
  const assetResponse = await fetchWithTimeout(
    fetchImpl,
    releaseAsset.url,
    { redirect: "follow", headers: { ...headers, Accept: "application/octet-stream" } },
    timeoutMs,
  );
  if (!assetResponse.ok) {
    throw new Error(`备用通道下载失败 HTTP ${assetResponse.status} — 请检查网络后重试`);
  }
  return Buffer.from(await assetResponse.arrayBuffer());
}

async function fetchWithTimeout(fetchImpl, requestUrl, options, timeoutMs) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  try {
    return await fetchImpl(requestUrl, { ...options, signal: controller.signal });
  } finally {
    clearTimeout(timer);
  }
}

// 把技能同步到本机所有已装 agent;失败只 warn,postinstall 不报错退出。
function syncSkills() {
  if (process.env.OPENYDT_NO_SKILLS_SYNC) {
    console.log("[openydt] 已按 OPENYDT_NO_SKILLS_SYNC 跳过 skills 同步");
    return;
  }
  try {
    execFileSync("npx", ["-y", "skills", "add", skillsSource(pkg.version), "-g", "--copy", "-y"], { stdio: "inherit" });
    writeSkillsState();
    console.log("[openydt] skills 已同步");
  } catch (e) {
    console.warn(`[openydt] skills 同步失败(可稍后手动:openydt skill sync):${e.message}`);
  }
}

function skillsSource(version, localDir = path.join(__dirname, "skills"), existsSync = fs.existsSync) {
  if (existsSync(localDir)) return localDir;
  return /^\d+\.\d+\.\d+$/.test(version)
    ? `https://github.com/${REPO}/tree/v${version}`
    : REPO;
}

// 写 skills-state.json,使二进制兜底的漂移检测有基准(XDG-aware)。
function writeSkillsState() {
  const base = process.env.XDG_CONFIG_HOME || path.join(require("os").homedir(), ".config");
  const dir = path.join(base, "openydt-cli");
  fs.mkdirSync(dir, { recursive: true });
  const state = {
    version: pkg.version,
    last_attempt_version: pkg.version,
    updated_at: new Date().toISOString(),
  };
  fs.writeFileSync(path.join(dir, "skills-state.json"), JSON.stringify(state, null, 2) + "\n");
}

if (require.main === module) {
  install().catch((e) => {
    console.error("[openydt]", e.message);
    process.exit(1);
  });
}

module.exports = { downloadReleaseAsset, fetchWithTimeout, skillsSource };
