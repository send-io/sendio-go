package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func withAttachments() {
	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	if apiKey == "" {
		panic("Api Key is missing")
	}

	// Read attachment file
	pwd, _ := os.Getwd()
	f, err := os.ReadFile(pwd + "/resources/invoice.pdf")
	if err != nil {
		panic(err)
	}

	client := sendio.NewClient(apiKey)

	// Create attachments objects
	pdfAttachmentFromLocalFile := &sendio.Attachment{
		Content:     f,
		Filename:    "invoice1.pdf",
		ContentType: "application/pdf",
	}

	pdfAttachmentFromRemotePath := &sendio.Attachment{
		Path:        "https://send.io/static/sdk/invoice.pdf",
		Filename:    "invoice2.pdf",
		ContentType: "application/pdf",
	}

	params := &sendio.SendEmailRequest{
		To:          []string{"delivered@send.io"},
		From:        "onboarding@yourdomain.com",
		Text:        "email with attachments !!",
		Html:        "<strong>email with attachments !!</strong>",
		Subject:     "Email with attachment",
		Attachments: []*sendio.Attachment{pdfAttachmentFromLocalFile, pdfAttachmentFromRemotePath},
	}

	sent, err := client.Emails.SendWithContext(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Println("Sent email ID:", sent.Id)

	fmt.Println("\nListing all attachments:")
	attachments, err := client.Emails.ListAttachments(sent.Id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d attachments\n", len(attachments.Data))
	for _, att := range attachments.Data {
		fmt.Printf("- Attachment ID: %s, Filename: %s, ContentType: %s\n",
			att.Id, att.Filename, att.ContentType)
	}

	fmt.Println("\nListing attachments with context:")
	attachmentsWithCtx, err := client.Emails.ListAttachmentsWithContext(ctx, sent.Id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d attachments\n", len(attachmentsWithCtx.Data))

	fmt.Println("\nListing attachments with options:")
	limit := 10
	listOptions := &sendio.ListOptions{
		Limit: &limit,
	}
	attachmentsWithOpts, err := client.Emails.ListAttachmentsWithOptions(ctx, sent.Id, listOptions)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d attachments (limit: %d)\n", len(attachmentsWithOpts.Data), listOptions.Limit)

	// Get a specific attachment (using the first attachment ID if available)
	if len(attachments.Data) > 0 {
		firstAttachmentId := attachments.Data[0].Id
		fmt.Printf("\nGetting attachment with ID: %s\n", firstAttachmentId)

		// Get attachment
		attachment, err := client.Emails.GetAttachment(sent.Id, firstAttachmentId)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Retrieved attachment: %s (%s)\n", attachment.Filename, attachment.ContentType)

	}
}
