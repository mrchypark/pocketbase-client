package generator

import "encoding/json"

// CollectionSchema는 두 가지 JSON 스키마 형식을 모두 처리할 수 있는 통합 구조체입니다.
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

	// 두 형식을 아우르기 위한 필드
	Schema []FieldSchema `json:"schema"`
	Fields []FieldSchema `json:"fields"`
}

// FieldSchema는 필드 정보를 정의합니다.
type FieldSchema struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	System      bool          `json:"system"`
	Required    bool          `json:"required"`
	Presentable bool          `json:"presentable"`
	Unique      bool          `json:"unique"`
	Hidden      bool          `json:"hidden"`
	Options     *FieldOptions `json:"options"`
}

// FieldOptions는 다양한 타입의 옵션을 처리합니다.
type FieldOptions struct {
	CollectionID  string `json:"collectionId"`
	CascadeDelete bool   `json:"cascadeDelete"`

	// int 또는 string이 될 수 있는 필드를 위해 json.RawMessage 사용
	Min json.RawMessage `json:"min"`
	Max json.RawMessage `json:"max"`

	MinSelect *int `json:"minSelect"`
	MaxSelect *int `json:"maxSelect"`

	Pattern string `json:"pattern"`

	// 기타 옵션들...
	MimeTypes []string `json:"mimeTypes"`
	Thumbs    []string `json:"thumbs"`
	MaxSize   int      `json:"maxSize"`
	Protected bool     `json:"protected"`
	Values    []string `json:"values"`
}

// UnmarshalJSON은 `CollectionSchema`를 위한 커스텀 파싱 로직입니다.
// 'schema' 또는 'fields' 프로퍼티를 'Fields'로 통합합니다.
func (cs *CollectionSchema) UnmarshalJSON(data []byte) error {
	// 재귀 호출을 피하기 위해 type alias 사용
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
			// 오류가 발생하면 반환하여 파싱 실패를 알림
			return err
		}
	} else if len(aux.FieldsRaw) > 0 && string(aux.FieldsRaw) != "null" {
		if err := json.Unmarshal(aux.FieldsRaw, &cs.Fields); err != nil {
			return err
		}
	}

	// Unmarshal 후 원래 struct의 Schema 필드는 필요 없으므로 비워줍니다.
	cs.Schema = nil

	return nil
}
