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
	Version            string   `json:"version"`              // last SUCCESSFUL sync version
	LastAttemptVersion string   `json:"last_attempt_version"` // last attempted version (debounce)
	Skills             []string `json:"skills,omitempty"`     // informational
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
