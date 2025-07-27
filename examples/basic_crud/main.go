package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

// Post는 컬렉션 스키마와 일치하는 구조체를 정의합니다.
// `json` 태그는 직렬화에 중요합니다.
type Post struct {
	pocketbase.BaseModel
	Title   string `json:"title"`
	Content string `json:"content"`
}

func main() {
	// PocketBase 클라이언트를 초기화합니다
	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))

	// 데이터 수정 권한을 위해 관리자(또는 사용자)로 인증합니다
	if _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password123"); err != nil {
		log.Fatalf("인증 실패: %v", err)
	}

	// posts 컬렉션을 위한 제네릭 레코드 서비스를 생성합니다
	postsService := pocketbase.NewRecordService[Post](client, "posts")

	// --- 1. 새 레코드 생성 ---
	fmt.Println("--- 새 레코드 생성 ---")
	newPost := &Post{
		Title:   "My First Post",
		Content: "Hello from the Go SDK!",
	}
	createdRecord, err := postsService.Create(context.Background(), newPost)
	if err != nil {
		log.Fatalf("레코드 생성 실패: %v", err)
	}
	fmt.Printf("생성된 레코드 ID: %s, 제목: '%s'\n\n", createdRecord.ID, createdRecord.Title)

	// --- 2. 레코드 목록 조회 ---
	fmt.Println("--- 레코드 목록 조회 ---")
	records, err := postsService.GetList(context.Background(), &pocketbase.ListOptions{})
	if err != nil {
		log.Fatalf("레코드 목록 조회 실패: %v", err)
	}
	fmt.Printf("총 %d개의 레코드를 찾았습니다.\n", records.TotalItems)
	for i, record := range records.Items {
		fmt.Printf("  %d: ID=%s, 제목='%s'\n", i+1, record.ID, record.Title)
	}
	fmt.Println()

	// --- 3. 단일 레코드 조회 ---
	fmt.Println("--- 단일 레코드 조회 ---")
	recordID := createdRecord.ID
	retrievedRecord, err := postsService.GetOne(context.Background(), recordID, nil)
	if err != nil {
		log.Fatalf("레코드 %s 조회 실패: %v", recordID, err)
	}
	fmt.Printf("조회된 레코드 제목: '%s', 내용: '%s'\n\n", retrievedRecord.Title, retrievedRecord.Content)

	// --- 4. 레코드 업데이트 ---
	fmt.Println("--- 레코드 업데이트 ---")
	updatePost := &Post{
		BaseModel: pocketbase.BaseModel{ID: recordID},
		Title:     "My Updated Post Title",
		Content:   retrievedRecord.Content, // 기존 내용 유지
	}
	updatedRecord, err := postsService.Update(context.Background(), recordID, updatePost)
	if err != nil {
		log.Fatalf("레코드 %s 업데이트 실패: %v", recordID, err)
	}
	fmt.Printf("레코드 제목이 '%s'로 업데이트되었습니다.\n\n", updatedRecord.Title)

	// --- 5. 레코드 삭제 ---
	fmt.Println("--- 레코드 삭제 ---")
	if err := postsService.Delete(context.Background(), recordID); err != nil {
		log.Fatalf("레코드 %s 삭제 실패: %v", recordID, err)
	}
	fmt.Printf("레코드 %s가 성공적으로 삭제되었습니다.\n", recordID)
}
