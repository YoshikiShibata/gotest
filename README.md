# gotest
A simple wrapper for go test command to specify xxx_test.go file. With `-run` option, you specify test file name such as xxxx_test.go instead of functions.

`gotest` enumerates all `Test_XXX` functions in the file and invokes `go test` command.

# usage

```
gotest [-v] -run=testFileName
```
