package cmdutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/xiaowen-0725/openydt-cli/internal/catalog"
	"github.com/xiaowen-0725/openydt-cli/internal/client"
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
	body = f.applySchemaDefaults(cmd, body) // 缺参 → 补平台文档声明的「默认N」(api 与生成命令同享)
	body = f.applyDefaults(cmd, body)       // 缺参 → 补 profile 默认值(dry-run 预览也反映)
	if err := f.guardWrite(cmd); err != nil {
		return err
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
	if !resp.OK() {
		ei := buildErrorInfo(cmd, body, resp)
		return Exit(output.RenderError(f.Out, f.Format(), ei, resp))
	}
	output.Render(f.Out, f.Format(), resp)
	return nil
}

var fieldRe = regexp.MustCompile(`([A-Za-z][A-Za-z0-9_]+)`)

// buildErrorInfo turns a failed business response into a structured, agent-friendly
// error: status/resultCode hints + nextCommands + retryClass + parameter location.
func buildErrorInfo(cmd, body string, resp *client.Response) *output.ErrorInfo {
	ei := &output.ErrorInfo{
		Cmd: cmd, Status: resp.Status, StatusText: client.StatusText(resp.Status),
		Message: resp.Message, Retriable: client.Retriable(resp.Status),
		RetryClass: client.RetryClass(resp.Status),
	}
	if resp.Status == client.StatusBizFail {
		ei.ResultCode = resp.ResultCode
		ei.ResultText = client.ResultText(resp.ResultCode)
		ei.Hint = client.ResultHint(resp.ResultCode)
		ei.NextCommands = client.ResultNextCommands(resp.ResultCode)
	} else {
		ei.NextCommands = client.StatusNextCommands(resp.Status)
	}
	if ei.Hint == "" {
		ei.Hint = client.StatusHint(resp.Status)
	}
	// nextCommands 里的 "schema <cmd>" 占位换成真实 cmd
	for i, nc := range ei.NextCommands {
		if nc == "schema <cmd>" {
			ei.NextCommands[i] = "schema " + cmd
		}
	}
	locateParam(ei, cmd, body, resp)
	return ei
}

// setParam fills the field/type/required/desc/enum fields of an ErrorInfo from a catalog Param.
func setParam(ei *output.ErrorInfo, p catalog.Param) {
	ei.Field = p.Name
	ei.FieldType = p.Type
	ei.FieldRequired = p.Required
	ei.FieldDesc = p.Desc
	ei.AllowedValues = p.EnumValues()
}

// locateParam attempts to locate the specific missing/wrong parameter that caused
// the failure. Primary strategy: for status=7 or resultCode=909, compare the set
// of required non-group scalar params against the keys present in body and return
// the first missing required param. Fallback: extract param name from Chinese error
// message patterns like "参数错误:carCode不能为空".
func locateParam(ei *output.ErrorInfo, cmd, body string, resp *client.Response) {
	cat, err := catalog.Embedded()
	if err != nil {
		return
	}
	it, ok := cat.Find(cmd)
	if !ok {
		return
	}
	// 1) 参数不完整(status=7)或 909:按必填集对比 body 实际键,定位首个缺失必填标量
	if resp.Status == client.StatusBadParams || resp.ResultCode == 909 {
		present := map[string]bool{}
		var m map[string]any
		if json.Unmarshal([]byte(body), &m) == nil {
			for k := range m {
				present[k] = true
			}
		}
		for _, p := range it.Params {
			if p.Required && p.Group == "" && !present[p.Name] {
				setParam(ei, p)
				return
			}
		}
	}
	// 2) 回退:中文文案("…不能为空/参数错误")提到的参数名
	if strings.Contains(resp.Message, "不能为空") || strings.Contains(resp.Message, "参数错误") {
		for _, tok := range fieldRe.FindAllString(resp.Message, -1) {
			if p, ok := it.FindParam(tok); ok {
				setParam(ei, p)
				return
			}
		}
	}
}

// ConfirmWrite guards a write operation: it requires --yes (or --dry-run).
func (f *Factory) ConfirmWrite(cmd string) error {
	if f.Yes || f.DryRun {
		return nil
	}
	return usageErr(fmt.Errorf("%q 是写操作,需加 --yes 确认 (或 --dry-run 预览)", cmd))
}

type readOnlyError struct{ cmd string }

func (e readOnlyError) Error() string {
	return fmt.Sprintf("只读模式(--read-only / OPENYDT_READ_ONLY)下拒绝写操作 %q;去掉只读开关再执行", e.cmd)
}

func isReadOnlyErr(err error) bool {
	var e readOnlyError
	return errors.As(err, &e)
}

// writeGuard decides whether a call may proceed. Pure (unit-testable).
func writeGuard(cmd string, isWrite, yes, dryRun, readOnly bool) error {
	if !isWrite {
		return nil
	}
	if readOnly {
		return ExitError{Code: output.ExitUsage, Err: readOnlyError{cmd}}
	}
	if yes || dryRun {
		return nil
	}
	return usageErr(fmt.Errorf("%q 是写操作,需加 --yes 确认 (或 --dry-run 预览)", cmd))
}

// guardWrite resolves write-ness from the embedded catalog and applies writeGuard.
func (f *Factory) guardWrite(cmd string) error {
	isWrite := false
	if cat, err := catalog.Embedded(); err == nil {
		isWrite = cat.IsWrite(cmd)
	}
	readOnly := f.ReadOnly
	if v := os.Getenv("OPENYDT_READ_ONLY"); v == "1" || v == "true" {
		readOnly = true
	}
	return writeGuard(cmd, isWrite, f.Yes, f.DryRun, readOnly)
}
