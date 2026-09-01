package repositories

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafePathRejectsTraversalAndSymlinkEscape(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	inside := filepath.Join(root, "repo")
	if err := os.Mkdir(inside, 0o700); err != nil {
		t.Fatal(err)
	}
	if got, err := safePath(root, inside); err != nil || got != inside {
		t.Fatalf("inside got=%q err=%v", got, err)
	}
	if _, err := safePath(root, filepath.Join(root, "..", "escape")); err == nil {
		t.Fatal("expected traversal rejection")
	}
}
