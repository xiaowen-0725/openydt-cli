package selfupdate

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

const (
	envOptOut     = "OPENYDT_NO_UPDATE_CHECK"
	envChild      = "OPENYDT_UPDATE_CHECK_CHILD"
	checkInterval = 24 * time.Hour
	retryInterval = time.Hour
	timeFormat    = time.RFC3339
	updateLogName = "update-check.log"
)

var (
	nowFunc   = time.Now
	spawnFunc = spawnBackgroundCheck
)

// MaybeTrigger surfaces a cached update once and starts a detached stale check.
// It performs no network request in the calling command.
func MaybeTrigger(currentVersion string) {
	setPending(nil)
	if shouldSkip(currentVersion) {
		return
	}
	state, err := ReadState()
	if err != nil {
		state = nil
	}
	if state == nil {
		state = &State{}
	}

	latest := normalizeVersion(state.LatestVersion)
	current := normalizeVersion(currentVersion)
	if IsNewer(latest, current) && normalizeVersion(state.NotifiedVersion) != latest {
		setPending(&Notice{CurrentVersion: current, LatestVersion: latest})
		state.NotifiedVersion = latest
		_ = WriteState(*state)
	}

	if isRecent(state.CheckedAt, checkInterval) || isRecent(state.LastAttemptAt, retryInterval) {
		return
	}
	state.LastAttemptAt = nowFunc().UTC().Format(timeFormat)
	_ = WriteState(*state)
	_ = spawnFunc()
}

func shouldSkip(version string) bool {
	if os.Getenv(envOptOut) != "" || os.Getenv(envChild) != "" || !isReleaseVersion(version) {
		return true
	}
	for _, key := range []string{"CI", "GITHUB_ACTIONS", "BUILD_NUMBER", "GITLAB_CI"} {
		if os.Getenv(key) != "" {
			return true
		}
	}
	return false
}

func isRecent(value string, window time.Duration) bool {
	checked, err := time.Parse(timeFormat, value)
	if err != nil {
		return false
	}
	age := nowFunc().Sub(checked)
	return age >= 0 && age < window
}

func spawnBackgroundCheck() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	dir, err := config.Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	logFile, err := os.OpenFile(filepath.Join(dir, updateLogName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	cmd := exec.Command(self, "update", "check", "--quiet")
	cmd.Env = append(os.Environ(), envChild+"=1")
	cmd.Stdin = nil
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = detachSysProcAttr()
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return err
	}
	return logFile.Close()
}
