package cmdutil

import (
	"encoding/json"
	"fmt"
	"regexp"
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

// defaultRe extracts a platform-documented default from a param desc, matching
// "默认1", "默认为1", "默认 1", "默认是0", etc. Only numeric defaults are honored —
// that is the form the platform uses for enum/flag/page fields.
var defaultRe = regexp.MustCompile(`默认\s*(?:为|是)?\s*(-?[0-9]+)`)

// schemaDefault returns the documented default embedded in a param desc, or ""
// if none. e.g. "记录类型（…）默认1" -> "1".
func schemaDefault(desc string) string {
	if m := defaultRe.FindStringSubmatch(desc); len(m) == 2 {
		return m[1]
	}
	return ""
}

// injectSchemaDefaults fills omitted body fields with the default the platform
// documents for them ("默认N" in the schema desc). It only touches top-level
// scalar params, never overwrites a value already present, coerces the default
// to the param's declared type, and returns the (possibly rewritten) compact
// body plus the field names it injected. Pure + unit-testable.
func injectSchemaDefaults(body string, params []catalog.Param) (string, []string) {
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		return body, nil
	}
	if m == nil {
		m = map[string]any{}
	}
	var injected []string
	for _, p := range params {
		if p.Group != "" { // nested object/array subfield — not a top-level key
			continue
		}
		raw := schemaDefault(p.Desc)
		if raw == "" {
			continue
		}
		if _, has := m[p.Name]; has {
			continue
		}
		v, err := coerce(p.Type, raw)
		if err != nil {
			continue
		}
		m[p.Name] = v
		injected = append(injected, p.Name)
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

// applySchemaDefaults consults the embedded catalog for the command's documented
// field defaults ("默认N") and injects any the body omits. Unlike applyDefaults
// (profile-scoped), this runs for every call — generated commands and `api`
// alike — so omitting a field the platform treats as required-with-default (e.g.
// supplementParkingRecordIn.recordType) no longer fails at the gateway. Any error
// (no catalog / unknown cmd / bad body) is non-fatal.
func (f *Factory) applySchemaDefaults(cmd, body string) string {
	cat, err := catalog.Embedded()
	if err != nil {
		return body
	}
	it, ok := cat.Find(cmd)
	if !ok {
		return body
	}
	out, injected := injectSchemaDefaults(body, it.Params)
	if len(injected) > 0 && f.Verbose {
		fmt.Fprintf(f.Err, "[openydt] 已按平台文档默认值补全: %s\n", strings.Join(injected, ", "))
	}
	return out
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
