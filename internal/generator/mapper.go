package generator

import (
	"fmt"
	"strings"
)

// MapPbTypeToGoType maps PocketBase field types to Go types and their corresponding getter methods.
// It returns the Go type and the getter method name.
func MapPbTypeToGoType(field FieldSchema, omitEmpty bool) (string, string) {
	var goType, getterMethod string

	// Pre-determine whether it's a multi-select field based on MaxSelect option
	isMulti := false

	// MaxSelect handling:
	// - nil or > 1: multi (nil means unlimited in PocketBase schema)
	// - 1: single
	// - 0: treated as single (0 is often used as default/unset value)
	if field.Options != nil && field.Options.MaxSelect != nil {
		maxSelect := *field.Options.MaxSelect
		if maxSelect > 1 {
			isMulti = true
		}
		// maxSelect <= 1 (including 0) is treated as single
	} else if field.Type == "relation" || field.Type == "file" || field.Type == "select" {
		// When MaxSelect is nil (when maxSelect is missing or null in schema),
		// relation/file/select types are considered multi by default.
		isMulti = true
	}

	switch field.Type {
	case "text", "email", "url", "editor", "password":
		goType = "string"
		getterMethod = "GetString"
		if omitEmpty {
			getterMethod = "GetStringPointer"
		}
	case "number":
		goType = "float64"
		getterMethod = "GetFloat"
		if omitEmpty {
			getterMethod = "GetFloatPointer"
		}
	case "bool":
		goType = "bool"
		getterMethod = "GetBool"
		if omitEmpty {
			getterMethod = "GetBoolPointer"
		}
	case "date", "autodate":
		goType = "pocketbase.DateTime"
		getterMethod = "GetDateTime"
		if omitEmpty {
			getterMethod = "GetDateTimePointer"
		}
	case "geoPoint":
		goType = "pocketbase.GeoPoint"
		getterMethod = "Get"
	case "json":
		goType = "json.RawMessage"
		getterMethod = "GetRawMessage"
	case "relation", "file", "select":
		if isMulti {
			goType = "[]string"
			getterMethod = "GetStringSlice"
		} else { // Single selection/file/relation
			goType = "string"
			getterMethod = "GetString"
			if omitEmpty {
				getterMethod = "GetStringPointer"
			}
		}
	default:
		goType = "any"
		getterMethod = "Get"
	}

	// Finally decide whether to apply pointer type (exclude if already pointer, slice, json.RawMessage, or any)
	if omitEmpty && !strings.HasPrefix(goType, "[]") && goType != "json.RawMessage" && goType != "any" && !strings.HasPrefix(goType, "*") {
		goType = "*" + goType
	}

	return goType, getterMethod
}

// ToPascalCase converts a string to PascalCase format.
// It handles common abbreviations like ID, URL, HTML, JSON properly.
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	for i, part := range parts {
		switch strings.ToLower(part) {
		case "id":
			parts[i] = "ID"
		case "url":
			parts[i] = "URL"
		case "html":
			parts[i] = "HTML"
		case "json":
			parts[i] = "JSON"
		default:
			// 수동으로 첫 글자를 대문자로 변환 (strings.Title 대체)
			if len(part) > 0 {
				parts[i] = strings.ToUpper(part[:1]) + part[1:]
			}
		}
	}

	return strings.Join(parts, "")
}

// BuildJSONTag returns a preformatted struct tag string for a JSON field.
func BuildJSONTag(name string, omitEmpty bool) string {
	if omitEmpty {
		return fmt.Sprintf("`json:\"%s,omitempty\"`", name)
	}
	return fmt.Sprintf("`json:\"%s\"`", name)
}

// ToConstantName converts a value to a valid Go constant name
func ToConstantName(collectionName, fieldName, value string) string {
	// Base name: CollectionFieldValue
	baseName := ToPascalCase(collectionName) + ToPascalCase(fieldName)

	// Convert value to PascalCase first, then clean
	cleanValue := ToPascalCase(value)

	return baseName + cleanValue
}
