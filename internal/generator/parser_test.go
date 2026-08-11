package generator

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoadSchema(t *testing.T) {
	// Create a dummy schema file for testing (latest flattened format)
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
			"exceptDomains": null,
			"onlyDomains": null
		},
		{
			"id": "_pb_users_auth_password",
			"name": "password",
			"type": "text",
			"required": true,
			"pattern": ""
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
			"pattern": ""
		},
		{
			"id": "_pb_users_auth_content",
			"name": "content",
			"type": "editor",
			"required": false,
			"convertUrls": false
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
	err = os.WriteFile(invalidJsonPath, []byte(`{"foo": "bar"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}
	defer os.Remove(invalidJsonPath)

	_, err = LoadSchema(invalidJsonPath)
	// Expect an error because the object is not a recognizable collection shape
	if err == nil {
		t.Errorf("Expected an error for invalid JSON, but got none")
	}

	// Test non-existent file
	_, err = LoadSchema("non_existent.json")
	if err == nil {
		t.Errorf("Expected an error for non-existent file, but got none")
	}
}

func TestParseSchemas(t *testing.T) {
	collectionJSON := `{"id": "c1", "name": "posts", "type": "base", "system": false, "fields": [{"id": "f1", "name": "title", "type": "text", "required": true}]}`

	tests := []struct {
		name      string
		jsonData  string
		wantCount int
		wantName  string
		wantErr   bool
	}{
		{
			name:      "plain array of collections",
			jsonData:  `[` + collectionJSON + `]`,
			wantCount: 1,
			wantName:  "posts",
		},
		{
			name:      "paginated wrapper from GET /api/collections",
			jsonData:  `{"page": 1, "perPage": 30, "totalItems": 1, "totalPages": 1, "items": [` + collectionJSON + `]}`,
			wantCount: 1,
			wantName:  "posts",
		},
		{
			name:      "single collection object from GET /api/collections/{id}",
			jsonData:  collectionJSON,
			wantCount: 1,
			wantName:  "posts",
		},
		{
			name:      "paginated wrapper with empty items",
			jsonData:  `{"page": 1, "perPage": 30, "totalItems": 0, "totalPages": 0, "items": []}`,
			wantCount: 0,
		},
		{
			name:      "single object without collection markers",
			jsonData:  `{"foo": "bar"}`,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:     "legacy array with schema key is rejected",
			jsonData: `[{"id": "c1", "name": "posts", "schema": [{"name": "title", "type": "text"}]}]`,
			wantErr:  true,
		},
		{
			name:     "legacy single object with schema key is rejected",
			jsonData: `{"id": "c1", "name": "posts", "schema": [{"name": "title", "type": "text"}]}`,
			wantErr:  true,
		},
		{
			name:     "legacy wrapper with schema key is rejected",
			jsonData: `{"page": 1, "perPage": 30, "items": [{"id": "c1", "name": "posts", "schema": [{"name": "title", "type": "text"}]}]}`,
			wantErr:  true,
		},
		{
			name:     "invalid json",
			jsonData: `{"items": [` + collectionJSON,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemas, err := parseSchemas([]byte(tt.jsonData))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSchemas() error = %v", err)
			}
			if len(schemas) != tt.wantCount {
				t.Fatalf("got %d collections, want %d", len(schemas), tt.wantCount)
			}
			if tt.wantCount > 0 && schemas[0].Name != tt.wantName {
				t.Errorf("collection name = %q, want %q", schemas[0].Name, tt.wantName)
			}
		})
	}
}

func TestBuildTemplateData(t *testing.T) {
	schemas := []CollectionSchema{
		{
			Name:   "posts",
			System: false,
			Fields: []FieldSchema{
				{Name: "title", Type: "text", Required: true},
				{Name: "content", Type: "editor", Required: false},
				{Name: "user_id", Type: "relation", Required: true, System: true},
			},
		},
		{
			Name:   "_superusers", // Should be skipped
			System: true,
		},
	}

	pkgName := "testpkg"
	tplData := BuildTemplateData(schemas, pkgName)

	if tplData.PackageName != pkgName {
		t.Errorf("Expected package name %q, got %q", pkgName, tplData.PackageName)
	}

	if len(tplData.Collections) != 1 {
		t.Fatalf("Expected 1 collection, got %d", len(tplData.Collections))
	}

	postsCollection := tplData.Collections[0]
	if postsCollection.CollectionName != "posts" {
		t.Errorf("Expected collection name 'posts', got %q", postsCollection.CollectionName)
	}
	if postsCollection.StructName != "Posts" {
		t.Errorf("Expected struct name 'Posts', got %q", postsCollection.StructName)
	}

	if len(postsCollection.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(postsCollection.Fields))
	}

	titleField := postsCollection.Fields[0]
	if titleField.JSONName != "title" {
		t.Errorf("Expected field json name 'title', got %q", titleField.JSONName)
	}
	if titleField.GoName != "Title" {
		t.Errorf("Expected field go name 'Title', got %q", titleField.GoName)
	}
	if titleField.GoType != "string" {
		t.Errorf("Expected field go type 'string', got %q", titleField.GoType)
	}
	if titleField.OmitEmpty {
		t.Error("Expected OmitEmpty to be false for required field")
	}
}

func TestCollectionSchema_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		jsonData   string
		wantErr    bool
		wantFields int
	}{
		{
			name:       "with fields field (latest format)",
			jsonData:   `{"name": "test", "fields": [{"name": "field1"}, {"name": "field2"}]}`,
			wantErr:    false,
			wantFields: 2,
		},
		{
			name:       "with no fields",
			jsonData:   `{"name": "test"}`,
			wantErr:    false,
			wantFields: 0,
		},
		{
			name:       "with null fields",
			jsonData:   `{"name": "test", "fields": null}`,
			wantErr:    false,
			wantFields: 0,
		},
		{
			name:     "invalid json",
			jsonData: `{"name": "test", "fields": [{"name": "field1"}]`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cs CollectionSchema
			err := json.Unmarshal([]byte(tt.jsonData), &cs)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(cs.Fields) != tt.wantFields {
				t.Errorf("UnmarshalJSON() got %d fields, want %d", len(cs.Fields), tt.wantFields)
			}
		})
	}
}

func TestFieldSchema_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		jsonData      string
		wantErr       bool
		wantName      string
		wantType      string
		wantRequired  bool
		wantMaxSelect *int
		wantValues    []string
	}{
		{
			name:         "text field with flattened pattern",
			jsonData:     `{"name": "title", "type": "text", "required": true, "pattern": "^test$"}`,
			wantErr:      false,
			wantName:     "title",
			wantType:     "text",
			wantRequired: true,
		},
		{
			name:          "relation field with flattened options",
			jsonData:      `{"name": "author", "type": "relation", "required": true, "collectionId": "col123", "maxSelect": 1, "cascadeDelete": false}`,
			wantErr:       false,
			wantName:      "author",
			wantType:      "relation",
			wantRequired:  true,
			wantMaxSelect: intPtr(1),
		},
		{
			name:       "select field with values",
			jsonData:   `{"name": "status", "type": "select", "values": ["active", "inactive", "pending"]}`,
			wantErr:    false,
			wantName:   "status",
			wantType:   "select",
			wantValues: []string{"active", "inactive", "pending"},
		},
		{
			name:     "file field with thumbs",
			jsonData: `{"name": "cover", "type": "file", "mimeTypes": ["image/jpeg", "image/png"], "thumbs": ["100x100", "300x300"]}`,
			wantErr:  false,
			wantName: "cover",
			wantType: "file",
		},
		{
			name:          "maxSelect at field level",
			jsonData:      `{"name": "tags", "type": "relation", "maxSelect": 5}`,
			wantErr:       false,
			wantName:      "tags",
			wantType:      "relation",
			wantMaxSelect: intPtr(5),
		},
		{
			name:     "autodate with onCreate/onUpdate",
			jsonData: `{"name": "created", "type": "autodate", "onCreate": true, "onUpdate": false}`,
			wantErr:  false,
			wantName: "created",
			wantType: "autodate",
		},
		{
			name:     "email with domain restrictions",
			jsonData: `{"name": "email", "type": "email", "onlyDomains": ["example.com"], "exceptDomains": ["spam.com"]}`,
			wantErr:  false,
			wantName: "email",
			wantType: "email",
		},
		{
			name:     "number with onlyInt",
			jsonData: `{"name": "count", "type": "number", "onlyInt": true}`,
			wantErr:  false,
			wantName: "count",
			wantType: "number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fs FieldSchema
			err := json.Unmarshal([]byte(tt.jsonData), &fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if fs.Name != tt.wantName {
					t.Errorf("Name = %q, want %q", fs.Name, tt.wantName)
				}
				if fs.Type != tt.wantType {
					t.Errorf("Type = %q, want %q", fs.Type, tt.wantType)
				}
				if fs.Required != tt.wantRequired {
					t.Errorf("Required = %v, want %v", fs.Required, tt.wantRequired)
				}
				if tt.wantMaxSelect != nil && (fs.Options == nil || fs.Options.MaxSelect == nil || *fs.Options.MaxSelect != *tt.wantMaxSelect) {
					want := 0
					if tt.wantMaxSelect != nil {
						want = *tt.wantMaxSelect
					}
					got := 0
					if fs.Options != nil && fs.Options.MaxSelect != nil {
						got = *fs.Options.MaxSelect
					}
					t.Errorf("MaxSelect = %d, want %d", got, want)
				}
				if tt.wantValues != nil {
					if fs.Options == nil || len(fs.Options.Values) != len(tt.wantValues) {
						t.Errorf("Values = %v, want %v", fs.Options.Values, tt.wantValues)
					}
				}
			}
		})
	}
}
