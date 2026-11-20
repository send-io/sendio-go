package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func inlineAttachmentExample() {
	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	if apiKey == "" {
		panic("Api Key is missing")
	}

	client := sendio.NewClient(apiKey)

	// Create attachments objects
	attachment := &sendio.Attachment{
		Path:      "https://app.send.io/icon-512.png",
		Filename:  "sendio-icon.png",
		ContentId: "my-test-image",
	}

	params := &sendio.SendEmailRequest{
		To:          []string{"delivered@send.io"},
		From:        "onboarding@yourdomain.com",
		Text:        "email with inline content attachment",
		Html:        "<p>This is an email with an <img width=100 height=40 src=\"cid:my-test-image\" /> embed image</p>",
		Subject:     "Email with inline attachment",
		Attachments: []*sendio.Attachment{attachment},
	}

	sent, err := client.Emails.SendWithContext(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Println(sent.Id)
}
