// Package config manages openydt-cli credentials and profiles.
//
// Config lives at ~/.config/openydt-cli/config.json and holds multiple
// authorized-merchant profiles (key/secret/env/sign). Environment variables
// override the resolved profile, which is handy for CI:
//
//	OPENYDT_PROFILE, OPENYDT_KEY, OPENYDT_SECRET, OPENYDT_ENV, OPENYDT_SIGN
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvBaseURL maps an environment name to its API base URL.
var EnvBaseURL = map[string]string{
	"test": "https://openapi-test.yidianting.com.cn",
	"dev":  "https://openapi-dev.yidianting.xin",
	"prod": "https://open.yidianting.xin",
}

const (
	DefaultEnv  = "test"
	DefaultSign = "v2"
)

// Profile is one authorized-merchant credential set.
type Profile struct {
	Name         string `json:"name"`
	Key          string `json:"key"`
	Secret       string `json:"secret"`
	Env          string `json:"env,omitempty"`          // test|dev|prod
	Sign         string `json:"sign,omitempty"`         // v2|v3
	DefaultPark  string `json:"defaultPark,omitempty"`  // 缺参时自动补的 parkCode
	DefaultCarNo string `json:"defaultCarNo,omitempty"` // 缺参时自动补的车牌
}

// Config is the on-disk configuration.
type Config struct {
	CurrentProfile string    `json:"currentProfile"`
	Profiles       []Profile `json:"profiles"`
}

// Dir returns the openydt-cli config directory, honoring XDG_CONFIG_HOME.
func Dir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "openydt-cli"), nil
}

// Path returns the config file path, honoring XDG_CONFIG_HOME.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config file. A missing file yields an empty Config.
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	return &c, nil
}

// Save writes the config file (0600, dir 0700).
func (c *Config) Save() error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

// Upsert adds or replaces a profile by name.
func (c *Config) Upsert(p Profile) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == p.Name {
			c.Profiles[i] = p
			return
		}
	}
	c.Profiles = append(c.Profiles, p)
}

// Find returns the profile with the given name.
func (c *Config) Find(name string) (Profile, bool) {
	for _, p := range c.Profiles {
		if p.Name == name {
			return p, true
		}
	}
	return Profile{}, false
}

// Active returns the profile selected by the same precedence as Resolve
// (flag > OPENYDT_PROFILE > CurrentProfile). Returns false when none resolves.
func (c *Config) Active(profileFlag string) (Profile, bool) {
	name := firstNonEmpty(profileFlag, os.Getenv("OPENYDT_PROFILE"), c.CurrentProfile)
	if name == "" {
		return Profile{}, false
	}
	return c.Find(name)
}

// Resolved is a fully-resolved credential context for a request.
type Resolved struct {
	Profile string
	Key     string
	Secret  string
	Env     string
	Sign    string
	BaseURL string
}

// Resolve merges (in increasing priority) defaults < profile < env vars <
// explicit flag overrides (empty overrides are ignored). profileFlag/envFlag/
// signFlag come from CLI flags and win when non-empty.
func (c *Config) Resolve(profileFlag, envFlag, signFlag string) (Resolved, error) {
	name := firstNonEmpty(profileFlag, os.Getenv("OPENYDT_PROFILE"), c.CurrentProfile)

	var p Profile
	if name != "" {
		if found, ok := c.Find(name); ok {
			p = found
		} else if os.Getenv("OPENYDT_KEY") == "" {
			return Resolved{}, fmt.Errorf("profile %q not found (run: openydt config set --profile %s --key ... --secret ...)", name, name)
		}
	}

	r := Resolved{
		Profile: name,
		Key:     firstNonEmpty(os.Getenv("OPENYDT_KEY"), p.Key),
		Secret:  firstNonEmpty(os.Getenv("OPENYDT_SECRET"), p.Secret),
		Env:     firstNonEmpty(envFlag, os.Getenv("OPENYDT_ENV"), p.Env, DefaultEnv),
		Sign:    strings.ToLower(firstNonEmpty(signFlag, os.Getenv("OPENYDT_SIGN"), p.Sign, DefaultSign)),
	}
	if r.Key == "" || r.Secret == "" {
		return Resolved{}, fmt.Errorf("missing key/secret: set a profile (openydt config set) or OPENYDT_KEY/OPENYDT_SECRET")
	}
	base, ok := EnvBaseURL[r.Env]
	if !ok {
		return Resolved{}, fmt.Errorf("unknown env %q (want test|dev|prod)", r.Env)
	}
	r.BaseURL = base
	return r, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
