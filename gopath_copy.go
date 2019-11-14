package main

import (
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/tools/refactor/importgraph"
)

// CopyGopath creates a new Gopath with a copy of a package
// and all of its dependencies.
func CopyGopath(packageName, newGopath string, keepTests bool) error {
	ctx := build.Default

	rootPkg, err := ctx.Import(packageName, "", 0)
	if err != nil {
		return err
	}

	allDeps, err := findDeps(packageName, &ctx)
	if err != nil {
		return err
	}

	for dep := range allDeps {
		pkg, err := build.Default.Import(dep, rootPkg.Dir, 0)
		if err != nil {
			return err
		}
		if pkg.Goroot {
			continue
		}
		if err := copyDep(pkg, newGopath, keepTests); err != nil {
			return err
		}
	}

	if !keepTests {
		ctx.GOPATH = newGopath
		allDeps, err = findDeps(packageName, &ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func findDeps(packageName string, ctx *build.Context) (map[string]bool, error) {
	forward, _, errs := importgraph.Build(ctx)
	if _, ok := forward[packageName]; !ok {
		if err, ok := errs[packageName]; ok {
			return nil, err
		}
		return nil, fmt.Errorf("package %s not found", packageName)
	}
	return forward.Search(packageName), nil
}

func copyDep(pkg *build.Package, newGopath string, keepTests bool) error {
	newPath := filepath.Join(newGopath, "src", pkg.ImportPath)
	err := os.MkdirAll(newPath, 0755)
	if err != nil {
		return err
	}

	srcFiles := [][]string{
		pkg.GoFiles,
		pkg.CgoFiles,
		pkg.CFiles,
		pkg.CXXFiles,
		pkg.MFiles,
		pkg.HFiles,
		pkg.FFiles,
		pkg.SFiles,
		pkg.SwigFiles,
		pkg.SwigCXXFiles,
		pkg.SysoFiles,
	}
	if keepTests {
		srcFiles = append(srcFiles, pkg.TestGoFiles, pkg.XTestGoFiles)
	}

	for _, list := range srcFiles {
		for _, file := range list {
			src := filepath.Join(pkg.Dir, file)
			dst := filepath.Join(newPath, file)
			if err := copyFile(src, dst); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	newFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer newFile.Close()
	oldFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer oldFile.Close()
	_, err = io.Copy(newFile, oldFile)
	return err
}
