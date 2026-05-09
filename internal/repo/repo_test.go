package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDiscoverInsideRepoReturnsContainingRepo(t *testing.T) {
	root := t.TempDir()
	initGit(t, root)

	subdir := filepath.Join(root, "subdir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}
	file := filepath.Join(subdir, "file.txt")
	if err := os.WriteFile(file, []byte("content"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	repositories, err := Discover(file)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertRepositories(t, repositories, []Repository{
		{Root: root, Name: filepath.Base(root)},
	})
}

func TestDiscoverOutsideRepoFindsChildRepos(t *testing.T) {
	base := t.TempDir()
	repoA := filepath.Join(base, "a")
	repoB := filepath.Join(base, "b")
	initGit(t, repoA)
	initGit(t, repoB)

	repositories, err := Discover(base)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertRepositories(t, repositories, []Repository{
		{Root: repoA, Name: "a"},
		{Root: repoB, Name: "b"},
	})
}

func TestDiscoverOutsideRepoFindsChildRepoWithGitFile(t *testing.T) {
	base := t.TempDir()
	repoRoot := filepath.Join(base, "worktree")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("mkdir repo root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, ".git"), []byte("gitdir: ../.git/worktrees/worktree\n"), 0o644); err != nil {
		t.Fatalf("write .git file: %v", err)
	}

	repositories, err := Discover(base)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertRepositories(t, repositories, []Repository{
		{Root: repoRoot, Name: "worktree"},
	})
}

func TestDiscoverOutsideRepoWithGitFileDoesNotIncludeNestedRepo(t *testing.T) {
	base := t.TempDir()
	repoRoot := filepath.Join(base, "worktree")
	nested := filepath.Join(repoRoot, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested repo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, ".git"), []byte("gitdir: ../.git/worktrees/worktree\n"), 0o644); err != nil {
		t.Fatalf("write .git file: %v", err)
	}
	initGit(t, nested)

	repositories, err := Discover(base)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertRepositories(t, repositories, []Repository{
		{Root: repoRoot, Name: "worktree"},
	})
}

func TestDiscoverOutsideRepoWithGitDirectoryDoesNotIncludeNestedRepo(t *testing.T) {
	base := t.TempDir()
	repoRoot := filepath.Join(base, "repo")
	nested := filepath.Join(repoRoot, "nested")
	initGit(t, repoRoot)
	initGit(t, nested)

	repositories, err := Discover(base)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertRepositories(t, repositories, []Repository{
		{Root: repoRoot, Name: "repo"},
	})
}

func TestDiscoverOutsideRepoSkipsVendorRepo(t *testing.T) {
	base := t.TempDir()
	initGit(t, filepath.Join(base, "vendor"))

	repositories, err := Discover(base)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertRepositories(t, repositories, nil)
}

func TestDiscoverOutsideRepoSkipsRepoUnderNodeModules(t *testing.T) {
	base := t.TempDir()
	initGit(t, filepath.Join(base, "node_modules", "pkg"))

	repositories, err := Discover(base)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertRepositories(t, repositories, nil)
}

func TestDiscoverInsideRepoDoesNotIncludeNestedRepo(t *testing.T) {
	root := t.TempDir()
	initGit(t, root)
	nested := filepath.Join(root, "submodule")
	initGit(t, nested)

	repositories, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertRepositories(t, repositories, []Repository{
		{Root: root, Name: filepath.Base(root)},
	})
}

func initGit(t *testing.T, dir string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir git dir: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git init: %v\n%s", err, output)
	}
}

func assertRepositories(t *testing.T, got, want []Repository) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(repositories) = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("repositories[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}
