// Copyright © 2020-2022 Yoshiki Shibata. All rights reserved.

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

const version = "1.6.1"

func main() {
	var (
		runFiles     = flag.String("run", "", "Go test file to run. Multiple files can be seperated by a comma.")
		verbose      = flag.Bool("v", false, "verbose")
		tags         = flag.String("tags", "", "tags")
		race         = flag.Bool("race", false, "race detection")
		cpu          = flag.Int("p", 0, "cpu")
		coverprofile = flag.String("coverprofile", "", "coverprofile")
		coverpkg     = flag.String("coverpkg", "", "coverpkg")
	)

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

	runFlag := createRunFlag(funcNames)
	cmdArgs := createCmdArgs(runFlag,
		*verbose,
		*tags,
		*race,
		*cpu,
		*coverprofile,
		*coverpkg,
	)

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

func createCmdArgs(
	runFlag string,
	verbose bool,
	tags string,
	race bool,
	cpu int,
	coverprofile string,
	coverpkg string,
) []string {
	args := []string{"test"}

	if verbose {
		args = append(args, "-v")
	}

	if tags != "" {
		args = append(args, fmt.Sprintf("-tags=%s", tags))
	}

	if race {
		args = append(args, "-race")
	}

	if cpu != 0 {
		args = append(args, fmt.Sprintf("-p=%d", cpu))
	}

	if coverprofile != "" {
		args = append(args, fmt.Sprintf("-coverprofile=%s", coverprofile))
	}
	if coverpkg != "" {
		args = append(args, fmt.Sprintf("-coverpkg=%s", coverpkg))
	}

	args = append(args, flag.Args()...)
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
