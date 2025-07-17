package generator

import (
	"strings"
)

// MapPbTypeToGoType maps PocketBase field types to Go types and their corresponding getter methods.
func MapPbTypeToGoType(field FieldSchema, omitEmpty bool) (string, string, string) {
	var goType, comment, getterMethod string

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
		goType = "interface{}"
		getterMethod = "Get"
	}

	// Finally decide whether to apply pointer type (exclude if already pointer, slice, json.RawMessage, or interface{})
	if omitEmpty && !strings.HasPrefix(goType, "[]") && goType != "json.RawMessage" && goType != "interface{}" && !strings.HasPrefix(goType, "*") {
		goType = "*" + goType
	}

	return goType, comment, getterMethod
}

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
			parts[i] = strings.Title(part)
		}
	}

	return strings.Join(parts, "")
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
func analyzeFileField(enhanced EnhancedFieldInfo, collectionName string) EnhancedFieldInfo {
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
