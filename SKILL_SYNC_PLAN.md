# openydt skills 自动同步 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 openydt-cli 在 `npm i/update` 时自动把 11 个技能同步到本机所有已装 AI agent,并在二进制检测到版本漂移时后台静默补同步,同时提供 `openydt skill sync` 手动命令。

**Architecture:** 三条路径共用一个 state 文件(`~/.config/openydt-cli/skills-state.json`):① npm `postinstall` 跑 `npx skills add`(主路径,install+update 触发);② Go 二进制 `root.PersistentPreRunE` → `skillsync.MaybeTrigger` 检测漂移 → detached 后台 `openydt skill sync --quiet`(兜底);③ `openydt skill sync [--force]` 手动命令。全程 best-effort,绝不阻断主流程或污染 stdout。

**Tech Stack:** Go 1.26 + spf13/cobra;外部 `npx skills`(vercel-labs 包管理器);state 走 `internal/config` 的 XDG 目录约定。

**与 spec 的两处规划期细化(见 `SKILL_SYNC_DESIGN.md`):**
1. 降级 notice 只打到 **stderr**(在 `cmd/root.go` 的 `Execute()` 里),不注入 JSON envelope → `internal/output/output.go` 不改动,stdout 对 agent 始终干净。
2. 触发用 `root.PersistentPreRunE`,`skill` 命令用自己的空 `PersistentPreRun` 退出触发 → 手动 `skill sync` 不会再 fork 一个后台同步。

---

## File Structure

| 文件 | 责任 |
|---|---|
| `internal/config/config.go`(改) | 抽出 `Dir()` 返回配置目录;`Path()` 复用之 |
| `internal/skillsync/state.go`(建) | `State` 类型、原子读写、`normalizeVersion`、`RecordSuccess`、`nowFunc` 时钟 seam |
| `internal/skillsync/skip.go`(建) | `shouldSkip` / `isCIEnv` / `isReleaseVersion` + env 常量 |
| `internal/skillsync/runner.go`(建) | `RunResult`、`RunSync(force)`、`runSkills`、`npxAvailable`、`Source`、`runnerOverride` seam |
| `internal/skillsync/notice.go`(建) | 降级 `Notice` 的进程级存取 + `Message()` |
| `internal/skillsync/spawn_unix.go`(建) | `//go:build !windows` 的 `detachSysProcAttr` |
| `internal/skillsync/spawn_windows.go`(建) | `//go:build windows` 的 `detachSysProcAttr` |
| `internal/skillsync/trigger.go`(建) | `MaybeTrigger`、`recordAttempt`、`spawnBackgroundSync`、`spawnFunc`/`npxAvailableFunc` seam |
| `cmd/skill/skill.go`(建) | `openydt skill` + `sync` 子命令(`--force`/`--quiet`,空 PersistentPreRun) |
| `cmd/root.go`(改) | 注册 skill 命令、挂 `PersistentPreRunE`、`Execute()` 末尾打降级 notice |
| `npm/install.js`(改) | 下载二进制后 best-effort `npx skills add` + 写 state |
| `README.md` / `CLAUDE.md` / `PROJECT_STATUS.md`(改) | 文档 |

---

## Task 1: `config.Dir()` 配置目录辅助

**Files:**
- Modify: `internal/config/config.go`(`Path()` 约在 45-56 行)
- Test: `internal/config/dir_test.go`

- [ ] **Step 1: 写失败测试**

Create `internal/config/dir_test.go`:
```go
package config

import (
	"path/filepath"
	"testing"
)

func TestDirHonorsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdghome")
	got, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}
	want := filepath.Join("/tmp/xdghome", "openydt-cli")
	if got != want {
		t.Fatalf("Dir() = %s, want %s", got, want)
	}
}

func TestPathUsesDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdghome")
	got, err := Path()
	if err != nil {
		t.Fatalf("Path() error: %v", err)
	}
	want := filepath.Join("/tmp/xdghome", "openydt-cli", "config.json")
	if got != want {
		t.Fatalf("Path() = %s, want %s", got, want)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/config/ -run TestDir -v`
Expected: 编译失败 `undefined: Dir`。

- [ ] **Step 3: 实现 `Dir()`,让 `Path()` 复用**

在 `internal/config/config.go` 把现有 `Path()` 替换为:
```go
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
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/config/ -run "TestDir|TestPath" -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/config/config.go internal/config/dir_test.go
git commit -m "refactor(config): 抽出 Dir() 供 skillsync 复用配置目录"
```

---

## Task 2: `skillsync` state 读写

**Files:**
- Create: `internal/skillsync/state.go`
- Test: `internal/skillsync/state_test.go`

- [ ] **Step 1: 写失败测试**

