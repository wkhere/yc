package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/printer"
	"github.com/kr/pretty"
	"gitlab.com/wkhere/argp"
)

type action struct {
	inputFile   string
	astDumpFile string
	do          struct{ astDump bool }

	help func()
}

func run(a *action) error {

	f, err := parser.ParseFile(a.inputFile, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	if a.do.astDump {
		dump, err := os.Create(a.astDumpFile)
		if err != nil {
			return fmt.Errorf("dump: %w", err)
		}
		pretty.Fprintf(dump, "%# v\n", f)
		err = dump.Close()
		if err != nil {
			return fmt.Errorf("dump: %w", err)
		}
	}

	var p printer.Printer
	// todo: there can be multiple yaml docs in one file, print all of them
	os.Stdout.Write(p.PrintNode(f.Docs[0]))

	return nil
}

func parseArgs(args []string) (a action, _ error) {
	const usage = `usage: yc [-d,--ast-dump=DUMPFILE] FILE`

	var p argp.PState
	var rest []string
flags:
	for ; len(args) > 0 && p.Err == nil; args = args[1:] {
		switch arg := args[0]; {

		case p.IsFlagExpr(arg, "-d", "--ast-dump"):
			a.do.astDump = true
			a.astDumpFile = p.Val

		case arg == "-h", arg == "--help":
			a.help = func() { fmt.Println(usage) }
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
