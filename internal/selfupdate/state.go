package selfupdate

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/xiaowen-0725/openydt-cli/internal/config"
)

const stateFile = "update-state.json"

// State caches update checks so ordinary CLI commands never wait on the network.
type State struct {
	LatestVersion   string `json:"latest_version,omitempty"`
	CheckedAt       string `json:"checked_at,omitempty"`
	LastAttemptAt   string `json:"last_attempt_at,omitempty"`
	NotifiedVersion string `json:"notified_version,omitempty"`
}

func statePath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, stateFile), nil
}

// ReadState reads the cached update state. A missing file is not an error.
func ReadState() (*State, error) {
	path, err := statePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// WriteState atomically persists the update state.
func WriteState(state State) error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := filepath.Join(dir, stateFile)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// RecordResult stores a successful registry check while preserving notification state.
func RecordResult(result Result) error {
	state, _ := ReadState()
	if state == nil {
		state = &State{}
	}
	state.LatestVersion = normalizeVersion(result.LatestVersion)
	state.CheckedAt = result.CheckedAt
	if state.CheckedAt == "" {
		state.CheckedAt = nowFunc().UTC().Format(timeFormat)
	}
	return WriteState(*state)
}
