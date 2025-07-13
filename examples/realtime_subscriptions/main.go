package main

import (
	"fmt"
	"log"
	"time"

	pocketbase "github.com/cypark/pocketbase-client"
)

func main() {
	client := pocketbase.NewClient("http://127.0.0.1:8090")

	// Subscribe to all events on the 'posts' collection
	unsubscribe, err := client.Realtime.Subscribe("posts", func(e *pocketbase.RealtimeEvent) {
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