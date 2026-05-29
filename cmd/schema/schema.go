// Package schema implements `openydt schema [cmd]` — discover interface params,
// required/optional, types, enum values, and a sample body from the embedded catalog.
package schema

import (
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
	cmd := &cobra.Command{
		Use:   "schema [cmd]",
		Short: "查看接口参数说明(必填/选填/类型/可选值/示例)",
		Long: `从内置接口目录查询某个业务编码(cmd)的入参说明,便于人和 AI Agent 自助发现参数。

  openydt schema getParkFee          # 查看某接口的参数表/枚举/示例
  openydt schema                     # 列出全部可调用接口(可加 --domain trade 过滤)
  openydt schema --domain coupon`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cat, err := catalog.Embedded()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				return showOne(f, cat, args[0])
			}
			return list(f, cat, domain)
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "按业务域过滤(trade/park/parking/device/ticket/coupon/...)")
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
	if strings.TrimSpace(it.SampleBody) != "" && it.SampleBody != "{}" {
		fmt.Fprintf(w, "\n示例 body:\n%s\n", it.SampleBody)
	}
	fmt.Fprintf(w, "\n调用: openydt %s %s --body '<json>'   或   openydt api %s --body '<json>'\n",
		domainOrApi(it), strutil.Kebab(it.Cmd), it.Cmd)
	return nil
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
