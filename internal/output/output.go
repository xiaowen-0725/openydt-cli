// Package output renders results and maps outcomes to process exit codes.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/xiaowen-0725/openydt-cli/internal/client"
)

// Exit codes (aligned with the Feishu CLI convention).
const (
	ExitOK       = 0
	ExitAPIError = 1 // business failure / non-success status
	ExitUsage    = 2 // bad arguments
	ExitAuth     = 4 // signature / key / auth failure
	ExitNetwork  = 5 // transport failure
)

// Format selects how a value is rendered.
type Format string

const (
	JSON  Format = "json"
	Table Format = "table"
)

// PrintJSON writes v as indented JSON.
func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// PrintRaw writes pre-serialized JSON bytes, indented if possible.
func PrintRaw(w io.Writer, raw []byte) error {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		_, err = w.Write(raw)
		return err
	}
	return PrintJSON(w, v)
}

// Render prints an API response in the requested format and returns the exit
// code that the command should propagate.
func Render(w io.Writer, format Format, resp *client.Response) int {
	switch format {
	case Table:
		renderTable(w, resp)
	default:
		_ = PrintRaw(w, resp.Raw)
	}
	return ExitFor(resp)
}

// ExitFor maps a business response to an exit code.
func ExitFor(resp *client.Response) int {
	switch resp.Status {
	case client.StatusSuccess:
		return ExitOK
	case client.StatusSignError, client.StatusKeyError, client.StatusNoAuth:
		return ExitAuth
	default:
		return ExitAPIError
	}
}

// ErrorInfo is a structured, AI-agent-friendly error description for a failed call.
type ErrorInfo struct {
	Cmd           string   `json:"cmd,omitempty"`
	Status        int      `json:"status"`
	StatusText    string   `json:"statusText"`
	ResultCode    int      `json:"resultCode,omitempty"`
	ResultText    string   `json:"resultText,omitempty"`
	Message       string   `json:"message"`
	Hint          string   `json:"hint,omitempty"`
	Retriable     bool     `json:"retriable"`
	Field         string   `json:"field,omitempty"`
	FieldType     string   `json:"fieldType,omitempty"`
	FieldRequired bool     `json:"fieldRequired,omitempty"`
	FieldDesc     string   `json:"fieldDesc,omitempty"`
	AllowedValues []string `json:"allowedValues,omitempty"`
}

// RenderError prints a structured error and returns the process exit code.
// JSON mode emits the response envelope plus an `_error` object that an agent
// can parse to self-correct; table mode prints human-readable lines.
func RenderError(w io.Writer, format Format, e *ErrorInfo, resp *client.Response) int {
	if format == Table {
		fmt.Fprintf(w, "✗ status %d (%s)\n", e.Status, e.StatusText)
		if e.ResultCode != 0 {
			fmt.Fprintf(w, "  resultCode : %d (%s)\n", e.ResultCode, e.ResultText)
		}
		fmt.Fprintf(w, "  message    : %s\n", e.Message)
		if e.Field != "" {
			req := "选填"
			if e.FieldRequired {
				req = "必填"
			}
			fmt.Fprintf(w, "  字段       : %s (%s, %s) %s\n", e.Field, e.FieldType, req, e.FieldDesc)
			if len(e.AllowedValues) > 0 {
				fmt.Fprintf(w, "  可选值     : %s\n", strings.Join(e.AllowedValues, " | "))
			}
		}
		if e.Hint != "" {
			fmt.Fprintf(w, "  建议       : %s\n", e.Hint)
		}
		if e.Retriable {
			fmt.Fprintf(w, "  可重试     : 是\n")
		}
	} else {
		env := map[string]any{
			"status": resp.Status, "resultCode": resp.ResultCode,
			"message": resp.Message, "_error": e,
		}
		if len(resp.Data) > 0 {
			env["data"] = json.RawMessage(resp.Data)
		}
		_ = PrintJSON(w, env)
	}
	return ExitFor(resp)
}

func renderTable(w io.Writer, resp *client.Response) {
	fmt.Fprintf(w, "status     : %d (%s)\n", resp.Status, client.StatusText(resp.Status))
	if resp.Status == client.StatusBizFail {
		fmt.Fprintf(w, "resultCode : %d (%s)\n", resp.ResultCode, client.ResultText(resp.ResultCode))
	}
	fmt.Fprintf(w, "message    : %s\n", resp.Message)
	if len(resp.Data) > 0 && string(resp.Data) != "null" {
		fmt.Fprintln(w, "data       :")
		_ = PrintRaw(w, resp.Data)
	}
}
