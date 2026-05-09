package module

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverModulesFindsRootAndNestedGoMods(t *testing.T) {
	repoRoot := t.TempDir()
	writeGoMod(t, repoRoot)
	nested := filepath.Join(repoRoot, "tools", "report")
	writeGoMod(t, nested)

	modules, err := Discover(repoRoot)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertModules(t, modules, []Module{
		{Root: repoRoot, Display: "."},
		{Root: nested, Display: "tools/report"},
	})
}

func TestDiscoverModulesSkipsExcludedDirectories(t *testing.T) {
	repoRoot := t.TempDir()
	writeGoMod(t, repoRoot)
	writeGoMod(t, filepath.Join(repoRoot, "vendor", "x"))
	writeGoMod(t, filepath.Join(repoRoot, ".git", "hooks"))
	writeGoMod(t, filepath.Join(repoRoot, "node_modules", "x"))

	modules, err := Discover(repoRoot)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertModules(t, modules, []Module{
		{Root: repoRoot, Display: "."},
	})
}

func TestDiscoverModulesReturnsEmptyWhenNoGoMod(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, "tools", "report"), 0o755); err != nil {
		t.Fatalf("mkdir nested dir: %v", err)
	}

	modules, err := Discover(repoRoot)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertModules(t, modules, nil)
}

func writeGoMod(t *testing.T, dir string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir module dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
}

func assertModules(t *testing.T, got, want []Module) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(modules) = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("modules[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}
