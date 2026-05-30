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
