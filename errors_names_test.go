package sendio

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// SENDIO-414: a send refused for credits answers 402 insufficient_credits,
// and a subaccount over its own quota answers 429 with a quota name. Both
// are readable by name.
func TestSendEmailInsufficientCredits(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/emails", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		fmt.Fprint(w, `{"statusCode":402,"name":"insufficient_credits","message":"Your credit balance is too low to send this email."}`)
	})

	_, err := client.Emails.Send(&SendEmailRequest{From: "x@example.com", To: []string{"d@example.com"}})

	var apiErr *APIError
	if assert.ErrorAs(t, err, &apiErr) {
		assert.Equal(t, http.StatusPaymentRequired, apiErr.StatusCode)
		assert.Equal(t, ErrorNameInsufficientCredits, apiErr.Name)
	}
}

func TestSendEmailSubaccountQuotaName(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/emails", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"statusCode":429,"name":"subaccount_daily_quota_exceeded","message":"This subaccount has reached its daily sending limit."}`)
	})

	_, err := client.Emails.Send(&SendEmailRequest{From: "x@example.com", To: []string{"d@example.com"}})

	var rl *RateLimitError
	if assert.ErrorAs(t, err, &rl) {
		assert.Equal(t, ErrorNameSubaccountDailyQuotaExceeded, rl.Name)
	}
	assert.ErrorIs(t, err, ErrRateLimit)
}
