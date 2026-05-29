// extract.mjs
// 把停车开放平台的 Vue2 接口文档 (Doc/**/*.vue) 抽成 catalog.json。
//
// 流程:
//   1. 遍历 DOC_ROOT 下所有 *.vue
//   2. @vue/compiler-sfc parse() 取 descriptor.script.content
//   3. @babel/parser 解析为 AST
//   4. 定位 export default -> ObjectExpression -> data() -> return ObjectExpression
//   5. 求值 requestBody / responseCode / infoData
//   6. 按规则生成 catalog 记录 (含 included / excludeReason)
//   7. 写 catalog/catalog.json 并打印统计
//
// 用法: node extract.mjs

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parse as parseSFC } from "@vue/compiler-sfc";
import { parse as parseBabel } from "@babel/parser";
import _traverse from "@babel/traverse";

// @babel/traverse 在 ESM 下是 { default: fn }
const traverse = _traverse.default || _traverse;

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// ---------------------------------------------------------------------------
// 路径配置
// ---------------------------------------------------------------------------
const DOC_ROOT =
  "/Users/zhoujw/develop/git/open-api-front/src/components/Doc";
const OUT_PATH =
  "/Users/zhoujw/develop/tmp/openydt-cli/catalog/catalog.json";

// ---------------------------------------------------------------------------
// domain 映射 (dir -> 命令分组 domain)
// ---------------------------------------------------------------------------
const DOMAIN_MAP = {
  "platform/trade": "trade",
  "platform/park": "park",
  "platform/park/area": "park",
  "platform/parking": "parking",
  "platform/parking/record": "parking",
  "platform/device": "device",
  "platform/device/scan": "device",
  "platform/ticket": "ticket",
  "platform/ticket/vip": "ticket",
  "platform/ticket/blacklist": "blacklist",
  "platform/ticket/redlist": "redlist",
  "platform/ticket/visitor": "visitor",
  "platform/dataAnalysis": "data",
  "platform/preferential/coupon": "coupon",
};

// 一等命令的目录白名单
const INCLUDE_DIRS = new Set([
  "platform/trade",
  "platform/park",
  "platform/park/area",
  "platform/parking",
  "platform/parking/record",
  "platform/device",
  "platform/device/scan",
  "platform/ticket",
  "platform/ticket/vip",
  "platform/ticket/blacklist",
  "platform/ticket/redlist",
  "platform/ticket/visitor",
  "platform/dataAnalysis",
  "platform/preferential/coupon",
]);

// write 前缀 (小写比较)
const WRITE_PREFIXES = [
  "add", "create", "pay", "payback", "cancel", "lock", "unlock", "op",
  "del", "delete", "edit", "update", "save", "send", "sell", "grant",
  "modify", "renew", "refund", "remove", "reduce", "set", "supply",
  "supplement", "mock", "retrieve", "print", "frozen", "deduct", "book",
  "sync", "redlistadd", "delredlist", "upload",
];

// ---------------------------------------------------------------------------
// 排除集 (命中 -> included=false, 记 excludeReason)
// ---------------------------------------------------------------------------
const EXCLUDE_TAG = new Set(["addCarTags", "delCarTags"]);
const EXCLUDE_CERTIFICATE = new Set([
  "getCertificateByInfo",
  "saveOrUpdateMonthTicketCretifi",
  "monthTicketCertifiRuleSave",
  "delMonthTicketCertifiRule",
  "getMonthTicketCertifiRuleList",
  "getMonthTicketCertificateInfoList",
]);
const EXCLUDE_APPOINTMENT_EXACT = new Set([
  "bookOrCancelReservation",
  "checkMonthTicketAppointment",
]);

function matchExcludeSet(cmd) {
  // 返回命中的排除原因(tag/authorize/certificate/appointment)或 null
  if (EXCLUDE_TAG.has(cmd)) return "tag";
  if (cmd.includes("AuthorizeVisitor") || cmd === "allowBuyMonthlyTicket")
    return "authorize";
  if (EXCLUDE_CERTIFICATE.has(cmd)) return "certificate";
  if (cmd.includes("Appointment") || EXCLUDE_APPOINTMENT_EXACT.has(cmd))
    return "appointment";
  return null;
}

