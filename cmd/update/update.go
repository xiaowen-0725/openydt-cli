// Package update provides openydt self-update commands.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/selfupdate"
	"github.com/xiaowen-0725/openydt-cli/internal/skillsync"
)

type checkFunc func(context.Context, string) (selfupdate.Result, error)
type installFunc func(context.Context, string, io.Writer, io.Writer) error

// New builds the update command tree.
func New(f *cmdutil.Factory) *cobra.Command {
	checker := selfupdate.DefaultChecker()
	return newCommand(f, checker.Check, selfupdate.Install)
}

func newCommand(f *cmdutil.Factory, check checkFunc, install installFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "检查并更新 openydt CLI 与 Skills",
		Long:  "从 npm Registry 检查最新正式版本；更新时安装精确版本，postinstall 会同步同版本 Git tag 的 Skills。",
		// 跳过根命令的后台更新检查，避免 update check 递归触发自身。
		PersistentPreRun: func(*cobra.Command, []string) {},
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := check(cmd.Context(), cmdutil.Version)
			if err != nil {
				return err
			}
			_ = selfupdate.RecordResult(result)
			if !result.UpdateAvailable {
				fmt.Fprintf(f.Out, "✓ 已是最新版本 v%s\n", result.CurrentVersion)
				return nil
			}
			fmt.Fprintf(f.Out, "正在更新 v%s → v%s ...\n", result.CurrentVersion, result.LatestVersion)
			if err := install(cmd.Context(), result.LatestVersion, f.Out, f.Err); err != nil {
				return err
			}
			fmt.Fprintf(f.Out, "✓ openydt CLI 已更新到 v%s\n", result.LatestVersion)
			if skillsync.IsSynced(result.LatestVersion) {
				fmt.Fprintln(f.Out, "✓ 同版本 Skills 已同步")
			} else {
				fmt.Fprintln(f.Err, "[openydt] Skills 同步未确认；可运行 openydt skill sync 重试")
			}
			return nil
		},
	}
	cmd.AddCommand(newCheckCommand(f, check))
	return cmd
}

func newCheckCommand(f *cmdutil.Factory, check checkFunc) *cobra.Command {
	var asJSON, quiet bool
	cmd := &cobra.Command{
		Use:   "check",
		Short: "检查 npm 上是否有新版本",
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := check(cmd.Context(), cmdutil.Version)
			if err != nil {
				return err
			}
			_ = selfupdate.RecordResult(result)
			if quiet {
				return nil
			}
			if asJSON {
				return json.NewEncoder(f.Out).Encode(result)
			}
			fmt.Fprintf(f.Out, "当前版本: v%s\n最新版本: v%s\n", result.CurrentVersion, result.LatestVersion)
			if result.UpdateAvailable {
				fmt.Fprintln(f.Out, "有新版本可用；运行 openydt update 更新 CLI 与 Skills")
			} else {
				fmt.Fprintln(f.Out, "✓ 已是最新版本")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "输出机器可读 JSON")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "静默检查并只更新本地缓存(供后台任务使用)")
	return cmd
}
