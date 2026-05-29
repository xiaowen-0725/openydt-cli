#!/usr/bin/env node
// openydt skill 格式校验脚本（Node 内置 fs，无第三方依赖）
// 校验项：
//   1. frontmatter 存在，且含非空 name 与 description；name 与目录名一致。     -> 不满足为 FAIL
//   2. 域技能（非 shared/explorer/maker）正文含对 ../openydt-shared/SKILL.md 的 Read 引用。 -> 不满足为 FAIL
//   3. 正文中出现的 `openydt <domain> <use>`，其 <use> 能在 cmd/gen/<domain>.go 中
//      找到（作为 Use 或 alias）；找不到（或无对应 .go 文件）列为 WARN。
//   每个 skill 输出 PASS/FAIL/WARN 摘要；末尾总计；有 FAIL 时退出码非 0。

'use strict';

const fs = require('fs');
const path = require('path');

const ROOT = path.resolve(__dirname, '..', '..');
const SKILLS_DIR = path.join(ROOT, 'skills');
const GEN_DIR = path.join(ROOT, 'cmd', 'gen');

// 非域技能（不要求 Read 共享基座）：shared / explorer / maker
const NON_DOMAIN_SKILLS = new Set([
  'openydt-shared',
  'openydt-api-explorer',
  'openydt-skill-maker',
]);

// --- 工具函数 -------------------------------------------------------------

// 解析 frontmatter：返回 { ok, raw, fields } —— 仅做行级 key: value 提取，够用即可。
function parseFrontmatter(content) {
  // frontmatter 必须在文件最开头，以 --- 起，以 --- 止。
  if (!content.startsWith('---')) {
    return { ok: false, fields: {} };
  }
  const lines = content.split(/\r?\n/);
  // 第 0 行是 ---，找下一个独立的 --- 作为结束。
  let endIdx = -1;
  for (let i = 1; i < lines.length; i++) {
    if (lines[i].trim() === '---') {
      endIdx = i;
      break;
    }
  }
  if (endIdx === -1) {
    return { ok: false, fields: {} };
  }
  const fields = {};
  // 仅解析顶层（无缩进）的 key: value 行，避免误读嵌套 metadata 子键。
  const re = /^([A-Za-z0-9_-]+):\s*(.*)$/;
  for (let i = 1; i < endIdx; i++) {
    const line = lines[i];
    if (/^\s/.test(line)) continue; // 跳过缩进行（嵌套）
    const m = line.match(re);
    if (m) {
      let val = m[2].trim();
      // 去掉成对引号
      if (
        (val.startsWith('"') && val.endsWith('"')) ||
        (val.startsWith("'") && val.endsWith("'"))
      ) {
        val = val.slice(1, -1);
      }
      fields[m[1]] = val;
    }
  }
  const body = lines.slice(endIdx + 1).join('\n');
  return { ok: true, fields, body };
}

// 从 cmd/gen/<domain>.go 中收集所有 Use 与 alias。带缓存。
const goUseCache = new Map(); // domain -> Set<string> | null（null 表示无文件）
function collectGoUses(domain) {
  if (goUseCache.has(domain)) return goUseCache.get(domain);
  const file = path.join(GEN_DIR, domain + '.go');
  if (!fs.existsSync(file)) {
    goUseCache.set(domain, null);
    return null;
  }
  const src = fs.readFileSync(file, 'utf8');
  const set = new Set();
  // Use: "xxx"
  const useRe = /Use:\s*"([^"]+)"/g;
  let m;
  while ((m = useRe.exec(src)) !== null) {
    // Use 可能形如 "get-park-fee" 或 "trade"，取第一个空格前的 token
    const tok = m[1].split(/\s+/)[0];
    set.add(tok);
  }
  // Aliases: []string{"a", "b"}
  const aliasRe = /Aliases:\s*\[\]string\{([^}]*)\}/g;
  while ((m = aliasRe.exec(src)) !== null) {
    const inner = m[1];
    const strRe = /"([^"]+)"/g;
    let sm;
    while ((sm = strRe.exec(inner)) !== null) {
      set.add(sm[1]);
    }
  }
  goUseCache.set(domain, set);
  return set;
}

// 从正文中提取 `openydt <domain> <use>` 命令调用。
// 仅当 domain 为纯小写字母、use 形如命令（小写字母/数字/连字符，含至少一个字母）时纳入。
function extractCommands(body) {
  const cmds = [];
  const re = /\bopenydt\s+([a-z][a-z0-9]*)\s+([a-z][a-z0-9-]*)/g;
  let m;
  while ((m = re.exec(body)) !== null) {
    cmds.push({ domain: m[1], use: m[2] });
  }
  return cmds;
}

