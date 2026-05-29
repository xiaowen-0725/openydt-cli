// Package api implements the generic passthrough command:
// `openydt api <cmd> --body '{...}'` signs and sends any business cmd.
package api

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
)

// New returns the `api` command.
func New(f *cmdutil.Factory) *cobra.Command {
	var body, bodyFile string
	cmd := &cobra.Command{
		Use:   "api <cmd>",
		Short: "通用调用:对任意业务编码 cmd 自动签名并发送 POST",
		Long: `对任意业务编码(cmd)自动加签名并发送 POST 请求。

示例:
  openydt api getParkFee --body '{"carCode":"粤EJW962"}'
  openydt api getAuthParkCodes
  echo '{"parkCode":"PTD2YBBZ"}' | openydt api getParkOnSiteCar --body-file -`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			b, err := resolveBody(body, bodyFile)
			if err != nil {
				return cmdutil.ExitError{Code: 2, Err: err}
			}
			return f.RunCall(args[0], b)
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "请求体 JSON 字符串")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "从文件读取请求体 JSON(- 表示 stdin)")
	return cmd
}

func resolveBody(body, bodyFile string) (string, error) {
	if body != "" && bodyFile != "" {
		return "", fmt.Errorf("--body 与 --body-file 不能同时使用")
	}
	if bodyFile == "" {
		return body, nil
	}
	if bodyFile == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	data, err := os.ReadFile(bodyFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
