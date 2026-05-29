// 生成「停车服务3.0」接口支持对照表 -> SUPPORTED_INTERFACES.html(自包含, 可搜索/筛选)。
// 仅含 platform/* (停车服务3.0); 城市运营/第三方接入/已废弃等不计入。
// 用法: node tools/gen-support-doc.mjs
import fs from "fs";
import path from "path";

const __root = path.resolve(path.dirname(new URL(import.meta.url).pathname), "..");
const cat = JSON.parse(fs.readFileSync(path.join(__root, "catalog/catalog.json"), "utf8"));
const all = cat.interfaces || [];
const items = all.filter((it) => (it.dir || "").startsWith("platform")); // 停车服务3.0

function kebab(s) {
  let o = "";
  for (let i = 0; i < s.length; i++) {
    const c = s[i];
    if (c >= "A" && c <= "Z") {
      if (i > 0) {
        const p = s[i - 1], n = s[i + 1] || "", nl = n >= "a" && n <= "z";
        if ((p >= "a" && p <= "z") || (p >= "0" && p <= "9") || (p >= "A" && p <= "Z" && nl)) o += "-";
      }
      o += c.toLowerCase();
    } else o += c;
  }
  return o;
}

function name(it) {
  let s = (it.explain || "").replace(/\s+/g, "");
  for (const p of ["第三方接入系统请求智慧停车开放平台", "第三方接入系统请求一点停开放平台",
    "智慧停车开放平台主动上报", "智慧停车开放平台主动请求第三方接入系统",
    "智慧停车开放平台调用第三方接口", "第三方接入系统主动推送", "第三方"])
    if (s.startsWith(p)) { s = s.slice(p.length); break; }
  return s || it.cmd;
}

const CAT = {
  "platform/trade": "停车缴费", "platform/park": "车场信息", "platform/park/area": "车场区域",
  "platform/parking": "停车记录", "platform/parking/record": "停车记录·明细",
  "platform/device": "设备控制", "platform/device/scan": "设备控制·扫码",
  "platform/ticket": "月票 / VIP", "platform/ticket/vip": "会员车类型",
  "platform/ticket/blacklist": "黑名单", "platform/ticket/redlist": "白名单",
  "platform/ticket/visitor": "访客", "platform/ticket/query": "月票查询",
  "platform/preferential/coupon": "电子券", "platform/preferential/score": "积分",
  "platform/preferential/thirdCoupon": "第三方券", "platform/dataAnalysis": "数据分析",
  "platform/invoice": "停车发票", "platform/upward": "上报 / 回调(平台推送)",
  "platform/cloud": "云车场", "platform/other": "其他", "platform/other/h5": "其他·H5",
};
const catName = (d) => CAT[d] || d;

const reasonCN = {
  "vems-only": "VEMS 专用", "out-of-scope-domain": "未纳入(发票/其它)",
  "tag": "车辆标签类", "authorize": "月票/访客授权类", "certificate": "月票凭证类", "appointment": "预约类",
};

function status(it) {
  if (it.included) return { key: "supported", label: "已支持", how: "openydt " + it.domain + " " + kebab(it.cmd) };
  if (it.direction === "webhook") return { key: "webhook", label: "Webhook", how: "平台主动推送回调, 需自建接收端验签(CLI 不主动调用)" };
  if (it.excludeReason === "no-endpoint") return { key: "none", label: "非接口", how: "文档/流程说明页(无可调用 cmd)" };
  if (it.excludeReason === "deprecated") return { key: "deprecated", label: "已废弃", how: "—" };
  return { key: "todo", label: "未做命令", how: "openydt api " + it.cmd + " · " + (reasonCN[it.excludeReason] || "待扩展") };
}

const rows = items.map((it) => {
  const st = status(it);
  return { category: catName(it.dir), dir: it.dir, cmd: it.cmd, name: name(it),
    rw: it.readwrite === "write" ? "写" : "读", status: st.key, statusLabel: st.label, how: st.how };
});

const counts = {};
for (const r of rows) counts[r.status] = (counts[r.status] || 0) + 1;

