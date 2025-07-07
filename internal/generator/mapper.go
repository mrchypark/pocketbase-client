package generator

import (
	"strings"
	"unicode"
)

// MapPbTypeToGoType maps PocketBase field types to Go types.
// It now takes the omitEmpty flag as an argument to determine if a pointer type should be used.
func MapPbTypeToGoType(field FieldSchema, omitEmpty bool) (string, string) {
	var goType string
	var comment string

	switch field.Type {
	case "text", "email", "url", "editor":
		goType = "string"
	case "number":
		goType = "float64"
	case "bool":
		goType = "bool"
	case "date", "authdate":
		goType = "types.DateTime"
	case "json":
		goType = "json.RawMessage"
	case "relation", "file":
		// If maxSelect option is not present or not equal to 1, treat as multiple selection.
		if field.Options != nil && field.Options.MaxSelect != nil && *field.Options.MaxSelect == 1 {
			goType = "string"
			comment = "// Single relation/file" // Translated
		} else {
			goType = "[]string"
			comment = "// Multiple relations/files" // Translated
		}
	case "select":
		if field.Options != nil && field.Options.MaxSelect != nil && *field.Options.MaxSelect == 1 {
			goType = "string"
		} else {
			goType = "[]string"
		}
	default:
		goType = "interface{}"
		comment = "// Unknown PocketBase type" // Translated
	}

	// If omitEmpty is true and the type can be a pointer (not a slice or map),
	// change the type to a pointer type.
	if omitEmpty {
		switch goType {
		case "string", "float64", "bool", "types.DateTime":
			goType = "*" + goType
		}
	}

	return goType, comment
}

// ToPascalCase converts snake_case to PascalCase.
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}
	// Handle common abbreviations
	switch strings.ToLower(s) {
	case "id":
		return "ID"
	case "url":
		return "URL"
	case "html":
		return "HTML"
	case "json":
		return "JSON"
	}

	var result strings.Builder
	capitalizeNext := true
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
