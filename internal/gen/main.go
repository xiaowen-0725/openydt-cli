// Command gen turns catalog/catalog.json into per-domain cobra commands under
// cmd/gen. Run from the repo root: `go run ./internal/gen`.
//
// Only interfaces with included==true && direction=="callable" become first-class
// commands. Everything else stays reachable via `openydt api <cmd>`.
package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xiaowen-0725/openydt-cli/internal/catalog"
	"github.com/xiaowen-0725/openydt-cli/internal/strutil"
)

type Param struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     string `json:"type"`
	Desc     string `json:"desc"`
	Group    string `json:"group"`
}

type Iface struct {
	Cmd           string  `json:"cmd"`
	Dir           string  `json:"dir"`
	Domain        string  `json:"domain"`
	Explain       string  `json:"explain"`
	FitSystem     string  `json:"fitSystem"`
	Pattern       string  `json:"pattern"`
	Direction     string  `json:"direction"`
	ReadWrite     string  `json:"readwrite"`
	Params        []Param `json:"params"`
	SampleBody    string  `json:"sampleBody"`
	Included      bool    `json:"included"`
	ExcludeReason string  `json:"excludeReason"`
}

type Catalog struct {
	Count      int     `json:"count"`
	Interfaces []Iface `json:"interfaces"`
}

// forceIncluded re-includes interfaces the extractor marked excluded (e.g.
// excludeReason=="deprecated") but that we still want as first-class commands,
// remapping each into a friendlier domain and (optionally) correcting an upstream
// Doc explain that is wrong/misattributed. Keeps catalog.json (a generated
// product) untouched and survives re-extraction.
type forceIncluded struct {
	Domain  string
	Explain string // overrides catalog explain when non-empty (fixes bad Doc text)
}

var forceInclude = map[string]forceIncluded{
	// 盘点离场(写) -> openydt parking inventory-car. 上游 Doc 的 explain 误写为
	// "获取城市地图车辆分布数据",此处校正。
	"inventoryCar": {Domain: "parking", Explain: "盘点离场:将 enterTimeEnd 之前的在场停车记录批量盘点离场"},
	// 查盘点记录(读) -> openydt parking get-inventory-record. catalog explain 正确,不覆盖。
	"getInventoryRecord": {Domain: "parking"},
}

var domainShort = map[string]string{
	"trade": "停车缴费", "park": "车场信息", "parking": "停车记录", "device": "设备控制",
	"ticket": "月票/VIP", "blacklist": "黑名单", "redlist": "白名单", "visitor": "访客",
	"data": "数据分析", "coupon": "电子券", "evcharge": "电动车充电",
}

// reserved long flag names that would clash with persistent/local flags.
var reserved = map[string]bool{
	"body": true, "profile": true, "env": true, "output": true,
	"sign": true, "yes": true, "dry-run": true, "verbose": true, "help": true,
	"read-only": true, "all-pages": true, "out": true,
}

var scalarTypes = map[string]bool{
	"string": true, "integer": true, "int": true, "long": true,
	"decimal": true, "double": true, "float": true, "number": true,
	"boolean": true, "bool": true, "bigdecimal": true,
}

// flagValue scans os.Args for --name <value> or --name=value and returns the
// value, or "" if not found. It does not consume from os.Args.
func flagValue(name string) string {
	for i, a := range os.Args {
		if a == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
		if strings.HasPrefix(a, name+"=") {
			return a[len(name)+1:]
		}
	}
	return ""
}

func main() {
	catalogPath := "catalog/catalog.json"
	outDir := "cmd/gen"
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "--") {
		catalogPath = os.Args[1]
	}
	if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "--") {
		outDir = os.Args[2]
	}

	raw, err := os.ReadFile(catalogPath)
	must(err)
	var cat Catalog
	must(json.Unmarshal(raw, &cat))

	for i := range cat.Interfaces {
		if fi, ok := forceInclude[cat.Interfaces[i].Cmd]; ok {
			cat.Interfaces[i].Included = true
			if fi.Domain != "" {
				cat.Interfaces[i].Domain = fi.Domain
			}
			if fi.Explain != "" {
				cat.Interfaces[i].Explain = fi.Explain
			}
		}
	}

	// --refs <dir>: render reference drafts for all included callable cmds,
	// then return without touching cmd/gen (normal codegen path unchanged).
	if refsDir := flagValue("--refs"); refsDir != "" {
		writeReferenceDrafts(refsDir, cat)
		return
	}

	byDomain := map[string][]Iface{}
	for _, it := range cat.Interfaces {
		if !it.Included || it.Direction != "callable" {
			continue
		}
		d := it.Domain
		if d == "" {
			d = "misc"
		}
		byDomain[d] = append(byDomain[d], it)
	}

	domains := make([]string, 0, len(byDomain))
	for d := range byDomain {
		domains = append(domains, d)
	}
	sort.Strings(domains)

	must(os.MkdirAll(outDir, 0o755))
	total := 0
	for _, d := range domains {
		list := byDomain[d]
		sort.Slice(list, func(i, j int) bool { return list[i].Cmd < list[j].Cmd })
		src := genDomainFile(d, list)
		writeGo(filepath.Join(outDir, d+".go"), src)
		total += len(list)
	}
	writeGo(filepath.Join(outDir, "registry.go"), genRegistry(domains))

	fmt.Printf("generated %d commands across %d domains -> %s\n", total, len(domains), outDir)
	for _, d := range domains {
		fmt.Printf("  %-12s %d\n", d, len(byDomain[d]))
	}
}

