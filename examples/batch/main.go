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

// newRandomID generates a 15-character random ID suitable for PocketBase (lowercase letters + digits).
func newRandomID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 15
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			// This error is unlikely to occur, so panic if it does.
			panic(fmt.Errorf("critical error generating random ID: %w", err))
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret)
}

func main() {
	// --- 1. Client Initialization and Authentication ---
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	ctx := context.Background()

	if _, err := client.HealthCheck(ctx); err != nil {
		log.Fatalf("❌ Could not connect to PocketBase server: %v\n   Please ensure the server is running at http://127.0.0.1:8090", err)
	}

	if _, err := client.AuthenticateAsAdmin(ctx, "admin@example.com", "1q2w3e4r5t"); err != nil {
		log.Fatalf("❌ Admin authentication failed: %v", err)
	}
	fmt.Println("✅ Admin authentication successful!")
	fmt.Println("------------------------------------")

	// --- 2. First Batch Operation: Create 5 Records ---
	fmt.Println("🚀 [Phase 1] Starting batch creation of 5 records...")
	var createRequests []*pocketbase.BatchRequest
	for i := 1; i <= 5; i++ {
		req, err := client.Records.NewCreateRequest("posts", map[string]any{
			"title": fmt.Sprintf("New Post #%d", i),
		})
		if err != nil {
			log.Fatalf("Create request %d failed: %v", i, err)
		}
		createRequests = append(createRequests, req)
	}

	createResults, err := client.Batch.Execute(ctx, createRequests)
	if err != nil {
		log.Fatalf("❌ First batch API execution failed: %v", err)
	}

	var createdIDs []string
	fmt.Println("\n✨ Creation Results:")
	for i, res := range createResults {
		// [Modified] res.StatusCode -> res.Status
		if res.Status == http.StatusOK {
			// [Modified] res.Data -> res.Body
			data, ok := res.Body.(map[string]interface{})
			if ok {
				id, _ := data["id"].(string)
				createdIDs = append(createdIDs, id)
				fmt.Printf("   - ✅ Request %d successful: New record ID '%s' created\n", i+1, id)
			}
		} else {
			// [Modified] Use res.ParsedError.Message instead of res.Error
			errMsg := "Unknown error"
			if res.ParsedError != nil {
				errMsg = res.ParsedError.Message
			}
			fmt.Printf("   - 🚨 Request %d failed: Status code %d, Error: %s\n", i+1, res.Status, errMsg)
		}
	}
	fmt.Println("------------------------------------")

	if len(createdIDs) < 4 {
		log.Fatalf("❌ Not enough records created to proceed with update and delete.")
	}

	// --- 3. Second Batch Operation: Update 3, Delete 1 ---
	fmt.Println("\n🚀 [Phase 2] Starting update (3) and delete (1) using created IDs...")
	var nextRequests []*pocketbase.BatchRequest

	for i := 0; i < 3; i++ {
		req, err := client.Records.NewUpdateRequest("posts", createdIDs[i], map[string]any{
			"title": fmt.Sprintf("Updated Post #%d (ID: %s)", i+1, createdIDs[i]),
		})
		if err != nil {
			log.Fatalf("Update request %d failed: %v", i+1, err)
		}
		nextRequests = append(nextRequests, req)
	}

	deleteReq, err := client.Records.NewDeleteRequest("posts", createdIDs[3])
	if err != nil {
		log.Fatalf("Delete request failed: %v", err)
	}
	nextRequests = append(nextRequests, deleteReq)

	updateDeleteResults, err := client.Batch.Execute(ctx, nextRequests)
	if err != nil {
		log.Fatalf("❌ Second batch API execution failed: %v", err)
	}

	fmt.Println("\n✨ Update/Delete Results:")
	for i, res := range updateDeleteResults {
		// [Modified] res.StatusCode -> res.Status
		if res.Status >= http.StatusBadRequest {
			errMsg := "Unknown error"
			if res.ParsedError != nil {
				errMsg = res.ParsedError.Message
			}
			fmt.Printf("   - 🚨 Request %d failed: Status code %d, Error: %s\n", i+1, res.Status, errMsg)
		} else {
			fmt.Printf("   - ✅ Request %d successful: Status code %d\n", i+1, res.Status)
		}
	}
	fmt.Println("------------------------------------")

	// --- 4. Third Batch Operation: 5 Upserts with 1-second interval ---
	fmt.Println("\n🚀 [Phase 3] Starting 5 Upserts with random IDs at 1-second intervals...")
	randomID := newRandomID()

	for i := 1; i <= 5; i++ {

		upsertReq, err := client.Records.NewUpsertRequest("posts", map[string]any{
			"id":    randomID,
			"title": fmt.Sprintf("Upsert Post (Attempt #%d)", i),
		})
		if err != nil {
			log.Fatalf("Upsert request creation failed: %v", err)
		}

		fmt.Printf("\n[Attempt %d/5] Upsert request with ID '%s'...\n", i, randomID)
		upsertResults, err := client.Batch.Execute(ctx, []*pocketbase.BatchRequest{upsertReq})
		if err != nil {
			log.Printf("   - 🚨 Upsert API execution failed: %v\n", err)
		} else if len(upsertResults) > 0 {
			res := upsertResults[0]
			// [Modified] res.StatusCode -> res.Status
			if res.Status == http.StatusCreated {
				fmt.Printf("   - ✅ Success: New record created (Status code: %d)\n", res.Status)
			} else if res.Status == http.StatusOK {
				fmt.Printf("   - ✅ Success: Existing record updated (Status code: %d)\n", res.Status)
			} else {
				errMsg := "Unknown error"
				if res.ParsedError != nil {
					errMsg = res.ParsedError.Message
				}
				fmt.Printf("   - 🚨 Failed: Status code %d, Error: %s\n", res.Status, errMsg)
			}
		}

		time.Sleep(1 * time.Second)
	}
	fmt.Println("\n🎉 All operations completed.")
}
