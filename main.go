package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
)

type funcDescription struct {
	Recv *ast.FieldList
	Type *ast.FuncType
	Body *ast.BlockStmt
}

var (
	receivers = flag.Bool("recv", false, "include receivers in scan")
	returns   = flag.Bool("ret", false, "include returns in scan")
	overwrite = flag.Bool("w", false, "overwrite file with changes")
)

func main() {

	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Usage:", os.Args[0], "[-w] <file[s]>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	for _, file := range flag.Args() {

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			fmt.Println(file, err)
			continue
		}

		altered := blankId(f)

		if *overwrite {
			if altered {
				fmt.Println(file, "overwritten")
				outfile, err := os.OpenFile(file, os.O_TRUNC|os.O_WRONLY, 0)
				if err != nil {
					log.Fatal(err)
				}
				if err := printer.Fprint(outfile, fset, f); err != nil {
					log.Fatal(err)
				}
				if err := outfile.Close(); err != nil {
					log.Fatal(err)
				}
			} else {
				fmt.Println(file, "unaltered")
			}
		} else {
			printer.Fprint(os.Stdout, fset, f)
		}

	}

}

func blankId(f ast.Node) bool {

	var altered bool

	ast.Inspect(f, func(n ast.Node) bool {

		switch x := n.(type) {
		case *ast.FuncDecl:
			if changed := scan(&funcDescription{
				Recv: x.Recv,
				Type: x.Type,
				Body: x.Body,
			}); changed {
				altered = true
			}
		case *ast.FuncLit:
			if changed := scan(&funcDescription{
				Type: x.Type,
				Body: x.Body,
			}); changed {
				altered = true
			}
		}

		return true

	})

	return altered

}

func scan(x *funcDescription) bool {

	if x.Body == nil {
		return false
	}

	type v struct {
		ident *ast.Ident
		used  bool
	}

	vars := map[string]*v{}

	// store receiver
	if *receivers && x.Recv != nil {
		// only if a named receiver
		if len(x.Recv.List[0].Names) > 0 {
			vars[x.Recv.List[0].Names[0].Name] = &v{ident: x.Recv.List[0].Names[0]}
		}
	}

	// store function params
	for _, p := range x.Type.Params.List {
		for _, ident := range p.Names {
			if ident.Name != "_" {
				vars[ident.Name] = &v{ident: ident}
			}
		}
	}

	// store return params
	if *returns && x.Type.Results != nil {
		for _, r := range x.Type.Results.List {
			for _, ident := range r.Names {
				if ident.Name != "_" {
					vars[ident.Name] = &v{ident: ident}
				}
			}
		}
	}

	// scan for idents that are used in the function body
	ast.Inspect(x.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if x.Name != "_" {
				if iv, exists := vars[x.Name]; exists {
					if x.Obj != nil {
						// matching ident declaration means same var
						if f, ok := x.Obj.Decl.(*ast.Field); ok && f != nil {
							for _, ident := range f.Names {
								if ident == iv.ident {
									iv.used = true
									return true
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	var altered bool

	// set to _ where unused
	for _, v := range vars {
		if !v.used {
			v.ident.Name = "_"
			altered = true
		}
	}

	return altered

}
