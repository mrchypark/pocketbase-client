package generator

import (
	"os"
	"testing"
)

func TestLoadSchema(t *testing.T) {
	// Create a dummy schema file for testing
	dummySchema := `[
	{
		"id": "_pb_users_auth_",
		"name": "users",
		"type": "auth",
		"system": true,
		"fields": [
		{
			"id": "_pb_users_auth_email",
			"name": "email",
			"type": "email",
			"required": true,
			"unique": true,
			"options": {
				"exceptDomains": null,
				"onlyDomains": null
			}
		},
		{
			"id": "_pb_users_auth_password",
			"name": "password",
			"type": "text",
			"required": true,
			"options": {
				"min": 8,
				"max": 72,
				"pattern": ""
			}
		}
		],
		"indexes": [],
		"listRule": null,
		"viewRule": null,
		"createRule": null,
		"updateRule": null,
		"deleteRule": null,
		"options": {
			"allowEmailAuth": true,
			"allowOAuth2Auth": true,
			"allowUsernameAuth": true,
			"exceptEmailDomains": null,
			"manageRule": null,
			"minPasswordLength": 8,
			"onlyEmailDomains": null,
			"requireEmailVerification": false,
			"requireOriginal": false,
			"tokenDuration": 3600,
			"autoVerification": true
		}
	},
	{
		"id": "_pb_users_auth_test",
		"name": "posts",
		"type": "base",
		"system": false,
		"fields": [
		{
			"id": "_pb_users_auth_title",
			"name": "title",
			"type": "text",
			"required": true,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		},
		{
			"id": "_pb_users_auth_content",
			"name": "content",
			"type": "editor",
			"required": false,
			"options": {
				"convertUrls": false
			}
		}
		],
		"indexes": [],
		"listRule": null,
		"viewRule": null,
		"createRule": null,
		"updateRule": null,
		"deleteRule": null,
		"options": {}
	}
	]
	`

	testFilePath := "test_schema.json"
	err := os.WriteFile(testFilePath, []byte(dummySchema), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy schema file: %v", err)
	}
	defer os.Remove(testFilePath)

	// Test successful parsing
	schemas, err := LoadSchema(testFilePath)
	if err != nil {
		t.Fatalf("LoadSchema failed: %v", err)
	}

	if len(schemas) != 2 {
		t.Errorf("Expected 2 schemas, got %d", len(schemas))
	}

	// Verify content of the first schema (users)
	userSchema := schemas[0]
	if userSchema.Name != "users" {
		t.Errorf("Expected first schema name to be 'users', got %s", userSchema.Name)
	}
	if userSchema.Type != "auth" {
		t.Errorf("Expected first schema type to be 'auth', got %s", userSchema.Type)
	}
	if len(userSchema.Fields) != 2 {
		t.Errorf("Expected 2 fields in user schema, got %d", len(userSchema.Fields))
	}

	// Verify content of the first field in user schema (email)
	userEmailField := userSchema.Fields[0]
	if userEmailField.Name != "email" {
		t.Errorf("Expected user email field name to be 'email', got %s", userEmailField.Name)
	}
	if userEmailField.Type != "email" {
		t.Errorf("Expected user email field type to be 'email', got %s", userEmailField.Type)
	}
	if !userEmailField.Required {
		t.Errorf("Expected user email field to be required")
	}

	// Test invalid JSON
	invalidJsonPath := "invalid.json"
	err = os.WriteFile(invalidJsonPath, []byte(`{"name": "test"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}
	defer os.Remove(invalidJsonPath)

	_, err = LoadSchema(invalidJsonPath)
	// Expect an error because the root is not an array
	if err == nil {
		t.Errorf("Expected an error for invalid JSON, but got none")
	}

	// Test non-existent file
	_, err = LoadSchema("non_existent.json")
	if err == nil {
		t.Errorf("Expected an error for non-existent file, but got none")
	}
}
