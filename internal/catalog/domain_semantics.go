package catalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

type DeduplicationRule struct {
	Key     string `json:"key"`
	OrderBy string `json:"orderBy"`
	Keep    string `json:"keep"`
}

type DomainSemanticValue struct {
	Code              int      `json:"code"`
	EventNature       string   `json:"eventNature"`
	TrafficTreatment  string   `json:"trafficTreatment"`
	DurationTreatment string   `json:"durationTreatment"`
	BusinessTags      []string `json:"businessTags,omitempty"`
}

type FieldSemantics struct {
	Field       string                `json:"field"`
	Description string                `json:"description"`
	Values      []DomainSemanticValue `json:"values"`
}

type DomainSemantics struct {
	RecordSet              string             `json:"recordSet"`
	Description            string             `json:"description"`
	PrimaryDepartureSource *bool              `json:"primaryDepartureSource,omitempty"`
	ImpliesDeparture       *bool              `json:"impliesDeparture,omitempty"`
	Deduplication          *DeduplicationRule `json:"deduplication,omitempty"`
	Fields                 []FieldSemantics   `json:"fields"`
}

type domainSemanticsCatalog struct {
	Interfaces map[string]DomainSemantics `json:"interfaces"`
}

//go:embed domain-semantics.json
var domainSemanticsJSON []byte

func DomainSemanticsFor(cmd string) (*DomainSemantics, error) {
	var source domainSemanticsCatalog
	if err := json.Unmarshal(domainSemanticsJSON, &source); err != nil {
		return nil, fmt.Errorf("解析领域语义: %w", err)
	}
	for name, semantics := range source.Interfaces {
		if strings.EqualFold(name, cmd) {
			return &semantics, nil
		}
	}
	return nil, nil
}
