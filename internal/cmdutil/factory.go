// Package cmdutil wires shared dependencies (config, client, IO, global flags)
// into a Factory that every command receives.
package cmdutil

import (
	"fmt"
	"io"
	"os"

	"github.com/xiaowen-0725/openydt-cli/internal/client"
	"github.com/xiaowen-0725/openydt-cli/internal/config"
	"github.com/xiaowen-0725/openydt-cli/internal/output"
	"github.com/xiaowen-0725/openydt-cli/internal/sign"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// Factory carries IO streams and resolved global flags.
type Factory struct {
	Out io.Writer
	Err io.Writer

	// Global flags (bound on the root command).
	Profile string
	Env     string
	Output  string // json|table
	Sign    string // v2|v3 (empty = profile/default)
	Yes     bool
	DryRun  bool
	Verbose bool
}

// NewFactory returns a Factory writing to stdout/stderr.
func NewFactory() *Factory {
	return &Factory{Out: os.Stdout, Err: os.Stderr, Output: string(output.JSON)}
}

// UserAgent identifies this client to the gateway.
func (f *Factory) UserAgent() string { return "openydt-cli/" + Version }

// Format returns the requested output format.
func (f *Factory) Format() output.Format {
	if f.Output == string(output.Table) {
		return output.Table
	}
	return output.JSON
}

// Client resolves the active profile/env/sign and builds an API client.
func (f *Factory) Client() (*client.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	r, err := cfg.Resolve(f.Profile, f.Env, f.Sign)
	if err != nil {
		return nil, err
	}
	if f.Verbose {
		fmt.Fprintf(f.Err, "[openydt] profile=%s env=%s sign=%s base=%s\n", r.Profile, r.Env, r.Sign, r.BaseURL)
	}
	return client.New(r.BaseURL, r.Key, r.Secret, sign.Version(r.Sign), f.UserAgent()), nil
}
