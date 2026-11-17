package sendio

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendioUserAgent(t *testing.T) {
	client := NewClient("123")
	req, err := client.NewRequest(context.TODO(), "GET", "domains", nil)
	assert.NoError(t, err)
	assert.Equal(t, "sendio-go:1.2.0", req.Header.Get("User-Agent"))
}

func TestSendioBaseURLFromEnvAtNewClient(t *testing.T) {
	for _, raw := range []string{
		"https://api.sendio.run",
		"https://api.sendio.run/",
		"http://localhost:8080/v1",
		"http://localhost:8080/v1/",
	} {
		t.Setenv("SENDIO_BASE_URL", raw)
		client := NewClient("123")
		req, err := client.NewRequest(context.TODO(), "GET", "/emails/abc", nil)
		assert.NoError(t, err)
		assert.Equal(t, strings.TrimSuffix(raw, "/")+"/emails/abc", req.URL.String())

		req, err = client.NewRequest(context.TODO(), "GET", "domains", nil)
		assert.NoError(t, err)
		assert.Equal(t, strings.TrimSuffix(raw, "/")+"/domains", req.URL.String())
	}
}

func TestSendioDefaultBaseURLWhenEnvEmpty(t *testing.T) {
	t.Setenv("SENDIO_BASE_URL", "")
	client := NewClient("123")
	assert.Equal(t, "https://api.send.io/", client.BaseURL.String())
}

func TestSendioSetBaseURL(t *testing.T) {
	client := NewClient("123")
	assert.NoError(t, client.SetBaseURL("http://127.0.0.1:9999/prefix"))
	req, err := client.NewRequest(context.TODO(), "GET", "domains", nil)
	assert.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:9999/prefix/domains", req.URL.String())
	assert.Error(t, client.SetBaseURL("not a url"))
}

func TestSendioAPIKeyFromEnv(t *testing.T) {
	t.Setenv("SENDIO_API_KEY", "mx_test")
	client := NewClient("")
	assert.Equal(t, "mx_test", client.ApiKey)
}
