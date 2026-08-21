package catalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

// ResponseEnumValue is one documented response code.
type ResponseEnumValue struct {
	Code          int      `json:"code"`
	Key           string   `json:"key"`
	Name          string   `json:"name"`
	Legacy        bool     `json:"legacy,omitempty"`
	LegacyAliases []string `json:"legacyAliases,omitempty"`
}

// ResponseEnum supplements response fields omitted or truncated by the
// upstream platform documentation.
type ResponseEnum struct {
	Name        string              `json:"name"`
	Fields      []string            `json:"fields"`
	Description string              `json:"description"`
	Source      string              `json:"source"`
	Notes       []string            `json:"notes,omitempty"`
	Values      []ResponseEnumValue `json:"values"`
}

type responseEnumDefinition struct {
	Description string              `json:"description"`
	Source      string              `json:"source"`
	Notes       []string            `json:"notes"`
	Values      []ResponseEnumValue `json:"values"`
}

type responseEnumBinding struct {
	Cmd        string   `json:"cmd"`
	Fields     []string `json:"fields"`
	Definition string   `json:"definition"`
}

type responseEnumCatalog struct {
	Definitions map[string]responseEnumDefinition `json:"definitions"`
	Bindings    []responseEnumBinding             `json:"bindings"`
}

//go:embed response-enums.json
var responseEnumsJSON []byte

// ResponseEnumsFor returns response enum supplements bound to cmd.
func ResponseEnumsFor(cmd string) ([]ResponseEnum, error) {
	var supplements responseEnumCatalog
	if err := json.Unmarshal(responseEnumsJSON, &supplements); err != nil {
		return nil, fmt.Errorf("解析响应枚举补充: %w", err)
	}
	var result []ResponseEnum
	for _, binding := range supplements.Bindings {
		if !strings.EqualFold(binding.Cmd, cmd) {
			continue
		}
		definition, ok := supplements.Definitions[binding.Definition]
		if !ok {
			return nil, fmt.Errorf("响应枚举定义 %q 不存在", binding.Definition)
		}
		result = append(result, ResponseEnum{
			Name:        binding.Definition,
			Fields:      binding.Fields,
			Description: definition.Description,
			Source:      definition.Source,
			Notes:       definition.Notes,
			Values:      definition.Values,
		})
	}
	return result, nil
}