Create `internal/skillsync/state_test.go`(完整可粘贴,单一 import 块,helper 在末尾):
```go
package skillsync

import (
	"os"
	"testing"
	"time"
)

func writeRaw(p string, b []byte) error { return os.WriteFile(p, b, 0o644) }

func TestNormalizeVersion(t *testing.T) {
	cases := map[string]string{
		"v0.1.1": "0.1.1", "V0.1.1": "0.1.1", "0.1.1": "0.1.1", " v0.1.1 ": "0.1.1",
	}
	for in, want := range cases {
		if got := normalizeVersion(in); got != want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStateRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if s, err := ReadState(); err != nil || s != nil {
		t.Fatalf("cold read: got (%v,%v), want (nil,nil)", s, err)
	}

	in := State{Version: "0.1.1", LastAttemptVersion: "0.1.1", Skills: []string{"openydt-billing"}, UpdatedAt: "2026-05-30T00:00:00Z"}
	if err := WriteState(in); err != nil {
		t.Fatalf("WriteState: %v", err)
	}
	got, err := ReadState()
	if err != nil || got == nil {
		t.Fatalf("ReadState after write: (%v,%v)", got, err)
	}
	if got.Version != "0.1.1" || got.LastAttemptVersion != "0.1.1" || len(got.Skills) != 1 {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

func TestReadCorruptStateErrors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := WriteState(State{Version: "0.1.1"}); err != nil {
		t.Fatal(err)
	}
	// corrupt it
	p, _ := statePath()
	if err := writeRaw(p, []byte("{not json")); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadState(); err == nil {
		t.Fatalf("expected error on corrupt state")
	}
}

func TestRecordSuccess(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	nowFunc = func() time.Time { return time.Unix(0, 0).UTC() }
	defer func() { nowFunc = time.Now }()
	if err := RecordSuccess("v0.1.2"); err != nil {
		t.Fatal(err)
	}
	got, _ := ReadState()
	if got.Version != "v0.1.2" || got.LastAttemptVersion != "v0.1.2" {
		t.Fatalf("RecordSuccess state = %+v", got)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/skillsync/ -run TestState -v`
Expected: 编译失败(包/类型未定义)。

- [ ] **Step 3: 实现 state.go**

Create `internal/skillsync/state.go`:
```go
// Package skillsync keeps the locally installed openydt AI-agent skills in
// sync with the running binary, via the external `npx skills` package manager.
package skillsync

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

const stateFile = "skills-state.json"

// nowFunc is the clock seam (overridable in tests).
var nowFunc = time.Now

// State records the last skill-sync attempt/success for drift detection.
type State struct {
	Version            string   `json:"version"`                 // last SUCCESSFUL sync version
	LastAttemptVersion string   `json:"last_attempt_version"`    // last attempted version (debounce)
	Skills             []string `json:"skills,omitempty"`        // informational
	UpdatedAt          string   `json:"updated_at"`
}

func statePath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, stateFile), nil
}

// ReadState returns the stored state. Missing file -> (nil, nil). Corrupt -> error.
func ReadState() (*State, error) {
	p, err := statePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// WriteState atomically writes the state file (tmp + rename).
func WriteState(s State) error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	p := filepath.Join(dir, stateFile)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// RecordSuccess persists a successful sync at version.
func RecordSuccess(version string) error {
	prev, _ := ReadState()
	s := State{}
	if prev != nil {
		s = *prev
	}
	s.Version = version
	s.LastAttemptVersion = version
	s.UpdatedAt = nowFunc().UTC().Format(time.RFC3339)
	return WriteState(s)
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/skillsync/ -run "TestNormalizeVersion|TestState|TestReadCorrupt|TestRecordSuccess" -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/skillsync/state.go internal/skillsync/state_test.go
git commit -m "feat(skillsync): state 读写 + 版本归一 + RecordSuccess"
```

---

## Task 3: 跳过规则 `skip.go`

**Files:**
- Create: `internal/skillsync/skip.go`
- Test: `internal/skillsync/skip_test.go`

- [ ] **Step 1: 写失败测试**

Create `internal/skillsync/skip_test.go`:
```go
package skillsync

import "testing"

func TestShouldSkip(t *testing.T) {
	// 清掉可能存在的 CI 变量,保证基线可控
	for _, k := range []string{"CI", "GITHUB_ACTIONS", "BUILD_NUMBER", "GITLAB_CI", "OPENYDT_NO_SKILLS_SYNC", "OPENYDT_SKILL_SYNC_CHILD"} {
		t.Setenv(k, "")
	}

	if shouldSkip("0.1.1") {
		t.Fatalf("release version on clean env should NOT skip")
	}

	t.Setenv("OPENYDT_NO_SKILLS_SYNC", "1")
	if !shouldSkip("0.1.1") {
		t.Fatalf("opt-out env should skip")
	}
	t.Setenv("OPENYDT_NO_SKILLS_SYNC", "")

	t.Setenv("OPENYDT_SKILL_SYNC_CHILD", "1")
	if !shouldSkip("0.1.1") {
		t.Fatalf("child marker should skip")
	}
	t.Setenv("OPENYDT_SKILL_SYNC_CHILD", "")

	t.Setenv("CI", "true")
	if !shouldSkip("0.1.1") {
		t.Fatalf("CI should skip")
	}
	t.Setenv("CI", "")

	for _, v := range []string{"", "dev", "DEV", "0.1.1-3-gabc", "v1.2", "garbage"} {
		if !shouldSkip(v) {
			t.Fatalf("non-release version %q should skip", v)
		}
	}
	for _, v := range []string{"0.1.1", "v0.1.1", "V2.0.0", "10.20.30"} {
		if shouldSkip(v) {
			t.Fatalf("release version %q should NOT skip", v)
		}
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/skillsync/ -run TestShouldSkip -v`
Expected: 编译失败 `undefined: shouldSkip`。

