package main

import (
	"errors"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/tools/refactor/importgraph"
)

// CopyGopath creates a new Gopath with a copy of a package
// and all of its dependencies.
func CopyGopath(packageName, newGopath string) bool {
	ctx := build.Default
	if ctx.GOPATH == "" {
		fmt.Fprintln(os.Stderr, "GOPATH not set.")
	}
	forward, _, errs := importgraph.Build(&ctx)
	allDeps := forward.Search(packageName)
	if len(allDeps) == 0 {
		fmt.Fprintln(os.Stderr, "Failed to build import graph:", packageName)
		if err, ok := errs[packageName]; ok {
			fmt.Fprintln(os.Stderr, " -> Error for package:", err)
		}
		return false
	}

	for dep := range allDeps {
		err := copyDep(dep, ctx.GOPATH, newGopath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to copy %s: %s\n", dep, err)
			return false
		}
	}
	return true
}

func copyDep(packagePath, oldGopath, newGopath string) error {
	oldPath := filepath.Join(oldGopath, "src", packagePath)
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil
	}
	newPath := filepath.Join(newGopath, "src", packagePath)
	if _, err := os.Stat(newPath); err == nil {
		os.RemoveAll(newPath)
	}
	createDir(newPath)
	return filepath.Walk(oldPath, func(source string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base, err := filepath.Rel(oldGopath, source)
		if err != nil {
			return err
		}
		newPath := filepath.Join(newGopath, base)
		if info.IsDir() {
			return createDir(newPath)
		} else {
			newFile, err := os.Create(newPath)
			if err != nil {
				return err
			}
			defer newFile.Close()
			oldFile, err := os.Open(source)
			if err != nil {
				return err
			}
			defer oldFile.Close()
			_, err = io.Copy(newFile, oldFile)
			return err
		}
	})
}

func createDir(dir string) error {
	if info, err := os.Stat(dir); err == nil {
		if info.IsDir() {
			return nil
		} else {
			return errors.New("file already exists: " + dir)
		}
	}
	if filepath.Dir(dir) != dir {
		parent := filepath.Dir(dir)
		if err := createDir(parent); err != nil {
			return err
		}
	}
	return os.Mkdir(dir, 0755)
}
