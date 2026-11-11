package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func mailboxesExample() {
	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	client := sendio.NewClient(apiKey)

	// Create a mailbox on a verified domain
	mailbox, err := client.Mailboxes.CreateWithContext(ctx, &sendio.CreateMailboxRequest{
		Address: "sales@yourdomain.com",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created mailbox: %s (%s)\n", mailbox.Address, mailbox.ImapSmtp)

	// Credentials are returned once, when IMAP/SMTP is ready.
	// The password is also an API key for this mailbox.
	if mailbox.Credentials != nil {
		fmt.Printf("IMAP %s:%d, SMTP %s:%d, user %s\n",
			mailbox.Credentials.Imap.Host, mailbox.Credentials.Imap.Port,
			mailbox.Credentials.Smtp.Host, mailbox.Credentials.Smtp.Port,
			mailbox.Credentials.Username)
	}

	// List
	mailboxes, err := client.Mailboxes.ListWithContext(ctx)
	if err != nil {
		panic(err)
	}
	for _, m := range mailboxes.Data {
		fmt.Printf("%s health=%s sent=%d\n", m.Address, m.Health, m.Sent)
	}

	// Turn warmup off
	updated, err := client.Mailboxes.UpdateWithContext(ctx, mailbox.Id, &sendio.UpdateMailboxRequest{
		WarmupEnabled: false,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Warmup enabled: %t\n", updated.WarmupEnabled)

	// Reset the password (the old one stops working)
	reset, err := client.Mailboxes.ResetPasswordWithContext(ctx, mailbox.Id)
	if err != nil {
		panic(err)
	}
	fmt.Println("New password for " + reset.Address + ": " + reset.Credentials.Password)

	// Delete the mailbox (and the agent inbox on it, if any)
	removed, err := client.Mailboxes.RemoveWithContext(ctx, mailbox.Id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Deleted mailbox %s: %t\n", removed.Id, removed.Deleted)
}
