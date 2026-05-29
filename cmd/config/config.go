// Package config implements `openydt config` (profile management).
package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/config"
	"github.com/xiaowen-0725/openydt-cli/internal/output"
)

// New returns the `config` command group.
func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "管理授权商 profile 与凭据",
	}
	cmd.AddCommand(newSet(f), newList(f), newUse(f), newPath(f))
	return cmd
}

func newSet(f *cmdutil.Factory) *cobra.Command {
	var p config.Profile
	cmd := &cobra.Command{
		Use:   "set",
		Short: "新增或更新一个授权商 profile",
		Example: `  openydt config set --profile demo --key test --secret 123456 --env test --sign v2`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if p.Name == "" || p.Key == "" || p.Secret == "" {
				return cmdutil.ExitError{Code: output.ExitUsage, Err: fmt.Errorf("--profile/--key/--secret 必填")}
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.Upsert(p)
			if cfg.CurrentProfile == "" {
				cfg.CurrentProfile = p.Name
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(f.Out, "已保存 profile %q (当前: %s)\n", p.Name, cfg.CurrentProfile)
			return nil
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&p.Name, "profile", "", "profile 名(授权商标识)")
	fl.StringVar(&p.Key, "key", "", "开放平台分配的 key")
	fl.StringVar(&p.Secret, "secret", "", "开放平台分配的 secret")
	fl.StringVar(&p.Env, "env", config.DefaultEnv, "默认环境 test|dev|prod")
	fl.StringVar(&p.Sign, "sign", config.DefaultSign, "默认签名版本 v2|v3")
	return cmd
}

func newList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有 profile(secret 脱敏)",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if len(cfg.Profiles) == 0 {
				fmt.Fprintln(f.Out, "(无 profile,先运行 openydt config set ...)")
				return nil
			}
			for _, p := range cfg.Profiles {
				marker := "  "
				if p.Name == cfg.CurrentProfile {
					marker = "* "
				}
				fmt.Fprintf(f.Out, "%s%-16s key=%s secret=%s env=%s sign=%s\n",
					marker, p.Name, p.Key, mask(p.Secret), orDefault(p.Env, config.DefaultEnv), orDefault(p.Sign, config.DefaultSign))
			}
			return nil
		},
	}
}

func newUse(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "切换当前 profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if _, ok := cfg.Find(args[0]); !ok {
				return cmdutil.ExitError{Code: output.ExitUsage, Err: fmt.Errorf("profile %q 不存在", args[0])}
			}
			cfg.CurrentProfile = args[0]
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(f.Out, "当前 profile: %s\n", args[0])
			return nil
		},
	}
}

func newPath(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "打印配置文件路径",
		RunE: func(_ *cobra.Command, _ []string) error {
			p, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Fprintln(f.Out, p)
			return nil
		},
	}
}

func mask(s string) string {
	if len(s) <= 2 {
		return "**"
	}
	return s[:1] + "***" + s[len(s)-1:]
}

func orDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
