package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mrchypark/pocketbase-client"
)

func main() {
	// 1. Create a PocketBase client instance
	pb := pocketbase.NewClient("http://127.0.0.1:8090")

	// 2. Define a callback function to handle real-time events.
	// This function is called by the library whenever an event occurs on the server.
	realtimeCallback := func(event *pocketbase.RealtimeEvent, err error) {
		if err != nil {
			log.Printf("Callback Error: %v\n", err)
			return
		}

		// Output the action and record data of the received event.
		// Depending on the library structure, event.Action or event.Name might be used.
		log.Printf("Received event: Action=%s, RecordID=%s, Message=%v",
			event.Action, // Although realtime.go does not have a Name field, use Action or Name depending on the RealtimeEvent struct
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
