package docs

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func RenderParametersTable(dir string) (string, error) {
	table, err := processDirectory(dir)
	if err != nil {
		return "", fmt.Errorf("error while process directory: %w", err)
	}

	return table, nil
}

func processDirectory(dir string) (string, error) {
	var sb strings.Builder
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") && !info.IsDir() {
			content, err := generateMarkdownPerFile(path)
			if err != nil {
				return err
			}

			sb.WriteString(content)
		}
		return nil
	})
	return sb.String(), err
}

// Generates markdown documentation for Go structs in a file.
func generateMarkdownPerFile(filePath string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	packageName := node.Name.Name

	for _, decl := range node.Decls {
		if genDecl, ok := processTypeDecl(decl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok && strings.HasSuffix(typeSpec.Name.Name, "Config") {
						generateStructMarkdown(&sb, packageName, typeSpec, structType)
					}
				}
			}
		}
	}
	return sb.String(), nil
}

func processTypeDecl(decl ast.Decl) (*ast.GenDecl, bool) {
	genDecl, ok := decl.(*ast.GenDecl)
	return genDecl, ok && genDecl.Tok == token.TYPE
}

// Helper function to generate markdown for a specific struct.
func generateStructMarkdown(sb *strings.Builder, packageName string, typeSpec *ast.TypeSpec, structType *ast.StructType) {
	sb.WriteString("## " + fmt.Sprintf("%s.%s", packageName, typeSpec.Name.Name) + "\n\n")
	sb.WriteString("| Field | Type | Description | Default |\n")
	sb.WriteString("|-------|------|-------------|---------|\n")

	for _, field := range structType.Fields.List {
		if isExportedField(field) {
			fieldName := getFieldName(field)
			fieldType := getFieldType(packageName, field)
			fieldDesc := getFieldDescription(field)
			fieldDefault := extractDefaultValue(fieldDesc)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", fieldName, fieldType, fieldDesc, fieldDefault))
		}
	}
	sb.WriteString("\n")
}

// Check if a field is exported and has a name.
func isExportedField(field *ast.Field) bool {
	return len(field.Names) > 0 && unicode.IsUpper(rune(field.Names[0].Name[0]))
}

// Resolve field name from tags, or use the struct field name.
func getFieldName(field *ast.Field) string {
	fieldName := strings.ToLower(field.Names[0].Name)

	if field.Tag == nil {
		return fieldName
	}

	jsonTag, ok := parseTag(field.Tag.Value, "json")
	if ok {
		if jsonTag == "-" {
			yamlTag, ok := parseTag(field.Tag.Value, "yaml")
			if ok {
				return strings.Split(yamlTag, ",")[0] + " (computed). Not possible to configure!"
			}
		}
		return strings.Split(jsonTag, ",")[0]
	}

	return fieldName
}

// Extract type information and link if necessary.
func getFieldType(packageName string, field *ast.Field) string {
	fieldType := resolveType(field.Type)
	if strings.Contains(fieldType, "*") {
		fullType := fmt.Sprintf("%s.%s", packageName, fieldType)
		if strings.Contains(fieldType, ".") {
			fullType = fieldType
		}
		return fmt.Sprintf("[%s](#%s)", fullType, toMarkdownLink(fullType))
	}
	return fieldType
}

// Consolidates comments from Doc and Comment fields.
func getFieldDescription(field *ast.Field) string {
	var parts []string
	if field.Doc != nil {
		parts = append(parts, field.Doc.Text())
	}
	if field.Comment != nil {
		parts = append(parts, field.Comment.Text())
	}
	return strings.ReplaceAll(strings.Join(parts, " "), "\n", " ")
}

// toMarkdownLink refactors the string replacements to be more idiomatic and maintainable.
func toMarkdownLink(src string) string {
	// Map of replacements; the key is what to look for, the value is what to replace with.
	replacements := map[string]string{
		"[]": "",
		".":  "",
		"*":  "",
	}

	// Apply all replacements from the map
	for old, new := range replacements {
		src = strings.ReplaceAll(src, old, new)
	}

	// Convert to lower case
	return strings.ToLower(src)
}

func extractDefaultValue(comment string) string {
	prefix := "Default is"
	startPos := strings.Index(comment, prefix)
	if startPos == -1 {
		return ""
	}
	defaultValue := comment[startPos+len(prefix):]
	return strings.Trim(strings.TrimSpace(defaultValue), ".")
}
