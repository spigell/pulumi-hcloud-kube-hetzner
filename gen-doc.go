package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := "./internal"

	// Start processing
	docs, err := processDirectory(dir)
	if err != nil {
		fmt.Println("Error processing directory:", err)
		return
	}

	// Check if docs is empty
	if docs == "" {
		fmt.Println("No documentation was generated. Please check the presence of Go files and structs.")
		return
	}

	// Write to a Markdown file
	err = ioutil.WriteFile("documentation.md", []byte(docs), 0644)
	if err != nil {
		fmt.Println("Error writing documentation:", err)
	}
}

func processDirectory(dir string) (string, error) {
	var sb strings.Builder
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") && !info.IsDir() {
			// fmt.Println("Processing file:", path) // Diagnostic output
			content, err := generateMarkdown(path)
			if err != nil {
				fmt.Println("Error processing file:", path, err) // Diagnostic output
				return err
			}
			if content != "" {
				sb.WriteString(content)
			}
		}
		return nil
	})
	return sb.String(), err
}

func generateMarkdown(filePath string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	packageName := node.Name.Name
	for _, d := range node.Decls {
		genDecl, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if strings.HasSuffix(typeSpec.Name.Name, "Config") || packageName == "config" {
				sb.WriteString("## " + fmt.Sprintf("%s.%s", packageName, typeSpec.Name.Name) + "\n\n")
				sb.WriteString("| Field | Type | Description |\n")
				sb.WriteString("|-------|------|-------------|\n")

				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						continue
					}

					if field.Tag == nil || !strings.Contains(field.Tag.Value, "doc:") {
						continue
					}

					fieldName := field.Names[0].Name
					fmt.Printf("%+v\n", field.Type)
					fieldType := resolveType(field.Type)
					if !strings.Contains(fieldType, ".") && strings.Contains(fieldType, "*") {
						fullType := fmt.Sprintf("%s.%s", packageName, fieldType)
						fieldType = fmt.Sprintf("[%s](#%s)", fullType, strings.ToLower(strings.Replace(fullType, ".*", "", -1)))
					}
					var fieldDesc string
					if field.Doc != nil {
						fieldDesc = strings.Join(strings.Fields(field.Doc.Text()), " ")
					} else if field.Comment != nil {
						fieldDesc = strings.Join(strings.Fields(field.Comment.Text()), " ")
					}

					sb.WriteString("| " + fieldName + " | " + fieldType + " | " + fieldDesc + " |\n")
				}
				sb.WriteString("\n")
			}
		}
	}
	return sb.String(), nil
}

// Resolve type expressions to string
func resolveType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name // simple types
	case *ast.SelectorExpr:
		return resolveType(t.X) + "." + t.Sel.Name // qualified (package) types
	case *ast.ArrayType:
		return "[]" + resolveType(t.Elt) // arrays
	case *ast.StarExpr:
		return "*" + resolveType(t.X) // pointer types
	case *ast.StructType:
		return "struct {...}" // embedded structs
	default:
		return fmt.Sprintf("(%T)[%T]", t, t) // use type name for unknown types
	}
}
