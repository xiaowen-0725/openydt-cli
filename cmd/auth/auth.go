// Package auth implements `openydt auth` (credential verification).
package auth

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xiaowen-0725/openydt-cli/internal/client"
	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/output"
)

// New returns the `auth` command group.
func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "验证凭据与签名链路",
	}
	cmd.AddCommand(newTest(f))
	return cmd
}

// newTest does a lightweight read (getAuthParkCodes) to confirm the active
// profile's key/secret and signature are accepted by the platform.
func newTest(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "冒烟验证:调用 getAuthParkCodes 确认凭据/签名可用",
		RunE: func(_ *cobra.Command, _ []string) error {
			c, err := f.Client()
			if err != nil {
				return cmdutil.ExitError{Code: output.ExitUsage, Err: err}
			}
			resp, err := c.Call(context.Background(), "getAuthParkCodes", "{}")
			if err != nil {
				return cmdutil.ExitError{Code: output.ExitNetwork, Err: err}
			}
			if resp.OK() {
				fmt.Fprintf(f.Out, "✓ 认证通过 (status=1)。authorized parks: %s\n", string(resp.Data))
				return nil
			}
			fmt.Fprintf(f.Err, "✗ 认证未通过: status=%d (%s) message=%q resultCode=%d\n",
				resp.Status, client.StatusText(resp.Status), resp.Message, resp.ResultCode)
			return cmdutil.Exit(output.ExitFor(resp))
		},
	}
}
