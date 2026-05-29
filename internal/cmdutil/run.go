package cmdutil

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xiaowen-0725/openydt-cli/internal/output"
)

// ExitError carries a process exit code up to main. A nil Err means the result
// was already rendered and only the code should propagate (no extra message).
type ExitError struct {
	Code int
	Err  error
}

func (e ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

func (e ExitError) Unwrap() error { return e.Err }

// Exit returns a silent ExitError carrying only a code.
func Exit(code int) error { return ExitError{Code: code} }

// usageErr wraps a user/argument error as exit code 2.
func usageErr(err error) error { return ExitError{Code: output.ExitUsage, Err: err} }

// RunCall is the shared path for invoking a single platform command: it builds
// the client, validates the body, honors --dry-run, sends, renders, and maps the
// business status to an exit code. Generated domain commands and `api` both use it.
func (f *Factory) RunCall(cmd, body string) error {
	if body == "" {
		body = "{}"
	}
	if !json.Valid([]byte(body)) {
		return usageErr(fmt.Errorf("--body 不是合法 JSON: %s", body))
	}
	c, err := f.Client()
	if err != nil {
		return usageErr(err)
	}
	if f.DryRun {
		p, err := c.Prepare(cmd, body)
		if err != nil {
			return usageErr(err)
		}
		return output.PrintJSON(f.Out, p)
	}
	resp, err := c.Call(context.Background(), cmd, body)
	if err != nil {
		return ExitError{Code: output.ExitNetwork, Err: err}
	}
	if code := output.Render(f.Out, f.Format(), resp); code != output.ExitOK {
		return Exit(code)
	}
	return nil
}

// ConfirmWrite guards a write operation: it requires --yes (or --dry-run).
func (f *Factory) ConfirmWrite(cmd string) error {
	if f.Yes || f.DryRun {
		return nil
	}
	return usageErr(fmt.Errorf("%q 是写操作,需加 --yes 确认 (或 --dry-run 预览)", cmd))
}
