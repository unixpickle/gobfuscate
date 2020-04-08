package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Command line arguments.
var (
	customPadding       string
	tags                string
	outputGopath        bool
	keepTests           bool
	winHide             bool
	noStaticLink        bool
	preservePackageName bool
	verbose             bool
)

func main() {
	flag.StringVar(&customPadding, "padding", "", "use a custom padding for hashing sensitive information (otherwise a random padding will be used)")
	flag.BoolVar(&outputGopath, "outdir", false, "output a full GOPATH")
	flag.BoolVar(&keepTests, "keeptests", false, "keep _test.go files")
	flag.BoolVar(&winHide, "winhide", false, "hide windows GUI")
	flag.BoolVar(&noStaticLink, "nostatic", false, "do not statically link")
	flag.BoolVar(&preservePackageName, "noencrypt", false,
		"no encrypted package name for go build command (works when main package has CGO code)")
	flag.BoolVar(&verbose, "verbose", false, "verbose mode")
	flag.StringVar(&tags, "tags", "", "tags are passed to the go compiler")

	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: gobfuscate [flags] pkg_name out_path")
		flag.PrintDefaults()
		os.Exit(1)
	}

	pkgName := flag.Args()[0]
	outPath := flag.Args()[1]

	if !obfuscate(pkgName, outPath) {
		os.Exit(1)
	}
}

func obfuscate(pkgName, outPath string) bool {
	var newGopath string
	if outputGopath {
		newGopath = outPath
		if err := os.Mkdir(newGopath, 0755); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create destination:", err)
			return false
		}
	} else {
		var err error
		newGopath, err = ioutil.TempDir("", "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create temp dir:", err)
			return false
		}
		defer os.RemoveAll(newGopath)
	}

	log.Println("Copying GOPATH...")

	if err := CopyGopath(pkgName, newGopath, keepTests); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to copy into a new GOPATH:", err)
		return false
	}
	var n NameHasher
	if customPadding == "" {
		buf := make([]byte, 32)
		rand.Read(buf)
		n = buf
	} else {
		n = []byte(customPadding)
	}

	log.Println("Obfuscating package names...")
	if err := ObfuscatePackageNames(newGopath, n); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to obfuscate package names:", err)
		return false
	}
	log.Println("Obfuscating strings...")
	if err := ObfuscateStrings(newGopath); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to obfuscate strings:", err)
		return false
	}
	log.Println("Obfuscating symbols...")
	if err := ObfuscateSymbols(newGopath, n); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to obfuscate symbols:", err)
		return false
	}

	if outputGopath {
		return true
	}

	ctx := build.Default

	newPkg := pkgName
	if !preservePackageName {
		newPkg = encryptComponents(pkgName, n)
	}

	ldflags := `-s -w`
	if winHide {
		ldflags += " -H=windowsgui"
	}
	if !noStaticLink {
		ldflags += ` -extldflags '-static'`
	}

	goCache := newGopath + "/cache"
	os.Mkdir(goCache, 0755)

	arguments := []string{"build", "-ldflags", ldflags, "-tags", tags, "-o", outPath, newPkg}
	environment := []string{
		"GOROOT=" + ctx.GOROOT,
		"GOARCH=" + ctx.GOARCH,
		"GOOS=" + ctx.GOOS,
		"GOPATH=" + newGopath,
		"PATH=" + os.Getenv("PATH"),
		"GOCACHE=" + goCache,
	}

	cmd := exec.Command("go", arguments...)
	cmd.Env = environment
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if verbose {
		fmt.Println()
		fmt.Println("[Verbose] Temporary path:", newGopath)
		fmt.Println("[Verbose] Go build command: go", strings.Join(arguments, " "))
		fmt.Println("[Verbose] Environment variables:")
		for _, envLine := range environment {
			fmt.Println(envLine)
		}
		fmt.Println()
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to compile:", err)
		return false
	}

	return true
}

func encryptComponents(pkgName string, n NameHasher) string {
	comps := strings.Split(pkgName, "/")
	for i, comp := range comps {
		comps[i] = n.Hash(comp)
	}
	return strings.Join(comps, "/")
}
