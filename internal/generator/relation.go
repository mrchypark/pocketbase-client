package generator

import (
	"fmt"
	"strings"
)

// RelationGenerator handles relation type generation for relation fields
type RelationGenerator struct{}

// NewRelationGenerator creates a new RelationGenerator instance
func NewRelationGenerator() *RelationGenerator {
	return &RelationGenerator{}
}

// GenerateRelationTypes generates relation type data for all collections
func (g *RelationGenerator) GenerateRelationTypes(collections []CollectionData, schemas []CollectionSchema) []RelationTypeData {
	// Performance optimization: pre-allocate slice with estimated size
	estimatedRelations := 0
	for _, schema := range schemas {
		for _, field := range schema.Fields {
			if !field.System && field.Type == "relation" {
				estimatedRelations++
			}
		}
	}

	relationTypes := make([]RelationTypeData, 0, estimatedRelations)
	seenRelations := make(map[string]bool) // Map for duplicate removal

	// Performance optimization: convert schemas to map for O(1) lookup
	schemaMap := make(map[string]CollectionSchema, len(schemas))
	collectionIDMap := make(map[string]string, len(schemas)) // ID -> Name mapping
	for _, schema := range schemas {
		schemaMap[schema.Name] = schema
		if schema.ID != "" {
			collectionIDMap[schema.ID] = schema.Name
		}
	}

	for _, collection := range collections {
		schema, exists := schemaMap[collection.CollectionName]
		if !exists {
			continue
		}

		for _, field := range schema.Fields {
			if field.System || field.Type != "relation" {
				continue
			}

			// Performance optimization: directly extract relation information
			if field.Options != nil && field.Options.CollectionID != "" {
				targetCollection, exists := collectionIDMap[field.Options.CollectionID]
				if exists {
					relationType := g.generateRelationTypeDataOptimized(field, collection.CollectionName, targetCollection)
					// Duplicate removal: check if same type name already exists
					if !seenRelations[relationType.TypeName] {
						seenRelations[relationType.TypeName] = true
						relationTypes = append(relationTypes, relationType)
					}
				}
			}
		}
	}

	return relationTypes
}

// GenerateRelationTypeData generates relation type data for a single relation field
func (g *RelationGenerator) GenerateRelationTypeData(enhanced EnhancedFieldInfo, collectionName string) RelationTypeData {
	methods := g.GenerateRelationMethods(enhanced)

	return RelationTypeData{
		TypeName:         enhanced.RelationTypeName,
		TargetCollection: enhanced.TargetCollection,
		TargetTypeName:   ToPascalCase(enhanced.TargetCollection),
		IsMulti:          enhanced.IsMultiRelation,
		Methods:          methods,
	}
}

// GenerateRelationMethods generates methods for a relation type
func (g *RelationGenerator) GenerateRelationMethods(enhanced EnhancedFieldInfo) []MethodData {
	var methods []MethodData

	// ID method
	idMethod := MethodData{
		Name:       "ID",
		ReturnType: "string",
		Body:       "return r.id",
	}
	methods = append(methods, idMethod)

	// Load method
	targetTypeName := ToPascalCase(enhanced.TargetCollection)
	loadMethod := MethodData{
		Name:       "Load",
		ReturnType: fmt.Sprintf("(*%s, error)", targetTypeName),
		Body: fmt.Sprintf(`if r.id == "" {
		return nil, nil
	}
	return Get%s(client, r.id, nil)`, targetTypeName),
	}
	methods = append(methods, loadMethod)

	// IsEmpty method
	isEmptyMethod := MethodData{
		Name:       "IsEmpty",
		ReturnType: "bool",
		Body:       `return r.id == ""`,
	}
	methods = append(methods, isEmptyMethod)

	return methods
}

// GenerateRelationTypeCode generates the complete Go code for a relation type
func (g *RelationGenerator) GenerateRelationTypeCode(relationType RelationTypeData) string {
	var code strings.Builder

	// Type definition
	code.WriteString(fmt.Sprintf("// %s represents a relation to %s collection\n",
		relationType.TypeName, relationType.TargetCollection))
	code.WriteString(fmt.Sprintf("type %s struct {\n", relationType.TypeName))
	code.WriteString("\tid string\n")
	code.WriteString("}\n\n")

	// Methods
	for _, method := range relationType.Methods {
		code.WriteString(g.GenerateMethodCode(relationType.TypeName, method))
		code.WriteString("\n")
	}

	// Constructor
	code.WriteString(g.GenerateConstructorCode(relationType))

	return code.String()
}

