package generator

import (
	"reflect"
	"testing"
)

func TestEnumGenerator_GenerateEnums(t *testing.T) {
	generator := NewEnumGenerator()

	// 테스트용 데이터 준비
	collections := []CollectionData{
		{
			CollectionName: "devices",
		},
		{
			CollectionName: "users",
		},
	}

	schemas := []CollectionSchema{
		{
			Name: "devices",
			Fields: []FieldSchema{
				{
					Name:   "status",
					Type:   "select",
					System: false,
					Options: &FieldOptions{
						Values: []string{"active", "inactive", "pending"},
					},
				},
				{
					Name:   "type",
					Type:   "select",
					System: false,
					Options: &FieldOptions{
						Values: []string{"sensor", "actuator"},
					},
				},
				{
					Name:   "name",
					Type:   "text",
					System: false,
				},
			},
		},
		{
			Name: "users",
			Fields: []FieldSchema{
				{
					Name:   "role",
					Type:   "select",
					System: false,
					Options: &FieldOptions{
						Values: []string{"admin", "user", "guest"},
					},
				},
				{
					Name:   "id",
					Type:   "text",
					System: true, // System field should be ignored
				},
			},
		},
	}

	result := generator.GenerateEnums(collections, schemas)

	// 예상 결과: 3개의 enum (devices.status, devices.type, users.role)
	expectedCount := 3
	if len(result) != expectedCount {
		t.Errorf("Expected %d enums, got %d", expectedCount, len(result))
	}

	// devices.status enum 검증
	var deviceStatusEnum *EnumData
	for i := range result {
		if result[i].CollectionName == "devices" && result[i].FieldName == "status" {
			deviceStatusEnum = &result[i]
			break
		}
	}

	if deviceStatusEnum == nil {
		t.Fatal("devices.status enum not found")
	}

	expectedConstants := []ConstantData{
		{Name: "DevicesStatusActive", Value: "active"},
		{Name: "DevicesStatusInactive", Value: "inactive"},
		{Name: "DevicesStatusPending", Value: "pending"},
	}

	if !reflect.DeepEqual(deviceStatusEnum.Constants, expectedConstants) {
		t.Errorf("devices.status constants mismatch.\nGot: %+v\nWant: %+v",
			deviceStatusEnum.Constants, expectedConstants)
	}
}

func TestNewEnumGenerator(t *testing.T) {
	generator := NewEnumGenerator()
	if generator == nil {
		t.Error("NewEnumGenerator() returned nil")
	}
}
