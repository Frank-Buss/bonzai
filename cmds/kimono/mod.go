package kimono

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwxrob/bonzai/futil"
	"github.com/rwxrob/bonzai/run"
)

// Tidy runs `go get -u` and `go mod tidy` on all supported Go
// modules in the current git repository.
func TidyAll(root string) error {
	return filepath.WalkDir(root, sanitizeWalkDirFn)
}

func TidyDependents() error {
	deps, err := dependencyGraph()
	if err != nil {
		return err
	}
	root, err := futil.HereOrAbove(`.git`)
	if err != nil {
		return err
	}
	root = filepath.Dir(root)
	modName := strings.TrimSpace(run.Out("go", "list", "-m"))
	dependents := getDependents(deps, modName)
	for _, dep := range dependents {
		if dep.path == "" {
			continue
		}
		rel, err := filepath.Rel(root, dep.path)
		if err != nil {
			return err
		}
		fmt.Printf("\n%s:\n",rel)
		os.Chdir(dep.path)
		update()
		tidy()
	}
	return nil
}

func TidyDependencies() error {
	deps, err := dependencyGraph()
	if err != nil {
		return err
	}
	root, err := futil.HereOrAbove(`.git`)
	if err != nil {
		return err
	}
	root = filepath.Dir(root)
	modName := strings.TrimSpace(run.Out("go", "list", "-m"))
	node := deps.nodes[modName]
	for _, dep := range node.dependencies {
		if dep.path == "" {
			continue
		}
		rel, err := filepath.Rel(root, dep.path)
		if err != nil {
			return err
		}
		fmt.Printf("\n%s:\n",rel)
		os.Chdir(dep.path)
		update()
		tidy()
	}
	return nil
}

func sanitizeWalkDirFn(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() {
		return nil
	}
	if d.Name() == ".git" || d.Name() == "vendor" {
		return filepath.SkipDir
	}
	if !futil.Exists(filepath.Join(path, "go.mod")) {
		return filepath.SkipDir
	}
	if err := os.Chdir(path); err != nil {
		return err
	}
	if !hasDependencies() {
		return filepath.SkipDir
	}
	fmt.Printf("\n%s:\n", path)
	_ = update()
	_ = tidy()
	return nil
}

func hasDependencies() bool {
	out := run.Out(`go`, `list`, `-m`, `all`)
	return len(strings.Split(out, "\n")) > 1
}

func update() error {
	return run.Exec("go", "get", "-u")
}

func tidy() error {
	return run.Exec("go", "mod", "tidy")
}
