package examples

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/send-io/sendio-go"
)

func sendEmailCustomClientExample() {
	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	httpClient := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Create a send.io client with the custom HTTP client
	client := sendio.NewCustomClient(httpClient, apiKey)

	// Send params
	params := &sendio.SendEmailRequest{
		To:      []string{"delivered@send.io"},
		From:    "onboarding@yourdomain.com",
		Text:    "hello world",
		Subject: "Hello from Golang",
		Cc:      []string{"cc@example.com"},
		Bcc:     []string{"ccc@example.com"},
		ReplyTo: "to@example.com",
	}

	sent, err := client.Emails.SendWithContext(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Sent basic email: %s\n", sent.Id)

	// Sending with IdempotencyKey
	options := &sendio.SendEmailOptions{
		IdempotencyKey: "unique-idempotency-key",
	}

	sent, err = client.Emails.SendWithOptions(ctx, params, options)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Sent email with idempotency key: %s\n", sent.Id)

	// Get Email
	email, err := client.Emails.GetWithContext(ctx, sent.Id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", email)
}
