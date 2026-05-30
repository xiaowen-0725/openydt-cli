package cmdutil

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xiaowen-0725/openydt-cli/internal/catalog"
	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

// carFields are candidate body field names for a license plate, in priority
// order. The command's catalog schema decides which (if any) applies.
var carFields = []string{"carCode", "carNo"}

// injectDefaults fills missing body fields from profile defaults. It only adds a
// field when (a) the body omits it and (b) hasParam reports the command declares
// it. Existing values are never overwritten. Returns the (possibly rewritten)
// compact body and the injected field names. Pure + unit-testable.
func injectDefaults(body, defaultPark, defaultCarNo string, hasParam func(string) bool) (string, []string) {
	if defaultPark == "" && defaultCarNo == "" {
		return body, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		return body, nil
	}
	if m == nil {
		m = map[string]any{}
	}
	var injected []string
	if defaultPark != "" {
		if _, has := m["parkCode"]; !has && hasParam("parkCode") {
			m["parkCode"] = defaultPark
			injected = append(injected, "parkCode")
		}
	}
	if defaultCarNo != "" {
		for _, fld := range carFields {
			if hasParam(fld) {
				if _, has := m[fld]; !has {
					m[fld] = defaultCarNo
					injected = append(injected, fld)
				}
				break // command uses at most one plate field name
			}
		}
	}
	if len(injected) == 0 {
		return body, nil
	}
	out, err := json.Marshal(m)
	if err != nil {
		return body, nil
	}
	return string(out), injected
}

// applyDefaults resolves the active profile's defaults, consults the embedded
// catalog for which fields this command accepts, and injects missing ones. Any
// error (no config / no profile / no catalog / unknown cmd) is non-fatal — the
// body is returned unchanged.
func (f *Factory) applyDefaults(cmd, body string) string {
	cfg, err := config.Load()
	if err != nil {
		return body
	}
	p, ok := cfg.Active(f.Profile)
	if !ok || (p.DefaultPark == "" && p.DefaultCarNo == "") {
		return body
	}
	cat, err := catalog.Embedded()
	if err != nil {
		return body
	}
	it, ok := cat.Find(cmd)
	if !ok {
		return body
	}
	has := func(name string) bool { _, ok := it.FindParam(name); return ok }
	out, injected := injectDefaults(body, p.DefaultPark, p.DefaultCarNo, has)
	if len(injected) > 0 && f.Verbose {
		fmt.Fprintf(f.Err, "[openydt] 已用 profile %q 默认值补全: %s\n", p.Name, strings.Join(injected, ", "))
	}
	return out
}
