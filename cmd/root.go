// Package cmd assembles the openydt root command and its subcommands.
package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	apicmd "github.com/xiaowen-0725/openydt-cli/cmd/api"
	authcmd "github.com/xiaowen-0725/openydt-cli/cmd/auth"
	configcmd "github.com/xiaowen-0725/openydt-cli/cmd/config"
	gencmd "github.com/xiaowen-0725/openydt-cli/cmd/gen"
	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/output"
)

// NewRootCmd builds the root command and binds global flags onto f.
func NewRootCmd(f *cmdutil.Factory) *cobra.Command {
	root := &cobra.Command{
		Use:           "openydt",
		Short:         "艾科智泊停车开放平台 CLI —— 为人和 AI Agent 而生",
		Version:       cmdutil.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `openydt 把艾科智泊智慧停车开放平台的接口封装成命令行工具。
自动处理签名鉴权(v2/v3)、多授权商 profile、多环境(test/dev/prod)。

常用:
  openydt config set --profile demo --key test --secret 123456   # 配置授权商
  openydt auth test                                              # 验证凭据
  openydt api getParkFee --body '{"carCode":"粤EJW962"}'         # 通用调用任意接口`,
	}

	pf := root.PersistentFlags()
	pf.StringVar(&f.Profile, "profile", "", "授权商 profile 名(默认当前 profile)")
	pf.StringVar(&f.Env, "env", "", "环境 test|dev|prod (默认 test)")
	pf.StringVarP(&f.Output, "output", "o", "json", "输出格式 json|table")
	pf.StringVar(&f.Sign, "sign", "", "签名版本 v2|v3 (默认按 profile,否则 v2)")
	pf.BoolVarP(&f.Yes, "yes", "y", false, "确认执行写操作")
	pf.BoolVar(&f.DryRun, "dry-run", false, "只打印将发送的签名请求,不实际发送")
	pf.BoolVarP(&f.Verbose, "verbose", "v", false, "输出调试信息到 stderr")

	root.AddCommand(
		configcmd.New(f),
		authcmd.New(f),
		apicmd.New(f),
	)
	// Generated per-domain catalog commands.
	root.AddCommand(gencmd.Commands(f)...)
	return root
}

// Execute runs the CLI and returns a process exit code.
func Execute() int {
	f := cmdutil.NewFactory()
	root := NewRootCmd(f)
	err := root.Execute()
	if err == nil {
		return output.ExitOK
	}
	var ee cmdutil.ExitError
	if errors.As(err, &ee) {
		if ee.Err != nil {
			fmt.Fprintln(f.Err, "Error:", ee.Err)
		}
		return ee.Code
	}
	fmt.Fprintln(f.Err, "Error:", err)
	return output.ExitUsage
}
