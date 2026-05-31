// Package catalog loads the generated interface catalog (catalog/catalog.json).
package catalog

import (
	"encoding/json"
	"os"
	"strings"
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

// IsWrite reports whether cmd is a write operation per the embedded catalog.
// Unknown cmds default to false (api 兜底未知 cmd 时不误拦,但已知写 cmd 必拦)。
func (c *Catalog) IsWrite(cmd string) bool {
	if it, ok := c.Find(cmd); ok {
		return it.ReadWrite == "write"
	}
	return false
}

// Hints holds per-command safety annotations (命令安全注解) derived from catalog
// fields. Inspired by tool-annotation conventions; this CLI is not an MCP server —
// hints surface in schema/--help to help agents decide before calling.
type Hints struct {
	ReadOnly       bool   `json:"readOnly"`
	Destructive    bool   `json:"destructive"`
	Idempotent     bool   `json:"idempotent"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// idemKeys are parameter names that carry a unique business key, indicating the
// operation is safe to retry with the same key (idempotent).
var idemKeys = []string{"billCode", "thirdBillCode", "thirdpartyBillCode", "uniqNo", "transationNum"}

// Hints derives safety annotations from the interface's existing catalog fields.
// ReadOnly is derived from readwrite=="read" (reads are naturally idempotent).
// For writes: Destructive is derived from the cmd name containing delete/del/cancel/
// remove/freeze/frozen/refund; Idempotent is derived from params containing a known
// idempotency key (billCode/thirdBillCode/thirdpartyBillCode/uniqNo/transationNum).
func (it Iface) Hints() Hints {
	h := Hints{ReadOnly: it.ReadWrite == "read"}
	if h.ReadOnly {
		h.Idempotent = true // 读天然幂等
		return h
	}
	lc := strings.ToLower(it.Cmd)
	for _, kw := range []string{"delete", "del", "cancel", "remove", "freeze", "frozen", "refund"} {
		if strings.Contains(lc, kw) {
			h.Destructive = true
			break
		}
	}
	for _, p := range it.Params {
		for _, k := range idemKeys {
			if p.Name == k {
				h.Idempotent = true
				h.IdempotencyKey = k
				return h
			}
		}
	}
	return h
}
