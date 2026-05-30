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
