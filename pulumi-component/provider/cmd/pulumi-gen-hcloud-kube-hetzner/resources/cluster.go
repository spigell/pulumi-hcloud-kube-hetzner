package resources

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/pkg/phkh"
)

const (
	ClusterTypePrefix         = "hcloud-kube-hetzner:cluster"
	ClusterServersOutputsName = ClusterTypePrefix + ":servers"
	ClusterConfigType         = "config" + "Config"

	pointerChar = "*"
)

var serversOutputs = schema.ComplexTypeSpec{
	ObjectTypeSpec: schema.ObjectTypeSpec{
		Type: "object",
		Properties: map[string]schema.PropertySpec{
			phkh.ServerIPKey: {
				TypeSpec: schema.TypeSpec{Type: "string"},
			},
			phkh.ServerInternalIPKey: {
				TypeSpec: schema.TypeSpec{Type: "string"},
			},
			phkh.ServerUserKey: {
				TypeSpec: schema.TypeSpec{Type: "string"},
			},
			phkh.ServerNameKey: {
				TypeSpec: schema.TypeSpec{Type: "string"},
			},
		},
	},
}

func GatherClusterTypes(dir string) (map[string]schema.ComplexTypeSpec, error) {
	types := make(map[string]schema.ComplexTypeSpec)

	types[ClusterServersOutputsName] = serversOutputs

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".go") || info.IsDir() {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		packageName := node.Name.Name
		for _, decl := range node.Decls {
			if genDecl, ok := processTypeDecl(decl); ok {
				processTypeSpecs(packageName, genDecl.Specs, types)
			}
		}
		return nil
	})

	return types, err
}

func processTypeSpecs(packageName string, specs []ast.Spec, types map[string]schema.ComplexTypeSpec) {
	for _, spec := range specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			processTypeSpec(packageName, typeSpec, types)
		}
	}
}

func processTypeSpec(packageName string, typeSpec *ast.TypeSpec, types map[string]schema.ComplexTypeSpec) {
	if structType, ok := typeSpec.Type.(*ast.StructType); ok && strings.HasSuffix(typeSpec.Name.Name, "Config") {
		if isType(structType) && exportedField(typeSpec.Name.Name) {
			t := schema.ComplexTypeSpec{
				ObjectTypeSpec: schema.ObjectTypeSpec{
					Type:       "object",
					Properties: make(map[string]schema.PropertySpec),
				},
			}
			for _, field := range structType.Fields.List {
				processField(packageName, field, t)
			}
			typeKey := fmt.Sprintf("hcloud-kube-hetzner:cluster:%s%s", packageName, typeSpec.Name.Name)
			types[typeKey] = t
		}
	}
}

func processField(packageName string, field *ast.Field, t schema.ComplexTypeSpec) {
	if len(field.Names) == 0 || !exportedField(field.Names[0].Name) {
		return
	}

	fieldName := fieldName(field)
	if fieldName == "" {
		return
	}

	tt := schema.PropertySpec{
		Description: fieldDescription(field),
		TypeSpec: schema.TypeSpec{
			Ref: fmt.Sprintf("#types/%s:%s", ClusterTypePrefix, strings.ReplaceAll(fieldType(packageName, field), "*", "")),
		},
	}

	if isSimpleType(field.Type) {
		ty := resolveType(field.Type)
		tt.TypeSpec.Ref = ""
		tt.TypeSpec.Type = ty

		if ty == "array" {
			tt.Items = &schema.TypeSpec{Type: "string"}

			fullType := fieldType(packageName, field)
			if strings.Contains(fullType, pointerChar) {
				tt.Items.Type = ""
				tt.Items.Ref = fmt.Sprintf(
					"#types/%s:%s%s",
					ClusterTypePrefix,
					packageName,
					strings.Split(fullType, pointerChar)[1],
				)
			}
		}
	}

	t.ObjectTypeSpec.Properties[fieldName] = tt
}

func processTypeDecl(decl ast.Decl) (*ast.GenDecl, bool) {
	genDecl, ok := decl.(*ast.GenDecl)
	return genDecl, ok && genDecl.Tok == token.TYPE
}

func fieldName(field *ast.Field) string {
	fieldName := field.Names[0].Name

	if field.Tag == nil {
		return fieldName
	}

	jsonTag, ok := parseTag(field.Tag.Value, "json")
	if ok {
		if jsonTag == "-" {
			return ""
		}
	}

	return fieldName
}

// Consolidates comments from Doc and Comment fields.
func fieldDescription(field *ast.Field) string {
	var parts []string
	if field.Doc != nil {
		parts = append(parts, field.Doc.Text())
	}
	if field.Comment != nil {
		parts = append(parts, field.Comment.Text())
	}
	return strings.ReplaceAll(strings.Join(parts, " "), "\n", " ")
}

// Parse a struct tag to get the value associated with a key.
func parseTag(tagValue string, key string) (string, bool) {
	tags := reflect.StructTag(strings.Trim(tagValue, "`"))
	value, ok := tags.Lookup(key)
	return value, ok
}

func isType(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.StarExpr:
		return true
	case *ast.StructType:
		return true
	default:
		return false
	}
}

func isSimpleType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return true
	case *ast.ArrayType:
		return true
	case *ast.StarExpr:
		if isBool(t.X) {
			return true
		}
	}

	return false
}

func resolveType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if isBool(t.X) {
			return "boolean"
		}
		return ""
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string"
		case "bool":
			return "boolean"
		case "int":
			return "number"
		default:
			return fmt.Sprintf("unknow simple type: %T", t.Name)
		}
	case *ast.ArrayType:
		return "array" // arrays
	default:
		return fmt.Sprintf("unknown type: %T", t)
	}
}

// Check if a field is exported and has a name.
func exportedField(name string) bool {
	return len(name) > 0 && unicode.IsUpper(rune(name[0]))
}

func isBool(t ast.Expr) bool {
	ident, ok := t.(*ast.Ident)

	if ok && ident.Name == "bool" {
		return true
	}

	return false
}

func resolve(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name // simple types
	case *ast.SelectorExpr:
		return (resolve(t.X) + t.Sel.Name) + pointerChar // qualified (package*) types
	case *ast.ArrayType:
		return "[]" + resolve(t.Elt) // arrays
	case *ast.StarExpr:
		return "*" + resolve(t.X) // pointer types
	case *ast.StructType:
		return "struct {...}" // embedded structs
	default:
		return fmt.Sprintf("(%T)[%T]", t, t) // use type name for unknown types
	}
}

func fieldType(packageName string, field *ast.Field) string {
	fieldType := resolve(field.Type)
	if strings.Contains(fieldType, pointerChar) {
		fullType := fmt.Sprintf("%s%s", packageName, fieldType)
		// This is the already full name.
		if strings.HasSuffix(fieldType, pointerChar) {
			fullType = strings.ReplaceAll(fieldType, pointerChar, "")
		}
		return fullType
	}
	return fieldType
}
