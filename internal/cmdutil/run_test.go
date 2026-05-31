package cmdutil

import "testing"

func TestWriteGuard(t *testing.T) {
	cases := []struct {
		name                    string
		write, yes, dry, ro bool
		wantErr                 bool
		wantReadOnly            bool // err 是只读拒绝
	}{
		{"read passes", false, false, false, false, false, false},
		{"write no-yes blocked", true, false, false, false, true, false},
		{"write yes ok", true, true, false, false, false, false},
		{"write dry-run ok", true, false, true, false, false, false},
		{"readonly blocks write even with yes", true, true, false, true, true, true},
		{"readonly allows read", false, false, false, true, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := writeGuard("someCmd", c.write, c.yes, c.dry, c.ro)
			if (err != nil) != c.wantErr {
				t.Fatalf("err=%v want %v", err, c.wantErr)
			}
			if c.wantReadOnly && err != nil && !isReadOnlyErr(err) {
				t.Fatalf("expected read-only error, got %v", err)
			}
		})
	}
}
