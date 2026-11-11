package sendio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const mailboxJSON = `{
	"object": "mailbox",
	"id": "mb_123",
	"address": "sales@example.com",
	"kind": "normal",
	"agent_inbox_id": null,
	"imap_smtp": "ready",
	"domain_id": "dom_1",
	"domain": "example.com",
	"warmup_enabled": false,
	"warmup_changed_at": null,
	"created_at": "2026-10-01T10:00:00Z",
	"sent": 12,
	"sent_total": 40,
	"delivered": 11,
	"bounced": 1,
	"complained": 0,
	"last_sent_at": "2026-10-20T09:00:00Z",
	"health": "healthy"
}`

const mailboxCredentialsJSON = `{
	"username": "sales@example.com",
	"password": "mbx_secret",
	"imap": {"host": "imap.send.io", "port": 993, "security": "ssl"},
	"smtp": {"host": "smtp.send.io", "port": 587, "security": "starttls"}
}`

func TestListMailboxes(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/mailboxes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Empty(t, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, `{
			"object": "list",
			"has_more": false,
			"data": [%s, {
				"object": "mailbox",
				"id": "mb_free",
				"address": "bob@free.send.io",
				"kind": "agent",
				"agent_inbox_id": "inb_9",
				"imap_smtp": "pending",
				"domain_id": null,
				"domain": null,
				"warmup_enabled": true,
				"warmup_changed_at": "2026-10-02T10:00:00Z",
				"created_at": "2026-10-01T10:00:00Z",
				"sent": 0, "sent_total": 0, "delivered": 0, "bounced": 0, "complained": 0,
				"last_sent_at": null,
				"health": "no_data",
				"pool": "free",
				"origin": "signup",
				"receive_only": true,
				"sending": {"enabled": false}
			}]
		}`, mailboxJSON)
	})

	resp, err := client.Mailboxes.List()
	if err != nil {
		t.Fatalf("Mailboxes.List returned error: %v", err)
	}
	assert.Equal(t, "list", resp.Object)
	assert.False(t, resp.HasMore)
	assert.Len(t, resp.Data, 2)

	mb := resp.Data[0]
	assert.Equal(t, "mailbox", mb.Object)
	assert.Equal(t, "mb_123", mb.Id)
	assert.Equal(t, "sales@example.com", mb.Address)
	assert.Equal(t, MailboxKindNormal, mb.Kind)
	assert.Nil(t, mb.AgentInboxId)
	assert.Equal(t, MailboxImapSmtpReady, mb.ImapSmtp)
	assert.Equal(t, "dom_1", *mb.DomainId)
	assert.Equal(t, "example.com", *mb.Domain)
	assert.False(t, mb.WarmupEnabled)
	assert.Nil(t, mb.WarmupChangedAt)
	assert.Equal(t, 12, mb.Sent)
	assert.Equal(t, 40, mb.SentTotal)
	assert.Equal(t, 11, mb.Delivered)
	assert.Equal(t, 1, mb.Bounced)
	assert.Equal(t, 0, mb.Complained)
	assert.Equal(t, "2026-10-20T09:00:00Z", *mb.LastSentAt)
	assert.Equal(t, MailboxHealthHealthy, mb.Health)
	assert.Nil(t, mb.Credentials)

	free := resp.Data[1]
	assert.Equal(t, MailboxKindAgent, free.Kind)
	assert.Equal(t, "inb_9", *free.AgentInboxId)
	assert.Nil(t, free.DomainId)
	assert.Nil(t, free.Domain)
	assert.True(t, free.WarmupEnabled)
	assert.Equal(t, MailboxHealthNoData, free.Health)
}