func genRegistry(domains []string) string {
	var b strings.Builder
	b.WriteString("// Code generated by internal/gen; DO NOT EDIT.\n")
	b.WriteString("package gen\n\n")
	b.WriteString("import (\n\t\"github.com/spf13/cobra\"\n\n\t\"github.com/xiaowen-0725/openydt-cli/internal/cmdutil\"\n)\n\n")
	b.WriteString("// Commands returns all generated domain command groups.\n")
	b.WriteString("func Commands(f *cmdutil.Factory) []*cobra.Command {\n\treturn []*cobra.Command{\n")
	for _, d := range domains {
		fmt.Fprintf(&b, "\t\tnew%sCmd(f),\n", title(d))
	}
	b.WriteString("\t}\n}\n")
	return b.String()
}

func genDomainFile(domain string, list []Iface) string {
	var b strings.Builder
	b.WriteString("// Code generated by internal/gen; DO NOT EDIT.\n")
	b.WriteString("package gen\n\n")
	b.WriteString("import (\n\t\"github.com/spf13/cobra\"\n\n\t\"github.com/xiaowen-0725/openydt-cli/internal/cmdutil\"\n)\n\n")

	short := domainShort[domain]
	if short == "" {
		short = domain
	}
	fmt.Fprintf(&b, "func new%sCmd(f *cmdutil.Factory) *cobra.Command {\n", title(domain))
	fmt.Fprintf(&b, "\tc := &cobra.Command{Use: %s, Short: %s}\n", strconv.Quote(domain), strconv.Quote(short))
	b.WriteString("\tc.AddCommand(\n")
	for _, it := range list {
		fmt.Fprintf(&b, "\t\t%s(f),\n", fnName(domain, it.Cmd))
	}
	b.WriteString("\t)\n\treturn c\n}\n\n")

	for _, it := range list {
		genCmdFunc(&b, domain, it)
	}
	return b.String()
}

// subAliases keeps the full kebab cmd and the raw business cmd as aliases so
// nothing that referenced the old names breaks, deduped against the visible use.
func subAliases(use, cmd string) []string {
	out := []string{}
	add := func(s string) {
		if s == "" || s == use {
			return
		}
		for _, e := range out {
			if e == s {
				return
			}
		}
		out = append(out, s)
	}
	add(strutil.Kebab(cmd))
	add(cmd)
	quoted := make([]string, len(out))
	for i, s := range out {
		quoted[i] = strconv.Quote(s)
	}
	return quoted
}

