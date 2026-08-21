const assert = require("node:assert/strict");
const test = require("node:test");

const { downloadReleaseAsset, skillsSource } = require("./install");

function response({ ok = true, status = 200, body = Buffer.alloc(0), json }) {
  return {
    ok,
    status,
    arrayBuffer: async () => body,
    json: async () => json,
  };
}

test("uses the direct GitHub release download when available", async () => {
  const calls = [];
  const content = Buffer.from("direct");
  const fetchImpl = async (url) => {
    calls.push(url);
    return response({ body: content });
  };

  const actual = await downloadReleaseAsset({
    version: "1.2.3",
    asset: "openydt.tar.gz",
    fetchImpl,
  });

  assert.deepEqual(actual, content);
  assert.deepEqual(calls, [
    "https://github.com/xiaowen-0725/openydt-cli/releases/download/v1.2.3/openydt.tar.gz",
  ]);
});

test("falls back to the GitHub API asset download when direct access fails", async () => {
  const calls = [];
  const content = Buffer.from("fallback");
  const fetchImpl = async (url, options) => {
    calls.push({ url, options });
    if (url.startsWith("https://github.com/")) throw new Error("connect timeout");
    if (url.endsWith("/releases/tags/v1.2.3")) {
      return response({
        json: { assets: [{ name: "openydt.tar.gz", url: "https://api.github.test/assets/42" }] },
      });
    }
    if (url === "https://api.github.test/assets/42") return response({ body: content });
    throw new Error(`unexpected URL ${url}`);
  };

  const actual = await downloadReleaseAsset({
    version: "1.2.3",
    asset: "openydt.tar.gz",
    fetchImpl,
  });

  assert.deepEqual(actual, content);
  assert.equal(calls.length, 3);
  assert.equal(calls[2].options.headers.Accept, "application/octet-stream");
  assert.equal(calls[2].options.redirect, "follow");
});

test("reports a missing asset from the fallback release", async () => {
  const fetchImpl = async (url) => {
    if (url.startsWith("https://github.com/")) return response({ ok: false, status: 503 });
    return response({ json: { assets: [] } });
  };

  await assert.rejects(
    downloadReleaseAsset({
      version: "1.2.3",
      asset: "missing.tar.gz",
      fetchImpl,
    }),
    /备用通道未找到安装包 missing\.tar\.gz/,
  );
});

test("pins skills to the npm package release tag", () => {
  assert.equal(
    skillsSource("0.4.2"),
    "https://github.com/xiaowen-0725/openydt-cli/tree/v0.4.2",
  );
  assert.equal(skillsSource("dev"), "xiaowen-0725/openydt-cli");
});

test("prefers skills bundled in the npm package", () => {
  assert.equal(
    skillsSource("0.4.2", "/package/skills", () => true),
    "/package/skills",
  );
});
