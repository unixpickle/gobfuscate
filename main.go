package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
)

func main() {
	var encKey string
	var outputGopath bool

	flag.StringVar(&encKey, "enckey", "", "rename encryption key")
	flag.BoolVar(&outputGopath, "outdir", false, "output a full GOPATH")

	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: gobfuscate [flags] pkg_name out_path")
		flag.PrintDefaults()
		os.Exit(1)
	}

	pkgName := flag.Args()[0]
	outPath := flag.Args()[1]

	if encKey == "" {
		buf := make([]byte, 32)
		rand.Read(buf)
		encKey = string(buf)
	}

	if !obfuscate(outputGopath, encKey, pkgName, outPath) {
		os.Exit(1)
	}
}

func obfuscate(outGopath bool, encKey, pkgName, outPath string) bool {
	var newGopath string
	if outGopath {
		newGopath = outPath
		os.Mkdir(newGopath, 0755)
		os.Mkdir(filepath.Join(newGopath, "src"), 0755)
	} else {
		var err error
		newGopath, err = ioutil.TempDir("", "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create temp dir:", err)
			return false
		}
		defer os.RemoveAll(newGopath)
	}

	if !CopyGopath(pkgName, newGopath) {
		return false
	}

	enc := &Encrypter{Key: encKey}
	if err := ObfuscatePackageNames(newGopath, enc); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to obfuscate package names:", err)
		return false
	}
	if err := ObfuscateSymbols(newGopath, enc); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to obfuscate symbols:", err)
		return false
	}

	// TODO: compile source here if requested.

	return true
}
