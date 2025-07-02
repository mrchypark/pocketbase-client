package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

// newRandomID는 PocketBase 규칙에 맞는 15자리의 랜덤 ID를 생성합니다. (소문자 + 숫자)
func newRandomID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 15
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			// 이 에러는 거의 발생하지 않으므로, 발생 시 panic 처리
			panic(fmt.Errorf("랜덤 ID 생성 중 심각한 오류 발생: %w", err))
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret)
}

func main() {
	// --- 1. 클라이언트 초기화 및 인증 ---
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	ctx := context.Background()

	if _, err := client.HealthCheck(ctx); err != nil {
		log.Fatalf("❌ PocketBase 서버에 연결할 수 없습니다: %v\n   서버가 http://127.0.0.1:8090 에서 실행 중인지 확인하세요.", err)
	}

	if _, err := client.AuthenticateAsAdmin(ctx, "admin@example.com", "1q2w3e4r5t"); err != nil {
		log.Fatalf("❌ 관리자 인증에 실패했습니다: %v", err)
	}
	fmt.Println("✅ 관리자 인증 성공!")
	fmt.Println("------------------------------------")

	// --- 2. 첫 번째 일괄 작업: 레코드 5개 생성 ---
	fmt.Println("🚀 [Phase 1] 레코드 5개 일괄 생성을 시작합니다...")
	var createRequests []*pocketbase.BatchRequest
	for i := 1; i <= 5; i++ {
		req, err := client.Records.NewCreateRequest("posts", map[string]any{
			"title": fmt.Sprintf("새 포스트 #%d", i),
		})
		if err != nil {
			log.Fatalf("생성 요청 %d 실패: %v", i, err)
		}
		createRequests = append(createRequests, req)
	}

	createResults, err := client.Batch.Execute(ctx, createRequests)
	if err != nil {
		log.Fatalf("❌ 첫 번째 일괄 처리 API 실행 실패: %v", err)
	}

	var createdIDs []string
	fmt.Println("\n✨ 생성 결과:")
	for i, res := range createResults {
		// [수정됨] res.StatusCode -> res.Status
		if res.Status == http.StatusOK {
			// [수정됨] res.Data -> res.Body
			data, ok := res.Body.(map[string]interface{})
			if ok {
				id, _ := data["id"].(string)
				createdIDs = append(createdIDs, id)
				fmt.Printf("   - ✅ 요청 %d 성공: 새 레코드 ID '%s' 생성\n", i+1, id)
			}
		} else {
			// [수정됨] res.Error 대신 res.ParsedError.Message 사용
			errMsg := "알 수 없는 오류"
			if res.ParsedError != nil {
				errMsg = res.ParsedError.Message
			}
			fmt.Printf("   - 🚨 요청 %d 실패: 상태 코드 %d, 에러: %s\n", i+1, res.Status, errMsg)
		}
	}
	fmt.Println("------------------------------------")

	if len(createdIDs) < 4 {
		log.Fatalf("❌ 수정 및 삭제를 진행하기에 충분한 레코드가 생성되지 않았습니다.")
	}

	// --- 3. 두 번째 일괄 작업: 3개 수정, 1개 삭제 ---
	fmt.Println("\n🚀 [Phase 2] 생성된 ID를 사용하여 수정(3개) 및 삭제(1개)를 시작합니다...")
	var nextRequests []*pocketbase.BatchRequest

	for i := 0; i < 3; i++ {
		req, err := client.Records.NewUpdateRequest("posts", createdIDs[i], map[string]any{
			"title": fmt.Sprintf("수정된 포스트 #%d (ID: %s)", i+1, createdIDs[i]),
		})
		if err != nil {
			log.Fatalf("수정 요청 %d 실패: %v", i+1, err)
		}
		nextRequests = append(nextRequests, req)
	}

	deleteReq, err := client.Records.NewDeleteRequest("posts", createdIDs[3])
	if err != nil {
		log.Fatalf("삭제 요청 실패: %v", err)
	}
	nextRequests = append(nextRequests, deleteReq)

	updateDeleteResults, err := client.Batch.Execute(ctx, nextRequests)
	if err != nil {
		log.Fatalf("❌ 두 번째 일괄 처리 API 실행 실패: %v", err)
	}

	fmt.Println("\n✨ 수정/삭제 결과:")
	for i, res := range updateDeleteResults {
		// [수정됨] res.StatusCode -> res.Status
		if res.Status >= http.StatusBadRequest {
			errMsg := "알 수 없는 오류"
			if res.ParsedError != nil {
				errMsg = res.ParsedError.Message
			}
			fmt.Printf("   - 🚨 요청 %d 실패: 상태 코드 %d, 에러: %s\n", i+1, res.Status, errMsg)
		} else {
			fmt.Printf("   - ✅ 요청 %d 성공: 상태 코드 %d\n", i+1, res.Status)
		}
	}
	fmt.Println("------------------------------------")

	// --- 4. 세 번째 일괄 작업: 1초 간격으로 Upsert 5회 수행 ---
	fmt.Println("\n🚀 [Phase 3] 1초 간격으로 랜덤 ID Upsert 5회를 시작합니다...")
	randomID := newRandomID()

	for i := 1; i <= 5; i++ {

		upsertReq, err := client.Records.NewUpsertRequest("posts", map[string]any{
			"id":    randomID,
			"title": fmt.Sprintf("Upsert 포스트 (시도 #%d)", i),
		})
		if err != nil {
			log.Fatalf("Upsert 요청 생성 실패: %v", err)
		}

		fmt.Printf("\n[시도 %d/5] ID '%s'로 Upsert 요청...\n", i, randomID)
		upsertResults, err := client.Batch.Execute(ctx, []*pocketbase.BatchRequest{upsertReq})
		if err != nil {
			log.Printf("   - 🚨 Upsert API 실행 실패: %v\n", err)
		} else if len(upsertResults) > 0 {
			res := upsertResults[0]
			// [수정됨] res.StatusCode -> res.Status
			if res.Status == http.StatusCreated {
				fmt.Printf("   - ✅ 성공: 새 레코드가 생성되었습니다 (상태 코드: %d)\n", res.Status)
			} else if res.Status == http.StatusOK {
				fmt.Printf("   - ✅ 성공: 기존 레코드가 수정되었습니다 (상태 코드: %d)\n", res.Status)
			} else {
				errMsg := "알 수 없는 오류"
				if res.ParsedError != nil {
					errMsg = res.ParsedError.Message
				}
				fmt.Printf("   - 🚨 실패: 상태 코드 %d, 에러: %s\n", res.Status, errMsg)
			}
		}

		time.Sleep(1 * time.Second)
	}
	fmt.Println("\n🎉 모든 작업이 완료되었습니다.")
}