// ---------------------------------------------------------------------------
// AST -> JS 值求值器
// ---------------------------------------------------------------------------
function isThisCommonUrls(node) {
  // 判断 node 是否为 this.commonUrls.test|dev|prod
  if (!node || node.type !== "MemberExpression") return false;
  const prop = node.property;
  if (prop.type !== "Identifier") return false;
  if (!["test", "dev", "prod"].includes(prop.name)) return false;
  const obj = node.object; // this.commonUrls
  if (!obj || obj.type !== "MemberExpression") return false;
  if (obj.property.type !== "Identifier" || obj.property.name !== "commonUrls")
    return false;
  if (obj.object.type !== "ThisExpression") return false;
  return true;
}

function keyName(prop) {
  // 键名: Identifier 或 StringLiteral
  const k = prop.key;
  if (!k) return null;
  if (k.type === "Identifier") return k.name;
  if (k.type === "StringLiteral") return k.value;
  if (k.type === "NumericLiteral") return String(k.value);
  return null;
}

// 哨兵: 表示无法求值
const UNRESOLVED = Symbol("unresolved");

function evalNode(node) {
  if (!node) return UNRESOLVED;
  switch (node.type) {
    case "StringLiteral":
      return node.value;
    case "NumericLiteral":
      return node.value;
    case "BooleanLiteral":
      return node.value;
    case "NullLiteral":
      return null;
    case "TemplateLiteral": {
      // 仅当无内嵌表达式时取 quasi 拼接
      if (node.expressions && node.expressions.length > 0) return UNRESOLVED;
      return node.quasis.map((q) => q.value.cooked ?? q.value.raw).join("");
    }
    case "UnaryExpression": {
      // 负数 (含 -0)、正号
      if (node.operator === "-" || node.operator === "+") {
        const arg = evalNode(node.argument);
        if (typeof arg === "number") {
          return node.operator === "-" ? -arg : +arg;
        }
      }
      return UNRESOLVED;
    }
    case "BinaryExpression": {
      if (node.operator === "+") {
        // cmd 模式: this.commonUrls.X + "字面量"  (任意一侧)
        const left = node.left;
        const right = node.right;
        if (isThisCommonUrls(left) && right.type === "StringLiteral") {
          return right.value;
        }
        if (isThisCommonUrls(right) && left.type === "StringLiteral") {
          return left.value;
        }
        // 两侧都是可求值字面量时做字符串/数字拼接
        const lv = evalNode(left);
        const rv = evalNode(right);
        if (
          lv !== UNRESOLVED &&
          rv !== UNRESOLVED &&
          (typeof lv === "string" || typeof lv === "number") &&
          (typeof rv === "string" || typeof rv === "number")
        ) {
          return lv + rv;
        }
      }
      return UNRESOLVED;
    }
    case "ArrayExpression": {
      const arr = [];
      for (const el of node.elements) {
        if (el === null) {
          arr.push(null); // 稀疏数组洞
          continue;
        }
        if (el.type === "SpreadElement") {
          const spread = evalNode(el.argument);
          if (Array.isArray(spread)) arr.push(...spread);
          // 无法求值的 spread 直接忽略
          continue;
        }
        const v = evalNode(el);
        arr.push(v === UNRESOLVED ? null : v);
      }
      return arr;
    }
    case "ObjectExpression": {
      const obj = {};
      for (const prop of node.properties) {
        if (prop.type === "SpreadElement" || prop.type === "SpreadProperty") {
          const spread = evalNode(prop.argument);
          if (spread && typeof spread === "object" && !Array.isArray(spread)) {
            Object.assign(obj, spread);
          }
          continue;
        }
        if (prop.type === "ObjectMethod") {
          // 方法不求值
          continue;
        }
        if (prop.type !== "ObjectProperty") continue;
        const name = keyName(prop);
        if (name === null) continue;
        const v = evalNode(prop.value);
        obj[name] = v === UNRESOLVED ? null : v;
      }
      return obj;
    }
    default:
      return UNRESOLVED;
  }
}

