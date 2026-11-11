//go:build ignore

// Live smoke test against a real send.io endpoint. Not part of the unit tests.
//
//	SENDIO_API_KEY=mx_... SENDIO_BASE_URL=https://api.send.io go run ./examples/smoke.go
//
// Optional: SMOKE_FROM (default ada@compat-sdk.sendio-test.dev), SMOKE_TO (default delivered@send.io).
package main

import (
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	client := sendio.NewClient(os.Getenv("SENDIO_API_KEY"))
	fmt.Println("base URL:", client.BaseURL.String())
	failed := false

	sent, err := client.Emails.Send(&sendio.SendEmailRequest{
		From:    env("SMOKE_FROM", "ada@compat-sdk.sendio-test.dev"),
		To:      []string{env("SMOKE_TO", "delivered@send.io")},
		Subject: "sendio-go smoke test",
		Html:    "<p>Hello from the sendio Go SDK</p>",
		Text:    "Hello from the sendio Go SDK",
	})
	if err != nil {
		fmt.Println("send: FAIL", err)
		os.Exit(1)
	}
	fmt.Println("send: OK id =", sent.Id)

	email, err := client.Emails.Get(sent.Id)
	if err != nil {
		fmt.Println("get: FAIL", err)
		failed = true
	} else {
		fmt.Printf("get: OK id=%s from=%s last_event=%s\n", email.Id, email.From, email.LastEvent)
	}

	domains, err := client.Domains.List()
	if err != nil {
		fmt.Println("domains.list: FAIL", err)
		failed = true
	} else {
		fmt.Printf("domains.list: OK %d domain(s)\n", len(domains.Data))
	}

	keys, err := client.ApiKeys.List()
	if err != nil {
		fmt.Println("api_keys.list: FAIL", err)
		failed = true
	} else {
		fmt.Printf("api_keys.list: OK %d key(s)\n", len(keys.Data))
	}

	if failed {
		os.Exit(1)
	}
}
