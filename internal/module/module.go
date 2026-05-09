package module

import (
	"os"
	"path/filepath"
	"sort"
)

type Module struct {
	Root    string
	Display string
}

func Discover(repoRoot string) ([]Module, error) {
	var modules []Module
	repoRoot = filepath.Clean(repoRoot)
	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != repoRoot {
			switch d.Name() {
			case ".git", "vendor", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "go.mod" {
			return nil
		}

		root := filepath.Dir(path)
		display, err := filepath.Rel(repoRoot, root)
		if err != nil {
			return err
		}
		if display == "." || display == "" {
			display = "."
		}
		modules = append(modules, Module{Root: root, Display: filepath.ToSlash(display)})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(modules, func(i, j int) bool { return modules[i].Root < modules[j].Root })
	return modules, nil
}
