package catalog

import (
	_ "embed"
	"regexp"
	"strings"
	"sync"
)

//go:embed catalog.json
var embedded []byte

var (
	embedOnce sync.Once
	embedCat  *Catalog
	embedErr  error
)

// Embedded returns the catalog compiled into the binary (catalog/catalog.json,
// kept in sync by `make generate`). Parsed once and cached.
func Embedded() (*Catalog, error) {
	embedOnce.Do(func() {
		embedCat, embedErr = parse(embedded)
	})
	return embedCat, embedErr
}

// FindParam returns the (case-insensitive) param by name for an interface,
// searching nested groups too.
func (it Iface) FindParam(name string) (Param, bool) {
	for _, p := range it.Params {
		if strings.EqualFold(p.Name, name) {
			return p, true
		}
	}
	return Param{}, false
}

var enumItemRe = regexp.MustCompile(`(?:^|[,，(（：:、])\s*(-?\d+)\s*([^,，)）、]+)`)

// EnumValues parses enum-like options from a param's description, e.g.
// "车牌颜色：0其他，1蓝色，2黄色" -> ["0 其他","1 蓝色","2 黄色"]. Returns nil if not enum-like.
func (p Param) EnumValues() []string {
	d := p.Desc
	if d == "" || !strings.ContainsAny(d, "0123456789") {
		return nil
	}
	ms := enumItemRe.FindAllStringSubmatch(d, -1)
	if len(ms) < 2 { // need at least two numbered options to look like an enum
		return nil
	}
	var out []string
	for _, m := range ms {
		label := strings.TrimSpace(m[2])
		label = strings.Trim(label, "：: 。.")
		if label == "" || len([]rune(label)) > 12 {
			continue // long text -> probably prose, not an enum label
		}
		out = append(out, m[1]+" "+label)
	}
	if len(out) < 2 {
		return nil
	}
	return out
}