- [ ] **Step 3: 实现 skip.go**

Create `internal/skillsync/skip.go`:
```go
package skillsync

import (
	"os"
	"regexp"
	"strings"
)

const (
	envOptOut    = "OPENYDT_NO_SKILLS_SYNC"
	envChildMark = "OPENYDT_SKILL_SYNC_CHILD"
)

// releaseVersionRe matches a clean release semver, optionally v/V-prefixed.
var releaseVersionRe = regexp.MustCompile(`^[vV]?\d+\.\d+\.\d+$`)

// shouldSkip reports whether automatic skill sync must be suppressed.
func shouldSkip(version string) bool {
	if os.Getenv(envOptOut) != "" {
		return true
	}
	if os.Getenv(envChildMark) != "" {
		return true
	}
	if isCIEnv() {
		return true
	}
	return !isReleaseVersion(version)
}

func isCIEnv() bool {
	for _, k := range []string{"CI", "GITHUB_ACTIONS", "BUILD_NUMBER", "GITLAB_CI"} {
		if os.Getenv(k) != "" {
			return true
		}
	}
	return false
}

func isReleaseVersion(v string) bool {
	v = strings.TrimSpace(v)
	switch v {
	case "", "dev", "DEV":
		return false
	}
	return releaseVersionRe.MatchString(v)
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/skillsync/ -run TestShouldSkip -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/skillsync/skip.go internal/skillsync/skip_test.go
git commit -m "feat(skillsync): 跳过规则(opt-out/CI/DEV/非release/子进程)"
```

---

## Task 4: npx runner `runner.go`

**Files:**
- Create: `internal/skillsync/runner.go`
- Test: `internal/skillsync/runner_test.go`

- [ ] **Step 1: 写失败测试**

Create `internal/skillsync/runner_test.go`:
```go
package skillsync

import (
	"strings"
	"testing"
)

func TestRunSyncArgs(t *testing.T) {
	var captured []string
	runnerOverride = func(args ...string) *RunResult {
		captured = args
		return &RunResult{}
	}
	defer func() { runnerOverride = nil }()

	RunSync(false)
	want := "-y skills add xiaowen-0725/openydt-cli -g -y"
	if got := strings.Join(captured, " "); got != want {
		t.Fatalf("RunSync(false) args = %q, want %q", got, want)
	}

	RunSync(true)
	wantForce := "-y skills add xiaowen-0725/openydt-cli -g --all"
	if got := strings.Join(captured, " "); got != wantForce {
		t.Fatalf("RunSync(true) args = %q, want %q", got, wantForce)
	}
}

func TestSourceConst(t *testing.T) {
	if Source() != "xiaowen-0725/openydt-cli" {
		t.Fatalf("Source() = %q", Source())
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/skillsync/ -run "TestRunSync|TestSource" -v`
Expected: 编译失败(`RunResult`/`RunSync`/`Source`/`runnerOverride` 未定义)。

- [ ] **Step 3: 实现 runner.go**

Create `internal/skillsync/runner.go`:
```go
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
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/skillsync/ -run "TestRunSync|TestSource" -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/skillsync/runner.go internal/skillsync/runner_test.go
git commit -m "feat(skillsync): npx skills runner(RunSync/Source/可注入 override)"
```

---

## Task 5: 降级 notice `notice.go`

**Files:**
- Create: `internal/skillsync/notice.go`
- Test: `internal/skillsync/notice_test.go`

- [ ] **Step 1: 写失败测试**

Create `internal/skillsync/notice_test.go`:
```go
package skillsync

import (
	"strings"
	"testing"
)

func TestNoticePending(t *testing.T) {
	setPending(nil)
	if Pending() != nil {
		t.Fatalf("expected nil pending")
	}
	setPending(&Notice{Reason: "未检测到 npx/node"})
	n := Pending()
	if n == nil || !strings.Contains(n.Message(), "skills add") {
		t.Fatalf("notice message missing fix command: %+v", n)
	}
	if !strings.Contains(n.Message(), "未检测到 npx/node") {
		t.Fatalf("notice message missing reason: %s", n.Message())
	}
	setPending(nil)
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/skillsync/ -run TestNoticePending -v`
Expected: 编译失败(`Notice`/`setPending`/`Pending` 未定义)。

