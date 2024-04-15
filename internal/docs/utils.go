package docs

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

// Parse a struct tag to get the value associated with a key.
func parseTag(tagValue string, key string) (string, bool) {
	tags := reflect.StructTag(strings.Trim(tagValue, "`"))
	value, ok := tags.Lookup(key)
	return value, ok
}

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
