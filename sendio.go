package sendio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	version     = "1.2.0"
	userAgent   = "sendio-go:" + version
	contentType = "application/json"

	// DefaultBaseURL is the send.io API endpoint used when SENDIO_BASE_URL is not set.
	DefaultBaseURL = "https://api.send.io/"
)

// resolveBaseURL returns the base URL to use for a new client. SENDIO_BASE_URL
// is read every time a client is created, so it can be changed at runtime
// (for example in tests). The value is accepted with or without a trailing slash.
func resolveBaseURL() string {
	if v := strings.TrimSpace(getEnv("SENDIO_BASE_URL", "")); v != "" {
		return v
	}
	return DefaultBaseURL
}

// parseBaseURL parses raw and guarantees a trailing slash on the path so that
// relative API paths are joined below it instead of replacing its last segment.
func parseBaseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("sendio: invalid base URL %q", raw)
	}
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	return u, nil
}

var defaultHTTPClient = &http.Client{
	Timeout: time.Minute,
}

// Options interface is used to define additional options that can be passed
// to the API methods.
type Options interface {
	// GetIdempotencyKey returns the idempotency key
	GetIdempotencyKey() string
}

// Client handles communication with the send.io API.
type Client struct {
	// HTTP client
	client *http.Client

	// Api Key
	ApiKey string

	// Base URL
	BaseURL *url.URL

	// User agent for client
	UserAgent string

	// HTTP headers
	headers map[string]string

	// Services
	Emails   *EmailsSvcImpl
	Batch    BatchSvc
	ApiKeys  ApiKeysSvc
	Domains  DomainsSvc
	Segments SegmentsSvc
	// Deprecated: Use Segments instead. Audiences have been renamed to Segments.
	Audiences         AudiencesSvc
	Contacts          *ContactsSvcImpl
	ContactProperties ContactPropertiesSvc
	Broadcasts        BroadcastsSvc
	Templates         TemplatesSvc
	// Transactional manages transactional emails (saved sends of a
	// published template, sent with SendEmailRequest.TransactionalId).
	Transactional TransactionalSvc
	Topics        TopicsSvc
	Webhooks      WebhooksSvc
	Logs          LogsSvc
	Automations   *AutomationsSvcImpl
	Events        EventsSvc
	OAuthGrants   OAuthGrantsSvc
	Suppressions  *SuppressionsSvcImpl
	Usage         UsageSvc
	Mailboxes     MailboxesSvc
}

// NewClient is the default client constructor. When apiKey is empty, the
// SENDIO_API_KEY environment variable is used.
func NewClient(apiKey string) *Client {
	key := strings.Trim(strings.TrimSpace(apiKey), "'")
	if key == "" {
		key = strings.Trim(strings.TrimSpace(getEnv("SENDIO_API_KEY", "")), "'")
	}
	return NewCustomClient(defaultHTTPClient, key)
}

// NewCustomClient builds a new send.io API client, using a provided Http client.
func NewCustomClient(httpClient *http.Client, apiKey string) *Client {
	if httpClient == nil {
		httpClient = defaultHTTPClient
	}

	baseURL, err := parseBaseURL(resolveBaseURL())
	if err != nil {
		baseURL, _ = parseBaseURL(DefaultBaseURL)
	}

	c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent}

	emailsSvc := &EmailsSvcImpl{client: c}
	emailsSvc.Receiving = &ReceivingSvcImpl{client: c}
	c.Emails = emailsSvc

	contactsSvc := &ContactsSvcImpl{client: c}
	contactsSvc.Topics = &ContactTopicsSvcImpl{client: c}
	contactsSvc.Segments = &ContactSegmentsSvcImpl{client: c}
	contactsSvc.Properties = &ContactPropertiesSvcImpl{client: c}
	contactsSvc.Imports = &ContactImportsSvcImpl{client: c}
	c.Contacts = contactsSvc

	automationsSvc := &AutomationsSvcImpl{client: c}
	automationsSvc.Templates = &AutomationTemplatesSvcImpl{client: c}
	c.Automations = automationsSvc

	suppressionsSvc := &SuppressionsSvcImpl{client: c}
	suppressionsSvc.Batch = &SuppressionsBatchSvcImpl{client: c}
	c.Suppressions = suppressionsSvc

	c.Batch = &BatchSvcImpl{client: c}
	c.ApiKeys = &ApiKeysSvcImpl{client: c}
	c.Domains = &DomainsSvcImpl{client: c}
	segmentsSvc := &SegmentsSvcImpl{client: c}
	c.Segments = segmentsSvc
	// Audiences is a deprecated alias for Segments for backward compatibility
	c.Audiences = &AudiencesSvcImpl{segments: segmentsSvc}
	c.ContactProperties = &ContactPropertiesSvcImpl{client: c}
	c.Broadcasts = &BroadcastsSvcImpl{client: c}
	c.Templates = &TemplatesSvcImpl{client: c}
	c.Transactional = &TransactionalSvcImpl{client: c}
	c.Topics = &TopicsSvcImpl{client: c}
	c.Webhooks = &WebhooksSvcImpl{client: c}
	c.Logs = &LogsSvcImpl{client: c}
	c.Events = &EventsSvcImpl{client: c}
	c.OAuthGrants = &OAuthGrantsSvcImpl{client: c}
	c.Usage = &UsageSvcImpl{client: c}
	c.Mailboxes = &MailboxesSvcImpl{client: c}

	c.ApiKey = apiKey
	c.headers = make(map[string]string)
	return c
}