// ---------------------------------------------------------------------------
// 定位 data() 的 return ObjectExpression
// ---------------------------------------------------------------------------
function findDataReturnObject(ast) {
  let result = null;

  traverse(ast, {
    ExportDefaultDeclaration(p) {
      const decl = p.node.declaration;
      if (!decl || decl.type !== "ObjectExpression") return;

      // 找 property 'data'
      let dataProp = null;
      for (const prop of decl.properties) {
        if (prop.type === "ObjectMethod") {
          if (keyName(prop) === "data") {
            dataProp = prop;
            break;
          }
        } else if (prop.type === "ObjectProperty") {
          if (keyName(prop) === "data") {
            const v = prop.value;
            if (
              v &&
              (v.type === "FunctionExpression" ||
                v.type === "ArrowFunctionExpression")
            ) {
              dataProp = prop;
              break;
            }
          }
        }
      }
      if (!dataProp) return;

      // 取函数体
      let body = null;
      if (dataProp.type === "ObjectMethod") {
        body = dataProp.body;
      } else {
        body = dataProp.value.body;
      }
      if (!body) return;

      // ArrowFunction 直接返回 ObjectExpression: data: () => ({...})
      if (body.type === "ObjectExpression") {
        result = body;
        p.stop();
        return;
      }

      if (body.type !== "BlockStatement") return;

      // 找 ReturnStatement
      for (const stmt of body.body) {
        if (stmt.type === "ReturnStatement") {
          let arg = stmt.argument;
          // 解开括号
          while (arg && arg.type === "ParenthesizedExpression") {
            arg = arg.expression;
          }
          if (arg && arg.type === "ObjectExpression") {
            result = arg;
            p.stop();
            return;
          }
        }
      }
    },
  });

  return result;
}

// 从 ObjectExpression 中取某个 top-level property 的 AST 节点
function getPropNode(objExpr, name) {
  if (!objExpr) return null;
  for (const prop of objExpr.properties) {
    if (prop.type === "ObjectProperty" && keyName(prop) === name) {
      return prop.value;
    }
    if (prop.type === "ObjectMethod" && keyName(prop) === name) {
      return null; // 方法不取值
    }
  }
  return null;
}

// 从 infoData 的 ObjectExpression 直接读取 cmd (t_url) 的 AST 模式
function extractCmdFromTUrl(infoObjExpr) {
  if (!infoObjExpr) return null;
  const tUrlNode = getPropNode(infoObjExpr, "t_url");
  if (!tUrlNode) return null;
  const v = evalNode(tUrlNode);
  if (typeof v === "string" && v.length > 0) return v;
  return null;
}

// ---------------------------------------------------------------------------
// params 展平
// ---------------------------------------------------------------------------
// isParamHeader returns true only when a group's tabledata is the standard
// 参数名/是否必填/类型/说明 column mapping. Other tables (e.g. 扣减规则说明,
// "注：" rows) are explanatory and must NOT be treated as request parameters.
function isParamHeader(tabledata) {
  if (!Array.isArray(tabledata)) return true; // 无表头信息时不强制(向后兼容)
  const byTd = {};
  for (const c of tabledata) {
    if (c && typeof c === "object" && c.tdName) byTd[c.tdName] = String(c.tableHead || "");
  }
  // info1 必须是参数名列
  return (byTd.info1 || "").includes("参数名");
}

function flattenParams(qParams) {
  const out = [];
  if (!Array.isArray(qParams)) return out;
  for (const group of qParams) {
    if (!group || typeof group !== "object") continue;
    if (!isParamHeader(group.tabledata)) continue; // 跳过解释性表格
    const groupName = group.titleName != null ? group.titleName : null;
    const td = group.td;
    if (!Array.isArray(td)) continue;
    for (const row of td) {
      if (!row || typeof row !== "object") continue;
      const name = row.info1 != null ? String(row.info1).trim() : "";
      if (name === "") continue; // 丢弃无参数名的说明行(如 "注：...")
      out.push({
        name,
        required: row.info2 != null && String(row.info2).trim() === "是",
        type: row.info3 != null ? String(row.info3).trim() : null,
        desc: row.info4 != null ? row.info4 : null,
        group: groupName,
      });
    }
  }
  return out;
}

// ---------------------------------------------------------------------------
// 工具
// ---------------------------------------------------------------------------
function walkVueFiles(root) {
  const result = [];
  function rec(dir) {
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const e of entries) {
      const full = path.join(dir, e.name);
      if (e.isDirectory()) rec(full);
      else if (e.isFile() && e.name.endsWith(".vue")) result.push(full);
    }
  }
  rec(root);
  result.sort();
  return result;
}

function relDir(fullPath) {
  const rel = path.relative(DOC_ROOT, fullPath);
  const dir = path.dirname(rel);
  if (dir === "." || dir === "") return "";
  return dir.split(path.sep).join("/");
}

function slugForDir(dir) {
  // 其它 dir: 取第二段(若存在)否则第一段
  const segs = dir.split("/").filter(Boolean);
  if (segs.length === 0) return "misc";
  if (segs.length >= 2) return segs[1];
  return segs[0];
}

