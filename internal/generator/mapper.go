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

	// Expect field.Options.MaxSelect to be correctly parsed as *int.
	if field.Options != nil && field.Options.MaxSelect != nil {
		if *field.Options.MaxSelect != 1 {
			isMulti = true
		}
	} else if field.Type == "relation" || field.Type == "file" || field.Type == "select" {
		// When MaxSelect is nil (when maxSelect is missing or null in schema),
		// relation/file/select types are considered multi by default.
		isMulti = true
	}

	switch field.Type {
	case "text", "email", "url", "editor":
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
		goType = "types.DateTime"
		getterMethod = "GetDateTime"
		if omitEmpty {
			getterMethod = "GetDateTimePointer"
		}
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

// BuildToMapBlock returns the ToMap assignment block for a field.
func BuildToMapBlock(jsonName, goName string, omitEmpty bool) string {
	if omitEmpty {
		return fmt.Sprintf("\tif val := m.%s(); val != nil {\n\t\tdata[\"%s\"] = val\n\t}\n", goName, jsonName)
	}
	return fmt.Sprintf("\t// For required fields, we always include them.\n\t// You can add more complex logic here if needed, e.g., checking for zero values.\n\tdata[\"%s\"] = m.%s()\n", jsonName, goName)
}

// BuildValueOrBlock returns the ValueOr method block for pointer fields.
func BuildValueOrBlock(structName, goName, jsonName, baseType string, isPointer bool) string {
	if !isPointer {
		return ""
	}
	return fmt.Sprintf("\n// %sValueOr returns the value of the '%s' field or the provided default value if nil.\nfunc (m *%s) %sValueOr(defaultValue %s) %s {\n\tif val := m.%s(); val != nil {\n\t\treturn *val\n\t}\n\treturn defaultValue\n}\n",
		goName, jsonName, structName, goName, baseType, baseType, goName)
}

// AnalyzeEnhancedField analyzes a field and returns enhanced information for code generation
func AnalyzeEnhancedField(field FieldSchema, collectionName string, allCollections []CollectionSchema) EnhancedFieldInfo {
	enhanced := EnhancedFieldInfo{
		FieldSchema: field,
	}

	switch field.Type {
	case "select":
		enhanced = analyzeSelectField(enhanced, collectionName)
	case "relation":
		enhanced = analyzeRelationField(enhanced, collectionName, allCollections)
	case "file":
		enhanced = analyzeFileField(enhanced, collectionName)
	}

	return enhanced
}

// analyzeSelectField analyzes select field for enum generation
func analyzeSelectField(enhanced EnhancedFieldInfo, collectionName string) EnhancedFieldInfo {
	if enhanced.Options != nil && len(enhanced.Options.Values) > 0 {
		enhanced.EnumValues = enhanced.Options.Values
		enhanced.EnumTypeName = ToPascalCase(collectionName) + ToPascalCase(enhanced.Name) + "Type"
	}
	return enhanced
}

// analyzeRelationField analyzes relation field for relation type generation
func analyzeRelationField(enhanced EnhancedFieldInfo, collectionName string, allCollections []CollectionSchema) EnhancedFieldInfo {
	if enhanced.Options != nil && enhanced.Options.CollectionID != "" {
		// Find target collection by ID
		for _, col := range allCollections {
			if col.ID == enhanced.Options.CollectionID {
				enhanced.TargetCollection = col.Name
				break
			}
		}

		if enhanced.TargetCollection != "" {
			enhanced.RelationTypeName = ToPascalCase(enhanced.TargetCollection) + "Relation"

			// Check if it's multi-relation
			if enhanced.Options.MaxSelect != nil && *enhanced.Options.MaxSelect > 1 {
				enhanced.IsMultiRelation = true
			}
		}
	}
	return enhanced
}

// analyzeFileField analyzes file field for file type generation
func analyzeFileField(enhanced EnhancedFieldInfo, _ string) EnhancedFieldInfo {
	enhanced.FileTypeName = ToPascalCase(enhanced.Name) + "File"

	if enhanced.Options != nil {
		// Check if it's multi-file
		if enhanced.Options.MaxSelect != nil && *enhanced.Options.MaxSelect > 1 {
			enhanced.IsMultiFile = true
		}

		// Check for thumbnails
		if len(enhanced.Options.Thumbs) > 0 {
			enhanced.HasThumbnails = true
			enhanced.ThumbnailSizes = enhanced.Options.Thumbs
		}
	}

	return enhanced
}

// ToConstantName converts a value to a valid Go constant name
func ToConstantName(collectionName, fieldName, value string) string {
	// Base name: CollectionFieldValue
	baseName := ToPascalCase(collectionName) + ToPascalCase(fieldName)

	// Convert value to PascalCase first, then clean
	cleanValue := ToPascalCase(value)

	return baseName + cleanValue
}
