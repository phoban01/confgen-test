package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
	"golang.org/x/exp/slices"
)

//
// TODO: prune fields not in schema that are in data file
// TODO: ensure comments are always set from schema
// TODO: Run cue fmt or equivelant to simplify and ensure consistent formatting

// if the system manages the field and the schema is on a new version:
// remove the field after reading data.cue and it will be updated

// if the user manages the field and the schema is on a new version:
// add a comment marking that the schema has changed the field

// if the field is unmanaged:
// do nothing

// required fields should have simply the field tpye as literal rather than string
// figure out how to get them as a type

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

	schemaValue := ctx.BuildFile(schema)
	schemaVersion, err := schemaValue.LookupPath(cue.ParsePath("#SchemaVersion")).String()
	if err != nil {
		log.Fatal(err)
	}

	data, _, err := parseFile(ctx, inputDataFile)
	if err != nil {
		log.Fatal(err)
	}

	// here we walk the ast and remove
	astutil.Apply(data, func(c astutil.Cursor) bool {
		f, ok := c.Node().(*ast.Field)
		if !ok {
			return true
		}
		for _, attr := range f.Attrs {
			k, v := attr.Split()
			if k != "manager" {
				continue
			}
			args := strings.Split(v, ",")
			if len(args) != 2 {
				continue
			}
			id := strings.Split(args[0], "=")[1]
			sv := strings.Trim(strings.Split(args[1], "=")[1], "\"")
			if id == "system" && sv != schemaVersion {
				fmt.Println("resetting", id, sv, schemaVersion)
				c.Delete()
			}
		}
		return true
	}, nil)

	if err := astutil.Sanitize(data); err != nil {
		log.Fatal(err)
	}

	missingFields := make([]cue.Path, 0)
	missingFields, err = getMissingFields(
		schemaValue,
		ctx.BuildFile(data),
		missingFields,
	)
	if err != nil {
		log.Fatal(err)
	}

	completed := &ast.File{
		Decls: generateDefaults(ctx.BuildFile(schema), missingFields, schemaVersion),
	}
	completed.SetComments(schemaComments)
	newData := ctx.BuildFile(completed, cue.Scope(ctx.BuildFile(data)))

	ufx := ctx.BuildFile(data).Unify(newData)

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

	if err := astutil.Sanitize(exportFile); err != nil {
		log.Fatal(err)
	}

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

func parseFile(ctx *cue.Context, p string) (*ast.File, []*ast.CommentGroup, error) {
	tree, err := parser.ParseFile(p, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	return tree, tree.Comments(), nil
}

func generateDefaults(input cue.Value, missingFields []cue.Path, schemaVersion string) []ast.Decl {
	result := make([]ast.Decl, 0)
	fields, err := input.Fields(cue.All(), cue.Hidden(false))
	if err != nil {
		log.Fatal(err)
	}

	var paths []string
	for _, p := range missingFields {
		paths = append(paths, p.String())
	}

	result = makeValues(schemaVersion, fields, paths, result)
	return result
}

func makeValues(
	schemaVersion string,
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
		attrs := v.Attributes(cue.ValueAttr)

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
				rx = makeValues(schemaVersion, f, paths, rx, sel...)
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

		f := &ast.Field{
			Label: ast.NewIdent(label),
		}

		if value != nil && !required {
			f.Value = value
		} else {
			switch vnode := v.Syntax(cue.Raw()).(type) {
			case *ast.BinaryExpr:
				f.Value = vnode
			default:
				f.Value = ast.NewIdent(v.IncompleteKind().String())
			}
		}

		for _, av := range attrs {
			ax := &ast.Attribute{
				Text: fmt.Sprintf(
					"@%s(%s,schema_version=\"%s\")",
					av.Name(),
					av.Contents(),
					schemaVersion,
				),
			}
			f.Attrs = append(f.Attrs, ax)
		}
		ast.SetComments(f, comments)
		result = append(result, f)
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
		attr := i.Value().Attribute("private")
		if err := attr.Err(); err == nil {
			continue
		}

		// need to ensure that if a nested
		// field is changed that we can update the individual field without
		// causing a conflict elswhere
		entry := v.LookupPath(path)

		if !entry.Exists() {
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
