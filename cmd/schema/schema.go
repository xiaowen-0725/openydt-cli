// Package schema implements `openydt schema [cmd]` — discover interface params,
// request enums, response enum supplements, and samples from the embedded catalog.
package schema

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xiaowen-0725/openydt-cli/internal/catalog"
	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/output"
	"github.com/xiaowen-0725/openydt-cli/internal/strutil"
)

// New returns the `schema` command.
func New(f *cmdutil.Factory) *cobra.Command {
	var domain string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "schema [cmd]",
		Short: "查看接口契约(参数/枚举/领域语义/示例)",
		Long: `从内置接口目录查询某个业务编码(cmd)的入参说明、响应枚举和领域语义,便于人和 AI Agent 自助发现契约。

  openydt schema getParkFee          # 查看某接口的参数表/枚举/示例
  openydt schema                     # 列出全部可调用接口(可加 --domain trade 过滤)
  openydt schema --domain coupon
  openydt schema getParkFee --json   # 机器可读 JSON 输出(含 hints 安全注解)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cat, err := catalog.Embedded()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				if asJSON {
					return showOneJSON(f, cat, args[0])
				}
				return showOne(f, cat, args[0])
			}
			if asJSON {
				return listJSON(f, cat, domain)
			}
			return list(f, cat, domain)
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "按业务域过滤(trade/park/parking/device/ticket/coupon/...)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "机器可读 JSON 输出(含命令安全注解 hints,不影响默认人读格式)")
	return cmd
}

func list(f *cmdutil.Factory, cat *catalog.Catalog, domain string) error {
	byDomain := map[string][]catalog.Iface{}
	for _, it := range cat.Included() {
		if domain != "" && it.Domain != domain {
			continue
		}
		byDomain[it.Domain] = append(byDomain[it.Domain], it)
	}
	domains := make([]string, 0, len(byDomain))
	for d := range byDomain {
		domains = append(domains, d)
	}
	sort.Strings(domains)
	for _, d := range domains {
		its := byDomain[d]
		sort.Slice(its, func(i, j int) bool { return its[i].Cmd < its[j].Cmd })
		fmt.Fprintf(f.Out, "\n# %s (%d)\n", d, len(its))
		for _, it := range its {
			rw := it.ReadWrite
			if rw == "write" {
				rw = "write*"
			}
			fmt.Fprintf(f.Out, "  %-36s [%s] %s\n", it.Cmd, rw, strutil.Clip(short(it.Explain), 40))
		}
	}
	fmt.Fprintf(f.Out, "\n(* = 写操作, 需 --yes;  openydt schema <cmd> 查参数)\n")
	return nil
}

func showOne(f *cmdutil.Factory, cat *catalog.Catalog, cmd string) error {
	var it catalog.Iface
	found := false
	for _, x := range cat.Interfaces {
		if strings.EqualFold(x.Cmd, cmd) {
			it, found = x, true
			break
		}
	}
	if !found {
		return cmdutil.ExitError{Code: output.ExitUsage, Err: fmt.Errorf("未知 cmd %q(openydt schema 列出全部)", cmd)}
	}
	w := f.Out
	fmt.Fprintf(w, "cmd:        %s\n", it.Cmd)
	fmt.Fprintf(w, "说明:       %s\n", short(it.Explain))
	fmt.Fprintf(w, "业务域:     %s\n", it.Domain)
	yes := ""
	if it.ReadWrite == "write" {
		yes = " (需 --yes)"
	}
	fmt.Fprintf(w, "读写:       %s%s\n", it.ReadWrite, yes)
	h := it.Hints()
	tags := []string{}
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
		fmt.Fprintf(w, "注解:       %s\n", strings.Join(tags, ", "))
	}
	if it.FitSystem != "" {
		fmt.Fprintf(w, "适用系统:   %s\n", short(it.FitSystem))
	}
	if it.Direction == "webhook" {
		fmt.Fprintf(w, "方向:       webhook(平台推送, CLI 不可主动调用)\n")
	}
	if !it.Included {
		fmt.Fprintf(w, "注:         未生成一等命令(excludeReason=%s), 可用 openydt api %s 调用\n", it.ExcludeReason, it.Cmd)
	}
	fmt.Fprintf(w, "\n参数:\n")
	for _, p := range it.Params {
		req := "选填"
		if p.Required {
			req = "必填"
		}
		grp := ""
		if p.Group != "" {
			grp = " [" + p.Group + " 子字段]"
		}
		fmt.Fprintf(w, "  %-24s %-10s %s  %s%s\n", p.Name, p.Type, req, strutil.Clip(short(p.Desc), 60), grp)
		if vals := p.EnumValues(); len(vals) > 0 {
			fmt.Fprintf(w, "      └ 可选值: %s\n", strings.Join(vals, " | "))
		}
	}
	responseEnums, err := catalog.ResponseEnumsFor(it.Cmd)
	if err != nil {
		return err
	}
	if len(responseEnums) > 0 {
		fmt.Fprintln(w, "\n响应枚举（补充）:")
		for _, enum := range responseEnums {
			fmt.Fprintf(w, "  %s  %s\n", strings.Join(enum.Fields, ", "), enum.Description)
			fmt.Fprintf(w, "      来源: %s\n", enum.Source)
			for _, value := range enum.Values {
				name := value.Name
				if len(value.LegacyAliases) > 0 {
					name += "（旧称: " + strings.Join(value.LegacyAliases, "/") + "）"
				}
				if value.Legacy {
					name += " [legacy]"
				}
				tags := []string{value.Key}
				fmt.Fprintf(w, "      %d %s [%s]\n", value.Code, name, strings.Join(tags, ", "))
			}
			for _, note := range enum.Notes {
				fmt.Fprintf(w, "      注: %s\n", note)
			}
		}
	}
	domainSemantics, err := catalog.DomainSemanticsFor(it.Cmd)
	if err != nil {
		return err
	}
	if domainSemantics != nil {
		printDomainSemantics(w, responseEnums, domainSemantics)
	}
	if strings.TrimSpace(it.SampleBody) != "" && it.SampleBody != "{}" {
		fmt.Fprintf(w, "\n示例 body:\n%s\n", it.SampleBody)
	}
	fmt.Fprintf(w, "\n调用: openydt %s %s --body '<json>'   或   openydt api %s --body '<json>'\n",
		domainOrApi(it), strutil.SubCmd(it.Domain, it.Cmd), it.Cmd)
	return nil
}

// showOneJSON outputs a single interface as machine-readable JSON including
// safety annotations (hints). Use --json flag to invoke.
func showOneJSON(f *cmdutil.Factory, cat *catalog.Catalog, cmd string) error {
	var it catalog.Iface
	found := false
	for _, x := range cat.Interfaces {
		if strings.EqualFold(x.Cmd, cmd) {
			it, found = x, true
			break
		}
	}
	if !found {
		return cmdutil.ExitError{Code: output.ExitUsage, Err: fmt.Errorf("未知 cmd %q(openydt schema 列出全部)", cmd)}
	}
	responseEnums, err := catalog.ResponseEnumsFor(it.Cmd)
	if err != nil {
		return err
	}
	domainSemantics, err := catalog.DomainSemanticsFor(it.Cmd)
	if err != nil {
		return err
	}
	return output.PrintJSON(f.Out, map[string]any{
		"cmd":             it.Cmd,
		"domain":          it.Domain,
		"readwrite":       it.ReadWrite,
		"direction":       it.Direction,
		"included":        it.Included,
		"hints":           it.Hints(),
		"params":          it.Params,
		"responseEnums":   responseEnums,
		"domainSemantics": domainSemantics,
		"sampleBody":      json.RawMessage(orEmptyObj(it.SampleBody)),
	})
}

func printDomainSemantics(w interface{ Write([]byte) (int, error) }, enums []catalog.ResponseEnum, semantics *catalog.DomainSemantics) {
	fmt.Fprintln(w, "\n业务语义:")
	fmt.Fprintf(w, "  记录集: %s\n", semantics.RecordSet)
	fmt.Fprintf(w, "  %s\n", semantics.Description)
	if semantics.Deduplication != nil {
		fmt.Fprintf(w, "  去重: 按 %s 分组，按 %s 保留 %s\n", semantics.Deduplication.Key, semantics.Deduplication.OrderBy, semantics.Deduplication.Keep)
	}
	names := responseEnumNames(enums)
	for _, field := range semantics.Fields {
		fmt.Fprintf(w, "  %s  %s\n", field.Field, field.Description)
		for _, value := range field.Values {
			name := names[value.Code]
			if name != "" {
				name += " "
			}
			tags := []string{eventNatureLabel(value.EventNature)}
			for _, tag := range value.BusinessTags {
				tags = append(tags, businessTagLabel(tag))
			}
			if value.TrafficTreatment == "exclude" {
				tags = append(tags, "经营统计排除")
			}
			fmt.Fprintf(w, "      %d %s[%s]\n", value.Code, name, strings.Join(tags, ", "))
		}
	}
}

func responseEnumNames(enums []catalog.ResponseEnum) map[int]string {
	result := map[int]string{}
	for _, enum := range enums {
		for _, value := range enum.Values {
			result[value.Code] = value.Name
		}
	}
	return result
}

func eventNatureLabel(value string) string {
	label := map[string]string{
		"physical_departure":    "物理离场",
		"logical_departure":     "逻辑离场",
		"logical_closure":       "逻辑闭环",
		"manual_reconciliation": "人工盘点闭环",
		"gate_operation":        "开闸操作",
		"legacy":                "遗留类型",
	}[value]
	if label == "" {
		return value
	}
	return label
}

func businessTagLabel(value string) string {
	label := map[string]string{
		"escape":                      "逃费",
		"abnormal_release":            "异常放行",
		"arrears_release":             "欠费放行",
		"physical_remote_controller":  "实体遥控器",
		"repeat_departure":            "重复离场保留最后一次",
		"suspicious_follow_operation": "可疑跟车操作",
	}[value]
	if label == "" {
		return value
	}
	return label
}

// listJSON outputs all matching interfaces as machine-readable JSON including hints.
func listJSON(f *cmdutil.Factory, cat *catalog.Catalog, domain string) error {
	type entry struct {
		Cmd       string        `json:"cmd"`
		Domain    string        `json:"domain"`
		ReadWrite string        `json:"readwrite"`
		Direction string        `json:"direction"`
		Included  bool          `json:"included"`
		Hints     catalog.Hints `json:"hints"`
		Explain   string        `json:"explain"`
	}
	var items []entry
	for _, it := range cat.Included() {
		if domain != "" && it.Domain != domain {
			continue
		}
		items = append(items, entry{
			Cmd:       it.Cmd,
			Domain:    it.Domain,
			ReadWrite: it.ReadWrite,
			Direction: it.Direction,
			Included:  it.Included,
			Hints:     it.Hints(),
			Explain:   short(it.Explain),
		})
	}
	return output.PrintJSON(f.Out, items)
}

// orEmptyObj returns s if non-empty, otherwise "{}".
func orEmptyObj(s string) string {
	if strings.TrimSpace(s) == "" {
		return "{}"
	}
	return s
}

func domainOrApi(it catalog.Iface) string {
	if it.Included {
		return it.Domain
	}
	return "api"
}

func short(s string) string {
	s = strings.TrimPrefix(strings.TrimSpace(s), "第三方接入系统请求智慧停车开放平台")
	s = strings.TrimPrefix(s, "第三方接入系统请求一点停开放平台")
	return strings.ReplaceAll(s, "\n", " ")
}
