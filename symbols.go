package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"golang.org/x/tools/refactor/importgraph"
	"golang.org/x/tools/refactor/rename"
)

const MainMethodName = "main"

type symbolRenameReq struct {
	OldName string
	NewName string
}

func ObfuscateSymbols(gopath string, enc *Encrypter) error {
	renames, err := topLevelRenames(gopath, enc)
	if err != nil {
		return fmt.Errorf("top-level renames: %s", err)
	}
	if err := runRenames(gopath, renames); err != nil {
		return fmt.Errorf("top-level renaming: %s", err)
	}
	renames, err = methodRenames(gopath, enc)
	if err != nil {
		return fmt.Errorf("method renames: %s", err)
	}
	if err := runRenames(gopath, renames); err != nil {
		return fmt.Errorf("method renaming: %s", err)
	}
	return nil
}

func runRenames(gopath string, renames []symbolRenameReq) error {
	ctx := build.Default
	ctx.GOPATH = gopath
	for _, r := range renames {
		if err := rename.Main(&ctx, "", r.OldName, r.NewName); err != nil {
			return err
		}
	}
	return nil
}

func topLevelRenames(gopath string, enc *Encrypter) ([]symbolRenameReq, error) {
	srcDir := filepath.Join(gopath, "src")
	var res []symbolRenameReq
	addRes := func(pkgPath, name string) {
		prefix := "\"" + pkgPath + "\"."
		oldName := prefix + name
		newName := enc.Encrypt(name)
		res = append(res, symbolRenameReq{oldName, newName})
	}
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) != GoExtension {
			return nil
		}
		pkgPath, err := filepath.Rel(srcDir, filepath.Dir(path))
		if err != nil {
			return err
		}
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.Name != MainMethodName && d.Recv == nil {
					addRes(pkgPath, d.Name.Name)
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						addRes(pkgPath, spec.Name.Name)
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							addRes(pkgPath, name.Name)
						}
					}
				}
			}
		}
		return nil
	})
	return res, err
}

func methodRenames(gopath string, enc *Encrypter) ([]symbolRenameReq, error) {
	exclude, err := interfaceMethods(gopath)
	if err != nil {
		return nil, err
	}

	srcDir := filepath.Join(gopath, "src")
	var res []symbolRenameReq
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) != GoExtension {
			return nil
		}
		pkgPath, err := filepath.Rel(srcDir, filepath.Dir(path))
		if err != nil {
			return err
		}
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			d, ok := decl.(*ast.FuncDecl)
			if !ok || exclude[d.Name.Name] || d.Recv == nil {
				continue
			}
			prefix := "\"" + pkgPath + "\"."
			for _, rec := range d.Recv.List {
				s, ok := rec.Type.(fmt.Stringer)
				if !ok {
					continue
				}
				oldName := prefix + s.String() + "." + d.Name.Name
				newName := enc.Encrypt(d.Name.Name)
				res = append(res, symbolRenameReq{oldName, newName})
			}
		}
		return nil
	})
	return res, err
}

func interfaceMethods(gopath string) (map[string]bool, error) {
	ctx := build.Default
	ctx.GOPATH = gopath
	forward, backward, _ := importgraph.Build(&ctx)
	pkgs := map[string]bool{}
	for _, m := range []importgraph.Graph{forward, backward} {
		for x := range m {
			pkgs[x] = true
		}
	}
	res := map[string]bool{}
	for pkgName := range pkgs {
		pkg, err := ctx.Import(pkgName, gopath, 0)
		if err != nil {
			return nil, fmt.Errorf("import %s: %s", pkgName, err)
		}
		for _, fileName := range pkg.GoFiles {
			sourcePath := filepath.Join(pkg.Dir, fileName)
			set := token.NewFileSet()
			file, err := parser.ParseFile(set, sourcePath, nil, 0)
			if err != nil {
				return nil, err
			}
			for _, decl := range file.Decls {
				d, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range d.Specs {
					spec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					t, ok := spec.Type.(*ast.InterfaceType)
					if !ok {
						continue
					}
					for _, field := range t.Methods.List {
						for _, name := range field.Names {
							res[name.Name] = true
						}
					}
				}
			}
		}
	}
	return res, nil
}
