package generator

import (
	"fmt"
	"strings"
)

// EnumGenerator handles enum constant generation for select fields
type EnumGenerator struct{}

// NewEnumGenerator creates a new EnumGenerator instance
func NewEnumGenerator() *EnumGenerator {
	return &EnumGenerator{}
}

// GenerateEnums generates enum data for all collections
func (g *EnumGenerator) GenerateEnums(collections []CollectionData, schemas []CollectionSchema) []EnumData {
	// Performance optimization: pre-allocate slice with estimated size
	estimatedEnums := 0
	for _, schema := range schemas {
		for _, field := range schema.Fields {
			if !field.System && field.Type == "select" {
				estimatedEnums++
			}
		}
	}

	enums := make([]EnumData, 0, estimatedEnums)

	// Performance optimization: convert schemas to map for O(1) lookup
	schemaMap := make(map[string]CollectionSchema, len(schemas))
	for _, schema := range schemas {
		schemaMap[schema.Name] = schema
	}

	for _, collection := range collections {
		schema, exists := schemaMap[collection.CollectionName]
		if !exists {
			continue
		}

		for _, field := range schema.Fields {
			if field.System || field.Type != "select" {
				continue
			}

			// Performance optimization: directly extract enum values to eliminate unnecessary function calls
			if field.Options != nil && len(field.Options.Values) > 0 {
				enumData := g.generateEnumDataOptimized(field, collection.CollectionName)
				enums = append(enums, enumData)
			}
		}
	}

	return enums
}

// GenerateEnumData generates enum data for a single select field
func (g *EnumGenerator) GenerateEnumData(enhanced EnhancedFieldInfo, collectionName string) EnumData {
	constants := make([]ConstantData, 0, len(enhanced.EnumValues))

	for _, value := range enhanced.EnumValues {
		constantName := ToConstantName(collectionName, enhanced.Name, value)
		constants = append(constants, ConstantData{
			Name:  constantName,
			Value: value,
		})
	}

	return EnumData{
		CollectionName: collectionName,
		FieldName:      enhanced.Name,
		EnumTypeName:   enhanced.EnumTypeName,
		Constants:      constants,
	}
}

// GenerateEnumConstants generates individual enum constants for a field
func (g *EnumGenerator) GenerateEnumConstants(field EnhancedFieldInfo, collectionName string) []ConstantData {
	if len(field.EnumValues) == 0 {
		return nil
	}

	constants := make([]ConstantData, 0, len(field.EnumValues))

	for _, value := range field.EnumValues {
		constantName := ToConstantName(collectionName, field.Name, value)
		constants = append(constants, ConstantData{
			Name:  constantName,
			Value: value,
		})
	}

	return constants
}

// GenerateEnumHelperFunction generates helper function code for enum values
func (g *EnumGenerator) GenerateEnumHelperFunction(enumData EnumData) string {
	functionName := enumData.EnumTypeName + "Values"

	var values []string
	for _, constant := range enumData.Constants {
		values = append(values, constant.Name)
	}

	return fmt.Sprintf(`// %s returns all possible values for %s
func %s() []string {
	return []string{%s}
}`, functionName, enumData.EnumTypeName, functionName, strings.Join(values, ", "))
}

// GenerateEnumValidationFunction generates validation function code for enum values
func (g *EnumGenerator) GenerateEnumValidationFunction(enumData EnumData) string {
	functionName := "IsValid" + enumData.EnumTypeName

	var cases []string
	for _, constant := range enumData.Constants {
		cases = append(cases, fmt.Sprintf(`case %s:`, constant.Name))
	}

	return fmt.Sprintf(`// %s checks if the given value is a valid %s
func %s(value string) bool {
	switch value {
	%s
		return true
	default:
		return false
	}
}`, functionName, enumData.EnumTypeName, functionName, strings.Join(cases, "\n\t"))
}

// ValidateEnumName ensures the enum name is a valid Go identifier
func (g *EnumGenerator) ValidateEnumName(name string) string {
	// Remove any invalid characters and ensure it starts with a letter
	cleaned := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, name)

	// Ensure it starts with a letter
	if len(cleaned) > 0 && cleaned[0] >= '0' && cleaned[0] <= '9' {
		cleaned = "Enum" + cleaned
	}

	if cleaned == "" {
		cleaned = "EnumValue"
	}

	return cleaned
}

// SanitizeConstantValue sanitizes a constant value for safe Go code generation
func (g *EnumGenerator) SanitizeConstantValue(value string) string {
	// Escape special characters in string literals
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, "\r", `\r`)
	value = strings.ReplaceAll(value, "\t", `\t`)

	return fmt.Sprintf(`"%s"`, value)
}

// generateEnumDataOptimized performance-optimized enum data generation function
func (g *EnumGenerator) generateEnumDataOptimized(field FieldSchema, collectionName string) EnumData {
	values := field.Options.Values
	constants := make([]ConstantData, 0, len(values))

	// Performance optimization: use string builder to minimize memory allocation
	var nameBuilder strings.Builder
	basePrefix := ToPascalCase(collectionName) + ToPascalCase(field.Name)

	for _, value := range values {
		nameBuilder.Reset()
		nameBuilder.WriteString(basePrefix)
		nameBuilder.WriteString(ToPascalCase(value))

		constants = append(constants, ConstantData{
			Name:  nameBuilder.String(),
			Value: value,
		})
	}

	return EnumData{
		CollectionName: collectionName,
		FieldName:      field.Name,
		EnumTypeName:   basePrefix + "Type",
		Constants:      constants,
	}
}
