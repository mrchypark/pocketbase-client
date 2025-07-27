package main

import (
	"context"
	"fmt"
	"log"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

// Post는 블로그 포스트를 나타내는 사용자 정의 타입입니다.
type Post struct {
	pocketbase.BaseModel
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

// User는 사용자를 나타내는 사용자 정의 타입입니다.
type User struct {
	pocketbase.BaseModel
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

func main() {
	// PocketBase 클라이언트 생성
	client := pocketbase.NewClient("http://127.0.0.1:8090")

	ctx := context.Background()

	// 예제 1: Post 타입을 위한 제네릭 서비스 생성
	fmt.Println("=== 제네릭 RecordService[T] 사용 예제 ===")
	postService := pocketbase.NewRecordService[Post](client, "posts")

	// 새로운 포스트 생성
	newPost := &Post{
		Title:   "제네릭 서비스 소개",
		Content: "PocketBase Go Client의 새로운 제네릭 RecordService[T]를 소개합니다.",
		Author:  "개발자",
	}

	createdPost, err := postService.Create(ctx, newPost, nil)
	if err != nil {
		log.Printf("포스트 생성 실패: %v", err)
	} else {
		fmt.Printf("생성된 포스트: %+v\n", createdPost)
	}

	// 포스트 목록 조회 (타입 안전)
	listOptions := &pocketbase.ListOptions{
		Page:    1,
		PerPage: 10,
		Sort:    "-created",
		Filter:  "author = '개발자'",
	}

	posts, err := postService.GetList(ctx, listOptions)
	if err != nil {
		log.Printf("포스트 목록 조회 실패: %v", err)
	} else {
		fmt.Printf("조회된 포스트 수: %d\n", len(posts.Items))
		for _, post := range posts.Items {
			fmt.Printf("- %s (작성자: %s)\n", post.Title, post.Author)
		}
	}

	// 예제 2: User 타입을 위한 제네릭 서비스 생성
	fmt.Println("\n=== 다중 타입 지원 예제 ===")
	userService := pocketbase.NewRecordService[User](client, "users")

	// 새로운 사용자 생성
	newUser := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Name:     "테스트 사용자",
	}

	createdUser, err := userService.Create(ctx, newUser, nil)
	if err != nil {
		log.Printf("사용자 생성 실패: %v", err)
	} else {
		fmt.Printf("생성된 사용자: %+v\n", createdUser)
	}

	// 예제 3: Admin 타입을 위한 제네릭 서비스 (기존 AdminService와 비교)
	fmt.Println("\n=== Admin 타입 제네릭 서비스 예제 ===")
	adminService := pocketbase.NewRecordService[pocketbase.Admin](client, "admins")

	// Admin 목록 조회 (제네릭 방식)
	adminListOptions := &pocketbase.ListOptions{
		Page:    1,
		PerPage: 5,
	}

	admins, err := adminService.GetList(ctx, adminListOptions)
	if err != nil {
		log.Printf("관리자 목록 조회 실패: %v", err)
	} else {
		fmt.Printf("조회된 관리자 수: %d\n", len(admins.Items))
		for _, admin := range admins.Items {
			fmt.Printf("- %s (%s)\n", admin.Email, admin.ID)
		}
	}

	// 예제 4: 타입 안전성 시연
	fmt.Println("\n=== 타입 안전성 시연 ===")
	demonstrateTypeSafety(postService, userService)

	// 예제 5: 기존 API와의 호환성
	fmt.Println("\n=== 기존 API와의 호환성 ===")
	demonstrateBackwardCompatibility(client, ctx)
}

// demonstrateTypeSafety는 제네릭 서비스의 타입 안전성을 시연합니다.
func demonstrateTypeSafety(postService *pocketbase.RecordService[Post], userService *pocketbase.RecordService[User]) {
	ctx := context.Background()

	// Post 서비스는 Post 타입만 반환
	post, err := postService.GetOne(ctx, "example_id", nil)
	if err == nil {
		// post는 *Post 타입으로 컴파일 타임에 보장됨
		fmt.Printf("포스트 제목: %s\n", post.Title) // 타입 안전한 필드 접근
	}

	// User 서비스는 User 타입만 반환
	user, err := userService.GetOne(ctx, "example_id", nil)
	if err == nil {
		// user는 *User 타입으로 컴파일 타임에 보장됨
		fmt.Printf("사용자 이름: %s\n", user.Username) // 타입 안전한 필드 접근
	}

	fmt.Println("✓ 컴파일 타임 타입 안전성 보장")
}

// demonstrateBackwardCompatibility는 기존 API와의 호환성을 시연합니다.
func demonstrateBackwardCompatibility(client *pocketbase.Client, ctx context.Context) {
	// 기존 방식 (여전히 지원됨)
	recordService := &pocketbase.RecordServiceLegacy{Client: client}

	// 기존 GetListAs 함수 사용 (deprecated이지만 여전히 작동)
	posts, err := pocketbase.GetListAs[Post](ctx, client, "posts", &pocketbase.ListOptions{
		Page:    1,
		PerPage: 5,
	})
	if err != nil {
		log.Printf("기존 GetListAs 함수 사용 실패: %v", err)
	} else {
		fmt.Printf("기존 API로 조회된 포스트 수: %d\n", len(posts.Items))
	}

	// 기존 GetOneAs 함수 사용 (deprecated이지만 여전히 작동)
	post, err := pocketbase.GetOneAs[Post](ctx, client, "posts", "example_id", nil)
	if err == nil {
		fmt.Printf("기존 API로 조회된 포스트: %s\n", post.Title)
	}

	// 기존 CreateAs 함수 사용 (deprecated이지만 여전히 작동)
	newPost := &Post{
		Title:   "기존 API 테스트",
		Content: "기존 CreateAs 함수를 사용한 포스트",
		Author:  "테스터",
	}

	createdPost, err := pocketbase.CreateAs[Post](ctx, recordService, "posts", newPost, nil)
	if err == nil {
		fmt.Printf("기존 API로 생성된 포스트: %s\n", createdPost.Title)
	}

	fmt.Println("✓ 기존 API와의 완전한 하위 호환성 유지")
}
