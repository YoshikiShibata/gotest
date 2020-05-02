// Copyright © 2020 Yoshiki Shibata. All rights reserved.

package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"strings"
)

const version = "1.0.0"

var runFile = flag.String("run", "", "Go test file to run")
var verbose = flag.Bool("v", false, "verbose")

func main() {
	flag.Parse()

	if *runFile == "" {
		fmt.Fprintf(os.Stderr, "usage: gotest [-v] -run=testfile\n")
		os.Exit(1)
	}

	funcNames := listFuncNames(*runFile, "Test")
	if len(funcNames) == 0 {
		fmt.Fprintf(os.Stderr,
			"%s doesn't containy any Test* functions\n",
			*runFile)
		os.Exit(1)
	}

	runFlag := createRunFlag(funcNames)
	cmdArgs := createCmdArgs(runFlag, *verbose)

	if *verbose {
		fmt.Fprintf(os.Stdout, "gotest version %s\n", version)
		fmt.Fprintf(os.Stdout, "go %s\n\n", strings.Join(cmdArgs, " "))
	}

	execGoTestCommand(cmdArgs)
}

func listFuncNames(filename, funcPrefix string) []string {
	var funcs []string

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse %s\n", filename)
		os.Exit(1)
	}

	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if strings.HasPrefix(funcDecl.Name.Name, funcPrefix) {
			funcs = append(funcs, funcDecl.Name.Name)
		}
	}
	return funcs
}

func createRunFlag(funcNames []string) string {
	var sb strings.Builder

	sb.WriteString("-run=")
	for i, funcName := range funcNames {
		if i != 0 {
			sb.WriteString("$|")
		}
		sb.WriteString(funcName)
	}
	sb.WriteString("$")
	return sb.String()
}

func createCmdArgs(runFlag string, verbose bool) []string {
	if verbose {
		return []string{"test", "-v", runFlag}
	}
	return []string{"test", runFlag}
}

func execGoTestCommand(cmdArgs []string) {
	cmd := exec.CommandContext(context.TODO(), "go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Start() failed: %v\n", err)
		os.Exit(1)
	}
	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Wait() failed: %v\n", err)
	}
}