- [ ] **Step 3: 实现 notice.go**

Create `internal/skillsync/notice.go`:
```go
package skillsync

import (
	"fmt"
	"sync/atomic"
)

// Notice is a degraded-mode hint surfaced (to stderr only) when automatic
// sync cannot run — e.g. npx/node missing.
type Notice struct {
	Reason string `json:"reason"`
}

// Message returns a one-line hint plus the manual fix command.
func (n *Notice) Message() string {
	return fmt.Sprintf("skills 未自动同步(%s);手动同步: npx skills add %s -g -y", n.Reason, skillsSource)
}

var pending atomic.Pointer[Notice]

func setPending(n *Notice) { pending.Store(n) }

// Pending returns the degraded-mode notice for this process, or nil.
func Pending() *Notice { return pending.Load() }
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/skillsync/ -run TestNoticePending -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/skillsync/notice.go internal/skillsync/notice_test.go
git commit -m "feat(skillsync): 降级 notice 进程级存取"
```

---

## Task 6: 跨平台 detached spawn 属性

**Files:**
- Create: `internal/skillsync/spawn_unix.go`
- Create: `internal/skillsync/spawn_windows.go`

> 这两个文件是 `syscall` 平台属性,不写单测(由 `go vet` + 跨平台 build 验证)。

- [ ] **Step 1: 实现 spawn_unix.go**

Create `internal/skillsync/spawn_unix.go`:
```go
//go:build !windows

package skillsync

import "syscall"

// detachSysProcAttr makes the child its own session leader so it survives the
// parent CLI process exiting.
func detachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
```

- [ ] **Step 2: 实现 spawn_windows.go**

Create `internal/skillsync/spawn_windows.go`:
```go
//go:build windows

package skillsync

import "syscall"

const (
	createNewProcessGroup = 0x00000200 // CREATE_NEW_PROCESS_GROUP
	detachedProcess       = 0x00000008 // DETACHED_PROCESS
)

// detachSysProcAttr detaches the child from the parent console on Windows.
func detachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: createNewProcessGroup | detachedProcess}
}
```

- [ ] **Step 3: 验证两平台都编译**

Run:
```bash
go build ./internal/skillsync/
GOOS=windows GOARCH=amd64 go build ./internal/skillsync/
```
Expected: 两条都无输出(成功)。

- [ ] **Step 4: 提交**

```bash
git add internal/skillsync/spawn_unix.go internal/skillsync/spawn_windows.go
git commit -m "feat(skillsync): 跨平台 detached spawn 属性(unix/windows)"
```

---

## Task 7: 漂移检测与后台触发 `trigger.go`

**Files:**
- Create: `internal/skillsync/trigger.go`
- Test: `internal/skillsync/trigger_test.go`

- [ ] **Step 1: 写失败测试**

Create `internal/skillsync/trigger_test.go`:
```go
package skillsync

import (
	"testing"
	"time"
)

// withTriggerSeams 安装可控的 spawn / npx / clock seam,并在测试结束还原。
func withTriggerSeams(t *testing.T, npx bool) *int {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, k := range []string{"CI", "GITHUB_ACTIONS", "BUILD_NUMBER", "GITLAB_CI", "OPENYDT_NO_SKILLS_SYNC", "OPENYDT_SKILL_SYNC_CHILD"} {
		t.Setenv(k, "")
	}
	calls := 0
	spawnFunc = func(string) error { calls++; return nil }
	npxAvailableFunc = func() bool { return npx }
	nowFunc = func() time.Time { return time.Unix(0, 0).UTC() }
	t.Cleanup(func() {
		spawnFunc = spawnBackgroundSync
		npxAvailableFunc = npxAvailable
		nowFunc = time.Now
		setPending(nil)
	})
	return &calls
}

func TestMaybeTriggerColdStartSpawns(t *testing.T) {
	calls := withTriggerSeams(t, true)
	MaybeTrigger("0.1.1")
	if *calls != 1 {
		t.Fatalf("cold start: spawn calls = %d, want 1", *calls)
	}
	// 第二次同版本应被防抖跳过
	MaybeTrigger("0.1.1")
	if *calls != 1 {
		t.Fatalf("debounce: spawn calls = %d, want 1", *calls)
	}
}

func TestMaybeTriggerInSyncNoSpawn(t *testing.T) {
	calls := withTriggerSeams(t, true)
	if err := RecordSuccess("0.1.1"); err != nil {
		t.Fatal(err)
	}
	MaybeTrigger("0.1.1")
	if *calls != 0 {
		t.Fatalf("in-sync: spawn calls = %d, want 0", *calls)
	}
}

func TestMaybeTriggerDriftSpawns(t *testing.T) {
	calls := withTriggerSeams(t, true)
	if err := RecordSuccess("0.1.0"); err != nil {
		t.Fatal(err)
	}
	MaybeTrigger("0.1.1")
	if *calls != 1 {
		t.Fatalf("drift: spawn calls = %d, want 1", *calls)
	}
}

func TestMaybeTriggerNoNpxSetsNotice(t *testing.T) {
	calls := withTriggerSeams(t, false)
	MaybeTrigger("0.1.1")
	if *calls != 0 {
		t.Fatalf("no npx: spawn calls = %d, want 0", *calls)
	}
	if Pending() == nil {
		t.Fatalf("no npx: expected degraded notice")
	}
}

func TestMaybeTriggerSkipsWhenOptOut(t *testing.T) {
	calls := withTriggerSeams(t, true)
	t.Setenv("OPENYDT_NO_SKILLS_SYNC", "1")
	MaybeTrigger("0.1.1")
	if *calls != 0 {
		t.Fatalf("opt-out: spawn calls = %d, want 0", *calls)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/skillsync/ -run TestMaybeTrigger -v`
