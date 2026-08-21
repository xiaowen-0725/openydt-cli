package selfupdate

import (
	"fmt"
	"sync/atomic"
)

// Notice is shown after a normal command when a newer npm release is cached.
type Notice struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
}

// Message returns the concise one-time update hint.
func (n *Notice) Message() string {
	return fmt.Sprintf("有新版本 v%s(当前 v%s);运行 openydt update 更新 CLI 与 Skills", n.LatestVersion, n.CurrentVersion)
}

var pending atomic.Pointer[Notice]

func setPending(notice *Notice) { pending.Store(notice) }

// Pending returns this process's update notice, if any.
func Pending() *Notice { return pending.Load() }
