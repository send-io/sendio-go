package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func sendEmailWithTemplateExample() {
	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	client := sendio.NewClient(apiKey)

	fmt.Println("Creating a template with variables:")
	templateParams := &sendio.CreateTemplateRequest{
		Name:    "user-welcome-template",
		Alias:   "welcome",
		From:    "onboarding@yourdomain.com",
		Subject: "Welcome to {{{companyName}}}, {{{userName}}}!",
		Html: `
		<html>
			<body>
				<h1>Hello {{{userName}}}!</h1>
				<p>Welcome to {{{companyName}}}. We're excited to have you on board.</p>
				<p>You currently have {{{messageCount}}} unread messages waiting for you.</p>
			</body>
		</html>
		`,
		Text: "Hello {{{userName}}}! Welcome to {{{companyName}}}.",
		Variables: []*sendio.TemplateVariable{
			{
				Key:           "userName",
				Type:          sendio.VariableTypeString,
				FallbackValue: "User",
			},
			{
				Key:           "companyName",
				Type:          sendio.VariableTypeString,
				FallbackValue: "Our Company",
			},
			{
				Key:           "messageCount",
				Type:          sendio.VariableTypeNumber,
				FallbackValue: 0,
			},
		},
	}

	template, err := client.Templates.CreateWithContext(ctx, templateParams)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✓ Created template: %s\n", template.Id)

	fmt.Println("\nPublishing template:")
	_, err = client.Templates.PublishWithContext(ctx, template.Id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✓ Published template: %s\n", template.Id)

	fmt.Println("\nSending email using template ID:")
	emailParams := &sendio.SendEmailRequest{
		To: []string{"delivered@send.io"},
		Template: &sendio.EmailTemplate{
			Id: template.Id,
			Variables: map[string]any{
				"userName":     "Alice Johnson",
				"companyName":  "Acme Corporation",
				"messageCount": 12,
			},
		},
	}

	sent, err := client.Emails.SendWithContext(ctx, emailParams)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✓ Sent email: %s\n", sent.Id)

	fmt.Println("\nSending email using template alias:")
	emailParams2 := &sendio.SendEmailRequest{
		To: []string{"delivered@send.io"},
		Template: &sendio.EmailTemplate{
			Id: "welcome", // Using alias instead of ID
			Variables: map[string]any{
				"userName":     "Bob Smith",
				"companyName":  "Tech Startup Inc",
				"messageCount": 3,
			},
		},
	}

	sent2, err := client.Emails.SendWithContext(ctx, emailParams2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✓ Sent email: %s\n", sent2.Id)

	fmt.Println("\nSending email with template and field overrides:")
	emailParams3 := &sendio.SendEmailRequest{
		From:    "support@yourdomain.com", // Override template's From
		To:      []string{"delivered@send.io"},
		Subject: "Custom Subject Override", // Override template's Subject
		Bcc:     []string{"bcc@example.com"},
		ReplyTo: "noreply@yourdomain.com",
		Template: &sendio.EmailTemplate{
			Id: template.Id,
			Variables: map[string]any{
				"userName":     "Charlie Brown",
				"companyName":  "Example LLC",
				"messageCount": 7,
			},
		},
	}

	sent3, err := client.Emails.SendWithContext(ctx, emailParams3)
	if err != nil {
		panic(err)
	}
	fmt.Printf("✓ Sent email with overrides: %s\n", sent3.Id)

	fmt.Println("\nCleaning up template:")
	removeResp, err := client.Templates.RemoveWithContext(ctx, template.Id)
	if err != nil {
		fmt.Printf("Warning: Could not remove template %s: %v\n", template.Id, err)
	} else {
		fmt.Printf("✓ Removed template: %s\n", removeResp.Id)
	}

	fmt.Println("\nExample completed successfully!")
}
