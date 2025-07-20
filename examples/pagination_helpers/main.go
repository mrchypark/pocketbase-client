package main

import (
	"context"
	"fmt"
	"log"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	// PocketBase 클라이언트 생성
	client := pocketbase.NewClient("http://localhost:8090")

	// 관리자 인증 (예제용)
	ctx := context.Background()
	_, err := client.WithAdminPassword(ctx, "admin@example.com", "password")
	if err != nil {
		log.Printf("인증 실패 (예제는 계속 진행): %v", err)
	}

	fmt.Println("=== PocketBase 페이지네이션 헬퍼 사용법 예제 ===")

	// 1. GetAll 메서드 사용 예제
	demonstrateGetAll(ctx, client)

	// 2. GetAllWithBatchSize 메서드 사용 예제
	demonstrateGetAllWithBatchSize(ctx, client)

	// 3. Iterator 패턴 사용 예제
	demonstrateIterator(ctx, client)

	// 4. Iterator with BatchSize 사용 예제
	demonstrateIteratorWithBatchSize(ctx, client)

	// 5. 에러 처리 예제
	demonstrateErrorHandling(ctx, client)

	// 6. 필터링과 정렬 옵션 사용 예제
	demonstrateWithOptions(ctx, client)
}

// GetAll 메서드 기본 사용법
func demonstrateGetAll(ctx context.Context, client *pocketbase.Client) {
	fmt.Println("1. GetAll 메서드 기본 사용법")
	fmt.Println("--------------------------------")

	// 모든 레코드를 한 번에 가져오기
	records, err := client.Records.GetAll(ctx, "posts", nil)
	if err != nil {
		fmt.Printf("에러 발생: %v\n", err)
		return
	}

	fmt.Printf("총 %d개의 레코드를 가져왔습니다.\n", len(records))

	// 처음 3개 레코드 출력
	for i, record := range records {
		if i >= 3 {
			fmt.Println("...")
			break
		}
		fmt.Printf("  - ID: %s, Created: %s\n", record.ID, record.GetString("created"))
	}
	fmt.Println()
}

// GetAllWithBatchSize 메서드 사용법
func demonstrateGetAllWithBatchSize(ctx context.Context, client *pocketbase.Client) {
	fmt.Println("2. GetAllWithBatchSize 메서드 사용법")
	fmt.Println("------------------------------------")

	// 배치 크기를 50으로 설정하여 모든 레코드 가져오기
	records, err := client.Records.GetAllWithBatchSize(ctx, "posts", nil, 50)
	if err != nil {
		fmt.Printf("에러 발생: %v\n", err)
		return
	}

	fmt.Printf("배치 크기 50으로 총 %d개의 레코드를 가져왔습니다.\n", len(records))
	fmt.Println()
}

// Iterator 패턴 기본 사용법
func demonstrateIterator(ctx context.Context, client *pocketbase.Client) {
	fmt.Println("3. Iterator 패턴 기본 사용법")
	fmt.Println("-----------------------------")

	// Iterator 생성
	iterator := client.Records.Iterate(ctx, "posts", nil)

	count := 0
	fmt.Println("Iterator를 사용하여 레코드 순회:")

	// 레코드 순회
	for iterator.Next() {
		record := iterator.Record()
		count++

		// 처음 5개만 출력
		if count <= 5 {
			fmt.Printf("  %d. ID: %s, Created: %s\n", count, record.ID, record.GetString("created"))
		}

		// 예제를 위해 10개만 처리
		if count >= 10 {
			break
		}
	}

	// 에러 확인
	if err := iterator.Error(); err != nil {
		fmt.Printf("Iterator 에러: %v\n", err)
		return
	}

	fmt.Printf("Iterator로 총 %d개의 레코드를 처리했습니다.\n", count)
	fmt.Println()
}

// Iterator with BatchSize 사용법
func demonstrateIteratorWithBatchSize(ctx context.Context, client *pocketbase.Client) {
	fmt.Println("4. Iterator with BatchSize 사용법")
	fmt.Println("----------------------------------")

	// 배치 크기 25로 Iterator 생성
	iterator := client.Records.IterateWithBatchSize(ctx, "posts", nil, 25)

	count := 0
	fmt.Println("배치 크기 25로 Iterator 사용:")

	for iterator.Next() {
		_ = iterator.Record()
		count++

		// 메모리 효율적인 처리 예제
		if count%25 == 0 {
			fmt.Printf("  %d개 레코드 처리 완료...\n", count)
		}

		// 예제를 위해 100개만 처리
		if count >= 100 {
			break
		}
	}

	if err := iterator.Error(); err != nil {
		fmt.Printf("Iterator 에러: %v\n", err)
		return
	}

	fmt.Printf("배치 처리로 총 %d개의 레코드를 처리했습니다.\n", count)
	fmt.Println()
}

// 에러 처리 예제
func demonstrateErrorHandling(ctx context.Context, client *pocketbase.Client) {
	fmt.Println("5. 에러 처리 예제")
	fmt.Println("------------------")

	// 존재하지 않는 컬렉션으로 테스트
	records, err := client.Records.GetAll(ctx, "nonexistent_collection", nil)
	if err != nil {
		// PaginationError 타입 확인
		if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
			fmt.Printf("페이지네이션 에러 발생:\n")
			fmt.Printf("  - 작업: %s\n", paginationErr.Operation)
			fmt.Printf("  - 페이지: %d\n", paginationErr.Page)
			fmt.Printf("  - 부분 데이터: %d개\n", len(paginationErr.GetPartialData()))
			fmt.Printf("  - 원본 에러: %v\n", paginationErr.OriginalErr)

			// 부분 데이터가 있으면 사용
			if len(paginationErr.GetPartialData()) > 0 {
				fmt.Printf("부분 데이터 %d개를 사용합니다.\n", len(paginationErr.GetPartialData()))
				records = paginationErr.GetPartialData()
			}
		} else {
			fmt.Printf("일반 에러: %v\n", err)
		}
	} else {
		fmt.Printf("성공적으로 %d개의 레코드를 가져왔습니다.\n", len(records))
	}
	fmt.Println()
}

// 필터링과 정렬 옵션 사용 예제
func demonstrateWithOptions(ctx context.Context, client *pocketbase.Client) {
	fmt.Println("6. 필터링과 정렬 옵션 사용 예제")
	fmt.Println("--------------------------------")

	// ListOptions 설정
	options := &pocketbase.ListOptions{
		Filter: "status = 'published'",
		Sort:   "-created",
		Expand: "author",
	}

	// 필터링된 모든 레코드 가져오기
	records, err := client.Records.GetAll(ctx, "posts", options)
	if err != nil {
		fmt.Printf("에러 발생: %v\n", err)
		return
	}

	fmt.Printf("필터링된 %d개의 게시물을 가져왔습니다.\n", len(records))

	// Iterator로도 동일한 옵션 사용 가능
	fmt.Println("\nIterator로 필터링된 데이터 처리:")
	iterator := client.Records.Iterate(ctx, "posts", options)

	count := 0
	for iterator.Next() {
		record := iterator.Record()
		count++

		if count <= 3 {
			fmt.Printf("  %d. ID: %s\n", count, record.ID)
		}

		if count >= 10 {
			break
		}
	}

	if err := iterator.Error(); err != nil {
		fmt.Printf("Iterator 에러: %v\n", err)
		return
	}

	fmt.Printf("Iterator로 %d개의 필터링된 레코드를 처리했습니다.\n", count)
	fmt.Println()
}
