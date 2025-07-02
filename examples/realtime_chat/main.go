package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mrchypark/pocketbase-client" // 실제 프로젝트 경로에 맞게 수정해주세요
)

func main() {
	// 1. PocketBase 클라이언트 인스턴스 생성
	pb := pocketbase.NewClient("http://127.0.0.1:8090")

	// 2. 실시간 이벤트를 처리할 콜백 함수 정의
	// 이 함수는 서버에서 이벤트가 발생할 때마다 라이브러리에 의해 호출됩니다.
	realtimeCallback := func(event *pocketbase.RealtimeEvent, err error) {
		if err != nil {
			log.Printf("Callback Error: %v\n", err)
			return
		}

		// 수신된 이벤트의 액션과 레코드 데이터를 출력합니다.
		// event.Action 대신 event.Name 을 사용해야 할 수도 있습니다. (라이브러리 구조 확인 필요)
		log.Printf("Received event: Action=%s, RecordID=%s, Message=%v",
			event.Action, // realtime.go를 보면 Name 필드는 없지만, RealtimeEvent 구조체에 따라 Action 또는 Name을 사용
			event.Record.ID,
			event.Record.Data["message"],
		)
	}

	// 3. "chat" 컬렉션 구독 및 콜백 함수 등록
	// Subscribe 함수는 구독을 취소할 수 있는 함수(unsubscribe)와 에러를 반환합니다.
	unsubscribe, err := pb.Realtime.Subscribe(
		context.Background(),
		[]string{"chat"},
		realtimeCallback,
	)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	// 프로그램 종료 시 구독 취소 함수 호출
	defer unsubscribe()

	log.Println("Subscribed; press Ctrl+C to exit")

	// 4. 프로그램이 바로 종료되지 않도록 대기
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Unsubscribing and exiting...")
}
