package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/printer"
	"github.com/kr/pretty"
	"gitlab.com/wkhere/argp"
)

type action struct {
	inputFile string
	do        struct {
		filter  bool
		merge   bool
		astDump bool
	}
	filter struct{ path string }
	merge  struct {
		root, file string
	}
	astDumpFile string

	help func()
}

func run(a *action) error {

	f, err := parser.ParseFile(a.inputFile, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if len(f.Docs) > 1 {
		return errors.New("multiple documents not handled")
	}
	var finalNode ast.Node

	switch {
	case a.do.filter:
		path, err := yaml.PathString(a.filter.path)
		if err != nil {
			return fmt.Errorf("filter path: %w", errors.Unwrap(err))
		}
		finalNode, err = path.FilterFile(f)
		if err != nil {
			return fmt.Errorf("filter: %w", err)
		}

	case a.do.merge:
		path, err := yaml.PathString(a.merge.root)
		if err != nil {
			return fmt.Errorf("merge root: %w", errors.Unwrap(err))
		}

		r, err := parser.ParseFile(a.merge.file, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("merged file: %w", err)
		}
		err = path.MergeFromFile(f, r)
		if err != nil {
			return fmt.Errorf("merge: %w", err)
		}
		fallthrough

	default:
		finalNode = f.Docs[0]
	}

	if a.do.astDump {
		dump, err := os.Create(a.astDumpFile)
		if err != nil {
			return fmt.Errorf("dump: %w", err)
		}
		pretty.Fprintf(dump, "%# v\n", finalNode)
		err = dump.Close()
		if err != nil {
			return fmt.Errorf("dump: %w", err)
		}
	}

	var p printer.Printer
	os.Stdout.Write(p.PrintNode(finalNode))

	return nil
}

func parseArgs(args []string) (a action, err error) {
	const usage = `usage: yc [FLAGS] FILE
    -f,--filter=YPATH
    -m,--merge=FILE
    -r,--merge-root=YPATH (default "$")
    -d,--ast-dump=DUMPFILE
`
	var p argp.PState
	var rest []string
flags:
	for ; len(args) > 0; args = args[1:] {
		switch arg := args[0]; {

		case p.IsFlagExpr(arg, "-f", "--filter"):
			a.do.filter = true
			a.filter.path = p.Val
			if a.filter.path[0] != '$' {
				a.filter.path = "$" + a.filter.path
			}

		case p.IsFlagExpr(arg, "-m", "--merge"):
			a.do.merge = true
			a.merge.file = p.Val
			if a.merge.root == "" {
				a.merge.root = "$"
			}

		case p.IsFlagExpr(arg, "-r", "--merge-root"):
			a.merge.root = p.Val
			if a.merge.root[0] != '$' {
				a.merge.root = "$" + a.merge.root
			}

		case p.IsFlagExpr(arg, "-d", "--ast-dump"):
			a.do.astDump = true
			a.astDumpFile = p.Val

		case arg == "-h", arg == "--help":
			a.help = func() { fmt.Print(usage) }
			return a, nil

		case arg == "--":
			rest = append(rest, args[1:]...)
			break flags

		case len(arg) > 1 && arg[0] == '-':
			p.Errorf("unknown flag %s", arg)

		default:
			rest = append(rest, arg)
		}
	}

	switch {
	case p.Err != nil:
		return a, p.Err

	case a.do.filter && a.do.merge:
		return a, errors.New("conflicting flags")

	case len(rest) != 1:
		return a, errors.New("expected input FILE")

	case len(rest) == 1:
		a.inputFile = rest[0]
	}
	return a, nil
}

func main() {
	a, err := parseArgs(os.Args[1:])
	if err != nil {
		die(2, err)
	}
	if a.help != nil {
		a.help()
		os.Exit(0)
	}

	err = run(&a)
	if err != nil {
		die(1, err)
	}
}

func die(exitcode int, err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(exitcode)
}
