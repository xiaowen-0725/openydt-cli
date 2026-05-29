package catalog

import (
	_ "embed"
	"strings"
	"sync"

	"github.com/xiaowen-0725/openydt-cli/internal/strutil"
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

// EnumValues parses enum-like options from a param's description, e.g.
// "车牌颜色：0其他，1蓝色，2黄色" -> ["0 其他","1 蓝色","2 黄色"]. Returns nil if not enum-like.
func (p Param) EnumValues() []string { return strutil.ParseEnum(p.Desc) }
