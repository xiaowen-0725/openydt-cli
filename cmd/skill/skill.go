// Package skill provides `openydt skill` — manage AI-agent skills.
package skill

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/output"
	"github.com/xiaowen-0725/openydt-cli/internal/skillsync"
)

// New builds the `openydt skill` command tree.
func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "管理 AI Agent 技能(同步到本机已装的各 agent)",
		Long: `把 openydt 的技能同步到本机所有已装 AI agent(Claude Code / Codex / Cursor / ... )。
底层调用 npx skills(需 Node.js)。日常 npm 安装/更新会自动同步,本命令用于手动或强制重装。`,
		// 空 PersistentPreRun:让 skill 子命令跳过根命令的自动同步触发,
		// 避免手动 sync 时再 fork 一个后台 sync。
		PersistentPreRun: func(*cobra.Command, []string) {},
	}
	cmd.AddCommand(newSyncCmd(f))
	return cmd
}

func newSyncCmd(f *cmdutil.Factory) *cobra.Command {
	var quiet, force bool
	c := &cobra.Command{
		Use:   "sync",
		Short: "把 openydt 技能同步到本机所有已装 agent(npx skills add)",
		RunE: func(_ *cobra.Command, _ []string) error {
			res := skillsync.RunSync(force)
			if res.Err != nil {
				if !quiet {
					fmt.Fprintf(f.Err, "✗ skills 同步失败: %v\n", res.Err)
				}
				return cmdutil.ExitError{Code: output.ExitAPIError, Err: res.Err}
			}
			if err := skillsync.RecordSuccess(cmdutil.Version); err != nil && !quiet {
				fmt.Fprintf(f.Err, "warning: 同步成功但 state 未写入: %v\n", err)
			}
			if !quiet {
				fmt.Fprintf(f.Err, "✓ skills 已同步(source: %s)\n", skillsync.Source())
			}
			return nil
		},
	}
	c.Flags().BoolVar(&quiet, "quiet", false, "静默(供后台自动同步子进程使用)")
	c.Flags().BoolVar(&force, "force", false, "全量重装所有技能到所有 agent")
	return c
}
