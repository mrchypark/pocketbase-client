package generator

import (
	"encoding/json"
	"fmt" // 에러 메시지 포맷팅을 위해 필요
)

// CollectionSchema 정의 (이전과 동일)
type CollectionSchema struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	System     bool     `json:"system"`
	Indexes    []string `json:"indexes"`
	ListRule   *string  `json:"listRule"`
	ViewRule   *string  `json:"viewRule"`
	CreateRule *string  `json:"createRule"`
	UpdateRule *string  `json:"updateRule"`
	DeleteRule *string  `json:"deleteRule"`
	Options    *struct {
		Query *string `json:"query"`
	} `json:"options"`

	Schema []FieldSchema `json:"schema"` // `json:"schema"` 태그 유지
	Fields []FieldSchema `json:"fields"` // `json:"fields"` 태그 유지
}

// CollectionSchema.UnmarshalJSON 메서드 (이전과 동일하게 유지)
// 이 메서드는 CollectionSchema의 'schema' 또는 'fields' 배열을 cs.Fields로 언마샬합니다.
func (cs *CollectionSchema) UnmarshalJSON(data []byte) error {
	type Alias CollectionSchema
	aux := &struct {
		SchemaRaw json.RawMessage `json:"schema"`
		FieldsRaw json.RawMessage `json:"fields"`
		*Alias
	}{
		Alias: (*Alias)(cs),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.SchemaRaw) > 0 && string(aux.SchemaRaw) != "null" {
		if err := json.Unmarshal(aux.SchemaRaw, &cs.Fields); err != nil {
			return err
		}
	} else if len(aux.FieldsRaw) > 0 && string(aux.FieldsRaw) != "null" {
		if err := json.Unmarshal(aux.FieldsRaw, &cs.Fields); err != nil {
			return err
		}
	}

	cs.Schema = nil // Fields 필드에 데이터가 채워졌으므로 Schema 필드는 비워둡니다.

	return nil
}

// FieldSchema 정의 (새로운 커스텀 UnmarshalJSON 포함)
type FieldSchema struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	System      bool   `json:"system"`
	Required    bool   `json:"required"`
	Presentable bool   `json:"presentable"`
	Unique      bool   `json:"unique"`
	Hidden      bool   `json:"hidden"`

	// options 필드는 FieldOptions 타입으로 유지 (포인터)
	Options *FieldOptions `json:"options"`

	// 필드 레벨에 직접 있는 maxSelect, minSelect를 위한 RawMessage 필드 (임시 파싱용)
	// 이 필드들은 FieldSchema.UnmarshalJSON 내부에서만 사용됩니다.
	MinSelectRaw json.RawMessage `json:"minSelect"` // schema에 따라 필드 레벨에 있을 수 있음
	MaxSelectRaw json.RawMessage `json:"maxSelect"` // schema에 따라 필드 레벨에 있을 수 있음
}

// UnmarshalJSON은 FieldSchema를 위한 커스텀 언마샬링 로직입니다.
// 이 메서드는 'options' 객체 내부 또는 필드 레벨에 직접 있는 minSelect/maxSelect를 모두 처리합니다.
func (fs *FieldSchema) UnmarshalJSON(data []byte) error {
	// 무한 재귀를 피하기 위해 임시 구조체를 사용하여 기본 필드와 원시 'options' 데이터를 언마샬합니다.
	type Alias FieldSchema // FieldSchema의 다른 모든 필드를 포함하는 별칭
	aux := &struct {
		OptionsRaw json.RawMessage `json:"options"` // 원시 'options' 객체를 json.RawMessage로 캡처
		*Alias
	}{
		Alias: (*Alias)(fs), // fs 인스턴스에 다른 필드를 자동으로 바인딩
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// fs.Options 초기화
	fs.Options = &FieldOptions{}

	// aux.OptionsRaw (원시 'options' 객체)가 존재하면 먼저 언마샬합니다.
	if len(aux.OptionsRaw) > 0 && string(aux.OptionsRaw) != "null" {
		if err := json.Unmarshal(aux.OptionsRaw, fs.Options); err != nil {
			return fmt.Errorf("failed to unmarshal FieldOptions from raw options: %w", err)
		}
	}

	// 이제 필드 레벨에 직접 있는 MaxSelectRaw/MinSelectRaw를 확인하고,
	// 이미 fs.Options에 값이 설정되지 않은 경우에만 덮어씁니다 (혹은 우선순위를 줍니다).
	// PocketBase 스키마에서는 `options` 내부의 값이 더 명시적일 수 있으므로,
	// `options` 내부의 값을 우선하고, 필드 레벨의 값을 보조적으로 사용합니다.
	// 여기서는 필드 레벨의 값을 무조건 덮어쓰도록 하겠습니다.
	if len(aux.MinSelectRaw) > 0 && string(aux.MinSelectRaw) != "null" {
		var val int
		if err := json.Unmarshal(aux.MinSelectRaw, &val); err == nil {
			fs.Options.MinSelect = &val // 필드 레벨의 값을 우선
		}
	}
	if len(aux.MaxSelectRaw) > 0 && string(aux.MaxSelectRaw) != "null" {
		var val int
		if err := json.Unmarshal(aux.MaxSelectRaw, &val); err == nil {
			fs.Options.MaxSelect = &val // 필드 레벨의 값을 우선
		}
	}

	return nil
}

// FieldOptions (단순 구조체로, MaxSelect/MinSelect는 *int 타입 유지, 커스텀 UnmarshalJSON 없음)
type FieldOptions struct {
	CollectionID  string `json:"collectionId"`
	CascadeDelete bool   `json:"cascadeDelete"`

	Min json.RawMessage `json:"min"`
	Max json.RawMessage `json:"max"`

	MinSelect *int `json:"minSelect"` // *int로 유지, FieldSchema.UnmarshalJSON에서 처리
	MaxSelect *int `json:"maxSelect"` // *int로 유지, FieldSchema.UnmarshalJSON에서 처리

	Pattern string `json:"pattern"`

	MimeTypes []string `json:"mimeTypes"`
	Thumbs    []string `json:"thumbs"`
	MaxSize   int      `json:"maxSize"`
	Protected bool     `json:"protected"`
	Values    []string `json:"values"`
}

// FieldOptions에 대한 UnmarshalJSON 메서드는 더 이상 필요하지 않습니다 (삭제됨).
// FieldSchema.UnmarshalJSON이 모든 파싱을 담당합니다.
