// Copyright © 2020 Yoshiki Shibata. All rights reserved.

package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

const version = "1.3.0"

var runFiles = flag.String("run", "", "Go test file to run. Multiple files can be seperated by a comma.")
var verbose = flag.Bool("v", false, "verbose")
var tags = flag.String("tags", "", "tags")
var shuffle = flag.Bool("shuffle", false, "shuffle the order of tests")

func main() {
	flag.Parse()

	if *runFiles == "" {
		fmt.Fprintf(os.Stderr, "usage: gotest [-v] -run=testfile\n")
		os.Exit(1)
	}

	files := strings.Split(*runFiles, ",")
	var funcNames []string
	for _, file := range files {
		funcs := listFuncNames(file, "Test")
		if len(funcs) == 0 {
			fmt.Fprintf(os.Stderr,
				"%s doesn't containy any Test* functions\n",
				*runFiles)
			os.Exit(1)
		}
		funcNames = append(funcNames, funcs...)
	}

	if *shuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(funcNames), func(i, j int) {
			funcNames[i], funcNames[j] = funcNames[j], funcNames[i]
		})
	}

	runFlag := createRunFlag(funcNames)
	cmdArgs := createCmdArgs(runFlag, *verbose, *tags)

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

func createCmdArgs(runFlag string, verbose bool, tags string) []string {
	args := []string{"test"}

	if verbose {
		args = append(args, "-v")
	}

	if tags != "" {
		args = append(args, fmt.Sprintf("-tags=%s", tags))
	}

	return append(args, runFlag)
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
		os.Exit(1)
	}
}
