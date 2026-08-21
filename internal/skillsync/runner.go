package skillsync

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	skillsSource = "xiaowen-0725/openydt-cli"
	syncTimeout  = 3 * time.Minute
)

// RunResult captures the outcome of an npx skills invocation.
type RunResult struct {
	Stdout bytes.Buffer
	Stderr bytes.Buffer
	Err    error
}

// runnerOverride intercepts the npx invocation in tests. Nil in production.
var runnerOverride func(args ...string) *RunResult

// Source returns bundled Skills when installed through npm. Standalone release
// binaries fall back to the matching Git tag; development builds use main.
func Source(version string) string {
	return selectSource(version, bundledSkillsDir())
}

func selectSource(version, bundled string) string {
	if bundled != "" {
		return bundled
	}
	return sourceWithoutBundle(version)
}

func sourceWithoutBundle(version string) string {
	if isReleaseVersion(version) {
		return "https://github.com/" + skillsSource + "/tree/v" + normalizeVersion(version)
	}
	return skillsSource
}

func bundledSkillsDir() string {
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	binDir := filepath.Dir(executable)
	for _, candidate := range []string{
		filepath.Join(binDir, "skills"),
		filepath.Join(filepath.Dir(binDir), "skills"),
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

// RunSync runs `npx -y skills add <source> -g {-y|--all}`.
// force=true uses --all (reinstall every skill to every agent).
func RunSync(force bool, version string) *RunResult {
	last := "-y"
	if force {
		last = "--all"
	}
	return runSkills("-y", "skills", "add", Source(version), "-g", "--copy", last)
}

func runSkills(args ...string) *RunResult {
	if runnerOverride != nil {
		return runnerOverride(args...)
	}
	r := &RunResult{}
	npx, err := exec.LookPath("npx")
	if err != nil {
		r.Err = fmt.Errorf("npx not found in PATH: %w", err)
		return r
	}
	ctx, cancel := context.WithTimeout(context.Background(), syncTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, npx, args...)
	cmd.Stdout = &r.Stdout
	cmd.Stderr = &r.Stderr
	r.Err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		r.Err = fmt.Errorf("skills sync timed out after %s", syncTimeout)
	}
	return r
}

// npxAvailable reports whether npx is resolvable on PATH.
func npxAvailable() bool {
	_, err := exec.LookPath("npx")
	return err == nil
}

// SetRunnerOverrideForTest installs (or clears with nil) the npx override.
// Intended for tests in other packages.
func SetRunnerOverrideForTest(fn func(args ...string) *RunResult) { runnerOverride = fn }
