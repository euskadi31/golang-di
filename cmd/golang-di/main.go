// Copyright 2018 Axel Etcheverry. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

// Identifier struct
type Identifier struct {
	Package   string
	IsPointer bool
	Name      string
}

// String implements fmt.Stringer
func (i Identifier) String() string {
	ident := ""

	if i.IsPointer {
		ident += "*"
	}

	ident += i.Package
	ident += "."
	ident += i.Name

	return ident
}

// Service struct
type Service struct {
	Identifier  Identifier
	FactoryName string
	Params      []Identifier
}

// Services type
type Services map[string]*Service

func main() {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, "./demo/", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	visitor := NewServiceVisistor()

	for _, pkg := range pkgs {
		ast.Walk(visitor, pkg)
	}

	g, err := NewGenerator()
	if err != nil {
		panic(err)
	}

	if err := g.Generate(visitor.Services); err != nil {
		panic(err)
	}
}

// ServiceVisitor struct
type ServiceVisitor struct {
	currentPackage string
	Services       Services
	currentService *Service
}

// NewServiceVisistor constructor
func NewServiceVisistor() *ServiceVisitor {
	return &ServiceVisitor{
		Services: make(Services),
	}
}

func (v *ServiceVisitor) isService(n *ast.FuncDecl) bool {
	for _, c := range n.Doc.List {
		for _, l := range strings.Split(c.Text, "\n") {
			l = strings.Trim(l, "/ ")

			if strings.Contains(l, "@Service") {
				return true
			}
		}
	}

	return false
}

func (v *ServiceVisitor) parseIdentifier(n *ast.FuncType) Identifier {
	i := Identifier{
		Package: v.currentPackage,
	}

	switch t := n.Results.List[0].Type.(type) {
	case *ast.StarExpr:
		ident := t.X.(*ast.Ident)

		i.IsPointer = true
		i.Name = ident.Name
	}

	return i
}

func (v *ServiceVisitor) parseParams(n *ast.FuncType) []Identifier {
	params := []Identifier{}

	for _, param := range n.Params.List {
		switch t := param.Type.(type) {
		case *ast.StarExpr:
			ident := t.X.(*ast.Ident)

			params = append(params, Identifier{
				Package:   v.currentPackage,
				IsPointer: true,
				Name:      ident.Name,
			})

			fmt.Printf("Name: %s\n", ident.Name)
		case *ast.Ident:
			params = append(params, Identifier{
				Package:   v.currentPackage,
				IsPointer: false,
				Name:      t.Name,
			})

			fmt.Printf("Name: %s\n", t.Name)
		}
	}

	fmt.Printf("Len: %d\n", len(params))

	return params
}

// Visit implements ast.Visitor
func (v *ServiceVisitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.Package:
		v.currentPackage = n.Name

		return v
	case *ast.File:
		return v
	case *ast.GenDecl:
		if n.Tok == token.TYPE {
			return v
		}
	case *ast.FuncDecl:
		if !v.isService(n) {
			return v
		}

		if !n.Name.IsExported() {
			log.Panicf("Service %s is not exportable", n.Name.Name)
		}

		service := &Service{
			Identifier:  v.parseIdentifier(n.Type),
			FactoryName: n.Name.Name,
			Params:      v.parseParams(n.Type),
		}

		fmt.Printf("service: %+v\n", service)

		v.Services[service.Identifier.String()] = service

		return v
	}

	return nil
}

// Generator struct
type Generator struct {
	tmpl *template.Template
}

// NewGenerator constructor
func NewGenerator() (*Generator, error) {
	tmpl, err := template.New("service").Parse(string(content))
	if err != nil {
		return nil, err
	}

	return &Generator{
		tmpl: tmpl,
	}, nil
}

// Generate services files
func (g Generator) Generate(services Services) error {
	fmt.Printf("services: %#v", services)

	f, err := os.Create("services.go")
	if err != nil {
		return err
	}

	f.Sync()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	err = g.tmpl.Execute(writer, struct {
		Services Services
	}{
		Services: services,
	})
	if err != nil {
		return err
	}

	writer.Flush()

	out, err := format.Source(b.Bytes())
	if err != nil {
		fmt.Println("Error in generated source :")
		fmt.Println(string(b.Bytes()))

		return err
	}

	_, err = f.Write(out)
	return err
}
