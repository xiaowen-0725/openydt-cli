package cmdutil

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// ParamDef describes one top-level request parameter for a generated command.
type ParamDef struct {
	Name     string // JSON key, e.g. "carCode"
	Flag     string // CLI flag, e.g. "car-code"
	Type     string // platform type: String/Integer/Long/Decimal/Boolean/...
	Required bool
}

// BuildBody assembles the request body for a generated command: it starts from
// an optional --body JSON base and overlays any param flags that were set,
// coercing each value to the declared type. Flag values win over the base.
func BuildBody(defs []ParamDef, cmd *cobra.Command, fields map[string]*string, base string) (string, error) {
	obj := map[string]any{}
	if strings.TrimSpace(base) != "" {
		if err := json.Unmarshal([]byte(base), &obj); err != nil {
			return "", fmt.Errorf("--body 不是合法 JSON: %w", err)
		}
	}
	for _, d := range defs {
		if cmd == nil || !cmd.Flags().Changed(d.Flag) {
			continue
		}
		holder, ok := fields[d.Name]
		if !ok || holder == nil {
			continue
		}
		v, err := coerce(d.Type, *holder)
		if err != nil {
			return "", fmt.Errorf("--%s: %w", d.Flag, err)
		}
		obj[d.Name] = v
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func coerce(typ, raw string) (any, error) {
	switch strings.ToLower(strings.TrimSpace(typ)) {
	case "integer", "int", "long":
		n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("需要整数,得到 %q", raw)
		}
		return n, nil
	case "decimal", "double", "float", "number", "bigdecimal":
		n, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			return nil, fmt.Errorf("需要数字,得到 %q", raw)
		}
		return n, nil
	case "boolean", "bool":
		b, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("需要 true/false,得到 %q", raw)
		}
		return b, nil
	default:
		return raw, nil
	}
}

// SP registers a string flag holder in fields and returns a pointer for StringVar.
func SP(fields map[string]*string, name string) *string {
	v := new(string)
	fields[name] = v
	return v
}
