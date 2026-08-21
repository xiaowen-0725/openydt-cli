package cmd

import (
	"bytes"
	"testing"

	"github.com/xiaowen-0725/openydt-cli/internal/cmdutil"
)

func TestRootHasSkillCommand(t *testing.T) {
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	root := NewRootCmd(f)
	sub, _, err := root.Find([]string{"skill", "sync"})
	if err != nil || sub.Use != "sync" {
		t.Fatalf("expected root -> skill -> sync; err=%v", err)
	}
}

func TestRootHasUpdateCommand(t *testing.T) {
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	root := NewRootCmd(f)
	sub, _, err := root.Find([]string{"update", "check"})
	if err != nil || sub.Use != "check" {
		t.Fatalf("expected root -> update -> check; err=%v", err)
	}
}

func TestRootHasAutoSyncPreRun(t *testing.T) {
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	root := NewRootCmd(f)
	if root.PersistentPreRunE == nil {
		t.Fatalf("root must wire PersistentPreRunE for skill auto-sync")
	}
}

func TestRootHasBulkPaginationFlags(t *testing.T) {
	f := &cmdutil.Factory{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	root := NewRootCmd(f)
	for _, name := range []string{"all-pages", "out"} {
		if root.PersistentFlags().Lookup(name) == nil {
			t.Fatalf("missing persistent flag --%s", name)
		}
	}
}