func genCmdFunc(b *strings.Builder, domain string, it Iface) {
	defs, flags := scalarFlags(it.Params)

	fmt.Fprintf(b, "func %s(f *cmdutil.Factory) *cobra.Command {\n", fnName(domain, it.Cmd))
	b.WriteString("\tvar body, bodyFile string\n")
	b.WriteString("\tfields := map[string]*string{}\n")
	b.WriteString("\tc := &cobra.Command{\n")
	use := strutil.SubCmd(domain, it.Cmd)
	fmt.Fprintf(b, "\t\tUse:     %s,\n", strconv.Quote(use))
	fmt.Fprintf(b, "\t\tAliases: []string{%s},\n", strings.Join(subAliases(use, it.Cmd), ", "))
	fmt.Fprintf(b, "\t\tShort:   %s,\n", strconv.Quote(shortText(it)))
	fmt.Fprintf(b, "\t\tLong:    %s,\n", strconv.Quote(longText(it, defs)))
	b.WriteString("\t\tArgs:    cobra.NoArgs,\n")
	b.WriteString("\t\tRunE: func(cc *cobra.Command, _ []string) error {\n")
	if it.ReadWrite == "write" {
		fmt.Fprintf(b, "\t\t\tif err := f.ConfirmWrite(%s); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n", strconv.Quote(it.Cmd))
	}
	b.WriteString("\t\t\tbase, err := cmdutil.ResolveBody(body, bodyFile)\n")
	b.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
	b.WriteString("\t\t\tb, err := cmdutil.BuildBody([]cmdutil.ParamDef{\n")
	for _, d := range defs {
		fmt.Fprintf(b, "\t\t\t\t{Name: %s, Flag: %s, Type: %s, Required: %v},\n",
			strconv.Quote(d.Name), strconv.Quote(d.Flag), strconv.Quote(d.Type), d.Required)
	}
	b.WriteString("\t\t\t}, cc, fields, base)\n")
	b.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
	fmt.Fprintf(b, "\t\t\treturn f.RunCall(%s, b)\n", strconv.Quote(it.Cmd))
	b.WriteString("\t\t},\n\t}\n")
	b.WriteString("\tc.Flags().StringVar(&body, \"body\", \"\", \"完整请求体 JSON(字段 flag 会合并覆盖)\")\n")
	b.WriteString("\tc.Flags().StringVar(&bodyFile, \"body-file\", \"\", \"从文件读取请求体 JSON(- 表示 stdin;与 --body 互斥)\")\n")
	for _, fl := range flags {
		req := ""
		if fl.Required {
			req = " 必填"
		}
		help := fl.Type + req + ": " + oneLine(fl.Desc)
		if vals := strutil.ParseEnum(fl.Desc); len(vals) > 0 {
			help = strutil.Clip(help, 110) + "  [可选: " + strings.Join(vals, " ") + "]"
		} else {
			help = strutil.Clip(help, 180)
		}
		fmt.Fprintf(b, "\tc.Flags().StringVar(cmdutil.SP(fields, %s), %s, \"\", %s)\n",
			strconv.Quote(fl.Name), strconv.Quote(fl.Flag), strconv.Quote(help))
	}
	b.WriteString("\treturn c\n}\n\n")
}

type flagDef struct {
	Name, Flag, Type, Desc string
	Required               bool
}

// scalarFlags returns ParamDefs and flag specs for top-level scalar params,
// deduped and avoiding reserved flag names.
func scalarFlags(params []Param) (defs []flagDef, flags []flagDef) {
	seen := map[string]bool{}
	for _, p := range params {
		if p.Group != "" {
			continue // nested object/array -> only via --body
		}
		if !scalarTypes[strings.ToLower(strings.TrimSpace(p.Type))] {
			continue
		}
		fl := strutil.Kebab(p.Name)
		if fl == "" || reserved[fl] || seen[fl] || seen[p.Name] {
			continue
		}
		seen[fl] = true
		seen[p.Name] = true
		d := flagDef{Name: p.Name, Flag: fl, Type: p.Type, Desc: p.Desc, Required: p.Required}
		defs = append(defs, d)
		flags = append(flags, d)
	}
	return defs, flags
}

func shortText(it Iface) string {
	s := strings.TrimSpace(it.Explain)
	s = strings.TrimPrefix(s, "第三方接入系统请求智慧停车开放平台")
	s = strings.TrimPrefix(s, "第三方接入系统请求一点停开放平台")
	if s == "" {
		s = it.Cmd
	}
	return strutil.Clip(oneLine(s), 60)
}

func longText(it Iface, defs []flagDef) string {
	var b strings.Builder
	b.WriteString(oneLine(it.Explain))
	b.WriteString("\n\ncmd: " + it.Cmd)
	if it.FitSystem != "" {
		b.WriteString("  | 适用: " + oneLine(it.FitSystem))
	}
	b.WriteString("  | " + it.ReadWrite)
	if it.ReadWrite == "write" {
		b.WriteString(" (需 --yes)")
	}
	if h := deriveHints(it); h.ReadOnly || h.Destructive || h.Idempotent {
		var tags []string
		if h.ReadOnly {
			tags = append(tags, "read-only")
		}
		if h.Destructive {
			tags = append(tags, "destructive")
		}
		if h.Idempotent {
			if h.IdempotencyKey != "" {
				tags = append(tags, "idempotent(key="+h.IdempotencyKey+")")
			} else {
				tags = append(tags, "idempotent")
			}
		}
		b.WriteString("  | 注解: " + strings.Join(tags, ", "))
	}
	if len(it.Params) > 0 {
		b.WriteString("\n\n参数:")
		for _, p := range it.Params {
			req := "可选"
			if p.Required {
				req = "必填"
			}
			tag := ""
			if p.Group != "" {
				tag = " [" + p.Group + " 子字段,仅 --body]"
			}
			fmt.Fprintf(&b, "\n  %-22s %-9s %s %s%s", p.Name, p.Type, req, strutil.Clip(oneLine(p.Desc), 80), tag)
		}
	}
	if strings.TrimSpace(it.SampleBody) != "" && it.SampleBody != "{}" {
		b.WriteString("\n\n示例 body:\n  " + it.SampleBody)
	}
	return b.String()
}

