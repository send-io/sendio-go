package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func scheduleEmail() {
	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	client := sendio.NewClient(apiKey)

	// Schedule the email
	params := &sendio.SendEmailRequest{
		To:          []string{"delivered@send.io"},
		From:        "onboarding@yourdomain.com",
		Text:        "hello world",
		Subject:     "Hello from Golang",
		ScheduledAt: "2024-09-05T11:52:01.858Z",
	}

	sent, err := client.Emails.SendWithContext(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Println(sent.Id)

	updateParams := &sendio.UpdateEmailRequest{
		Id:          sent.Id,
		ScheduledAt: "2024-11-05T11:52:01.858Z",
	}

	// Update the scheduled email
	updatedEmail, err := client.Emails.UpdateWithContext(ctx, updateParams)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", updatedEmail)

	canceled, err := client.Emails.CancelWithContext(ctx, "32723fee-8502-4b58-8b5e-bfd98f453ced")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", canceled)
}
