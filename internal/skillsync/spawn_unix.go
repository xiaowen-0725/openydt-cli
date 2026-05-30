//go:build !windows

package skillsync

import "syscall"

// detachSysProcAttr makes the child its own session leader so it survives the
// parent CLI process exiting.
func detachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
