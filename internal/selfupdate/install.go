package selfupdate

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// Install upgrades the globally installed npm package to an exact release.
func Install(ctx context.Context, version string, stdout, stderr io.Writer) error {
	version = normalizeVersion(version)
	if !isReleaseVersion(version) {
		return fmt.Errorf("拒绝安装无效版本 %q", version)
	}
	npm, err := exec.LookPath("npm")
	if err != nil {
		return fmt.Errorf("未找到 npm,请先安装 Node.js: %w", err)
	}
	cmd := exec.CommandContext(ctx, npm, "install", "-g", PackageName+"@"+version)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm 全局更新失败: %w", err)
	}
	return nil
}