Expected: 编译失败(`MaybeTrigger`/`spawnFunc`/`npxAvailableFunc`/`spawnBackgroundSync` 未定义)。

- [ ] **Step 3: 实现 trigger.go**

Create `internal/skillsync/trigger.go`:
```go
package skillsync

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

const logFileName = "skills-sync.log"

// Seams overridable in tests.
var (
	spawnFunc        = spawnBackgroundSync
	npxAvailableFunc = npxAvailable
)

// MaybeTrigger checks whether installed skills drifted from the running binary
// version and, if so, launches a detached background sync. Local-only and
// non-blocking — it never writes to the calling command's stdout/stderr.
func MaybeTrigger(version string) {
	setPending(nil)
	if shouldSkip(version) {
		return
	}
	state, err := ReadState()
	if err != nil {
		state = nil // corrupt/unreadable -> treat as cold start
	}
	if state != nil && normalizeVersion(state.Version) == normalizeVersion(version) {
		return // in sync
	}
	if state != nil && normalizeVersion(state.LastAttemptVersion) == normalizeVersion(version) {
		return // already attempted at this version (debounce)
	}
	if !npxAvailableFunc() {
		setPending(&Notice{Reason: "未检测到 npx/node"})
		return
	}
	// Record the attempt BEFORE spawning so concurrent/repeated invocations at
	// this version do not re-spawn.
	recordAttempt(state, version)
	_ = spawnFunc(version)
}

func recordAttempt(prev *State, version string) {
	s := State{}
	if prev != nil {
		s = *prev
	}
	s.LastAttemptVersion = version
	s.UpdatedAt = nowFunc().UTC().Format(time.RFC3339)
	_ = WriteState(s) // best-effort
}

// spawnBackgroundSync launches `openydt skill sync --quiet` detached, with
// output redirected to the sync log. It never blocks (no Wait).
func spawnBackgroundSync(version string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	lp, err := syncLogPath()
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(lp), 0o700)
	lf, err := os.OpenFile(lp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	cmd := exec.Command(self, "skill", "sync", "--quiet")
	cmd.Env = append(os.Environ(), envChildMark+"=1")
	cmd.Stdin = nil
	cmd.Stdout = lf
	cmd.Stderr = lf
	cmd.SysProcAttr = detachSysProcAttr()
	if err := cmd.Start(); err != nil {
		lf.Close()
		return err
	}
	lf.Close() // child holds its own descriptor
	return nil
}

func syncLogPath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, logFileName), nil
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/skillsync/ -v`
Expected: 全部 PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/skillsync/trigger.go internal/skillsync/trigger_test.go
git commit -m "feat(skillsync): 漂移检测 + 防抖 + detached 后台触发"
```

---

## Task 8: `openydt skill sync` 命令

**Files:**
- Create: `cmd/skill/skill.go`
- Test: `cmd/skill/skill_test.go`

- [ ] **Step 1: 写失败测试**

Create `cmd/skill/skill_test.go`:
```go
package skill

import (
	"bytes"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/skillsync"
)

func TestSkillCommandWiring(t *testing.T) {
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	cmd := New(f)
	if cmd.Use != "skill" {
		t.Fatalf("Use = %q", cmd.Use)
	}
	sub, _, err := cmd.Find([]string{"sync"})
	if err != nil || sub.Use != "sync" {
		t.Fatalf("sync subcommand missing: %v", err)
	}
	// skill 命令必须有空 PersistentPreRun 以退出根命令的自动触发
	if cmd.PersistentPreRun == nil {
		t.Fatalf("skill cmd must define a (no-op) PersistentPreRun to opt out of auto-trigger")
	}
}

