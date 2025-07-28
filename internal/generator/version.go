package generator

import "encoding/json"

// DetectPocketBaseVersion detects PocketBase version based on schema structure.
// Returns true if it's PocketBase 0.22+ (uses "schema" field), false if newer (uses "fields" field).
func DetectPocketBaseVersion(schemas []CollectionSchema) bool {
	if len(schemas) == 0 {
		return false // Default to newer version
	}

	// Check the first collection's raw JSON structure
	// We need to check the original JSON to see which field was used
	for _, schema := range schemas {
		// If Schema field is populated, it means the original JSON used "schema"
		// But since UnmarshalJSON clears Schema field, we need a different approach

		// For now, we'll use a heuristic: check if any collection has created/updated fields
		// In 0.22+, these are not included in schema, in newer versions they might be
		hasCreatedUpdated := false
		for _, field := range schema.Fields {
			if field.Name == "created" || field.Name == "updated" {
				hasCreatedUpdated = true
				break
			}
		}

		// If no created/updated fields found, likely 0.22+ version
		if !hasCreatedUpdated {
			return true // 0.22+ version
		}
	}

	return false // Newer version
}

// DetectPocketBaseVersionFromRaw detects version from raw JSON data.
// This is more accurate as it checks the actual JSON structure.
func DetectPocketBaseVersionFromRaw(data []byte) (bool, error) {
	var rawSchemas []map[string]interface{}
	if err := json.Unmarshal(data, &rawSchemas); err != nil {
		return false, err
	}

	if len(rawSchemas) == 0 {
		return false, nil
	}

	// Check if the first collection uses "schema" field
	firstCollection := rawSchemas[0]
	_, hasSchema := firstCollection["schema"]
	_, hasFields := firstCollection["fields"]

	// If "schema" field exists, it's 0.22+
	// If "fields" field exists, it's newer
	if hasSchema && !hasFields {
		return true, nil // 0.22+ version
	}

	return false, nil // Newer version
}
