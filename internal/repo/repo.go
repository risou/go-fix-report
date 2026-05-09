package repo

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type Repository struct {
	Root string
	Name string
}

func Discover(path string) ([]Repository, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if root, ok, err := containingRepository(absolutePath); err != nil {
		return nil, err
	} else if ok {
		return []Repository{newRepository(root)}, nil
	}

	repositories, err := childRepositories(absolutePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	sort.Slice(repositories, func(i, j int) bool {
		return repositories[i].Root < repositories[j].Root
	})

	return repositories, nil
}

func containingRepository(path string) (string, bool, error) {
	current := path
	if info, err := os.Stat(current); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	} else if !info.IsDir() {
		current = filepath.Dir(current)
	}

	for {
		if info, err := os.Stat(filepath.Join(current, ".git")); err == nil && info.IsDir() {
			return current, true, nil
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", false, err
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false, nil
		}
		current = parent
	}
}

func childRepositories(path string) ([]Repository, error) {
	var repositories []Repository

	err := filepath.WalkDir(path, func(current string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}

		name := entry.Name()
		if name == ".git" {
			repositories = append(repositories, newRepository(filepath.Dir(current)))
			return filepath.SkipDir
		}
		if name == "vendor" || name == "node_modules" {
			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return repositories, nil
}

func newRepository(root string) Repository {
	return Repository{
		Root: root,
		Name: filepath.Base(root),
	}
}
