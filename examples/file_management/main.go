// Package main demonstrates file management operations with PocketBase Go client.
// This example shows how to upload, download, generate URLs, and delete files
// associated with PocketBase records.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	// PocketBase 클라이언트 생성
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	ctx := context.Background()

	// 관리자 인증 (실제 환경에서는 환경변수 사용 권장)
	_, err := client.WithAdminPassword(ctx, "admin@example.com", "password123")
	if err != nil {
		log.Fatalf("관리자 인증 실패: %v", err)
	}

	fmt.Println("=== 파일 관리 예제 ===")

	// 1. 파일 업로드 예제
	fmt.Println("\n1. 파일 업로드")
	err = uploadFileExample(ctx, client)
	if err != nil {
		log.Printf("파일 업로드 실패: %v", err)
	}

	// 2. 파일 다운로드 예제
	fmt.Println("\n2. 파일 다운로드")
	err = downloadFileExample(ctx, client)
	if err != nil {
		log.Printf("파일 다운로드 실패: %v", err)
	}

	// 3. 파일 URL 생성 예제
	fmt.Println("\n3. 파일 URL 생성")
	fileURLExample(client)

	// 4. 파일 삭제 예제
	fmt.Println("\n4. 파일 삭제")
	err = deleteFileExample(ctx, client)
	if err != nil {
		log.Printf("파일 삭제 실패: %v", err)
	}
}

func uploadFileExample(ctx context.Context, client *pocketbase.Client) error {
	// 테스트용 파일 내용 생성
	fileContent := strings.NewReader("이것은 테스트 파일 내용입니다.")

	// 먼저 레코드를 생성합니다 (posts 컬렉션이 있다고 가정)
	recordData := map[string]interface{}{
		"title":   "파일 업로드 테스트",
		"content": "파일이 첨부된 게시물입니다.",
	}

	record, err := client.Records.Create(ctx, "posts", recordData)
	if err != nil {
		return fmt.Errorf("레코드 생성 실패: %w", err)
	}

	fmt.Printf("레코드 생성됨: ID = %s\n", record.ID)

	// 파일을 레코드의 'image' 필드에 업로드
	updatedRecord, err := client.Files.Upload(ctx, "posts", record.ID, "image", "test.txt", fileContent)
	if err != nil {
		return fmt.Errorf("파일 업로드 실패: %w", err)
	}

	fmt.Printf("파일 업로드 성공: %s\n", updatedRecord.ID)
	if imageField := updatedRecord.Get("image"); imageField != nil {
		fmt.Printf("업로드된 파일: %v\n", imageField)
	}

	return nil
}

func downloadFileExample(ctx context.Context, client *pocketbase.Client) error {
	// 먼저 파일이 있는 레코드를 찾습니다
	records, err := client.Records.GetList(ctx, "posts", &pocketbase.ListOptions{
		Page:    1,
		PerPage: 1,
		Filter:  "image != ''", // 이미지 필드가 비어있지 않은 레코드
	})
	if err != nil {
		return fmt.Errorf("레코드 조회 실패: %w", err)
	}

	if len(records.Items) == 0 {
		fmt.Println("다운로드할 파일이 있는 레코드가 없습니다.")
		return nil
	}

	record := records.Items[0]
	imageField := record.Get("image")
	if imageField == nil {
		fmt.Println("이미지 필드가 없습니다.")
		return nil
	}

	var filename string
	switch v := imageField.(type) {
	case string:
		filename = v
	case []interface{}:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				filename = str
			}
		}
	}

	if filename == "" {
		fmt.Println("파일명을 찾을 수 없습니다.")
		return nil
	}

	fmt.Printf("파일 다운로드 시작: %s\n", filename)

	// 파일 다운로드
	reader, err := client.Files.Download(ctx, "posts", record.ID, filename, nil)
	if err != nil {
		return fmt.Errorf("파일 다운로드 실패: %w", err)
	}
	defer reader.Close()

	// 파일 내용 읽기
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("파일 내용 읽기 실패: %w", err)
	}

	fmt.Printf("다운로드된 파일 크기: %d bytes\n", len(content))
	fmt.Printf("파일 내용 (처음 100자): %s\n", string(content[:min(100, len(content))]))

	return nil
}

func fileURLExample(client *pocketbase.Client) {
	// 예제 파일 URL 생성
	collection := "posts"
	recordID := "example_record_id"
	filename := "example.jpg"

	// 기본 파일 URL
	basicURL := client.Files.GetFileURL(collection, recordID, filename, nil)
	fmt.Printf("기본 파일 URL: %s\n", basicURL)

	// 썸네일 URL
	thumbURL := client.Files.GetFileURL(collection, recordID, filename, &pocketbase.FileDownloadOptions{
		Thumb: "100x100",
	})
	fmt.Printf("썸네일 URL: %s\n", thumbURL)

	// 다운로드 강제 URL
	downloadURL := client.Files.GetFileURL(collection, recordID, filename, &pocketbase.FileDownloadOptions{
		Download: true,
	})
	fmt.Printf("다운로드 URL: %s\n", downloadURL)

	// 썸네일 + 다운로드 URL
	combinedURL := client.Files.GetFileURL(collection, recordID, filename, &pocketbase.FileDownloadOptions{
		Thumb:    "200x200",
		Download: true,
	})
	fmt.Printf("썸네일 + 다운로드 URL: %s\n", combinedURL)
}

func deleteFileExample(ctx context.Context, client *pocketbase.Client) error {
	// 파일이 있는 레코드를 찾습니다
	records, err := client.Records.GetList(ctx, "posts", &pocketbase.ListOptions{
		Page:    1,
		PerPage: 1,
		Filter:  "image != ''", // 이미지 필드가 비어있지 않은 레코드
	})
	if err != nil {
		return fmt.Errorf("레코드 조회 실패: %w", err)
	}

	if len(records.Items) == 0 {
		fmt.Println("삭제할 파일이 있는 레코드가 없습니다.")
		return nil
	}

	record := records.Items[0]
	imageField := record.Get("image")
	if imageField == nil {
		fmt.Println("이미지 필드가 없습니다.")
		return nil
	}

	var filename string
	switch v := imageField.(type) {
	case string:
		filename = v
	case []interface{}:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				filename = str
			}
		}
	}

	if filename == "" {
		fmt.Println("삭제할 파일명을 찾을 수 없습니다.")
		return nil
	}

	fmt.Printf("파일 삭제 시작: %s\n", filename)

	// 파일 삭제
	updatedRecord, err := client.Files.Delete(ctx, "posts", record.ID, "image", filename)
	if err != nil {
		return fmt.Errorf("파일 삭제 실패: %w", err)
	}

	fmt.Printf("파일 삭제 완료. 업데이트된 레코드 ID: %s\n", updatedRecord.ID)
	if imageField := updatedRecord.Get("image"); imageField != nil {
		fmt.Printf("삭제 후 이미지 필드: %v\n", imageField)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
