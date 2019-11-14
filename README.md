# gobfuscate

When you compile a Go binary, it contains a lot of information about your source code: field names, strings, package paths, etc. If you want to ship a binary without leaking this kind of information, what are you to do?

With gobfuscate, you can compile a Go binary from obfuscated source code. This makes a lot of information difficult or impossible to decipher from the binary.

# How to use
```
go get -u github.com/unixpickle/gobfuscate
gobfuscate [flags] pkg_name out_path
```
`pkg_name` is the path relative from your $GOPATH/src to the package to obfuscate (typically something like domain.tld/user/repo)

`out_path` is the path where the binary will be written to

### Flags
```
Usage: gobfuscate [flags] pkg_name out_path
  -keeptests
    	keep _test.go files
  -noencrypt
    	no encrypted package name for go build command (works when main package has CGO code)
  -nostatic
    	do not statically link
  -outdir
    	output a full GOPATH
  -padding string
    	use a custom padding for hashing sensitive information (otherwise a random padding will be used)
  -tags string
    	tags are passed to the go compiler
  -verbose
    	verbose mode
  -winhide
    	hide windows GUI
```


# What it does

Currently, gobfuscate manipulates package names, global variable and function names, type names, method names, and strings.

### Package name obfuscation

When gobfuscate builds your program, it constructs a copy of a subset of your GOPATH. It then refactors this GOPATH by hashing package names and paths. As a result, a package like "github.com/unixpickle/deleteme" becomes something like "jiikegpkifenppiphdhi/igijfdokiaecdkihheha/jhiofoppieegdaif". This helps get rid of things like Github usernames from the executable.

**Limitation:** currently, packages which use CGO cannot be renamed. I suspect this is due to a bug in Go's refactoring API.

### Global names

Gobfuscate hashes the names of global vars, consts, and funcs. It also hashes the names of any newly-defined types.

Due to restrictions in the refactoring API, this does not work for packages which contain assembly files or use CGO. It also does not work for names which appear multiple times because of build constraints.

### Struct methods

Gobfuscate hashes the names of most struct methods. However, it does not rename methods whose names match methods of any imported interfaces. This is mostly due to internal constraints from the refactoring engine. Theoretically, most interfaces could be obfuscated as well (except for those in the standard library).

Due to restrictions in the refactoring API, this does not work for packages which contain assembly files or use CGO. It also does not work for names which appear multiple times because of build constraints.

### Strings

Strings are obfuscated by replacing them with functions. A string will be turned into an expression like the following:

```go
(func() string {
	mask := []byte{33, 15, 199}
	maskedStr := []byte{73, 106, 190}
	res := make([]byte, 3)
	for i, m := range mask {
		res[i] = m ^ maskedStr[i]
	}
	return string(res)
}())
```

Since `const` declarations cannot include function calls, gobfuscate tries to change any `const` strings into `var`s. It works for declarations like any of the following:

```
const MyStr = "hello"
const MyStr1 = MyStr + "yoyo"
const MyStr2 = MyStr + (MyStr1 + "hello1")

const (
  MyStr3 = "hey there"
  MyStr4 = MyStr1 + "yo"
)
```

However, it does not work for mixed const/int blocks:

```
const (
  MyStr = "hey there"
  MyNum = 3
)
```

# License

This is under a BSD 2-clause license. See [LICENSE](LICENSE).
