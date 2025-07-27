// Package main demonstrates a real-time chat application using PocketBase Go client.
// This example shows real-time subscriptions and Record-based message handling.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	// Get username from command-line arguments or prompt
	var username string
	if len(os.Args) > 1 {
		username = os.Args[1]
	} else {
		fmt.Print("Enter your username: ")
		reader := bufio.NewReader(os.Stdin)
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)
	}

	if username == "" {
		log.Fatal("Username cannot be empty.")
	}

	// --- Initialize PocketBase Client ---
	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))

	// Create Record service for chat collection
	chatService := client.Records("chat")

	// --- Subscribe to the 'chat' collection ---
	// We subscribe to all events ('*') on the 'chat' collection.
	unsubscribe, err := client.Realtime.Subscribe(context.Background(), []string{"chat"}, func(e *pocketbase.RealtimeEvent, err error) {
		// This callback function is triggered for every event.
		if err != nil {
			log.Printf("Subscription error: %v\n", err)
			return
		}

		// We only care about 'create' events, which represent new messages.
		if e.Action == "create" {
			// e.Record contains the newly created chat message record.
			// We can use GetString to safely access its fields.
			msgUser := e.Record.GetString("user")
			msgText := e.Record.GetString("text")
			msgTime := e.Record.GetDateTime("created")

			// Don't print the user's own messages back to them
			if msgUser != username {
				timeStr := msgTime.String()[:19] // YYYY-MM-DD HH:MM:SS format
				fmt.Printf("\n[%s] %s: %s\n> ", timeStr, msgUser, msgText)
			}
		}
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to 'chat' collection: %v", err)
	}
	defer unsubscribe() // Ensure we unsubscribe when main exits

	fmt.Printf("--- Welcome to the Real-time Chat, %s! ---\n", username)
	fmt.Println("Type your message and press Enter to send. Type 'exit' to quit.")
	fmt.Println("----------------------------------------------------")

	// --- Main loop to read user input and send messages ---
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break // Exit on Ctrl+D or other read errors
		}
		text := scanner.Text()

		if strings.ToLower(text) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if text == "" {
			continue
		}

		// Create a new record in the 'chat' collection.
		// This will trigger the real-time event for all subscribed clients.
		message := &pocketbase.Record{}
		message.Set("user", username)
		message.Set("text", text)

		if _, err := chatService.Create(context.Background(), message); err != nil {
			log.Printf("Error sending message: %v", err)
		}

		// A small delay to prevent the prompt from overlapping with received messages.
		time.Sleep(50 * time.Millisecond)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}