const DATA = JSON.stringify(rows);
const html = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>停车服务3.0 · 接口支持对照表 — openydt-cli</title>
<style>
:root{--bg:#f5f7fa;--card:#fff;--line:#e8ecf2;--txt:#1f2937;--mut:#6b7280;--brand:#2f7bff}
*{box-sizing:border-box}body{margin:0;font:14px/1.6 -apple-system,BlinkMacSystemFont,"PingFang SC","Microsoft YaHei",sans-serif;background:var(--bg);color:var(--txt)}
header{background:linear-gradient(90deg,#1e63e9,#2f7bff);color:#fff;padding:18px 28px}
header h1{margin:0;font-size:20px;font-weight:700}header .sub{opacity:.9;font-size:13px;margin-top:4px}
.wrap{max-width:1180px;margin:0 auto;padding:20px 28px 60px}
.cards{display:flex;gap:12px;flex-wrap:wrap;margin:18px 0}
.card{background:var(--card);border:1px solid var(--line);border-radius:10px;padding:12px 16px;min-width:120px}
.card .n{font-size:24px;font-weight:700}.card .l{color:var(--mut);font-size:12px}
.toolbar{display:flex;gap:10px;align-items:center;flex-wrap:wrap;margin:14px 0;position:sticky;top:0;background:var(--bg);padding:8px 0;z-index:5}
input#q{flex:1;min-width:220px;padding:9px 12px;border:1px solid var(--line);border-radius:8px;font-size:14px}
.chip{padding:6px 12px;border:1px solid var(--line);border-radius:20px;background:#fff;cursor:pointer;font-size:13px;user-select:none}
.chip.on{border-color:var(--brand);color:var(--brand);background:#eef4ff;font-weight:600}
h2.cat{font-size:15px;margin:26px 0 8px;padding-left:10px;border-left:4px solid var(--brand)}
h2.cat .c{color:var(--mut);font-weight:400;font-size:13px;margin-left:8px}
table{width:100%;border-collapse:collapse;background:var(--card);border:1px solid var(--line);border-radius:10px;overflow:hidden}
th,td{text-align:left;padding:9px 12px;border-bottom:1px solid var(--line);vertical-align:top}
th{background:#fafbfd;color:var(--mut);font-weight:600;font-size:12px}
tr:last-child td{border-bottom:none}
code{background:#f1f3f7;padding:1px 6px;border-radius:5px;font-family:ui-monospace,Menlo,monospace;font-size:12.5px}
.b{display:inline-block;padding:2px 9px;border-radius:12px;font-size:12px;font-weight:600;white-space:nowrap}
.b.supported{background:#e6f6ec;color:#13864a}.b.todo{background:#fff3e0;color:#b26a00}
.b.webhook{background:#eef0ff;color:#4a45c9}.b.deprecated{background:#f3f4f6;color:#888}.b.none{background:#f3f4f6;color:#888}
.rw{color:var(--mut);font-size:12px}.how{color:var(--mut);font-size:12.5px}
.hidden{display:none}.muted{color:var(--mut)}
</style></head><body>
<header><h1>智慧停车开放平台 · 停车服务3.0 接口支持对照表</h1>
<div class="sub">openydt-cli 支持情况 · 共 ${rows.length} 个接口 · 仅含停车服务3.0(platform/*),不含城市运营/第三方接入</div></header>
<div class="wrap">
<div class="cards">
<div class="card"><div class="n" style="color:#13864a">${counts.supported||0}</div><div class="l">✅ 已支持(一等命令)</div></div>
<div class="card"><div class="n" style="color:#b26a00">${counts.todo||0}</div><div class="l">⚪ 未做命令(api 兜底/待扩展)</div></div>
<div class="card"><div class="n" style="color:#4a45c9">${counts.webhook||0}</div><div class="l">🔔 Webhook(平台推送)</div></div>
<div class="card"><div class="n" style="color:#888">${counts.none||0}</div><div class="l">➖ 非接口(流程页)</div></div>
</div>
<div class="toolbar">
<input id="q" placeholder="搜索 接口名称 / cmd …">
<span class="chip on" data-f="all">全部</span>
<span class="chip" data-f="supported">✅ 已支持</span>
<span class="chip" data-f="todo">⚪ 未做命令</span>
<span class="chip" data-f="webhook">🔔 Webhook</span>
<span class="chip" data-f="none">➖ 非接口</span>
</div>
<div id="list"></div>
<p class="muted" style="margin-top:30px">扩展指引:从「⚪ 未做命令」里挑业务需要的接口 → 在 <code>tools/extractor/extract.mjs</code> 的纳入规则放开对应目录/cmd → <code>make catalog generate</code> 即自动生成一等命令与 skill。本表由 <code>node tools/gen-support-doc.mjs</code> 从 <code>catalog/catalog.json</code> 生成。</p>
</div>
<script>
const DATA=${DATA};
let filter="all",kw="";
const list=document.getElementById("list");
function esc(s){return (s||"").replace(/[&<>]/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;"}[c]))}
function render(){
  const rows=DATA.filter(r=>(filter==="all"||r.status===filter)&&(!kw||(r.name+r.cmd).toLowerCase().includes(kw)));
  const cats=[...new Set(rows.map(r=>r.category))];
  list.innerHTML=cats.map(cat=>{
    const rs=rows.filter(r=>r.category===cat);
    const sup=rs.filter(r=>r.status==="supported").length;
    return '<h2 class="cat">'+esc(cat)+'<span class="c">'+rs.length+' 个 · 已支持 '+sup+'</span></h2>'+
      '<table><thead><tr><th style="width:24%">接口名称</th><th style="width:22%">cmd</th><th style="width:6%">读写</th><th style="width:12%">状态</th><th>调用方式 / 原因</th></tr></thead><tbody>'+
      rs.map(r=>'<tr><td>'+esc(r.name)+'</td><td><code>'+esc(r.cmd)+'</code></td><td class="rw">'+r.rw+'</td><td><span class="b '+r.status+'">'+r.statusLabel+'</span></td><td class="how">'+esc(r.how)+'</td></tr>').join("")+
      '</tbody></table>';
  }).join("")||'<p class="muted">无匹配结果</p>';
}
document.querySelectorAll(".chip").forEach(c=>c.onclick=()=>{document.querySelectorAll(".chip").forEach(x=>x.classList.remove("on"));c.classList.add("on");filter=c.dataset.f;render();});
document.getElementById("q").oninput=e=>{kw=e.target.value.trim().toLowerCase();render();};
render();
</script></body></html>`;

fs.writeFileSync(path.join(__root, "SUPPORTED_INTERFACES.html"), html);
console.log("wrote SUPPORTED_INTERFACES.html (停车服务3.0)");
console.log("rows:", rows.length, "counts:", JSON.stringify(counts));