// 只读前缀。默认偏保守:不是明确只读前缀的一律视为 write(需 --yes),
// 这样 correct*/cloudOpenGate/xxxEdit/xxxSave/roadside* 等带副作用接口不会漏判。
const READ_PREFIXES = [
  "get", "query", "check", "list", "count", "find",
  "search", "validate", "exist", "common", "is", "has",
];
function isWriteCmd(cmd) {
  const lc = String(cmd).toLowerCase();
  return !READ_PREFIXES.some((p) => lc.startsWith(p));
}

function isWebhook(pattern) {
  if (typeof pattern !== "string") return false;
  const provides =
    pattern.includes("第三方接入系统提供接口") ||
    pattern.includes("第三方提供接口");
  const platformActive = pattern.includes("智慧停车开放平台主动请求");
  return provides && platformActive;
}

function isVemsOnly(fitSystem) {
  // VEMS-only = 适用系统明确是 VEMS/传统, 且没有任何云端适用标记。
  // 注意:平台云产品名包括「云停车场」与「智汇云车场/智汇云/云车场」, 不能只认「云停车场」精确串,
  // 否则智汇云车场这类云接口会被误判为 VEMS。
  if (typeof fitSystem !== "string") return false;
  const s = fitSystem.trim();
  if (s.length === 0) return false;
  const isCloud = /云停车场|智汇云|云车场/.test(s);
  const isVems = /VEMS|传统/i.test(s);
  return isVems && !isCloud;
}

function safeStringify(value) {
  if (value === UNRESOLVED || value === undefined) return "";
  try {
    return JSON.stringify(value, null, 4);
  } catch {
    return "";
  }
}