func TestSyncSuccess(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	skillsync.SetRunnerOverrideForTest(func(args ...string) *skillsync.RunResult {
		return &skillsync.RunResult{}
	})
	defer skillsync.SetRunnerOverrideForTest(nil)

	errOut := &bytes.Buffer{}
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: errOut}
	cmd := New(f)
	cmd.SetArgs([]string{"sync"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync execute: %v", err)
	}
	st, _ := skillsync.ReadState()
	if st == nil || st.Version != cmdutil.Version {
		t.Fatalf("sync should RecordSuccess; state=%+v", st)
	}
}

func TestSyncFailureReturnsError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	skillsync.SetRunnerOverrideForTest(func(args ...string) *skillsync.RunResult {
		r := &skillsync.RunResult{}
		r.Err = errForTest("boom")
		return r
	})
	defer skillsync.SetRunnerOverrideForTest(nil)

	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	cmd := New(f)
	cmd.SetArgs([]string{"sync", "--quiet"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("sync failure should return error")
	}
}

type errForTest string

func (e errForTest) Error() string { return string(e) }
```

> 这里测试需要从外部包设置 `runnerOverride`。因此在 runner.go 暴露一个仅测试用的 setter。

- [ ] **Step 2: 在 runner.go 增加测试用 setter**

Append to `internal/skillsync/runner.go`:
```go
// SetRunnerOverrideForTest installs (or clears with nil) the npx override.
// Intended for tests in other packages.
func SetRunnerOverrideForTest(fn func(args ...string) *RunResult) { runnerOverride = fn }
```

- [ ] **Step 3: 跑测试确认失败**

Run: `go test ./cmd/skill/ -v`
Expected: 编译失败(`New`/`SetRunnerOverrideForTest` 未定义)。

- [ ] **Step 4: 实现 skill.go**

Create `cmd/skill/skill.go`:
```go
// Package skill provides `openydt skill` — manage AI-agent skills.
package skill

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
	"github.com/xiaowen-0725/openydt-cli/internal/output"
	"github.com/xiaowen-0725/openydt-cli/internal/skillsync"
)

// New builds the `openydt skill` command tree.
func New(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "管理 AI Agent 技能(同步到本机已装的各 agent)",
		Long: `把 openydt 的技能同步到本机所有已装 AI agent(Claude Code / Codex / Cursor / ... )。
底层调用 npx skills(需 Node.js)。日常 npm 安装/更新会自动同步,本命令用于手动或强制重装。`,
		// 空 PersistentPreRun:让 skill 子命令跳过根命令的自动同步触发,
		// 避免手动 sync 时再 fork 一个后台 sync。
		PersistentPreRun: func(*cobra.Command, []string) {},
	}
	cmd.AddCommand(newSyncCmd(f))
	return cmd
}

func newSyncCmd(f *cmdutil.Factory) *cobra.Command {
	var quiet, force bool
	c := &cobra.Command{
		Use:   "sync",
		Short: "把 openydt 技能同步到本机所有已装 agent(npx skills add)",
		RunE: func(_ *cobra.Command, _ []string) error {
			res := skillsync.RunSync(force)
			if res.Err != nil {
				if !quiet {
					fmt.Fprintf(f.Err, "✗ skills 同步失败: %v\n", res.Err)
				}
				return cmdutil.ExitError{Code: output.ExitAPIError, Err: res.Err}
			}
			if err := skillsync.RecordSuccess(cmdutil.Version); err != nil && !quiet {
				fmt.Fprintf(f.Err, "warning: 同步成功但 state 未写入: %v\n", err)
			}
			if !quiet {
				fmt.Fprintf(f.Err, "✓ skills 已同步(source: %s)\n", skillsync.Source())
			}
			return nil
		},
	}
	c.Flags().BoolVar(&quiet, "quiet", false, "静默(供后台自动同步子进程使用)")
	c.Flags().BoolVar(&force, "force", false, "全量重装所有技能到所有 agent")
	return c
}
```

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./cmd/skill/ ./internal/skillsync/ -v`
Expected: 全部 PASS。

- [ ] **Step 6: 提交**

```bash
git add cmd/skill/skill.go cmd/skill/skill_test.go internal/skillsync/runner.go
git commit -m "feat(cmd/skill): openydt skill sync 手动命令(--force/--quiet)"
```

---

## Task 9: 接到 root 命令 + 降级 notice 输出

**Files:**
- Modify: `cmd/root.go`
- Test: `cmd/root_test.go`

- [ ] **Step 1: 写失败测试**

Create `cmd/root_test.go`:
```go
package cmd

import (
	"bytes"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
)

func TestRootHasSkillCommand(t *testing.T) {
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	root := NewRootCmd(f)
	sub, _, err := root.Find([]string{"skill", "sync"})
	if err != nil || sub.Use != "sync" {
		t.Fatalf("expected root -> skill -> sync; err=%v", err)
	}
}

func TestRootHasAutoSyncPreRun(t *testing.T) {
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	root := NewRootCmd(f)
	if root.PersistentPreRunE == nil {
		t.Fatalf("root must wire PersistentPreRunE for skill auto-sync")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./cmd/ -run "TestRootHas" -v`
