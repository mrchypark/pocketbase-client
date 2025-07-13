package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mrchypark/pocketbase-client"
	"github.com/pocketbase/pocketbase/tools/types"
)

func main() {
	client := pocketbase.NewClient("http://127.0.0.1:8090") // PocketBase 서버 URL

	ctx := context.Background()

	// 1. 관리자 계정으로 인증합니다.
	// 실제 환경에서는 환경 변수 등을 사용하여 안전하게 관리해야 합니다.
	if _, err := client.WithAdminPassword(ctx, "admin@example.com", "1q2w3e4r5t"); err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	fmt.Println("--- Admin authenticated successfully. ---")

	// --- RelatedCollection 레코드 생성 (AllTypes의 Relation Single/Multi 테스트를 위함) ---
	fmt.Println("\n--- Creating a RelatedCollection record for testing relations ---")
	newRelated := NewRelatedCollection()
	newRelated.SetName("Example Related Item")
	createdRelated, err := client.Records.Create(ctx, "related_collection", newRelated)
	if err != nil {
		log.Fatalf("Failed to create related record: %v", err)
	}
	fmt.Printf("Created RelatedCollection with ID: %s, Name: '%s'\n", createdRelated.ID, newRelated.Name())

	// --- 타입-세이프(Type-Safe) 모델을 사용한 AllTypes 레코드 생성 ---
	fmt.Println("\n--- Creating a new AllTypes record using generated types ---")

	newAllTypes := NewAllTypes() // 새로운 AllTypes 인스턴스 생성

	// Required 필드 설정
	newAllTypes.SetTextRequired("This is a required text.")
	newAllTypes.SetNumberRequired(123.45)
	newAllTypes.SetBoolRequired(true)
	newAllTypes.SetEmailRequired("test@example.com")
	newAllTypes.SetURLRequired("https://example.com")
	newAllTypes.SetDateRequired(types.NowDateTime())

	// 필수 select 필드에 스키마에 정의된 유효한 값 할당
	newAllTypes.SetSelectSingleRequired([]string{"a"})     // "a", "b", "c" 중 하나
	newAllTypes.SetSelectMultiRequired([]string{"a", "b"}) // "a", "b", "c" 중 하나 이상

	jsonContent := json.RawMessage(`{"key": "value", "number": 123}`)
	newAllTypes.SetJSONRequired(jsonContent)

	// Optional 필드 설정 (포인터 값을 사용)
	optionalText := "This is an optional text."
	newAllTypes.SetTextOptional(&optionalText)
	optionalNumber := 987.65
	newAllTypes.SetNumberOptional(&optionalNumber)
	optionalBool := false
	newAllTypes.SetBoolOptional(&optionalBool)
	optionalEmail := "optional@example.com"
	newAllTypes.SetEmailOptional(&optionalEmail)
	optionalURL := "https://optional.dev"
	newAllTypes.SetURLOptional(&optionalURL)
	optionalDate := types.NowDateTime().Add(24 * time.Hour)
	newAllTypes.SetDateOptional(&optionalDate)
	newAllTypes.SetSelectSingleOptional([]string{"x"}) // "x", "y", "z" 중 하나
	smo := []string{"y"}
	newAllTypes.SetSelectMultiOptional(&smo) // "x", "y", "z" 중 하나 이상
	optionalJSONContent := json.RawMessage(`{"another_key": "another_value"}`)
	newAllTypes.SetJSONOptional(optionalJSONContent)

	// File 및 Relation 필드 (예제에서는 실제 파일 업로드/ID 참조는 생략하고 빈 슬라이스 또는 생성된 ID 사용)
	// 실제 환경에서는 pocketbase-client의 File 관련 메서드를 사용해야 합니다.
	newAllTypes.SetRelationSingle([]string{createdRelated.ID}) // 위에서 생성한 RelatedCollection ID 참조
	newAllTypes.SetRelationMulti([]string{})                   // 여러 RelatedCollection ID들을 여기에 추가

	// 레코드 생성
	createdAllTypeRecord, err := client.Records.Create(ctx, "all_types", newAllTypes)
	if err != nil {
		log.Fatalf("Failed to create type-safe AllTypes record: %v", err)
	}

	// 생성된 레코드 확인 (타입-세이프 Getters 사용)
	createdAllTypes := ToAllTypes(createdAllTypeRecord)
	fmt.Printf("Created AllTypes record with ID: %s\n", createdAllTypes.ID)
	fmt.Printf("  TextRequired: '%s'\n", createdAllTypes.TextRequired())
	if txt := createdAllTypes.TextOptional(); txt != nil {
		fmt.Printf("  TextOptional: '%s'\n", *txt)
	}
	fmt.Printf("  NumberRequired: %f\n", createdAllTypes.NumberRequired())
	if num := createdAllTypes.NumberOptional(); num != nil {
		fmt.Printf("  NumberOptional: %f\n", *num)
	}
	fmt.Printf("  BoolRequired: %t\n", createdAllTypes.BoolRequired())
	if b := createdAllTypes.BoolOptional(); b != nil {
		fmt.Printf("  BoolOptional: %t\n", *b)
	}
	fmt.Printf("  EmailRequired: '%s'\n", createdAllTypes.EmailRequired())
	if email := createdAllTypes.EmailOptional(); email != nil {
		fmt.Printf("  EmailOptional: '%s'\n", *email)
	}
	fmt.Printf("  URLRequired: '%s'\n", createdAllTypes.URLRequired())
	if url := createdAllTypes.URLOptional(); url != nil {
		fmt.Printf("  URLOptional: '%s'\n", *url)
	}
	fmt.Printf("  DateRequired: '%s'\n", createdAllTypes.DateRequired().String())
	if dt := createdAllTypes.DateOptional(); dt != nil {
		fmt.Printf("  DateOptional: '%s'\n", dt.String())
	}
	fmt.Printf("  SelectSingleRequired: %v\n", createdAllTypes.SelectSingleRequired())
	fmt.Printf("  SelectSingleOptional: %v\n", createdAllTypes.SelectSingleOptional())
	fmt.Printf("  SelectMultiRequired: %v\n", createdAllTypes.SelectMultiRequired())
	fmt.Printf("  SelectMultiOptional: %v\n", createdAllTypes.SelectMultiOptional())
	fmt.Printf("  JSONRequired: %s\n", string(createdAllTypes.JSONRequired()))
	fmt.Printf("  JSONOptional: %s\n", string(createdAllTypes.JSONOptional()))
	fmt.Printf("  FileSingle (IDs): %v\n", createdAllTypes.FileSingle())
	fmt.Printf("  FileMulti (IDs): %v\n", createdAllTypes.FileMulti())
	fmt.Printf("  RelationSingle (IDs): %v\n", createdAllTypes.RelationSingle())
	fmt.Printf("  RelationMulti (IDs): %v\n", createdAllTypes.RelationMulti())

	// --- 타입-세이프(Type-Safe) 헬퍼를 사용한 AllTypes 목록 조회 ---
	fmt.Println("\n--- Listing AllTypes records using generated helper ---")

	allTypesCollection, err := GetAllTypesList(client.Records, &pocketbase.ListOptions{
		Page:    1,
		PerPage: 10,
		Sort:    "-created", // 최신 생성된 레코드부터 정렬
	})
	if err != nil {
		log.Fatalf("Failed to list AllTypes: %v", err)
	}

	fmt.Printf("Found %d AllTypes records. (Page %d/%d)\n", allTypesCollection.TotalItems, allTypesCollection.Page, allTypesCollection.TotalPages)
	for i, item := range allTypesCollection.Items {
		// 반복문 내에서도 타입-세이프 Getters를 사용해 안전하게 필드에 접근합니다.
		fmt.Printf("  [%d] ID: %s, TextRequired: '%s'\n", i+1, item.ID, item.TextRequired())
		// 더 많은 필드를 출력하려면 여기에 추가
	}

	// --- 단일 AllTypes 레코드 조회 ---
	fmt.Println("\n--- Fetching a single AllTypes record by ID ---")
	fetchedAllTypes, err := GetAllTypes(client.Records, createdAllTypeRecord.ID, nil)
	if err != nil {
		log.Fatalf("Failed to fetch single AllTypes record: %v", err)
	}
	fmt.Printf("Fetched AllTypes record (ID: %s) TextRequired: '%s'\n", fetchedAllTypes.ID, fetchedAllTypes.TextRequired())

	// --- 생성된 레코드 삭제 (클린업) ---
	fmt.Printf("\n--- Cleaning up created AllTypes record (ID: %s) ---\n", createdAllTypeRecord.ID)
	if err := client.Records.Delete(ctx, "all_types", createdAllTypeRecord.ID); err != nil {
		log.Printf("Failed to delete AllTypes record %s during cleanup: %v", createdAllTypeRecord.ID, err)
	} else {
		fmt.Println("AllTypes record cleanup complete.")
	}

	fmt.Printf("\n--- Cleaning up created RelatedCollection record (ID: %s) ---\n", createdRelated.ID)
	if err := client.Records.Delete(ctx, "related_collection", createdRelated.ID); err != nil {
		log.Printf("Failed to delete RelatedCollection record %s during cleanup: %v", createdRelated.ID, err)
	} else {
		fmt.Println("RelatedCollection record cleanup complete.")
	}

	fmt.Println("\n--- Example execution finished. ---")
}
