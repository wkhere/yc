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
)

func run(a *action) error {

	f, err := parser.ParseFile(a.inputFile, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if len(f.Docs) > 1 {
		return errors.New("multiple documents not handled")
	}
	var rootNode = f.Docs[0]
	var finalNode ast.Node

	switch {
	case a.do.filter:
		path, err := yaml.PathString(a.filter.path)
		if err != nil {
			return fmt.Errorf("filter path: %w", errors.Unwrap(err))
		}
		node, err := path.FilterFile(f)
		if err != nil {
			return fmt.Errorf("filter: %w", err)
		}
		finalNode = ast.Parent(rootNode, node)

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
		finalNode = rootNode
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