Expected: FAIL(`skill` 命令未注册 / `PersistentPreRunE` 为 nil)。

- [ ] **Step 3: 修改 cmd/root.go**

在 import 块加入:
```go
	skillcmd "github.com/xiaowen-0725/openydt-cli/cmd/skill"
	"github.com/xiaowen-0725/openydt-cli/internal/skillsync"
```

在 `NewRootCmd` 里,给 `root` 结构体补 `PersistentPreRunE`(放在 `root := &cobra.Command{...}` 内,`Long:` 之后):
```go
		// 每条普通命令执行前做一次本地版本比对;漂移则后台静默补同步。
		// skill 子命令用自己的空 PersistentPreRun 退出此触发。
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			skillsync.MaybeTrigger(cmdutil.Version)
			return nil
		},
```

在 `root.AddCommand(...)` 那组里加入 `skillcmd.New(f)`:
```go
	root.AddCommand(
		configcmd.New(f),
		authcmd.New(f),
		apicmd.New(f),
		schemacmd.New(f),
		skillcmd.New(f),
	)
```

在 `Execute()` 中,`err := root.Execute()` 之后、错误分支之前,加降级 notice 输出:
```go
	err := root.Execute()
	if n := skillsync.Pending(); n != nil {
		fmt.Fprintln(f.Err, "[openydt]", n.Message())
	}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./cmd/ -run "TestRootHas" -v`
Expected: PASS。

- [ ] **Step 5: 全量构建 + 冒烟**

Run:
```bash
make build
OPENYDT_NO_SKILLS_SYNC=1 ./bin/openydt skill sync --help
OPENYDT_NO_SKILLS_SYNC=1 ./bin/openydt --help >/dev/null && echo "root ok"
```
Expected: `skill sync --help` 打印用法;`root ok`。(用 opt-out 避免冒烟时真跑 npx。)

- [ ] **Step 6: 提交**

```bash
git add cmd/root.go cmd/root_test.go
git commit -m "feat(cmd): 注册 skill 命令 + PersistentPreRunE 自动同步 + 降级 notice"
```

---

## Task 10: npm postinstall 自动同步

**Files:**
- Modify: `npm/install.js`(在末尾安装完成处追加)

> 无单测(postinstall 脚本);用 `node -c` 语法检查 + 人工走查验证。

- [ ] **Step 1: 修改 npm/install.js**

把现有的 IIFE 结尾(`console.log("[openydt] 安装完成");` 之后)改为追加技能同步 + 写 state。将:
```js
  fs.unlinkSync(archive);
  console.log("[openydt] 安装完成");
})().catch((e) => {
```
替换为:
```js
  fs.unlinkSync(archive);
  console.log("[openydt] 安装完成");

  // best-effort: 同步 AI Agent 技能到本机已装的各 agent(失败绝不影响安装)
  syncSkills();
})().catch((e) => {
```

并在文件末尾(最后的 `});` 之后)追加两个辅助函数:
```js

// 把技能同步到本机所有已装 agent;失败只 warn,postinstall 不报错退出。
function syncSkills() {
  try {
    execFileSync("npx", ["-y", "skills", "add", REPO, "-g", "-y"], { stdio: "inherit" });
    writeSkillsState();
    console.log("[openydt] skills 已同步");
  } catch (e) {
    console.warn(`[openydt] skills 同步失败(可稍后手动:openydt skill sync):${e.message}`);
  }
}

// 写 skills-state.json,使二进制兜底的漂移检测有基准(XDG-aware)。
function writeSkillsState() {
  const base = process.env.XDG_CONFIG_HOME || path.join(require("os").homedir(), ".config");
  const dir = path.join(base, "openydt-cli");
  fs.mkdirSync(dir, { recursive: true });
  const state = {
    version: pkg.version,
    last_attempt_version: pkg.version,
    updated_at: new Date().toISOString(),
  };
  fs.writeFileSync(path.join(dir, "skills-state.json"), JSON.stringify(state, null, 2) + "\n");
}
```

- [ ] **Step 2: 语法检查**

Run: `node -c npm/install.js && echo "syntax ok"`
Expected: `syntax ok`。

- [ ] **Step 3: 人工走查**

确认:`syncSkills()` 的 `catch` 只 `console.warn` 且 **不** `process.exit(非0)`;state 的 `version` 用 `pkg.version`(与二进制 ldflags 版本经 `normalizeVersion` 可比)。

- [ ] **Step 4: 提交**

```bash
git add npm/install.js
git commit -m "feat(npm): postinstall best-effort 同步 skills + 写 state"
```

---

## Task 11: 文档更新

**Files:**
- Modify: `README.md`(安装段,约 11-21 行)
- Modify: `CLAUDE.md`(新增「技能同步」节)
- Modify: `PROJECT_STATUS.md`(待办第 5 项)

