package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func sendBatchEmails() {

	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	client := sendio.NewClient(apiKey)

	// Batch Send
	var batchEmails = []*sendio.SendEmailRequest{
		{
			To:      []string{"delivered@send.io"},
			From:    "onboarding@yourdomain.com",
			Text:    "hey",
			Subject: "Hello from emails",
		},
		{
			To:      []string{"delivered@send.io"},
			From:    "onboarding@yourdomain.com",
			Text:    "hellooo",
			Subject: "Hello from batch emails 2",
		},
	}

	// Regular send without options
	sent, err := client.Batch.SendWithContext(ctx, batchEmails)
	if err != nil {
		panic(err)
	}
	fmt.Println("Sent without options")
	fmt.Println(sent.Data)

	// Send with options: IdempotencyKey
	options := &sendio.BatchSendEmailOptions{
		IdempotencyKey: "68656c6c6f2d776f726c64",
	}
	sent, err = client.Batch.SendWithOptions(ctx, batchEmails, options)
	if err != nil {
		panic(err)
	}
	fmt.Println("Sent with idempotency key")
	fmt.Println(sent.Data)

	// Send with permissive validation mode
	// This allows partial success - valid emails will be sent even if some are invalid
	batchEmailsWithErrors := []*sendio.SendEmailRequest{
		{
			To:      []string{"delivered@send.io"},
			From:    "onboarding@yourdomain.com",
			Text:    "This email is valid",
			Subject: "Valid email",
		},
		{
			To:      []string{}, // Missing 'to' field - will fail validation
			From:    "onboarding@yourdomain.com",
			Text:    "This email has no recipient",
			Subject: "Invalid email",
		},
		{
			To:      []string{"another@yourdomain.com"},
			From:    "onboarding@yourdomain.com",
			Text:    "This email is also valid",
			Subject: "Another valid email",
		},
	}

	permissiveOptions := &sendio.BatchSendEmailOptions{
		BatchValidation: sendio.BatchValidationPermissive,
	}
	sent, err = client.Batch.SendWithOptions(ctx, batchEmailsWithErrors, permissiveOptions)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nSent with permissive validation mode:")
	fmt.Printf("Successfully sent %d emails\n", len(sent.Data))
	for _, email := range sent.Data {
		fmt.Printf("  - Email ID: %s\n", email.Id)
	}

	if len(sent.Errors) > 0 {
		fmt.Printf("Failed to send %d emails:\n", len(sent.Errors))
		for _, err := range sent.Errors {
			fmt.Printf("  - Index %d: %s\n", err.Index, err.Message)
		}
	}

	// Send with strict validation mode (default behavior)
	// All emails must be valid or the entire batch fails
	strictOptions := &sendio.BatchSendEmailOptions{
		BatchValidation: sendio.BatchValidationStrict, // This is the default, shown for clarity
	}
	sent, err = client.Batch.SendWithOptions(ctx, batchEmails, strictOptions)
	if err != nil {
		// In strict mode, if any email is invalid, an error is returned
		fmt.Printf("Batch send failed in strict mode: %v\n", err)
	} else {
		fmt.Println("\nSent with strict validation mode (all emails valid):")
		fmt.Printf("Successfully sent %d emails\n", len(sent.Data))
	}
}
