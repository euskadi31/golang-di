// Copyright 2018 Axel Etcheverry. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

// ServiceType struct
type ServiceType struct {
	IsPointer bool
	Name      string
}

// Service struct
type Service struct {
	PackageName string
	FactoryName string
	Params      []string
	Type        ServiceType
}

// Identifier of service
func (s Service) Identifier() string {
	ident := ""

	if s.Type.IsPointer {
		ident += "*"
	}

	ident += s.PackageName
	ident += "."
	ident += s.Type.Name

	return ident
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

	fmt.Printf("services: %#v", visitor.Services)
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

func (v *ServiceVisitor) parseType(n *ast.FuncType) ServiceType {
	st := ServiceType{}

	switch t := n.Results.List[0].Type.(type) {
	case *ast.StarExpr:
		ident := t.X.(*ast.Ident)

		st.IsPointer = true
		st.Name = ident.Name
	}

	return st
}

func (v *ServiceVisitor) parseParams(n *ast.FuncType) []string {
	params := make([]string, n.Params.NumFields())

	for _, param := range n.Params.List {
		switch t := param.Type.(type) {
		case *ast.StarExpr:
			ident := t.X.(*ast.Ident)

			params = append(params, fmt.Sprintf("*%s.%s", v.currentPackage, ident.Name))
		}
	}

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
			PackageName: v.currentPackage,
			FactoryName: n.Name.Name,
			Type:        v.parseType(n.Type),
			Params:      v.parseParams(n.Type),
		}

		fmt.Printf("service: %+v\n", service)

		v.Services[service.Identifier()] = service

		return v

	case *ast.TypeSpec:
		//fmt.Println(n.Name.Name)
	case *ast.CommentGroup:
		/*for _, c := range n.List {
			for _, l := range strings.Split(c.Text, "\n") {
				l = strings.Trim(l, "/ ")

				if !strings.Contains(l, "@Service") {
					continue
				}

				fmt.Println(l)
			}

		}*/
	case *ast.Comment:
		/*for _, l := range strings.Split(n.Text, "\n") {
			fmt.Println(l)
		}*/
	}

	return nil
}
