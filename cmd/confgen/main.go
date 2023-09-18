package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
	"golang.org/x/exp/slices"
)

//
// TODO: prune fields not in schema that are in data file
// TODO: ensure comments are always set from schema
// TODO: Run cue fmt or equivelant to simplify and ensure consistent formatting

func main() {
	inputDataFile, err := filepath.Abs("data.cue")
	if err != nil {
		log.Fatal(err)
	}
	ctx := cuecontext.New()

	schema, schemaComments, err := parseFile(ctx, "schema.cue")
	if err != nil {
		log.Fatal(err)
	}

	data, _, err := parseFile(ctx, inputDataFile)
	if err != nil {
		log.Fatal(err)
	}

	missingFields := make([]cue.Path, 0)
	missingFields, err = getMissingFields(schema, data, missingFields)
	if err != nil {
		log.Fatal(err)
	}

	completed := &ast.File{Decls: generateDefaults(schema, missingFields)}
	completed.SetComments(schemaComments)
	newData := ctx.BuildFile(completed, cue.Scope(data))

	ufx := data.Unify(newData)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	tmpFile, err := os.CreateTemp(cwd, "mpas-data-*")
	if err != nil {
		log.Fatal(err)
	}

	exportFile := &ast.File{
		Decls: ufx.Syntax(cue.Docs(true), cue.All()).(*ast.StructLit).Elts,
	}
	exportFile.SetComments(schemaComments)
	exportBytes, err := format.Node(exportFile, format.Simplify())
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(string(exportBytes))

	if _, err := tmpFile.Write(exportBytes); err != nil {
		log.Fatal(err)
	}

	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	if err := os.Rename(tmpFile.Name(), inputDataFile); err != nil {
		log.Fatal(err)
	}
}

func parseFile(ctx *cue.Context, p string) (cue.Value, []*ast.CommentGroup, error) {
	tree, err := parser.ParseFile(p, nil, parser.ParseComments)
	if err != nil {
		return cue.Value{}, nil, err
	}

	return ctx.BuildFile(tree), tree.Comments(), nil
}

func generateDefaults(input cue.Value, missingFields []cue.Path) []ast.Decl {
	result := make([]ast.Decl, 0)
	fields, err := input.Fields(cue.All())
	if err != nil {
		log.Fatal(err)
	}

	var paths []string
	for _, p := range missingFields {
		paths = append(paths, p.String())
	}

	result = makeValues(fields, paths, result)
	return result
}

func makeValues(
	i *cue.Iterator,
	paths []string,
	result []ast.Decl,
	parents ...cue.Selector,
) []ast.Decl {
	for i.Next() {
		var value ast.Expr
		var required bool
		v := i.Value()
		label, _ := v.Label()
		sel := append(parents, i.Selector())
		path := cue.MakePath(sel...)
		comments := v.Doc()

		if !slices.Contains(paths, path.String()) {
			continue
		}

		field, hasDefaultValue := v.Default()

		if !hasDefaultValue && v.IsConcrete() {
			switch v.Syntax(cue.Raw()).(type) {
			case *ast.StructLit:
				var rx []ast.Decl
				f, err := i.Value().Fields()
				if err != nil {
					log.Fatal(err)
				}
				rx = makeValues(f, paths, rx, sel...)
				value = &ast.StructLit{
					Elts: rx,
				}
			default:
				value = field.Syntax(cue.Raw()).(ast.Expr)
			}
		} else if !hasDefaultValue {
			required = true
		} else {
			value = field.Syntax(cue.Raw()).(ast.Expr)
		}

		if value != nil && !required {
			f := &ast.Field{
				Label: ast.NewIdent(label),
				Value: value,
			}
			ast.SetComments(f, comments)
			result = append(result, f)
		} else if value == nil {
			f := &ast.Field{
				Label: ast.NewIdent(label),
			}
			switch v.Syntax().(type) {
			case *ast.BinaryExpr:
				f.Value = v.Syntax(cue.Raw()).(*ast.BinaryExpr)
			default:
				f.Value = ast.NewString(fmt.Sprintf("<%s>", v.IncompleteKind()))
			}
			ast.SetComments(f, comments)
			result = append(result, f)
		}
	}
	return result
}

func getMissingFields(
	c cue.Value,
	v cue.Value,
	m []cue.Path,
	parents ...cue.Selector,
) ([]cue.Path, error) {
	i, err := c.Fields(cue.All())
	if err != nil {
		return nil, err
	}

	for i.Next() {
		sel := append(parents, i.Selector())
		path := cue.MakePath(sel...)

		// need to ensure that if a nested
		// field is changed that we can update the individual field without
		// causing a conflict elswhere

		if !v.LookupPath(path).Exists() {
			m = append(m, path)
		}

		switch i.Value().Syntax().(type) {
		case *ast.StructLit:
			// make a copy to avoid iterator issues
			x := c.LookupPath(path)

			// iterate over the struct fields
			var n []cue.Path
			n, err = getMissingFields(x, v, n, sel[1:]...)
			if err != nil {
				return nil, err
			}

			// restore the selector prefix to the path
			for _, nv := range n {
				nsel := sel[:]
				for _, nvs := range nv.Selectors() {
					nsel = append(nsel, nvs)
				}
				m = append(m, cue.MakePath(nsel...))
			}
		}
	}

	return m, nil
}
