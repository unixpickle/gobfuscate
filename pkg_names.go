package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/tools/refactor/rename"
)

func ObfuscatePackageNames(gopath string, enc *Encrypter) error {
	ctx := build.Default
	ctx.GOPATH = gopath

	level := 1
	srcDir := filepath.Join(gopath, "src")

	doneChan := make(chan struct{})
	defer close(doneChan)

	for {
		resChan := make(chan string)
		go func() {
			scanLevel(srcDir, level, resChan, doneChan)
			close(resChan)
		}()
		var gotAny bool
		for dirPath := range resChan {
			gotAny = true
			encPath := encryptPackageName(dirPath, enc)
			srcPkg, err := filepath.Rel(srcDir, dirPath)
			if err != nil {
				return err
			}
			dstPkg, err := filepath.Rel(srcDir, encPath)
			if err != nil {
				return err
			}
			if err := rename.Move(&ctx, srcPkg, dstPkg, ""); err != nil {
				return fmt.Errorf("package move: %s", err)
			}
		}
		if !gotAny {
			break
		}
		level++
	}

	return nil
}

func scanLevel(dir string, depth int, res chan<- string, done <-chan struct{}) {
	if depth == 0 {
		select {
		case res <- dir:
		case <-done:
			return
		}
		return
	}
	listing, _ := ioutil.ReadDir(dir)
	for _, item := range listing {
		if item.IsDir() {
			scanLevel(filepath.Join(dir, item.Name()), depth-1, res, done)
		}
		select {
		case <-done:
			return
		default:
		}
	}
}

func encryptPackageName(dir string, enc *Encrypter) string {
	subDir, base := filepath.Split(dir)
	return filepath.Join(subDir, enc.Encrypt(base))
}
