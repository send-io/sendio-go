package sendio

import (
	"context"
	"net/http"
	"net/url"
)

// A transactional email is a named, saved send of a published template: a
// sign-in code, a password reset, a receipt. It stores the envelope (From,
// subject, preview text, reply-to and the open and click tracking, off by
// default) and a template; the content is always the template's published
// version at send time. Send it with Emails.Send and
// SendEmailRequest.TransactionalId, passing only the recipient and its
// Variables.

// TransactionalEmail is a transactional email as the API returns it.
type TransactionalEmail struct {
	Object string `json:"object"`
	// Id is the transactional email id ("tx_" and 27 letters and digits).
	Id           string `json:"id"`
	Name         string `json:"name"`
	TemplateId   string `json:"template_id"`
	TemplateName string `json:"template_name"`
	// TemplateAlias is the template's alias ("" when it has none).
	TemplateAlias string `json:"template_alias"`
	// TemplateDeleted is true when the template was deleted: sends then fail
	// until another template is chosen.
	TemplateDeleted bool `json:"template_deleted"`
	// PublishedVersionId and PublishedVersionNumber are the template version
	// sends use now (nil when the template has no published version).
	PublishedVersionId     *string  `json:"published_version_id"`
	PublishedVersionNumber *int     `json:"published_version_number"`
	From                   string   `json:"from"`
	Subject                string   `json:"subject"`
	PreviewText            *string  `json:"preview_text"`
	ReplyTo                []string `json:"reply_to"`
	TrackOpens             bool     `json:"track_opens"`
	TrackClicks            bool     `json:"track_clicks"`
	// CreatedVia is "api" or "dashboard".
	CreatedVia string `json:"created_via"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	// Variables are the variables of the published version (Get only): what
	// a send's Variables fills.
	Variables []*TransactionalEmailVariable `json:"variables,omitempty"`
}

// TransactionalEmailVariable is one variable of the template version a
// transactional email sends.
type TransactionalEmailVariable struct {
	Key           string `json:"key"`
	Type          string `json:"type"`
	FallbackValue any    `json:"fallback_value"`
}

// CreateTransactionalEmailRequest is the request payload for creating a
// transactional email. Name, TemplateId (id or alias of a published
// template), From (on a verified sending domain) and Subject are required.
type CreateTransactionalEmailRequest struct {
	Name        string   `json:"name"`
	TemplateId  string   `json:"template_id"`
	From        string   `json:"from"`
	Subject     string   `json:"subject"`
	PreviewText string   `json:"preview_text,omitempty"`
	ReplyTo     []string `json:"reply_to,omitempty"`
	// TrackOpens and TrackClicks default to off; sendio.Bool(true) turns them on.
	TrackOpens  *bool `json:"track_opens,omitempty"`
	TrackClicks *bool `json:"track_clicks,omitempty"`
}

// UpdateTransactionalEmailRequest is the request payload for updating a
// transactional email: only the fields set are sent. Set
// ExpectedUpdatedAt to the UpdatedAt you read to get a 409 conflict instead
// of overwriting a change made since.
type UpdateTransactionalEmailRequest struct {
	Name        string  `json:"name,omitempty"`
	TemplateId  string  `json:"template_id,omitempty"`
	From        string  `json:"from,omitempty"`
	Subject     string  `json:"subject,omitempty"`
	PreviewText *string `json:"preview_text,omitempty"`
	// ReplyTo replaces the reply-to addresses when not nil; an empty slice clears them.
	ReplyTo           *[]string `json:"reply_to,omitempty"`
	TrackOpens        *bool     `json:"track_opens,omitempty"`
	TrackClicks       *bool     `json:"track_clicks,omitempty"`
	ExpectedUpdatedAt string    `json:"expected_updated_at,omitempty"`
}

// TransactionalEmailOptions are the options of a create.
type TransactionalEmailOptions struct {
	// IdempotencyKey makes a retried create return the first answer.
	IdempotencyKey string `json:"-"`
}

// GetIdempotencyKey returns the idempotency key.
func (o TransactionalEmailOptions) GetIdempotencyKey() string { return o.IdempotencyKey }

// ListTransactionalEmailsOptions filters and pages a list.
type ListTransactionalEmailsOptions struct {
	Limit  *int
	After  *string
	Before *string
	// Search matches the name and subject (case-insensitive).
	Search string
	// TemplateId lists only the transactional emails of this template (id or alias).
	TemplateId string
}

// ListTransactionalEmailsResponse is the response from listing transactional emails.
type ListTransactionalEmailsResponse struct {
	Object  string                `json:"object"`
	HasMore bool                  `json:"has_more"`
	Data    []*TransactionalEmail `json:"data"`
}

// RemoveTransactionalEmailResponse is the response from deleting a transactional email.
type RemoveTransactionalEmailResponse struct {
	Object  string `json:"object"`
	Id      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// SendTestTransactionalEmailRequest sends a transactional email to 1 to 5
// addresses, with "[Test] " before the subject. Tests are not counted in stats.
type SendTestTransactionalEmailRequest struct {
	To        []string       `json:"to"`
	Variables map[string]any `json:"variables,omitempty"`
}

// TransactionalEmailStatsOptions selects the period of Stats. Every field is optional.
type TransactionalEmailStatsOptions struct {
	// From and To: a date (2026-11-01) or an ISO 8601 time; at most 90 days apart.
	From string
	To   string
	// Timezone is an IANA time zone (default UTC).
	Timezone string
	// Version narrows to one template version (an id, or a version_number of the current template).
	Version string
	// Granularity is "daily" (default) or "hourly".
	Granularity string
}

// TransactionalEmailStats are the numbers of a transactional email's sends.
// Totals, Daily, PerVersion and TopLinks are the template stats' shapes.
type TransactionalEmailStats struct {
	Object          string           `json:"object"`
	TransactionalId string           `json:"transactional_id"`
	TemplateId      string           `json:"template_id"`
	Range           map[string]any   `json:"range"`
	Version         map[string]any   `json:"version"`
	LiveVersion     map[string]any   `json:"live_version"`
	Totals          map[string]any   `json:"totals"`
	Daily           []map[string]any `json:"daily"`
	PerVersion      []map[string]any `json:"per_version"`
	TopLinks        []map[string]any `json:"top_links"`
}

// TransactionalSvc handles transactional emails (/transactional).
type TransactionalSvc interface {
	CreateWithContext(ctx context.Context, params *CreateTransactionalEmailRequest) (*TransactionalEmail, error)
	CreateWithOptions(ctx context.Context, params *CreateTransactionalEmailRequest, options *TransactionalEmailOptions) (*TransactionalEmail, error)
	Create(params *CreateTransactionalEmailRequest) (*TransactionalEmail, error)
	ListWithContext(ctx context.Context, options *ListTransactionalEmailsOptions) (*ListTransactionalEmailsResponse, error)
	List(options *ListTransactionalEmailsOptions) (*ListTransactionalEmailsResponse, error)
	GetWithContext(ctx context.Context, id string) (*TransactionalEmail, error)
	Get(id string) (*TransactionalEmail, error)
	UpdateWithContext(ctx context.Context, id string, params *UpdateTransactionalEmailRequest) (*TransactionalEmail, error)
	Update(id string, params *UpdateTransactionalEmailRequest) (*TransactionalEmail, error)
	RemoveWithContext(ctx context.Context, id string) (*RemoveTransactionalEmailResponse, error)
	Remove(id string) (*RemoveTransactionalEmailResponse, error)
	SendTestWithContext(ctx context.Context, id string, params *SendTestTransactionalEmailRequest) (*SendEmailResponse, error)
	SendTest(id string, params *SendTestTransactionalEmailRequest) (*SendEmailResponse, error)
	StatsWithContext(ctx context.Context, id string, options *TransactionalEmailStatsOptions) (*TransactionalEmailStats, error)
	Stats(id string, options *TransactionalEmailStatsOptions) (*TransactionalEmailStats, error)
}

// TransactionalSvcImpl is the implementation of the TransactionalSvc interface.
type TransactionalSvcImpl struct {
	client *Client
}

func transactionalPath(id string, rest ...string) string {
	p := "transactional/" + url.PathEscape(id)
	for _, r := range rest {
		p += "/" + r
	}
	return p
}

// do sends one request and decodes the answer into ret.
func (s *TransactionalSvcImpl) do(req *http.Request, err error, failed error, ret any) error {
	if err != nil {
		return failed
	}
	_, err = s.client.Perform(req, ret)
	return err
}

// CreateWithContext creates a transactional email.
// https://developer.send.io/docs/api#create-transactional-email
func (s *TransactionalSvcImpl) CreateWithContext(ctx context.Context, params *CreateTransactionalEmailRequest) (*TransactionalEmail, error) {
	return s.CreateWithOptions(ctx, params, nil)
}

// CreateWithOptions creates a transactional email, with an Idempotency-Key when options sets one.
// https://developer.send.io/docs/api#create-transactional-email
func (s *TransactionalSvcImpl) CreateWithOptions(ctx context.Context, params *CreateTransactionalEmailRequest, options *TransactionalEmailOptions) (*TransactionalEmail, error) {
	var req *http.Request
	var err error
	if options != nil {
		req, err = s.client.NewRequestWithOptions(ctx, http.MethodPost, "transactional", params, options)
	} else {
		req, err = s.client.NewRequest(ctx, http.MethodPost, "transactional", params)
	}
	out := new(TransactionalEmail)
	if err := s.do(req, err, ErrFailedToCreateTransactionalCreateRequest, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create creates a transactional email. See CreateWithContext.
func (s *TransactionalSvcImpl) Create(params *CreateTransactionalEmailRequest) (*TransactionalEmail, error) {
	return s.CreateWithContext(context.Background(), params)
}

// ListWithContext lists transactional emails, newest first.
// https://developer.send.io/docs/api#list-transactional-emails
func (s *TransactionalSvcImpl) ListWithContext(ctx context.Context, options *ListTransactionalEmailsOptions) (*ListTransactionalEmailsResponse, error) {
	path := "transactional"
	if options != nil {
		q := paginationValues(options.Limit, options.After, options.Before)
		if options.Search != "" {
			q.Set("search", options.Search)
		}
		if options.TemplateId != "" {
			q.Set("template_id", options.TemplateId)
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	out := new(ListTransactionalEmailsResponse)
	if err := s.do(req, err, ErrFailedToCreateTransactionalListRequest, out); err != nil {
		return nil, err
	}
	return out, nil
}

// List lists transactional emails. See ListWithContext.
func (s *TransactionalSvcImpl) List(options *ListTransactionalEmailsOptions) (*ListTransactionalEmailsResponse, error) {
	return s.ListWithContext(context.Background(), options)
}

// GetWithContext gets a transactional email with the variables of the template version its sends use now.
// https://developer.send.io/docs/api#retrieve-transactional-email
func (s *TransactionalSvcImpl) GetWithContext(ctx context.Context, id string) (*TransactionalEmail, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, transactionalPath(id), nil)
	out := new(TransactionalEmail)
	if err := s.do(req, err, ErrFailedToCreateTransactionalGetRequest, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get gets a transactional email. See GetWithContext.
func (s *TransactionalSvcImpl) Get(id string) (*TransactionalEmail, error) {
	return s.GetWithContext(context.Background(), id)
}

// UpdateWithContext changes a transactional email; sends use the change within 10 seconds.
// https://developer.send.io/docs/api#update-transactional-email
func (s *TransactionalSvcImpl) UpdateWithContext(ctx context.Context, id string, params *UpdateTransactionalEmailRequest) (*TransactionalEmail, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPatch, transactionalPath(id), params)
	out := new(TransactionalEmail)
	if err := s.do(req, err, ErrFailedToCreateTransactionalUpdateRequest, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Update changes a transactional email. See UpdateWithContext.
func (s *TransactionalSvcImpl) Update(id string, params *UpdateTransactionalEmailRequest) (*TransactionalEmail, error) {
	return s.UpdateWithContext(context.Background(), id, params)
}

// RemoveWithContext deletes a transactional email; sends of it then fail with 404.
// https://developer.send.io/docs/api#delete-transactional-email
func (s *TransactionalSvcImpl) RemoveWithContext(ctx context.Context, id string) (*RemoveTransactionalEmailResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, transactionalPath(id), nil)
	out := new(RemoveTransactionalEmailResponse)
	if err := s.do(req, err, ErrFailedToCreateTransactionalRemoveRequest, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Remove deletes a transactional email. See RemoveWithContext.
func (s *TransactionalSvcImpl) Remove(id string) (*RemoveTransactionalEmailResponse, error) {
	return s.RemoveWithContext(context.Background(), id)
}

// SendTestWithContext sends a transactional email to 1 to 5 addresses now.
// https://developer.send.io/docs/api#test-transactional-email
func (s *TransactionalSvcImpl) SendTestWithContext(ctx context.Context, id string, params *SendTestTransactionalEmailRequest) (*SendEmailResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, transactionalPath(id, "test"), params)
	out := new(SendEmailResponse)
	if err := s.do(req, err, ErrFailedToCreateTransactionalSendTestRequest, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SendTest sends a test of a transactional email. See SendTestWithContext.
func (s *TransactionalSvcImpl) SendTest(id string, params *SendTestTransactionalEmailRequest) (*SendEmailResponse, error) {
	return s.SendTestWithContext(context.Background(), id, params)
}

// StatsWithContext reads a transactional email's numbers in a period.
// https://developer.send.io/docs/api#transactional-email-stats
func (s *TransactionalSvcImpl) StatsWithContext(ctx context.Context, id string, options *TransactionalEmailStatsOptions) (*TransactionalEmailStats, error) {
	path := transactionalPath(id, "stats")
	if options != nil {
		q := url.Values{}
		for k, v := range map[string]string{"from": options.From, "to": options.To, "timezone": options.Timezone, "version": options.Version, "granularity": options.Granularity} {
			if v != "" {
				q.Set(k, v)
			}
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	out := new(TransactionalEmailStats)
	if err := s.do(req, err, ErrFailedToCreateTransactionalStatsRequest, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Stats reads a transactional email's numbers. See StatsWithContext.
func (s *TransactionalSvcImpl) Stats(id string, options *TransactionalEmailStatsOptions) (*TransactionalEmailStats, error) {
	return s.StatsWithContext(context.Background(), id, options)
}
