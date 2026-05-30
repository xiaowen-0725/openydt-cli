package skillsync

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
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

// Source returns the skills source the CLI installs from.
func Source() string { return skillsSource }

// RunSync runs `npx -y skills add <source> -g {-y|--all}`.
// force=true uses --all (reinstall every skill to every agent).
func RunSync(force bool) *RunResult {
	last := "-y"
	if force {
		last = "--all"
	}
	return runSkills("-y", "skills", "add", skillsSource, "-g", last)
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