// ---------------------------------------------------------------------------
// 主流程
// ---------------------------------------------------------------------------
function main() {
  const files = walkVueFiles(DOC_ROOT);
  const interfaces = [];
  let parseErrorCount = 0;

  for (const file of files) {
    const fileName = path.basename(file, ".vue");
    const dir = relDir(file);
    let cmd = fileName;
    let explain = null;
    let fitSystem = null;
    let pattern = null;
    let params = [];
    let sampleBody = "";
    let sampleResponse = "";
    let parsedOK = false;
    let hasEndpoint = false; // 是否有真实的 t_url(this.commonUrls.X + "cmd") — 流程说明页没有

    try {
      const src = fs.readFileSync(file, "utf8");
      const { descriptor } = parseSFC(src, { filename: file });
      const scriptContent =
        (descriptor.script && descriptor.script.content) ||
        (descriptor.scriptSetup && descriptor.scriptSetup.content) ||
        "";

      if (scriptContent.trim().length > 0) {
        const ast = parseBabel(scriptContent, {
          sourceType: "module",
          plugins: ["objectRestSpread"],
        });

        const dataReturn = findDataReturnObject(ast);
        if (dataReturn) {
          parsedOK = true;

          // requestBody / responseCode
          const reqNode = getPropNode(dataReturn, "requestBody");
          const respNode = getPropNode(dataReturn, "responseCode");
          if (reqNode) sampleBody = safeStringify(evalNode(reqNode));
          if (respNode) sampleResponse = safeStringify(evalNode(respNode));

          // infoData
          const infoNode = getPropNode(dataReturn, "infoData");
          if (infoNode && infoNode.type === "ObjectExpression") {
            // cmd from t_url
            const tCmd = extractCmdFromTUrl(infoNode);
            if (tCmd) {
              cmd = tCmd;
              hasEndpoint = true;
            }

            const info = evalNode(infoNode);
            if (info && typeof info === "object") {
              explain = info.explain != null ? info.explain : null;
              fitSystem = info.fitSystem != null ? info.fitSystem : null;
              pattern = info.pattern != null ? info.pattern : null;
              params = flattenParams(info.Q_params);
            }
          }
        }
      }
    } catch (err) {
      parseErrorCount++;
      // 不抛错, 记录后继续 (cmd fallback 文件名)
      interfaces.__lastError = err.message;
    }

    if (!parsedOK) parseErrorCount += 0; // parsedOK 计数在汇总时统计

    // domain
    let domain = DOMAIN_MAP[dir];
    let inIncludeDir = INCLUDE_DIRS.has(dir);
    if (!domain) {
      domain = slugForDir(dir);
    }

    // direction
    const direction = isWebhook(pattern) ? "webhook" : "callable";

    // readwrite
    const readwrite = isWriteCmd(cmd) ? "write" : "read";

    // included + excludeReason (按优先级)
    const isDeprecated = dir.startsWith("deprecated") || dir === "hidden";
    const excludeHit = matchExcludeSet(cmd);
    const vemsOnly = isVemsOnly(fitSystem);

    let included = false;
    let excludeReason = "";

    // included 规则
    included =
      hasEndpoint && // 流程说明页(无 t_url)不是可调用接口
      direction === "callable" &&
      inIncludeDir &&
      excludeHit === null &&
      !vemsOnly &&
      !isDeprecated;

    if (included) {
      excludeReason = "";
    } else {
      // excludeReason 优先级:
      // no-endpoint > deprecated > webhook > 排除集 > vems-only > out-of-scope-domain
      if (!hasEndpoint) {
        excludeReason = "no-endpoint";
      } else if (isDeprecated) {
        excludeReason = "deprecated";
      } else if (direction === "webhook") {
        excludeReason = "webhook";
      } else if (excludeHit !== null) {
        excludeReason = excludeHit;
      } else if (vemsOnly) {
        excludeReason = "vems-only";
      } else if (!inIncludeDir) {
        excludeReason = "out-of-scope-domain";
      } else {
        // 在 includeDir、callable、非排除、非 vems、非 deprecated 却仍未 included
        // 理论上不会到这里; 兜底
        excludeReason = "out-of-scope-domain";
      }
    }

    interfaces.push({
      cmd,
      dir,
      domain,
      explain,
      fitSystem,
      pattern,
      direction,
      readwrite,
      params,
      sampleBody,
      sampleResponse,
      included,
      excludeReason,
    });
  }

  // 写出 catalog.json
  const catalog = {
    generatedFrom: "open-api-front Doc",
    count: interfaces.length,
    interfaces,
  };
  fs.mkdirSync(path.dirname(OUT_PATH), { recursive: true });
  fs.writeFileSync(OUT_PATH, JSON.stringify(catalog, null, 2) + "\n", "utf8");

  // -------------------------------------------------------------------------
  // 统计打印
  // -------------------------------------------------------------------------
  const includedCount = interfaces.filter((i) => i.included).length;
  const excludedCount = interfaces.length - includedCount;

  const byDomainIncluded = {};
  for (const i of interfaces) {
    if (i.included) {
      byDomainIncluded[i.domain] = (byDomainIncluded[i.domain] || 0) + 1;
    }
  }

  const byDirection = {};
  for (const i of interfaces) {
    byDirection[i.direction] = (byDirection[i.direction] || 0) + 1;
  }

  const byExcludeReason = {};
  for (const i of interfaces) {
    const r = i.included ? "(included)" : i.excludeReason || "(empty)";
    byExcludeReason[r] = (byExcludeReason[r] || 0) + 1;
  }

  console.log("==================================================");
  console.log("catalog.json written to:", OUT_PATH);
  console.log("total interfaces      :", interfaces.length);
  console.log("included              :", includedCount);
  console.log("excluded              :", excludedCount);
  console.log("parse errors          :", parseErrorCount);
  console.log("--------------------------------------------------");
  console.log("included by domain    :", JSON.stringify(byDomainIncluded));
  console.log("--------------------------------------------------");
  console.log("direction distribution:", JSON.stringify(byDirection));
  console.log("--------------------------------------------------");
  console.log("excludeReason dist    :", JSON.stringify(byExcludeReason, null, 2));
  console.log("==================================================");

  // 关键 cmd 校验
  const keyCmds = [
    "getParkFee", "payParkFee", "addOnlineMonthTicketType", "addBlackListCar",
    "addVisitorCarNew", "addSpecialCarType", "createTrader",
    "createCouponTemplate", "sellCoupon", "sendCoupon",
    "supplyCarIn", "supplementParkingRecordIn", "mockInOut",
  ];
  console.log("KEY CMD CHECK:");
  for (const kc of keyCmds) {
    const rec = interfaces.find((i) => i.cmd === kc);
    if (!rec) {
      console.log(`  ${kc.padEnd(28)} -> NOT FOUND`);
    } else {
      console.log(
        `  ${kc.padEnd(28)} -> dir=${rec.dir} domain=${rec.domain} ` +
          `dir=${rec.direction} rw=${rec.readwrite} included=${rec.included} ` +
          `params=${rec.params.length} excl=${rec.excludeReason || "-"}`
      );
    }
  }
  console.log("==================================================");
}

main();
