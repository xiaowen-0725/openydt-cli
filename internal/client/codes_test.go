package client

import (
	"strings"
	"testing"
)

func TestResultNextCommands(t *testing.T) {
	if got := ResultNextCommands(904); len(got) == 0 || got[0] != "park get-auth-park-codes" {
		t.Fatalf("904 nextCommands=%v", got)
	}
	if got := ResultNextCommands(912); len(got) == 0 || got[0] != "trade get-park-fee" {
		t.Fatalf("912 nextCommands=%v", got)
	}
	if got := ResultNextCommands(907); len(got) != 0 {
		t.Fatalf("907 应无重发建议(幂等命中),got %v", got)
	}
}

func TestMessageHint(t *testing.T) {
	if h := MessageHint("会话已过期"); h == "" {
		t.Fatal("「会话已过期」应给出文案级提示")
	}
	if !strings.Contains(MessageHint("会话已过期"), "channel-snap") {
		t.Fatal("「会话已过期」提示应指向先 channel-snap")
	}
	if h := MessageHint("业务成功"); h != "" {
		t.Fatalf("无匹配文案应返回空, got %q", h)
	}
}

func TestRetryClass(t *testing.T) {
	if RetryClass(StatusSysError) != "server_indeterminate" {
		t.Fatal("status3 应 server_indeterminate")
	}
	if RetryClass(StatusSuccess) != "" {
		t.Fatal("success 无 retry class")
	}
}
