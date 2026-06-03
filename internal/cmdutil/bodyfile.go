package cmdutil

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// ResolveBody resolves the request-body base from the mutually-exclusive
// --body / --body-file flags. bodyFile == "-" reads stdin; any other value is
// a file path. The result is the JSON base that BuildBody overlays param flags
// onto (an empty string means "no base"). Shared by the generic `api` command
// and every generated domain command so body sourcing is uniform.
func ResolveBody(body, bodyFile string) (string, error) {
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
