package cmdutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveBody(t *testing.T) {
	// 1) 都为空 -> ""
	if got, err := ResolveBody("", ""); err != nil || got != "" {
		t.Fatalf("空入参应返回空 base, got %q err=%v", got, err)
	}
	// 2) 仅 --body -> 原样返回
	if got, err := ResolveBody(`{"a":1}`, ""); err != nil || got != `{"a":1}` {
		t.Fatalf("仅 --body 应原样返回, got %q err=%v", got, err)
	}
	// 3) --body 与 --body-file 互斥
	if _, err := ResolveBody(`{"a":1}`, "-"); err == nil {
		t.Fatal("同时给 --body 与 --body-file 应报错")
	}
	// 4) 从文件读取并 trim 首尾空白
	dir := t.TempDir()
	p := filepath.Join(dir, "body.json")
	if err := os.WriteFile(p, []byte("  {\"k\":\"v\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := ResolveBody("", p); err != nil || got != `{"k":"v"}` {
		t.Fatalf("文件读取应 trim 后返回, got %q err=%v", got, err)
	}
	// 5) 文件不存在 -> 错误
	if _, err := ResolveBody("", filepath.Join(dir, "nope.json")); err == nil {
		t.Fatal("不存在的 body-file 应报错")
	}
}
