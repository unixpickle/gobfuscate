# gobfuscate

When you compile a Go binary, it contains a lot of information about your source code: field names, method names, package paths, etc. If you want to ship a binary without leaking this kind of information, what are you to do? Enter gobfuscate.

With gobfuscate, you can compile a Go binary from obfuscated source code. This prevents any unwanted information difficult or impossible to decipher from the binary.
