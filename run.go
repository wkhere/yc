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

func run(a *action) (rerr error) {

	f, err := parser.ParseFile(a.input.file, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if len(f.Docs) > 1 {
		return errors.New("multiple documents not supported")
	}
	var rootNode ast.Node = f.Docs[0]
	var finalNode = rootNode

	if a.do.merge {
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
	}

	if a.do.filter {
		path, err := yaml.PathString(a.filter.path)
		if err != nil {
			return fmt.Errorf("filter path: %w", errors.Unwrap(err))
		}
		node, err := path.FilterFile(f)
		if err != nil {
			return fmt.Errorf("filter: %w", err)
		}
		finalNode = ast.Parent(rootNode, node)
	}

	if a.do.astDump {
		err := dumpNode(a.astDump.file, finalNode)
		if err != nil {
			a.log.Printf("skipped dump: %v", err)
		}
	}

	var p printer.Printer
	if a.inplace {
		w, err := os.Create(a.input.file)
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
		defer safeClose(w, &rerr)
		_, err = w.Write(p.PrintNode(finalNode))
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
		return nil
	}
	os.Stdout.Write(p.PrintNode(finalNode))
	return nil
}

func dumpNode(dumpFile string, node ast.Node) (rerr error) {
	dump, err := os.Create(dumpFile)
	if err != nil {
		return err
	}
	defer safeClose(dump, &rerr)
	_, err = pretty.Fprintf(dump, "%# v\n", node)
	return err

}

func safeClose(f *os.File, perr *error) {
	cerr := f.Close()
	if cerr != nil && *perr == nil {
		*perr = fmt.Errorf("close: %w", cerr)
	}
}