func TestCreateMailbox(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/mailboxes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		assert.Equal(t, map[string]any{"local_part": "sales", "domain_id": "dom_1"}, body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{
			"object": "mailbox",
			"id": "mb_123",
			"address": "sales@example.com",
			"kind": "normal",
			"imap_smtp": "ready",
			"warmup_enabled": false,
			"health": "no_data",
			"credentials": %s
		}`, mailboxCredentialsJSON)
	})

	mb, err := client.Mailboxes.Create(&CreateMailboxRequest{LocalPart: "sales", DomainId: "dom_1"})
	if err != nil {
		t.Fatalf("Mailboxes.Create returned error: %v", err)
	}
	assert.Equal(t, "mb_123", mb.Id)
	assert.Equal(t, "sales@example.com", mb.Address)
	if assert.NotNil(t, mb.Credentials) {
		assert.Equal(t, "sales@example.com", mb.Credentials.Username)
		assert.Equal(t, "mbx_secret", mb.Credentials.Password)
		assert.Equal(t, MailboxServer{Host: "imap.send.io", Port: 993, Security: "ssl"}, mb.Credentials.Imap)
		assert.Equal(t, MailboxServer{Host: "smtp.send.io", Port: 587, Security: "starttls"}, mb.Credentials.Smtp)
	}
}

func TestCreateFreeMailbox(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/mailboxes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		assert.Equal(t, map[string]any{"username": "bob", "org": "acme", "display_name": "Bob"}, body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"object": "mailbox", "id": "mb_free", "address": "bob@acme.send.io", "imap_smtp": "pending", "pool": "free"}`)
	})

	mb, err := client.Mailboxes.Create(&CreateMailboxRequest{Username: "bob", Org: "acme", DisplayName: "Bob"})
	if err != nil {
		t.Fatalf("Mailboxes.Create returned error: %v", err)
	}
	assert.Equal(t, "mb_free", mb.Id)
	assert.Equal(t, MailboxImapSmtpPending, mb.ImapSmtp)
	assert.Nil(t, mb.Credentials)
}

func TestUpdateMailboxSendsWarmupFalse(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/mailboxes/mb_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		b, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, `{"warmup_enabled": false}`, string(b))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, mailboxJSON)
	})

	mb, err := client.Mailboxes.UpdateWithContext(context.Background(), "mb_123", &UpdateMailboxRequest{WarmupEnabled: false})
	if err != nil {
		t.Fatalf("Mailboxes.Update returned error: %v", err)
	}
	assert.Equal(t, "mb_123", mb.Id)
	assert.False(t, mb.WarmupEnabled)
}

func TestUpdateMailboxSendsWarmupTrue(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/mailboxes/mb_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		b, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, `{"warmup_enabled": true}`, string(b))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"object": "mailbox", "id": "mb_123", "warmup_enabled": true}`)
	})

	mb, err := client.Mailboxes.Update("mb_123", &UpdateMailboxRequest{WarmupEnabled: true})
	if err != nil {
		t.Fatalf("Mailboxes.Update returned error: %v", err)
	}
	assert.True(t, mb.WarmupEnabled)
}

func TestResetMailboxPassword(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/mailboxes/mb_123/password", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		b, _ := io.ReadAll(r.Body)
		assert.Empty(t, b)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"object": "mailbox_credentials",
			"id": "mb_123",
			"address": "sales@example.com",
			"credentials": %s,
			"note": "The old password no longer works."
		}`, mailboxCredentialsJSON)
	})

	resp, err := client.Mailboxes.ResetPassword("mb_123")
	if err != nil {
		t.Fatalf("Mailboxes.ResetPassword returned error: %v", err)
	}
	assert.Equal(t, "mailbox_credentials", resp.Object)
	assert.Equal(t, "mb_123", resp.Id)
	assert.Equal(t, "sales@example.com", resp.Address)
	assert.Equal(t, "mbx_secret", resp.Credentials.Password)
	assert.Equal(t, 993, resp.Credentials.Imap.Port)
	assert.Equal(t, "starttls", resp.Credentials.Smtp.Security)
	assert.Equal(t, "The old password no longer works.", resp.Note)
}

func TestRemoveMailbox(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/mailboxes/mb_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"object": "mailbox", "id": "mb_123", "deleted": true}`)
	})

	resp, err := client.Mailboxes.Remove("mb_123")
	if err != nil {
		t.Fatalf("Mailboxes.Remove returned error: %v", err)
	}
	assert.Equal(t, "mailbox", resp.Object)
	assert.Equal(t, "mb_123", resp.Id)
	assert.True(t, resp.Deleted)
}
