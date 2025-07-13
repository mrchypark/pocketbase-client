package main

import (
	"bytes"
	"fmt"
	"log"

	pocketbase "github.com/cypark/pocketbase-client"
)

func main() {
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	client.AuthWithPassword("users", "testuser", "1234567890")

	// The record that will hold the file
	recordData := map[string]interface{}{
		"title": "Record with a file",
	}

	// The file content and name
	fileContent := bytes.NewReader([]byte("this is the file content"))
	fileName := "example.txt"

	// Upload the file and create the record in one go
	createdRecord, err := client.Records.CreateWithFile("posts", recordData, "document", fileName, fileContent)
	if err != nil {
		log.Fatalf("Failed to create record with file: %v", err)
	}

	fmt.Printf("Successfully created record %s with file %s\n", createdRecord.ID, createdRecord.GetString("document"))
}