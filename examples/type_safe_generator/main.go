package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mrchypark/pocketbase-client"
)

func main() {
	// 클라이언트를 초기화합니다.
	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))
	ctx := context.Background()

	// 관리자 계정으로 인증합니다. (실제 환경에 맞게 수정해주세요)
	if _, err := client.WithAdminPassword(ctx, "admin@example.com", "password123"); err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	fmt.Println("Admin authenticated successfully.")

	// --- 타입-세이프(Type-Safe) 모델을 사용한 레코드 생성 ---
	fmt.Println("\n--- Creating a new record using generated types ---")

	// 1. pbc-gen으로 생성된 NewPosts() 함수를 사용해 새로운 Post 인스턴스를 생성합니다.
	// 이 인스턴스는 pocketbase.Record를 내장하고 있어, 레코드의 모든 필드에 접근할 수 있습니다.
	newPost := NewPosts() // 빈 레코드로 초기화

	// 2. 생성된 타입-세이프 Setters를 사용해 데이터를 설정합니다.
	// 이렇게 하면 필드 이름에 오타가 발생하는 것을 방지하고, 타입 안정성을 보장합니다.
	title := "My Type-Safe Post"
	newPost.SetTitle(&title) // 포인터 타입의 Setter 사용

	// 3. Records.Create 메서드에 생성된 인스턴스를 전달합니다.
	// newPost는 Mappable 인터페이스를 구현하고 있으므로, 자동으로 map[string]any 형태로 변환됩니다.
	createdRecord, err := client.Records.Create(ctx, "posts", newPost)
	if err != nil {
		log.Fatalf("Failed to create type-safe record: %v", err)
	}

	// 4. 생성된 레코드의 필드에 접근할 때는 타입-세이프 Getters를 사용합니다.
	createdPost := ToPosts(createdRecord)
	fmt.Printf("Created record with ID: %s, Title: '%s'\n", createdPost.ID, *createdPost.Title())

	// --- 타입-세이프(Type-Safe) 헬퍼를 사용한 목록 조회 ---
	fmt.Println("\n--- Listing records using generated helper ---")

	// 5. pbc-gen으로 생성된 GetPostsList 헬퍼 함수를 사용해 레코드 목록을 가져옵니다.
	// 이 함수는 내부적으로 client.Records.GetList를 호출하고, 결과를 []*Posts 타입으로 변환해줍니다.
	postsCollection, err := GetPostsList(client.Records, &pocketbase.ListOptions{
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		log.Fatalf("Failed to list posts: %v", err)
	}

	fmt.Printf("Found %d posts. (Page %d/%d)\n", postsCollection.TotalItems, postsCollection.Page, postsCollection.TotalPages)
	for _, post := range postsCollection.Items {
		// 6. 반복문 내에서도 타입-세이프 Getters를 사용해 안전하게 필드에 접근합니다.
		fmt.Printf("  - ID: %s, Title: '%s'\n", post.ID, *post.Title())
	}

	// --- 생성된 레코드 삭제 ---
	fmt.Printf("\n--- Cleaning up created record (ID: %s) ---\n", createdRecord.ID)
	if err := client.Records.Delete(ctx, "posts", createdRecord.ID); err != nil {
		log.Printf("Failed to delete record %s during cleanup: %v", createdRecord.ID, err)
	} else {
		fmt.Println("Cleanup complete.")
	}
}