- [ ] **Step 1: README 安装段**

把 README 中:
```
**AI Agent Skills**(让智能体直接会用):`cp -r skills/openydt-* ~/.claude/skills/`(或你的 Agent 技能目录)。
```
替换为:
```
**AI Agent Skills**(让智能体直接会用):`npm i -g @openydt/openydt-cli` 会自动把技能同步到本机已装的各 AI agent(Claude Code / Codex / Cursor / Gemini CLI / OpenCode 等,经 `npx skills`)。手动同步:`openydt skill sync`(或 `npx skills add xiaowen-0725/openydt-cli -g -y`)。关闭自动同步:设 `OPENYDT_NO_SKILLS_SYNC=1`。
```

- [ ] **Step 2: CLAUDE.md 新增节**

在 CLAUDE.md 的「## 网关 / 客户端」节之前(或文件末尾「## 安全」之后)插入:
```markdown
## 技能同步(skillsync)
- 真相源 = 仓库 `skills/openydt-*/SKILL.md`(11 个);分发交给外部 `npx skills`(vercel-labs 包管理器,自动探测本机已装 agent)。
- **主路径**:`npm i/update -g @openydt/openydt-cli` 的 postinstall 跑 `npx skills add xiaowen-0725/openydt-cli -g -y`(best-effort,失败不阻断安装),并写 `~/.config/openydt-cli/skills-state.json`。
- **兜底**:Go 二进制每条普通命令前 `skillsync.MaybeTrigger` 本地比对版本;漂移则 detached 后台跑 `openydt skill sync --quiet`(输出进 `skills-sync.log`,每版本只 fork 一次,不碰 stdout)。
- **手动**:`openydt skill sync [--force]`。
- **opt-out**:`OPENYDT_NO_SKILLS_SYNC=1`;CI / DEV / 非 release 版本自动跳过。
- `internal/skillsync/*` 有单测锚定;真实 `npx` 不进单测(用 override 隔离)。
```

- [ ] **Step 3: PROJECT_STATUS 待办**

把 PROJECT_STATUS.md「## 8. 待办 / 下一步」中关于扩展接口的项不动,在列表里把"子 Agent 平台对齐"作为已完成补一行(放在第 1 项之后):
```
2. **子 Agent 平台对齐(已落地)**:接入 `npx skills` 包管理器 —— npm 安装/更新自动同步技能到本机各 agent + 二进制后台兜底 + `openydt skill sync` 手动命令。详见 `SKILL_SYNC_DESIGN.md`。
```

- [ ] **Step 4: 提交**

```bash
git add README.md CLAUDE.md PROJECT_STATUS.md
git commit -m "docs: 技能同步说明(README/CLAUDE/PROJECT_STATUS)"
```

---

## Task 12: 全量验证

- [ ] **Step 1: 全量测试 + vet**

Run:
```bash
go test ./... && go vet ./...
```
Expected: 全绿,无 vet 警告。

- [ ] **Step 2: 跨平台编译**

Run:
```bash
GOOS=windows GOARCH=amd64 go build ./... && echo "windows build ok"
GOOS=linux GOARCH=arm64 go build ./... && echo "linux build ok"
```
Expected: 两行 ok。

- [ ] **Step 3: 技能格式校验未受影响**

Run: `node scripts/skill-format-check/index.js`
Expected: 全 PASS(本次不改 SKILL.md)。

- [ ] **Step 4:(可选)真机端到端冒烟**

Run(会真跑 npx,联网):
```bash
make build
./bin/openydt skill sync
cat ~/.config/openydt-cli/skills-state.json
```
Expected: 打印 `✓ skills 已同步`;state 文件含当前版本。

---

## Self-Review(已执行)

- **Spec 覆盖**:① postinstall → Task 10;② 二进制后台兜底 → Task 7+9;③ 手动命令 → Task 8;state 模型 → Task 2;skip/opt-out → Task 3;跨平台 detached → Task 6;降级 notice → Task 5+9;文档 → Task 11。spec §4「output.go 注入」改为 Task 9 的 stderr 输出(已在 plan 顶部标注细化)。
- **占位符**:无 TBD/TODO;每个代码步骤含完整代码与精确命令。
- **类型一致**:`State{Version,LastAttemptVersion,Skills,UpdatedAt}`、`RunResult{Stdout,Stderr,Err}`、`RunSync(force)`、`Source()`、`Pending()/setPending`、`MaybeTrigger`/`spawnFunc`/`npxAvailableFunc`/`nowFunc`、`detachSysProcAttr()`、`SetRunnerOverrideForTest` 在各 Task 间名称统一;`cmdutil.ExitError{Code,Err}` 与 `output.ExitAPIError` 与现有代码一致。
- **跨包测试 seam**:`cmd/skill` 测试需要的 override 通过 Task 8 Step 2 的 `SetRunnerOverrideForTest` 暴露,已闭环。
