package skillsync

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

const (
	logFileName = "skills-sync.log"
	retryWindow = 6 * time.Hour
)

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
	if state != nil && normalizeVersion(state.LastAttemptVersion) == normalizeVersion(version) && !attemptStale(state) {
		return // attempted recently at this version (debounce); stale attempts self-heal
	}
	if !npxAvailableFunc() {
		if state == nil || normalizeVersion(state.NoticedVersion) != normalizeVersion(version) {
			setPending(&Notice{Reason: "未检测到 npx/node"})
			recordNoticed(state, version)
		}
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

// attemptStale reports whether the last attempt is old enough to retry, so a
// transient sync failure self-heals instead of parking auto-sync until the next
// version bump. A missing/unparseable timestamp counts as stale.
func attemptStale(s *State) bool {
	if s == nil || s.UpdatedAt == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339, s.UpdatedAt)
	if err != nil {
		return true
	}
	return nowFunc().Sub(t) > retryWindow
}

// recordNoticed persists that the degraded (npx-missing) notice was shown for
// this version, so it is not repeated on every invocation.
func recordNoticed(prev *State, version string) {
	s := State{}
	if prev != nil {
		s = *prev
	}
	s.NoticedVersion = version
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
