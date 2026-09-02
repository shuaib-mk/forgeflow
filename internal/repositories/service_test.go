package repositories

import (
	"context"
	"os"
	"os/exec"
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

func TestInitializeRepositoryCreatesManagedCheckout(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	root := t.TempDir()
	if err := initializeRepository(context.Background(), root, filepath.Join("team", "project")); err != nil {
		t.Fatal(err)
	}
	repository := filepath.Join(root, "team", "project")
	if _, err := os.Stat(filepath.Join(repository, ".git")); err != nil {
		t.Fatalf("managed repository was not initialized: %v", err)
	}
	branch, err := gitOutput(context.Background(), repository, "symbolic-ref", "--short", "HEAD")
	if err != nil || branch != "main\n" {
		t.Fatalf("branch=%q err=%v", branch, err)
	}
}
