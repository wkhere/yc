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
		err = merge(f, a.merge.root, a.merge.file)
		if err != nil {
			return fmt.Errorf("merge: %w", err)
		}
	}

	if a.do.filter {
		err = filter(f, &finalNode, rootNode, a.filter.path)
		if err != nil {
			return fmt.Errorf("filter: %w", err)
		}
	}

	if a.do.astDump {
		err := dumpNode(a.astDump.file, finalNode)
		if err != nil {
			a.log.Printf("skipped dump: %v", err)
		}
	}

	var p printer.Printer
	if a.inplace {
		err = writeFile(a.input.file, p.PrintNode(finalNode))
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
		return nil
	}
	os.Stdout.Write(p.PrintNode(finalNode))
	return nil
}

func merge(f *ast.File, mergeRoot, mergeFile string) error {
	path, err := yaml.PathString(mergeRoot)
	if err != nil {
		return errors.Unwrap(err)
	}
	r, err := parser.ParseFile(mergeFile, parser.ParseComments)
	if err != nil {
		return err
	}
	return path.MergeFromFile(f, r)
}

func filter(f *ast.File, dest *ast.Node, root ast.Node, filterPath string) error {
	path, err := yaml.PathString(filterPath)
	if err != nil {
		return errors.Unwrap(err)
	}
	node, err := path.FilterFile(f)
	if err != nil {
		return err
	}
	*dest = ast.Parent(root, node)
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

func writeFile(file string, buf []byte) (rerr error) {
	w, err := os.Create(file)
	if err != nil {
		return err
	}
	defer safeClose(w, &rerr)
	_, err = w.Write(buf)
	return err
}

func safeClose(f *os.File, perr *error) {
	cerr := f.Close()
	if cerr != nil && *perr == nil {
		*perr = fmt.Errorf("close: %w", cerr)
	}
}
