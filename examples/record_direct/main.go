package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	// PocketBase 클라이언트를 초기화합니다
	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))

	// 데이터 수정 권한을 위해 관리자(또는 사용자)로 인증합니다
	if _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password123"); err != nil {
		log.Fatalf("인증 실패: %v", err)
	}

	// posts 컬렉션을 위한 Record 서비스를 생성합니다
	recordsService := client.Records("posts")

	// --- 1. Record 객체를 직접 사용하여 새 레코드 생성 ---
	fmt.Println("--- Record 객체로 새 레코드 생성 ---")
	newRecord := &pocketbase.Record{}
	newRecord.Set("title", "Direct Record Post")
	newRecord.Set("content", "Record 객체를 직접 사용한 예제입니다!")

	createdRecord, err := recordsService.Create(context.Background(), newRecord)
	if err != nil {
		log.Fatalf("레코드 생성 실패: %v", err)
	}
	fmt.Printf("생성된 레코드 ID: %s, 제목: '%s'\n\n", createdRecord.ID, createdRecord.GetString("title"))

	// --- 2. 레코드 목록 조회 ---
	fmt.Println("--- 레코드 목록 조회 ---")
	records, err := recordsService.GetList(context.Background(), &pocketbase.ListOptions{
		PerPage: 10,
		Sort:    "-created", // 최신순 정렬
	})
	if err != nil {
		log.Fatalf("레코드 목록 조회 실패: %v", err)
	}
	fmt.Printf("총 %d개의 레코드를 찾았습니다.\n", records.TotalItems)
	for i, record := range records.Items {
		fmt.Printf("  %d: ID=%s, 제목='%s', 생성일=%s\n",
			i+1,
			record.ID,
			record.GetString("title"),
			record.GetDateTime("created").String())
	}
	fmt.Println()

	// --- 3. 단일 레코드 조회 ---
	fmt.Println("--- 단일 레코드 조회 ---")
	recordID := createdRecord.ID
	retrievedRecord, err := recordsService.GetOne(context.Background(), recordID, &pocketbase.GetOneOptions{
		Expand: "user", // 관련 레코드 확장 (있는 경우)
	})
	if err != nil {
		log.Fatalf("레코드 %s 조회 실패: %v", recordID, err)
	}
	fmt.Printf("조회된 레코드:\n")
	fmt.Printf("  ID: %s\n", retrievedRecord.ID)
	fmt.Printf("  제목: %s\n", retrievedRecord.GetString("title"))
	fmt.Printf("  내용: %s\n", retrievedRecord.GetString("content"))
	fmt.Printf("  생성일: %s\n", retrievedRecord.GetDateTime("created").String())
	fmt.Printf("  수정일: %s\n\n", retrievedRecord.GetDateTime("updated").String())

	// --- 4. Record 객체를 사용한 업데이트 ---
	fmt.Println("--- Record 객체로 레코드 업데이트 ---")
	updateRecord := &pocketbase.Record{}
	updateRecord.Set("title", "업데이트된 제목")
	updateRecord.Set("content", "Record 객체로 업데이트된 내용입니다.")

	updatedRecord, err := recordsService.Update(context.Background(), recordID, updateRecord)
	if err != nil {
		log.Fatalf("레코드 %s 업데이트 실패: %v", recordID, err)
	}
	fmt.Printf("업데이트된 레코드:\n")
	fmt.Printf("  제목: %s\n", updatedRecord.GetString("title"))
	fmt.Printf("  내용: %s\n\n", updatedRecord.GetString("content"))

	// --- 5. 필터링을 사용한 레코드 검색 ---
	fmt.Println("--- 필터링을 사용한 레코드 검색 ---")
	filteredRecords, err := recordsService.GetList(context.Background(), &pocketbase.ListOptions{
		Filter: "title ~ '업데이트'",   // 제목에 '업데이트'가 포함된 레코드 검색
		Fields: "id,title,created", // 특정 필드만 조회
	})
	if err != nil {
		log.Fatalf("필터링된 레코드 조회 실패: %v", err)
	}
	fmt.Printf("필터링 결과: %d개의 레코드\n", len(filteredRecords.Items))
	for _, record := range filteredRecords.Items {
		fmt.Printf("  ID: %s, 제목: %s\n", record.ID, record.GetString("title"))
	}
	fmt.Println()

	// --- 6. Record의 다양한 데이터 타입 사용 예제 ---
	fmt.Println("--- 다양한 데이터 타입 사용 예제 ---")
	complexRecord := &pocketbase.Record{}
	complexRecord.Set("title", "복합 데이터 예제")
	complexRecord.Set("content", "다양한 타입의 데이터를 포함합니다")
	complexRecord.Set("is_published", true)
	complexRecord.Set("view_count", 42)
	complexRecord.Set("rating", 4.5)
	complexRecord.Set("tags", []string{"golang", "pocketbase", "example"})

	complexCreated, err := recordsService.Create(context.Background(), complexRecord)
	if err != nil {
		log.Fatalf("복합 레코드 생성 실패: %v", err)
	}

	fmt.Printf("복합 레코드 생성됨:\n")
	fmt.Printf("  제목: %s\n", complexCreated.GetString("title"))
	fmt.Printf("  게시됨: %t\n", complexCreated.GetBool("is_published"))
	fmt.Printf("  조회수: %.0f\n", complexCreated.GetFloat("view_count"))
	fmt.Printf("  평점: %.1f\n", complexCreated.GetFloat("rating"))
	fmt.Printf("  태그: %v\n\n", complexCreated.GetStringSlice("tags"))

	// --- 7. 레코드 삭제 ---
	fmt.Println("--- 레코드 삭제 ---")
	if err := recordsService.Delete(context.Background(), recordID); err != nil {
		log.Fatalf("레코드 %s 삭제 실패: %v", recordID, err)
	}
	fmt.Printf("레코드 %s가 성공적으로 삭제되었습니다.\n", recordID)

	if err := recordsService.Delete(context.Background(), complexCreated.ID); err != nil {
		log.Fatalf("복합 레코드 %s 삭제 실패: %v", complexCreated.ID, err)
	}
	fmt.Printf("복합 레코드 %s가 성공적으로 삭제되었습니다.\n", complexCreated.ID)
}