// ---- reference draft rendering (--refs mode) ----

// writeReferenceDrafts writes one Markdown reference draft per included
// callable interface into <dir>/<domain>/<kebab-cmd>.md.
// It never writes into skills/; the caller controls dir.
func writeReferenceDrafts(dir string, cat Catalog) {
	count := 0
	for _, it := range cat.Interfaces {
		if !it.Included || it.Direction != "callable" {
			continue
		}
		domain := it.Domain
		if domain == "" {
			domain = "misc"
		}
		domainDir := filepath.Join(dir, domain)
		must(os.MkdirAll(domainDir, 0o755))

		kebab := strutil.Kebab(it.Cmd)
		destPath := filepath.Join(domainDir, kebab+".md")
		content := renderReferenceDraft(it)
		must(os.WriteFile(destPath, []byte(content), 0o644))
		count++
	}
	fmt.Printf("reference drafts: %d files -> %s\n", count, dir)
}

// renderReferenceDraft produces a seven-section Markdown draft for a single interface.
func renderReferenceDraft(it Iface) string {
	var b strings.Builder
	kebab := strutil.Kebab(it.Cmd)

	// ① Title + explain
	fmt.Fprintf(&b, "# %s\n\n", it.Cmd)
	if explain := oneLine(it.Explain); explain != "" {
		fmt.Fprintf(&b, "%s\n\n", explain)
	}

	// ② Command
	b.WriteString("## 命令\n\n")
	if it.Included {
		fmt.Fprintf(&b, "```bash\nopenydt %s %s\n```\n\n", it.Domain, kebab)
	} else {
		fmt.Fprintf(&b, "```bash\nopenydt api %s\n```\n\n", it.Cmd)
	}

	// ③ 参数表
	b.WriteString("## 参数\n\n")
	if len(it.Params) > 0 {
		b.WriteString("| Name | Type | 必填 | Desc | Group |\n")
		b.WriteString("|------|------|------|------|-------|\n")
		for _, p := range it.Params {
			req := "否"
			if p.Required {
				req = "是"
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s | %s |\n",
				p.Name, p.Type, req, oneLine(p.Desc), p.Group)
		}
	} else {
		b.WriteString("_(无参数)_\n")
	}
	b.WriteString("\n")

	// ④ sampleBody
	if sb := strings.TrimSpace(it.SampleBody); sb != "" && sb != "{}" {
		b.WriteString("## 示例 body\n\n")
		b.WriteString("```json\n")
		b.WriteString(sb)
		b.WriteString("\n```\n\n")
	}

	// ⑤ API 路径
	b.WriteString("## API 路径\n\n")
	fmt.Fprintf(&b, "```\nPOST /openydt/api/v3/%s\n```\n\n", it.Cmd)

	// ⑥ 命令安全注解
	b.WriteString("## 安全注解\n\n")
	h := deriveHints(it)
	var tags []string
	if h.ReadOnly {
		tags = append(tags, "read-only")
	}
	if h.Destructive {
		tags = append(tags, "destructive")
	}
	if h.Idempotent {
		if h.IdempotencyKey != "" {
			tags = append(tags, "idempotent(key="+h.IdempotencyKey+")")
		} else {
			tags = append(tags, "idempotent")
		}
	}
	if len(tags) > 0 {
		fmt.Fprintf(&b, "- 注解: %s\n", strings.Join(tags, ", "))
	} else {
		b.WriteString("- 注解: (无)\n")
	}
	fmt.Fprintf(&b, "- readwrite: %s\n", it.ReadWrite)
	b.WriteString("\n")

	// ⑦ 坑点占位
	b.WriteString("## 踩坑 / 字段解读\n\n")
	b.WriteString("<!-- 人工补:踩坑/字段解读 -->\n")

	return b.String()
}

// ---- hints (delegates to catalog.DeriveHints; no local copy needed) ----

// deriveHints is a thin adapter that collects param names from a local Iface
// and calls catalog.DeriveHints, keeping the derivation logic in one place.
func deriveHints(it Iface) catalog.Hints {
	names := make([]string, 0, len(it.Params))
	for _, p := range it.Params {
		names = append(names, p.Name)
	}
	return catalog.DeriveHints(it.Cmd, it.ReadWrite, names)
}

// ---- helpers ----

func fnName(domain, cmd string) string { return "cmd" + title(domain) + "_" + cmd }

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}

func writeGo(path, src string) {
	formatted, err := format.Source([]byte(src))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: gofmt failed for %s: %v (writing unformatted)\n", path, err)
		formatted = []byte(src)
	}
	must(os.WriteFile(path, formatted, 0o644))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "gen error:", err)
		os.Exit(1)
	}
}
