package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "github.com/mrchypark/pocketbase-client"
)

// ProgressWriter는 스트리밍 진행률을 추적하는 Writer입니다
type ProgressWriter struct {
	writer     io.Writer
	totalBytes int64
	onProgress func(bytes int64)
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.totalBytes += int64(n)
	if pw.onProgress != nil {
		pw.onProgress(pw.totalBytes)
	}
	return
}

func main() {
	// PocketBase 클라이언트 생성
	client := pb.NewClient("http://127.0.0.1:8090")
	ctx := context.Background()

	fmt.Println("=== PocketBase Streaming API 예제 ===")

	// 관리자 인증 (선택사항 - 필요한 경우)
	fmt.Println("1. 관리자 인증 시도...")
	_, err := client.WithAdminPassword(ctx, "admin@example.com", "password")
	if err != nil {
		fmt.Printf("인증 실패 (무시하고 계속): %v\n", err)
	} else {
		fmt.Println("인증 성공!")
	}

	// 예제 1: 메모리 버퍼로 스트리밍
	fmt.Println("\n2. 메모리 버퍼로 스트리밍...")
	if err := streamToBuffer(client); err != nil {
		log.Printf("버퍼 스트리밍 실패: %v", err)
	}

	// 예제 2: 파일로 스트리밍
	fmt.Println("\n3. 파일로 스트리밍...")
	if err := streamToFile(client); err != nil {
		log.Printf("파일 스트리밍 실패: %v", err)
	}

	// 예제 3: 진행률 추적과 함께 스트리밍
	fmt.Println("\n4. 진행률 추적과 함께 스트리밍...")
	if err := streamWithProgress(client); err != nil {
		log.Printf("진행률 스트리밍 실패: %v", err)
	}

	// 예제 4: 실시간 스트리밍 (짧은 시간)
	fmt.Println("\n5. 실시간 스트리밍 (10초간)...")
	if err := setupRealtimeStream(client); err != nil {
		log.Printf("실시간 스트리밍 실패: %v", err)
	}

	fmt.Println("\n=== 모든 예제 완료 ===")
}

// streamToBuffer는 응답을 메모리 버퍼로 스트리밍합니다
func streamToBuffer(client *pb.Client) error {
	var buf bytes.Buffer

	// SendWithOptions를 사용하여 스트리밍
	err := client.SendWithOptions(
		context.Background(),
		"GET",
		"/api/collections",
		nil,
		nil, // responseData는 WithResponseWriter 사용 시 nil이어야 함
		pb.WithResponseWriter(&buf),
	)
	if err != nil {
		return fmt.Errorf("스트리밍 실패: %w", err)
	}

	fmt.Printf("   스트리밍된 데이터 크기: %d bytes\n", buf.Len())

	// 처음 200자만 출력
	content := buf.String()
	if len(content) > 200 {
		content = content[:200] + "..."
	}
	fmt.Printf("   내용 미리보기: %s\n", content)

	return nil
}

// streamToFile은 응답을 파일로 직접 스트리밍합니다
func streamToFile(client *pb.Client) error {
	// 출력 파일 생성
	file, err := os.Create("collections_stream.json")
	if err != nil {
		return fmt.Errorf("파일 생성 실패: %w", err)
	}
	defer file.Close()

	// 버퍼링된 Writer 사용 (성능 향상)
	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	// 타임아웃 설정
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 파일로 직접 스트리밍
	err = client.SendWithOptions(
		ctx,
		"GET",
		"/api/collections",
		nil,
		nil,
		pb.WithResponseWriter(bufferedWriter),
	)
	if err != nil {
		return fmt.Errorf("파일 스트리밍 실패: %w", err)
	}

	// 파일 크기 확인
	fileInfo, _ := file.Stat()
	fmt.Printf("   파일로 스트리밍 완료: collections_stream.json (%d bytes)\n", fileInfo.Size())

	return nil
}

// streamWithProgress는 진행률을 추적하면서 스트리밍합니다
func streamWithProgress(client *pb.Client) error {
	file, err := os.Create("progress_stream.json")
	if err != nil {
		return fmt.Errorf("파일 생성 실패: %w", err)
	}
	defer file.Close()

	// 진행률 추적 Writer 생성
	progressWriter := &ProgressWriter{
		writer: file,
		onProgress: func(bytes int64) {
			fmt.Printf("\r   진행률: %d bytes 스트리밍됨", bytes)
		},
	}

	err = client.SendWithOptions(
		context.Background(),
		"GET",
		"/api/collections",
		nil,
		nil,
		pb.WithResponseWriter(progressWriter),
	)

	fmt.Println() // 진행률 출력 후 새 줄

	if err != nil {
		return fmt.Errorf("진행률 스트리밍 실패: %w", err)
	}

	fmt.Printf("   진행률 스트리밍 완료: progress_stream.json (%d bytes)\n", progressWriter.totalBytes)
	return nil
}

// setupRealtimeStream은 실시간 스트리밍을 설정합니다
func setupRealtimeStream(client *pb.Client) error {
	// 10초 타임아웃 설정
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 실시간 구독 설정
	unsubscribe, err := client.Realtime.Subscribe(
		ctx,
		[]string{"*"}, // 모든 컬렉션 구독
		func(event *pb.RealtimeEvent, err error) {
			if err != nil {
				log.Printf("   실시간 에러: %v", err)
				return
			}
			fmt.Printf("   실시간 이벤트: %s 액션이 %s 레코드에서 발생\n",
				event.Action, event.Record.ID)
		},
	)
	if err != nil {
		return fmt.Errorf("실시간 구독 실패: %w", err)
	}
	defer unsubscribe()

	fmt.Println("   실시간 이벤트 수신 중... (10초간)")

	// 컨텍스트가 완료될 때까지 대기
	<-ctx.Done()

	fmt.Println("   실시간 스트리밍 완료")
	return nil
}