// GenerateMethodCode generates Go code for a single method
func (g *RelationGenerator) GenerateMethodCode(typeName string, method MethodData) string {
	var code strings.Builder

	// Method signature
	if method.Name == "Load" {
		code.WriteString(fmt.Sprintf("// %s fetches the related %s record\n",
			method.Name, strings.TrimSuffix(typeName, "Relation")))
		code.WriteString(fmt.Sprintf("func (r %s) %s(ctx context.Context, client pocketbase.RecordServiceAPI) %s {\n",
			typeName, method.Name, method.ReturnType))
	} else {
		code.WriteString(fmt.Sprintf("// %s returns the %s\n",
			method.Name, strings.ToLower(method.Name)))
		code.WriteString(fmt.Sprintf("func (r %s) %s() %s {\n",
			typeName, method.Name, method.ReturnType))
	}

	// Method body
	lines := strings.Split(method.Body, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			code.WriteString(fmt.Sprintf("\t%s\n", line))
		}
	}

	code.WriteString("}\n")

	return code.String()
}

// GenerateConstructorCode generates constructor function for relation type
func (g *RelationGenerator) GenerateConstructorCode(relationType RelationTypeData) string {
	return fmt.Sprintf(`// New%s creates a new %s
func New%s(id string) %s {
	return %s{id: id}
}`, relationType.TypeName, relationType.TypeName, relationType.TypeName,
		relationType.TypeName, relationType.TypeName)
}

// GenerateMultiRelationTypeCode generates code for multi-relation types
func (g *RelationGenerator) GenerateMultiRelationTypeCode(relationType RelationTypeData) string {
	if !relationType.IsMulti {
		return g.GenerateRelationTypeCode(relationType)
	}

	var code strings.Builder

	// Multi-relation type definition
	multiTypeName := relationType.TypeName + "s" // Pluralize
	code.WriteString(fmt.Sprintf("// %s represents multiple relations to %s collection\n",
		multiTypeName, relationType.TargetCollection))
	code.WriteString(fmt.Sprintf("type %s []%s\n\n", multiTypeName, relationType.TypeName))

	// Multi-relation methods
	code.WriteString(g.GenerateMultiRelationMethods(relationType, multiTypeName))

	return code.String()
}

// GenerateMultiRelationMethods generates methods for multi-relation types
func (g *RelationGenerator) GenerateMultiRelationMethods(relationType RelationTypeData, multiTypeName string) string {
	var code strings.Builder

	// IDs method
	code.WriteString(fmt.Sprintf(`// IDs returns all relation IDs
func (r %s) IDs() []string {
	ids := make([]string, len(r))
	for i, rel := range r {
		ids[i] = rel.ID()
	}
	return ids
}

`, multiTypeName))

	// LoadAll method
	targetTypeName := relationType.TargetTypeName
	code.WriteString(fmt.Sprintf(`// LoadAll fetches all related %s records
func (r %s) LoadAll(ctx context.Context, client pocketbase.RecordServiceAPI) ([]*%s, error) {
	if len(r) == 0 {
		return nil, nil
	}
	
	var results []*%s
	for _, rel := range r {
		record, err := rel.Load(ctx, client)
		if err != nil {
			return nil, err
		}
		if record != nil {
			results = append(results, record)
		}
	}
	return results, nil
}

`, targetTypeName, multiTypeName, targetTypeName, targetTypeName))

	// IsEmpty method
	code.WriteString(fmt.Sprintf(`// IsEmpty returns true if there are no relations
func (r %s) IsEmpty() bool {
	return len(r) == 0
}

`, multiTypeName))

	return code.String()
}

// ValidateRelationName ensures the relation name is a valid Go identifier
func (g *RelationGenerator) ValidateRelationName(name string) string {
	// Remove any invalid characters and ensure it starts with a letter
	cleaned := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, name)

	// Ensure it starts with a letter
	if len(cleaned) > 0 && cleaned[0] >= '0' && cleaned[0] <= '9' {
		cleaned = "Relation" + cleaned
	}

	if cleaned == "" {
		cleaned = "RelationType"
	}

	return cleaned
}

// generateRelationTypeDataOptimized performance-optimized relation type data generation function
func (g *RelationGenerator) generateRelationTypeDataOptimized(field FieldSchema, _, targetCollection string) RelationTypeData {
	// Performance optimization: use pre-calculated values
	relationTypeName := ToPascalCase(targetCollection) + "Relation"
	targetTypeName := ToPascalCase(targetCollection)

	// Check if it's a multiple relationship
	isMulti := field.Options.MaxSelect != nil && *field.Options.MaxSelect > 1

	// Generate methods (using pre-allocated slice)
	methods := make([]MethodData, 0, 3)

	// ID method
	methods = append(methods, MethodData{
		Name:       "ID",
		ReturnType: "string",
		Body:       "return r.id",
	})

	// Load method
	methods = append(methods, MethodData{
		Name:       "Load",
		ReturnType: fmt.Sprintf("(*%s, error)", targetTypeName),
		Body: fmt.Sprintf(`if r.id == "" {
		return nil, nil
	}
	return Get%s(client, r.id, nil)`, targetTypeName),
	})

	// IsEmpty method
	methods = append(methods, MethodData{
		Name:       "IsEmpty",
		ReturnType: "bool",
		Body:       `return r.id == ""`,
	})

	return RelationTypeData{
		TypeName:         relationTypeName,
		TargetCollection: targetCollection,
		TargetTypeName:   targetTypeName,
		IsMulti:          isMulti,
		Methods:          methods,
	}
}
