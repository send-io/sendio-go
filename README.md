# send.io Go SDK

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/send-io/sendio-go.svg)](https://pkg.go.dev/github.com/send-io/sendio-go)

The official Go library for the [send.io](https://send.io) email API.

- Module: `github.com/send-io/sendio-go`
- Package: `sendio`
- API reference: https://developer.send.io/docs/api

## Installation

```sh
go get github.com/send-io/sendio-go
```

Requires Go 1.23 or newer.

## Setup

Create an API key in the send.io dashboard. Keys start with `mx_`. Keep it out of source code and pass it through the environment:

```sh
export SENDIO_API_KEY="mx_..."
```

## Quick start: send an email

```go
package main

import (
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func main() {
	client := sendio.NewClient(os.Getenv("SENDIO_API_KEY"))

	sent, err := client.Emails.Send(&sendio.SendEmailRequest{
		From:    "Acme <onboarding@yourdomain.com>",
		To:      []string{"delivered@send.io"},
		Subject: "Hello from send.io",
		Html:    "<p>It works!</p>",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(sent.Id)
}
```

`sendio.NewClient("")` falls back to the `SENDIO_API_KEY` environment variable.

Every method has a `...WithContext(ctx, ...)` variant for deadlines and cancellation, and `SendWithOptions` accepts an idempotency key:

```go
sent, err := client.Emails.SendWithOptions(ctx, params, &sendio.SendEmailOptions{
	IdempotencyKey: "welcome-user-123",
})
```

`ReplyTo` takes one address. For several, use `ReplyToAddresses` (sent with `ReplyTo`, if set, as an array):

```go
params := &sendio.SendEmailRequest{
	From:             "Acme <onboarding@yourdomain.com>",
	To:               []string{"delivered@send.io"},
	Subject:          "Hello",
	Text:             "Hi!",
	ReplyToAddresses: []string{"support@yourdomain.com", "sales@yourdomain.com"},
}
```

## Batch sending

Send up to 100 emails in one request:

```go
batch, err := client.Batch.Send([]*sendio.SendEmailRequest{
	{
		From:    "Acme <onboarding@yourdomain.com>",
		To:      []string{"delivered@send.io"},
		Subject: "First",
		Html:    "<p>one</p>",
	},
	{
		From:    "Acme <onboarding@yourdomain.com>",
		To:      []string{"delivered@send.io"},
		Subject: "Second",
		Html:    "<p>two</p>",
	},
})
if err != nil {
	panic(err)
}
for _, e := range batch.Data {
	fmt.Println(e.Id)
}
```

## Transactional emails

A transactional email (a sign-in code, a password reset, a receipt) is a saved
send of a published template: it keeps the From, subject, reply-to and
tracking (off by default), and its content is the template's published
version. Create it once, then send it by id with only the recipient and the
variables.

```go
tx, err := client.Transactional.Create(&sendio.CreateTransactionalEmailRequest{
	Name:       "Sign-in code",
	TemplateId: "signin-code", // a published template, by id or alias
	From:       "Acme <security@mail.example.com>",
	Subject:    "Your code is {{CODE}}",
})
if err != nil {
	panic(err)
}

sent, err := client.Emails.Send(&sendio.SendEmailRequest{
	TransactionalId: tx.Id,
	To:              []string{"ada@example.com"},
	Variables:       map[string]any{"CODE": "123456"},
})
```

`Transactional` also has `List` (filter by `Search` or `TemplateId`), `Get`,
`Update` (set `ExpectedUpdatedAt` to refuse a stale save with 409), `Remove`,
`SendTest` and `Stats`.

## Retrieve an email

```go
email, err := client.Emails.Get("4ef9a417-02e9-4d39-ad75-9611e0fcc33c")
if err != nil {
	panic(err)
}
fmt.Println(email.Subject, email.LastEvent)
```

## Domains

```go
domain, err := client.Domains.Create(&sendio.CreateDomainRequest{Name: "yourdomain.com"})
if err != nil {
	panic(err)
}
for _, r := range domain.Records {
	fmt.Println(r.Type, r.Name, r.Value) // add these DNS records
}

_, err = client.Domains.Verify(domain.Id)

list, err := client.Domains.List()
for _, d := range list.Data {
	fmt.Println(d.Name, d.Status)
}
```

To receive email on a domain, add it with receiving enabled (the records then include an MX record):

```go
domain, err := client.Domains.Create(&sendio.CreateDomainRequest{
	Name:         "yourdomain.com",
	Capabilities: &sendio.DomainCapabilities{Receiving: sendio.DomainCapabilityStatusEnabled},
})
```

Turn receiving off or back on later with `Update`. Off stops send.io accepting email for the domain; remove its MX record afterwards. Turning it back on asks for the domain's MX record; publish it to start receiving email. Sending can't be disabled.

```go
_, err = client.Domains.Update(domain.Id, &sendio.UpdateDomainRequest{
	Capabilities: &sendio.DomainCapabilities{Receiving: sendio.DomainCapabilityStatusDisabled},
})
```

## Mailboxes

Mailboxes are email accounts with an IMAP/SMTP login. The domain must be verified (the API answers 409 if the address exists):

```go
mailbox, err := client.Mailboxes.Create(&sendio.CreateMailboxRequest{Address: "sales@yourdomain.com"})
if err != nil {
	panic(err)
}
if mailbox.Credentials != nil { // returned once, when IMAP/SMTP is ready
	fmt.Println(mailbox.Credentials.Username, mailbox.Credentials.Password)
	fmt.Println(mailbox.Credentials.Imap.Host, mailbox.Credentials.Smtp.Host)
}

list, err := client.Mailboxes.List()
for _, m := range list.Data {
	fmt.Println(m.Address, m.ImapSmtp, m.Health)
}

_, err = client.Mailboxes.Update(mailbox.Id, &sendio.UpdateMailboxRequest{WarmupEnabled: false})

reset, err := client.Mailboxes.ResetPassword(mailbox.Id) // the old password stops working
fmt.Println(reset.Credentials.Password)

_, err = client.Mailboxes.Remove(mailbox.Id) // also removes the agent inbox on it
```

The password is also an API key for that mailbox, so keep it secret.

The client also covers API keys, segments, contacts (with topics, segments, properties and imports), contact properties, broadcasts, templates, topics, webhooks, receiving (inbound emails and attachments), logs, automations (with automation templates), events, suppressions, OAuth grants and usage. See the [examples](./examples) folder for one file per resource.

## Configuration

| Setting | How |
|---|---|
| API key | `sendio.NewClient(key)`, or `SENDIO_API_KEY` when the key is empty |
| Base URL | `SENDIO_BASE_URL` environment variable, read every time a client is created (default `https://api.send.io/`) |
| Base URL per client | `client.SetBaseURL("http://localhost:8080")` |
| HTTP client / timeouts | `sendio.NewCustomClient(&http.Client{Timeout: 10 * time.Second}, key)` |

The base URL is accepted with or without a trailing slash, and a path prefix (for example `http://localhost:8080/v1`) is kept. Overriding it is mainly useful for tests against a mock server or a staging environment.

Every request sends `User-Agent: sendio-go:1.2.0`.

## Error handling

Non-2xx responses are returned as `*sendio.APIError`, with the HTTP `StatusCode`, the API's error `Name` (for example `validation_error`) and `Message`. Rate limits are returned as `*sendio.RateLimitError`, which carries the limit headers:

```go
sent, err := client.Emails.Send(params)
if err != nil {
	var rl *sendio.RateLimitError
	if errors.As(err, &rl) {
		fmt.Println("rate limited, retry after", rl.RetryAfter, "seconds")
		return
	}
	if errors.Is(err, sendio.ErrRateLimit) {
		// same condition, sentinel form
	}
	var apiErr *sendio.APIError
	if errors.As(err, &apiErr) {
		fmt.Println("send failed:", apiErr.StatusCode, apiErr.Name, apiErr.Message)
		return
	}
	fmt.Println("send failed:", err) // network error, cancelled context, ...
	return
}
```

`err.Error()` reads `[ERROR]: <message>`, as in earlier versions.

## Testing

```sh
go test ./...
```

A live smoke test (sends one real email) lives in `examples/smoke.go`:

```sh
SENDIO_API_KEY=mx_... SENDIO_BASE_URL=https://api.send.io go run ./examples/smoke.go
```

## Documentation

Full API reference: https://developer.send.io/docs/api

## License

MIT. See [LICENSE.md](./LICENSE.md) and [NOTICE](./NOTICE).
