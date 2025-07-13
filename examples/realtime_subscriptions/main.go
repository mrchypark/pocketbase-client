package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	ctx := context.Background()

	// Subscribe to all events on the 'posts' collection
	unsubscribe, err := client.Realtime.Subscribe(ctx, []string{"posts"}, func(e *pocketbase.RealtimeEvent, err error) {
		fmt.Printf("Received event: Action=%s, RecordID=%s\n", e.Action, e.Record.ID)
		// You can access the record data with e.Record
	})
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	fmt.Println("Subscribed to 'posts' collection. Waiting for events...")

	// Keep the application running to receive events
	time.Sleep(5 * time.Minute)
}
