package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func broadcastExamples() {
	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	client := sendio.NewClient(apiKey)

	// Create Broadcast
	params := &sendio.CreateBroadcastRequest{
		SegmentId: "ca4e37c5-a82a-4199-a3b8-bf912a6472aa",
		From:      "onboarding@yourdomain.com",
		Html:      "<html><body><h1>Hello, world!</h1></body></html>",
		Name:      "Test Broadcast",
		Subject:   "Hello, world!",
		// The line inboxes show after the subject, and the topic it belongs
		// to: contacts who opted out of the topic are skipped.
		PreviewText: "What is new this month",
		TopicId:     "b6d24b8e-af0b-4c3c-be0c-359bbd97381e",
	}

	broadcast, err := client.Broadcasts.CreateWithContext(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Println("Created broadcast with entry id: " + broadcast.Id)

	// Get Broadcast
	retrievedBroadcast, err := client.Broadcasts.GetWithContext(ctx, broadcast.Id)
	if err != nil {
		panic(err)
	}

	fmt.Println("ID: " + retrievedBroadcast.Id)
	fmt.Println("Name: " + retrievedBroadcast.Name)
	fmt.Println("Segment ID: " + retrievedBroadcast.SegmentId)
	fmt.Println("From: " + retrievedBroadcast.From)
	fmt.Println("Subject: " + retrievedBroadcast.Subject)
	fmt.Println("Preview Text: " + retrievedBroadcast.PreviewText)
	fmt.Println("Status: " + retrievedBroadcast.Status)
	fmt.Println("Created At: " + retrievedBroadcast.CreatedAt)
	fmt.Println("Scheduled At: " + retrievedBroadcast.ScheduledAt)
	fmt.Println("Sent At: " + retrievedBroadcast.SentAt)
	fmt.Println("Html: " + retrievedBroadcast.Html)
	fmt.Println("Text: " + retrievedBroadcast.Text)

	updateParams := &sendio.UpdateBroadcastRequest{
		BroadcastId: retrievedBroadcast.Id,
		Name:        "Updated Test Broadcast, Go SDK",
	}

	updatedBroadcast, err := client.Broadcasts.UpdateWithContext(ctx, updateParams)
	if err != nil {
		panic(err)
	}
	fmt.Println("Updated broadcast with entry id: " + updatedBroadcast.Id)

	// Estimate an audience before saving it
	estimate, err := client.Broadcasts.EstimateWithContext(ctx, &sendio.EstimateBroadcastRequest{
		SegmentIds: []string{"ca4e37c5-a82a-4199-a3b8-bf912a6472aa"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Reaches %d contacts (%d skipped as invalid)\n", estimate.Recipients, estimate.Skipped.Invalid)

	// Send a test to up to 5 addresses
	if _, err := client.Broadcasts.SendTestWithContext(ctx, broadcast.Id, &sendio.SendBroadcastTestRequest{To: []string{"you@yourdomain.com"}}); err != nil {
		panic(err)
	}

	// Saved versions: every save of a draft is one
	versions, err := client.Broadcasts.ListVersions(broadcast.Id)
	if err != nil {
		panic(err)
	}
	for _, v := range versions.Data {
		fmt.Printf("Version %d (current: %t)\n", v.VersionNumber, v.Current)
	}
	label := "Before the rename"
	if _, err := client.Broadcasts.UpdateVersionWithContext(ctx, broadcast.Id, "1", &sendio.UpdateBroadcastVersionRequest{Name: &label}); err != nil {
		panic(err)
	}
	if _, err := client.Broadcasts.RestoreVersionWithContext(ctx, broadcast.Id, "1"); err != nil {
		panic(err)
	}

	// Duplicate Broadcast
	duplicated, err := client.Broadcasts.DuplicateWithContext(ctx, broadcast.Id)
	if err != nil {
		panic(err)
	}
	fmt.Println("Duplicated broadcast with entry id: " + duplicated.Id)

	// Send Broadcast
	// sendParams := &sendio.SendBroadcastRequest{
	// 	BroadcastId: broadcast.Id,
	// }

	// sendResponse, err := client.Broadcasts.SendWithContext(ctx, sendParams)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Sent broadcast with entry id: " + sendResponse.Id)

	// List draft Broadcasts whose name contains "Test"
	listResponse, err := client.Broadcasts.ListWithFilters(ctx, &sendio.ListBroadcastsOptions{Status: "draft", Search: "Test"})
	if err != nil {
		panic(err)
	}

	for _, b := range listResponse.Data {
		fmt.Println("ID: " + b.Id)
		fmt.Println("Name: " + b.Name)
		fmt.Println("Segment ID: " + b.SegmentId)
		fmt.Println("From: " + b.From)
		fmt.Println("Subject: " + b.Subject)
		fmt.Println("Preview Text: " + b.PreviewText)
		fmt.Println("Status: " + b.Status)
		fmt.Println("Created At: " + b.CreatedAt)
		fmt.Println("Scheduled At: " + b.ScheduledAt)
		fmt.Println("Sent At: " + b.SentAt)
		fmt.Println("Html: " + b.Html)
		fmt.Println("Text: " + b.Text)
	}

	// Remove Broadcast (Only Draft Broadcasts can be deleted)
	removeResponse, err := client.Broadcasts.RemoveWithContext(ctx, broadcast.Id)
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted broadcast with entry id: " + removeResponse.Id)

	removeDuplicateResponse, err := client.Broadcasts.RemoveWithContext(ctx, duplicated.Id)
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted duplicated broadcast with entry id: " + removeDuplicateResponse.Id)

	// Stats of a sent broadcast
	// stats, err := client.Broadcasts.StatsWithContext(ctx, sentBroadcastId)
	// fmt.Printf("Delivered %d, open rate %.2f\n", stats.Delivered, stats.OpenRate)
}
