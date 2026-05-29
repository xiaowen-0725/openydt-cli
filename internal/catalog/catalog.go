// Package catalog loads the generated interface catalog (catalog/catalog.json).
package catalog

import (
	"encoding/json"
	"os"
)

// Param is one request parameter.
type Param struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     string `json:"type"`
	Desc     string `json:"desc"`
	Group    string `json:"group"`
}

// Iface is one platform interface.
type Iface struct {
	Cmd           string  `json:"cmd"`
	Dir           string  `json:"dir"`
	Domain        string  `json:"domain"`
	Explain       string  `json:"explain"`
	FitSystem     string  `json:"fitSystem"`
	Pattern       string  `json:"pattern"`
	Direction     string  `json:"direction"` // callable | webhook
	ReadWrite     string  `json:"readwrite"` // read | write
	Params        []Param `json:"params"`
	SampleBody    string  `json:"sampleBody"`
	SampleResp    string  `json:"sampleResponse"`
	Included      bool    `json:"included"`
	ExcludeReason string  `json:"excludeReason"`
}

// Catalog is the whole inventory.
type Catalog struct {
	Count      int     `json:"count"`
	Interfaces []Iface `json:"interfaces"`
}

// Load reads a catalog.json from path.
func Load(path string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parse(data)
}

func parse(data []byte) (*Catalog, error) {
	var c Catalog
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Find returns the interface with the given cmd.
func (c *Catalog) Find(cmd string) (Iface, bool) {
	for _, it := range c.Interfaces {
		if it.Cmd == cmd {
			return it, true
		}
	}
	return Iface{}, false
}

// Included returns all first-class (callable + in-scope) interfaces.
func (c *Catalog) Included() []Iface {
	var out []Iface
	for _, it := range c.Interfaces {
		if it.Included && it.Direction == "callable" {
			out = append(out, it)
		}
	}
	return out
}