// --- 主流程 ---------------------------------------------------------------

function main() {
  if (!fs.existsSync(SKILLS_DIR)) {
    console.error('skills 目录不存在: ' + SKILLS_DIR);
    process.exit(2);
  }

  const entries = fs
    .readdirSync(SKILLS_DIR, { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name)
    .sort();

  let totalPass = 0;
  let totalFail = 0;
  let totalWarn = 0;
  const failedSkills = [];
  const warnedSkills = [];

  console.log('openydt skill 格式校验');
  console.log('skills 目录: ' + SKILLS_DIR);
  console.log('cmd/gen 目录: ' + GEN_DIR);
  console.log('='.repeat(72));

  for (const dirName of entries) {
    const skillDir = path.join(SKILLS_DIR, dirName);
    const skillMd = path.join(skillDir, 'SKILL.md');

    const fails = [];
    const warns = [];

    if (!fs.existsSync(skillMd)) {
      fails.push('缺少 SKILL.md');
      printSkill(dirName, 'FAIL', fails, warns);
      totalFail++;
      failedSkills.push(dirName);
      continue;
    }

    const content = fs.readFileSync(skillMd, 'utf8');
    const fm = parseFrontmatter(content);

    // 1. frontmatter 校验
    if (!fm.ok) {
      fails.push('frontmatter 缺失或未正确闭合（需以 --- 开始并以 --- 结束）');
    } else {
      const name = fm.fields.name;
      const desc = fm.fields.description;
      if (!name || name.trim() === '') {
        fails.push('frontmatter 缺少非空 name');
      } else if (name.trim() !== dirName) {
        fails.push(
          'name "' + name.trim() + '" 与目录名 "' + dirName + '" 不一致'
        );
      }
      if (!desc || desc.trim() === '') {
        fails.push('frontmatter 缺少非空 description');
      }
    }

    const body = fm.ok ? fm.body : content;

    // 2. 域技能须含对共享基座的 Read 引用
    const isDomainSkill = !NON_DOMAIN_SKILLS.has(dirName);
    if (isDomainSkill) {
      const hasSharedRef = /openydt-shared\/SKILL\.md/.test(body);
      // 进一步确认是“Read 引用”：正文出现 Read 字样且指向共享基座，
      // 或出现指向 ../openydt-shared/SKILL.md 的链接（约定即为 Read 共享基座）。
      const mentionsRead = /Read/.test(body) || /读取/.test(body);
      if (!hasSharedRef) {
        fails.push('域技能正文缺少对 ../openydt-shared/SKILL.md 的引用');
      } else if (!mentionsRead) {
        fails.push(
          '域技能引用了 openydt-shared/SKILL.md 但未体现 Read/读取 指令'
        );
      }
    }

    // 3. 命令 use 校验（找不到 -> WARN）
    const cmds = extractCommands(body);
    const seen = new Set();
    for (const { domain, use } of cmds) {
      const key = domain + ' ' + use;
      if (seen.has(key)) continue;
      seen.add(key);
      const uses = collectGoUses(domain);
      if (uses === null) {
        warns.push(
          'openydt ' + domain + ' ' + use + '：无 cmd/gen/' + domain + '.go（无法校验）'
        );
      } else if (!uses.has(use)) {
        warns.push(
          'openydt ' + domain + ' ' + use + '：在 cmd/gen/' + domain + '.go 中未找到 Use/alias "' + use + '"'
        );
      }
    }

    // 汇总该 skill 状态
    let status;
    if (fails.length > 0) {
      status = 'FAIL';
      totalFail++;
      failedSkills.push(dirName);
    } else if (warns.length > 0) {
      status = 'WARN';
      totalWarn++;
      warnedSkills.push(dirName);
    } else {
      status = 'PASS';
      totalPass++;
    }
    printSkill(dirName, status, fails, warns);
  }

  console.log('='.repeat(72));
  console.log(
    '总计: ' +
      entries.length +
      ' 个 skill — PASS ' +
      totalPass +
      ' / WARN ' +
      totalWarn +
      ' / FAIL ' +
      totalFail
  );
  if (failedSkills.length > 0) {
    console.log('FAIL: ' + failedSkills.join(', '));
  }
  if (warnedSkills.length > 0) {
    console.log('WARN: ' + warnedSkills.join(', '));
  }

  process.exit(totalFail > 0 ? 1 : 0);
}

function printSkill(name, status, fails, warns) {
  console.log('[' + status + '] ' + name);
  for (const f of fails) {
    console.log('    FAIL: ' + f);
  }
  for (const w of warns) {
    console.log('    WARN: ' + w);
  }
}

main();