// SetBaseURL points the client at a different API endpoint (for example a
// local mock server in tests). The URL is accepted with or without a trailing slash.
func (c *Client) SetBaseURL(rawURL string) error {
	u, err := parseBaseURL(rawURL)
	if err != nil {
		return err
	}
	c.BaseURL = u
	return nil
}

// NewRequestWithOptions builds and returns a new HTTP request object
// based on the given arguments and options
// It is used to set additional options like idempotency key
func (c *Client) NewRequestWithOptions(ctx context.Context, method, path string, params any, options Options) (*http.Request, error) {
	req, err := c.NewRequest(ctx, method, path, params)

	if err != nil {
		return nil, err
	}

	// Set the idempotency key if provided and the method is POST for now.
	if options != nil {
		if options.GetIdempotencyKey() != "" && method == http.MethodPost {
			req.Header.Set("Idempotency-Key", options.GetIdempotencyKey())
		}

		// Handle batch-specific options
		if batchOptions, ok := options.(*BatchSendEmailOptions); ok {
			if batchOptions.GetBatchValidation() != "" {
				req.Header.Set("x-batch-validation", batchOptions.GetBatchValidation())
			}
		}
	}

	return req, nil
}

// NewRequest builds and returns a new HTTP request object
// based on the given arguments
func (c *Client) NewRequest(ctx context.Context, method, path string, params any) (*http.Request, error) {
	u, err := c.BaseURL.Parse(strings.TrimPrefix(path, "/"))
	if err != nil {
		return nil, err
	}

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		buf := new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(params)
		if err != nil {
			return nil, err
		}

		req.Body = io.NopCloser(buf)
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	req.Header.Set("Accept", contentType)
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	return req, nil
}

// Perform sends the request to the send.io API
func (c *Client) Perform(req *http.Request, ret any) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle possible errors.
	// Any 2xx status code is considered success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, handleError(resp)
	}

	if resp.StatusCode != http.StatusNoContent && ret != nil {
		if w, ok := ret.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			if resp.Body != nil {
				err = json.NewDecoder(resp.Body).Decode(ret)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return resp, err
}

// handleError turns a non-2xx response into an error: *RateLimitError for 429,
// *APIError for everything else.
func handleError(resp *http.Response) error {
	apiErr := &APIError{}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(resp.Body).Decode(apiErr); err != nil {
			apiErr.Message = resp.Status
		}
	} else {
		apiErr.Message = resp.Status
	}
	// The HTTP status is authoritative, whatever the body says.
	apiErr.StatusCode = resp.StatusCode

	if resp.StatusCode == http.StatusTooManyRequests {
		return &RateLimitError{
			Name:       apiErr.Name,
			Message:    apiErr.Message,
			Limit:      resp.Header.Get("ratelimit-limit"),
			Remaining:  resp.Header.Get("ratelimit-remaining"),
			Reset:      resp.Header.Get("ratelimit-reset"),
			RetryAfter: resp.Header.Get("retry-after"),
		}
	}
	return apiErr
}

// InvalidRequestError is the shape of an API error body.
//
// Deprecated: API errors are returned as *APIError, which has the same fields.
type InvalidRequestError struct {
	StatusCode int    `json:"statusCode"`
	Name       string `json:"name"`
	Message    string `json:"message"`
}

// DefaultError holds the message of an API error body.
//
// Deprecated: API errors are returned as *APIError.
type DefaultError struct {
	Message string `json:"message"`
}

// ListOptions contains pagination parameters for list methods
type ListOptions struct {
	Limit  *int    `json:"limit,omitempty"`
	After  *string `json:"after,omitempty"`
	Before *string `json:"before,omitempty"`
}

// paginationValues returns the limit, after and before query parameters that are set.
func paginationValues(limit *int, after, before *string) url.Values {
	query := make(url.Values)
	if limit != nil {
		query.Set("limit", fmt.Sprintf("%d", *limit))
	}
	if after != nil {
		query.Set("after", *after)
	}
	if before != nil {
		query.Set("before", *before)
	}
	return query
}

// buildPaginationQuery constructs query parameters for pagination
func buildPaginationQuery(options *ListOptions) string {
	if options == nil {
		return ""
	}

	query := paginationValues(options.Limit, options.After, options.Before)
	if len(query) > 0 {
		return "?" + query.Encode()
	}
	return ""
}
