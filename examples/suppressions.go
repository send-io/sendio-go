package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func suppressionsExample() {
	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	client := sendio.NewClient(apiKey)

	// Add a single email address to the suppression list.
	suppression, err := client.Suppressions.AddWithContext(ctx, &sendio.AddSuppressionRequest{
		Email: "steve.wozniak@gmail.com",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Created suppression id: " + suppression.Id)

	// List suppressions, optionally filtered by origin.
	limit := 20
	list, err := client.Suppressions.ListWithContext(ctx, &sendio.ListSuppressionsOptions{
		Origin: sendio.SuppressionOriginBounce,
		Limit:  &limit,
	})
	if err != nil {
		panic(err)
	}
	for _, entry := range list.Data {
		sourceId := "none"
		if entry.SourceId != nil {
			sourceId = *entry.SourceId
		}
		fmt.Printf("%s (%s) source: %s\n", entry.Email, entry.Origin, sourceId)
	}
	fmt.Printf("Has more: %v\n", list.HasMore)

	// Get accepts either a suppression ID or an email address.
	retrieved, err := client.Suppressions.GetWithContext(ctx, "steve.wozniak@gmail.com")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Retrieved suppression: %v\n", retrieved)

	// Remove accepts either a suppression ID or an email address.
	removed, err := client.Suppressions.RemoveWithContext(ctx, suppression.Id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Removed suppression %s: %v\n", removed.Id, removed.Deleted)

	// Add up to 100 suppressions at once.
	batchAdded, err := client.Suppressions.Batch.AddWithContext(ctx, &sendio.BatchAddSuppressionsRequest{
		Emails: []string{"steve.wozniak@gmail.com", "steve.jobs@gmail.com"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Batch added %d suppressions\n", len(batchAdded.Data))

	// Remove up to 100 suppressions at once, by email address or by ID.
	batchRemoved, err := client.Suppressions.Batch.RemoveWithContext(ctx, &sendio.BatchRemoveSuppressionsRequest{
		Emails: []string{"steve.wozniak@gmail.com", "steve.jobs@gmail.com"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Batch removed %d suppressions\n", len(batchRemoved.Data))
}
