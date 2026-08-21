// Package selfupdate checks and installs new openydt CLI releases.
package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	PackageName       = "@openydt/openydt-cli"
	RegistryLatestURL = "https://registry.npmjs.org/@openydt%2Fopenydt-cli/latest"
)

var versionPattern = regexp.MustCompile(`^[vV]?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?$`)

// Result is the machine-readable outcome of an update check.
type Result struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	CheckedAt       string `json:"checkedAt,omitempty"`
}

// Checker queries the npm latest endpoint. URL and Client are configurable for tests.
type Checker struct {
	URL    string
	Client *http.Client
}

// DefaultChecker returns the production npm registry checker.
func DefaultChecker() Checker {
	return Checker{
		URL:    RegistryLatestURL,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Check gets the latest published npm version and compares it with current.
func (c Checker) Check(ctx context.Context, current string) (Result, error) {
	if c.URL == "" {
		c.URL = RegistryLatestURL
	}
	if c.Client == nil {
		c.Client = &http.Client{Timeout: 10 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "openydt-cli/"+strings.TrimSpace(current))
	resp, err := c.Client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("检查 npm 最新版本失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("检查 npm 最新版本失败: HTTP %d", resp.StatusCode)
	}
	var payload struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Result{}, fmt.Errorf("解析 npm 版本响应失败: %w", err)
	}
	latest := normalizeVersion(payload.Version)
	if latest == "" {
		return Result{}, fmt.Errorf("npm 版本响应缺少 version")
	}
	return Result{
		CurrentVersion:  normalizeVersion(current),
		LatestVersion:   latest,
		UpdateAvailable: IsNewer(latest, current),
		CheckedAt:       time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// IsNewer reports whether latest is strictly newer than current.
func IsNewer(latest, current string) bool {
	l, lok := parseVersion(latest)
	c, cok := parseVersion(current)
	if !lok {
		return false
	}
	if !cok {
		return normalizeVersion(latest) != normalizeVersion(current)
	}
	for i := range l.numbers {
		if l.numbers[i] != c.numbers[i] {
			return l.numbers[i] > c.numbers[i]
		}
	}
	if l.prerelease == c.prerelease {
		return false
	}
	if l.prerelease == "" {
		return true // stable is newer than a prerelease with equal numbers
	}
	if c.prerelease == "" {
		return false
	}
	return l.prerelease > c.prerelease
}

type parsedVersion struct {
	numbers    [3]int
	prerelease string
}

func parseVersion(value string) (parsedVersion, bool) {
	match := versionPattern.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		return parsedVersion{}, false
	}
	var parsed parsedVersion
	for i := range parsed.numbers {
		n, err := strconv.Atoi(match[i+1])
		if err != nil {
			return parsedVersion{}, false
		}
		parsed.numbers[i] = n
	}
	parsed.prerelease = match[4]
	return parsed, true
}

func normalizeVersion(value string) string {
	return strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(value), "v"), "V")
}

func isReleaseVersion(value string) bool {
	_, ok := parseVersion(value)
	return ok
}
