package sendio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type SendBroadcastRequest struct {
	// BroadcastId goes in the path only, never in the JSON body.
	BroadcastId string `json:"-"`

	// Schedule email to be sent later. The date should be a relative time (e.g.: in 30 minutes, in 2 days)
	// or ISO 8601 format (e.g: 2024-08-05T11:52:01.858Z).
	ScheduledAt string `json:"scheduled_at"`
}

type CreateBroadcastRequest struct {
	SegmentId  string   `json:"segment_id,omitempty"`
	AudienceId string   `json:"audience_id,omitempty"` // Deprecated: Use SegmentId instead
	From       string   `json:"from,omitempty"`
	Subject    string   `json:"subject,omitempty"`
	ReplyTo    []string `json:"reply_to,omitempty"`
	Html       string   `json:"html,omitempty"`
	Text       string   `json:"text,omitempty"`
	Name       string   `json:"name,omitempty"`

	// PreviewText is the snippet inboxes show after the subject (preheader).
	PreviewText string `json:"preview_text,omitempty"`

	// TopicId scopes the broadcast to a topic: contacts who opted out of it
	// are skipped, and unsubscribing leaves the topic.
	TopicId string `json:"topic_id,omitempty"`

	// send.io extensions.

	// SegmentIds sends to several segments; a contact in more than one gets one
	// email. It can be combined with SegmentId.
	SegmentIds []string `json:"segment_ids,omitempty"`

	// ExcludeSegmentIds: members of these segments never get the broadcast.
	ExcludeSegmentIds []string `json:"exclude_segment_ids,omitempty"`

	// Audience is "segments" (the default) or "all" (every contact; not
	// combined with SegmentId or SegmentIds).
	Audience string `json:"audience,omitempty"`

	// UnsubscribeScope is "all" (the default: unsubscribing leaves everything,
	// or the topic) or "list" (it leaves the broadcast's segments only).
	UnsubscribeScope string `json:"unsubscribe_scope,omitempty"`

	// Tracking counts opens and clicks (the default, on). Set it with
	// sendio.Bool(false) to turn tracking off; nil keeps the default.
	Tracking *bool `json:"tracking,omitempty"`

	// TemplateEngine is "simple" (merge tags, the default) or "handlebars",
	// rendered for each recipient with the contact as data.
	TemplateEngine string `json:"template_engine,omitempty"`

	// TemplateId makes the broadcast from a published template (id or
	// alias): it copies the version's content, and its subject, From,
	// preview text and reply-to where the request has none. Not with Html,
	// Text or TemplateEngine.
	TemplateId string `json:"template_id,omitempty"`
	// TemplateVersionId picks the published version to copy (default: the
	// one published now).
	TemplateVersionId string `json:"template_version_id,omitempty"`

	// Send the broadcast immediately upon creation instead of creating a draft.
	Send bool `json:"send,omitempty"`

	// Schedule email to be sent later. The date should be a relative time (e.g.: in 30 minutes, in 2 days)
	// or ISO 8601 format (e.g: 2024-08-05T11:52:01.858Z).
	// Only valid when Send is true.
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

type UpdateBroadcastRequest struct {
	// BroadcastId goes in the path only, never in the JSON body: once a
	// broadcast is sending or sent, the API refuses every key but name.
	BroadcastId string   `json:"-"`
	SegmentId   string   `json:"segment_id,omitempty"`
	AudienceId  string   `json:"audience_id,omitempty"` // Deprecated: Use SegmentId instead
	From        string   `json:"from,omitempty"`
	Subject     string   `json:"subject,omitempty"`
	ReplyTo     []string `json:"reply_to,omitempty"`
	Html        string   `json:"html,omitempty"`
	Text        string   `json:"text,omitempty"`
	Name        string   `json:"name,omitempty"`

	// PreviewText is the snippet inboxes show after the subject (preheader).
	PreviewText string `json:"preview_text,omitempty"`

	// TopicId scopes the broadcast to a topic: contacts who opted out of it
	// are skipped, and unsubscribing leaves the topic.
	TopicId string `json:"topic_id,omitempty"`

	// send.io extensions.

	// Fields left empty (or nil) are not sent, so Update leaves them as
	// they are; an empty SegmentIds or ExcludeSegmentIds cannot clear a list.

	// SegmentIds sends to several segments; a contact in more than one gets one
	// email. It can be combined with SegmentId.
	SegmentIds []string `json:"segment_ids,omitempty"`

	// ExcludeSegmentIds: members of these segments never get the broadcast.
	ExcludeSegmentIds []string `json:"exclude_segment_ids,omitempty"`

	// Audience is "segments" (the default) or "all" (every contact; not
	// combined with SegmentId or SegmentIds).
	Audience string `json:"audience,omitempty"`

	// UnsubscribeScope is "all" (the default: unsubscribing leaves everything,
	// or the topic) or "list" (it leaves the broadcast's segments only).
	UnsubscribeScope string `json:"unsubscribe_scope,omitempty"`

	// Tracking counts opens and clicks (the default, on). Set it with
	// sendio.Bool(false) to turn tracking off; nil leaves it as it is.
	Tracking *bool `json:"tracking,omitempty"`

	// TemplateEngine is "simple" (merge tags, the default) or "handlebars",
	// rendered for each recipient with the contact as data.
	TemplateEngine string `json:"template_engine,omitempty"`

	// TemplateId links the broadcast to a published template (id or alias)
	// and copies its version's content. It fills subject, From, preview text
	// and reply-to only where the broadcast has none. Sending Html, Text or
	// TemplateEngine instead gives the broadcast its own content and
	// detaches it from its template. Use DetachTemplate to detach it
	// without changing the content.
	TemplateId string `json:"template_id,omitempty"`
	// TemplateVersionId picks the published version to copy; alone, it
	// switches to another version of the broadcast's template.
	TemplateVersionId string `json:"template_version_id,omitempty"`
	// DetachTemplate sends template_id: null (the content stays).
	DetachTemplate bool `json:"-"`
}

// MarshalJSON sends template_id: null for DetachTemplate.
func (r UpdateBroadcastRequest) MarshalJSON() ([]byte, error) {
	type plain UpdateBroadcastRequest
	if !r.DetachTemplate {
		return json.Marshal(plain(r))
	}
	return json.Marshal(struct {
		plain
		TemplateId *string `json:"template_id"`
	}{plain: plain(r)})
}

type CreateBroadcastResponse struct {
	Id string `json:"id"`
}

type UpdateBroadcastResponse struct {
	Id string `json:"id"`
}

type SendBroadcastResponse struct {
	Id string `json:"id"`
}

type CancelBroadcastResponse struct {
	Object string `json:"object"`
	Id     string `json:"id"`
}

type DuplicateBroadcastResponse struct {
	Object string `json:"object"`
	Id     string `json:"id"`
}

type RemoveBroadcastResponse struct {
	Object  string `json:"object"`
	Id      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

type ListBroadcastsResponse struct {
	Object  string      `json:"object"`
	Data    []Broadcast `json:"data"`
	HasMore bool        `json:"has_more"`
}

// BroadcastRecipientEventType is the recipient event type to filter by when listing a
// broadcast's recipients.
type BroadcastRecipientEventType = string

const (
	BroadcastRecipientEventTypeSent         BroadcastRecipientEventType = "sent"
	BroadcastRecipientEventTypeDelivered    BroadcastRecipientEventType = "delivered"
	BroadcastRecipientEventTypeOpened       BroadcastRecipientEventType = "opened"
	BroadcastRecipientEventTypeClicked      BroadcastRecipientEventType = "clicked"
	BroadcastRecipientEventTypeBounced      BroadcastRecipientEventType = "bounced"
	BroadcastRecipientEventTypeComplained   BroadcastRecipientEventType = "complained"
	BroadcastRecipientEventTypeUnsubscribed BroadcastRecipientEventType = "unsubscribed"
	BroadcastRecipientEventTypeSuppressed   BroadcastRecipientEventType = "suppressed"
)

// BroadcastRecipientBounceType is the bounce type to filter by. Only meaningful when
// ListBroadcastRecipientsOptions.Type is BroadcastRecipientEventTypeBounced.
type BroadcastRecipientBounceType = string

const (
	BroadcastRecipientBounceTypePermanent    BroadcastRecipientBounceType = "permanent"
	BroadcastRecipientBounceTypeTransient    BroadcastRecipientBounceType = "transient"
	BroadcastRecipientBounceTypeUndetermined BroadcastRecipientBounceType = "undetermined"
)

// BroadcastRecipientClickedLink is a link clicked by a recipient. Only present when Type is
// BroadcastRecipientEventTypeClicked.
type BroadcastRecipientClickedLink struct {
	Url    string `json:"url"`
	Clicks int    `json:"clicks"`
}

// BroadcastRecipient is a single recipient row for the requested event type. Count, BounceType
// and ClickedLinks are only populated depending on ListBroadcastRecipientsOptions.Type.
type BroadcastRecipient struct {
	// Id is an opaque cursor identifying this row, used for pagination. It is not a real entity id.
	Id        string  `json:"id"`
	ContactId *string `json:"contact_id"`
	Email     string  `json:"email"`

	// Count is the number of times this recipient triggered the event. Only present when Type is
	// BroadcastRecipientEventTypeOpened or BroadcastRecipientEventTypeClicked.
	Count int `json:"count,omitempty"`

	// BounceType is only present when Type is BroadcastRecipientEventTypeBounced.
	BounceType BroadcastRecipientBounceType `json:"bounce_type,omitempty"`

	// ClickedLinks is only present when Type is BroadcastRecipientEventTypeClicked.
	ClickedLinks []BroadcastRecipientClickedLink `json:"clicked_links,omitempty"`
}

type ListBroadcastRecipientsResponse struct {
	Object  string               `json:"object"`
	HasMore bool                 `json:"has_more"`
	Data    []BroadcastRecipient `json:"data"`
}

// ListBroadcastRecipientsOptions contains the filter and pagination parameters for listing a
// broadcast's recipients.
type ListBroadcastRecipientsOptions struct {
	// Type is the recipient event type to filter by. Required.
	Type BroadcastRecipientEventType

	// Email filters recipients whose email address contains this substring.
	Email string

	// BounceType filters bounced recipients by bounce type. Only meaningful when Type is
	// BroadcastRecipientEventTypeBounced.
	BounceType BroadcastRecipientBounceType

	Limit  *int
	After  *string
	Before *string
}

type ListBroadcastClickedLinksResponse struct {
	Object  string                 `json:"object"`
	Data    []BroadcastClickedLink `json:"data"`
	HasMore bool                   `json:"has_more"`
}

type BroadcastClickedLink struct {
	// Id is an opaque cursor for this row, used only for pagination. It does
	// not identify any entity in send.io.
	Id           string `json:"id"`
	Url          string `json:"url"`
	Clicks       int    `json:"clicks"`
	UniqueClicks int    `json:"unique_clicks"`
}

// Broadcast is a broadcast as Get returns it (List returns a subset of the
// fields). Fields the API returns as null read as their zero value, except
// TopicId and VersionNumber, which are nil.
type Broadcast struct {
	Object      string   `json:"object"`
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	SegmentId   string   `json:"segment_id"`
	AudienceId  string   `json:"audience_id"` // Deprecated: Use SegmentId instead
	From        string   `json:"from"`
	Subject     string   `json:"subject"`
	ReplyTo     []string `json:"reply_to"`
	PreviewText string   `json:"preview_text"`
	TopicId     *string  `json:"topic_id"`
	Status      string   `json:"status"`
	CreatedAt   string   `json:"created_at"`
	ScheduledAt string   `json:"scheduled_at"`
	SentAt      string   `json:"sent_at"`
	UpdatedAt   string   `json:"updated_at"`
	Html        string   `json:"html"`
	Text        string   `json:"text"`

	// FailedReason says why a broadcast failed (status "failed").
	FailedReason string `json:"failed_reason"`

	// HoldReason and HoldUntil are set while a queued broadcast is held back.
	HoldReason string `json:"hold_reason"`
	HoldUntil  string `json:"hold_until"`

	// VersionNumber is the broadcast's current saved version (Get only).
	VersionNumber *int `json:"version_number"`

	// send.io extensions.
	SegmentIds        []string `json:"segment_ids"`
	ExcludeSegmentIds []string `json:"exclude_segment_ids"`
	Audience          string   `json:"audience"`
	UnsubscribeScope  string   `json:"unsubscribe_scope"`
	Tracking          bool     `json:"tracking"`
	TemplateEngine    string   `json:"template_engine"`

	// The template the broadcast was made from (nil for its own content):
	// the exact published version it copied (Get only).
	TemplateId            *string `json:"template_id"`
	TemplateVersionId     *string `json:"template_version_id"`
	TemplateName          *string `json:"template_name"`
	TemplateVersionNumber *int    `json:"template_version_number"`
	TemplateVersionName   *string `json:"template_version_name"`
	TemplateDeleted       bool    `json:"template_deleted"`
	// UnfillableVariables (Get, for a broadcast made from a template) are
	// the template variables with no value in a broadcast. While it is not
	// empty, Send refuses with 422 unfillable_variables.
	UnfillableVariables []string `json:"unfillable_variables"`
}

type BroadcastsSvc interface {
	CreateWithContext(ctx context.Context, params *CreateBroadcastRequest) (CreateBroadcastResponse, error)
	Create(params *CreateBroadcastRequest) (CreateBroadcastResponse, error)

	UpdateWithContext(ctx context.Context, params *UpdateBroadcastRequest) (UpdateBroadcastResponse, error)
	Update(params *UpdateBroadcastRequest) (UpdateBroadcastResponse, error)

	ListWithOptions(ctx context.Context, options *ListOptions) (ListBroadcastsResponse, error)
	ListWithContext(ctx context.Context) (ListBroadcastsResponse, error)
	List() (ListBroadcastsResponse, error)

	ClickedLinksWithOptions(ctx context.Context, broadcastId string, options *ListOptions) (ListBroadcastClickedLinksResponse, error)
	ClickedLinksWithContext(ctx context.Context, broadcastId string) (ListBroadcastClickedLinksResponse, error)
	ClickedLinks(broadcastId string) (ListBroadcastClickedLinksResponse, error)

	GetWithContext(ctx context.Context, broadcastId string) (Broadcast, error)
	Get(broadcastId string) (Broadcast, error)

	SendWithContext(ctx context.Context, params *SendBroadcastRequest) (SendBroadcastResponse, error)
	Send(params *SendBroadcastRequest) (SendBroadcastResponse, error)

	CancelWithContext(ctx context.Context, broadcastId string) (CancelBroadcastResponse, error)
	Cancel(broadcastId string) (CancelBroadcastResponse, error)

	DuplicateWithContext(ctx context.Context, broadcastId string) (DuplicateBroadcastResponse, error)
	Duplicate(broadcastId string) (DuplicateBroadcastResponse, error)

	RemoveWithContext(ctx context.Context, broadcastId string) (RemoveBroadcastResponse, error)
	Remove(broadcastId string) (RemoveBroadcastResponse, error)

	RecipientsWithContext(ctx context.Context, broadcastId string, options *ListBroadcastRecipientsOptions) (ListBroadcastRecipientsResponse, error)
	Recipients(broadcastId string, options *ListBroadcastRecipientsOptions) (ListBroadcastRecipientsResponse, error)

	ListWithFilters(ctx context.Context, options *ListBroadcastsOptions) (ListBroadcastsResponse, error)

	EstimateWithContext(ctx context.Context, params *EstimateBroadcastRequest) (BroadcastEstimate, error)
	Estimate(params *EstimateBroadcastRequest) (BroadcastEstimate, error)

	SendTestWithContext(ctx context.Context, broadcastId string, params *SendBroadcastTestRequest) (SendBroadcastTestResponse, error)
	SendTest(broadcastId string, params *SendBroadcastTestRequest) (SendBroadcastTestResponse, error)

	StatsWithContext(ctx context.Context, broadcastId string) (BroadcastStats, error)
	Stats(broadcastId string) (BroadcastStats, error)

	ListVersionsWithOptions(ctx context.Context, broadcastId string, options *ListOptions) (ListBroadcastVersionsResponse, error)
	ListVersions(broadcastId string) (ListBroadcastVersionsResponse, error)

	GetVersionWithContext(ctx context.Context, broadcastId, version string) (BroadcastVersion, error)
	GetVersion(broadcastId, version string) (BroadcastVersion, error)

	UpdateVersionWithContext(ctx context.Context, broadcastId, version string, params *UpdateBroadcastVersionRequest) (BroadcastVersion, error)
	UpdateVersion(broadcastId, version string, params *UpdateBroadcastVersionRequest) (BroadcastVersion, error)

	RestoreVersionWithContext(ctx context.Context, broadcastId, version string) (BroadcastVersion, error)
	RestoreVersion(broadcastId, version string) (BroadcastVersion, error)
}

type BroadcastsSvcImpl struct {
	client *Client
}

// CreateWithContext creates a new Broadcast based on the given params
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) CreateWithContext(ctx context.Context, params *CreateBroadcastRequest) (CreateBroadcastResponse, error) {
	path := "broadcasts"

	if params.SegmentId == "" && params.AudienceId == "" && len(params.SegmentIds) == 0 && params.Audience != "all" {
		return CreateBroadcastResponse{}, errors.New(`[ERROR]: Either SegmentId, AudienceId or SegmentIds must be provided, or Audience must be "all"`)
	}

	// A broadcast made from a template takes the template's From and
	// subject when it has them; the API says when neither has one.
	if params.From == "" && params.TemplateId == "" {
		return CreateBroadcastResponse{}, errors.New("[ERROR]: From cannot be empty")
	}

	if params.Subject == "" && params.TemplateId == "" {
		return CreateBroadcastResponse{}, errors.New("[ERROR]: Subject cannot be empty")
	}

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, params)
	if err != nil {
		return CreateBroadcastResponse{}, ErrFailedToCreateBroadcastCreateRequest
	}

	// Build response recipient obj
	broadcastResp := new(CreateBroadcastResponse)

	// Send Request
	_, err = s.client.Perform(req, broadcastResp)

	if err != nil {
		return CreateBroadcastResponse{}, err
	}

	return *broadcastResp, nil
}

// Create creates a new Broadcast based on the given params
func (s *BroadcastsSvcImpl) Create(params *CreateBroadcastRequest) (CreateBroadcastResponse, error) {
	return s.CreateWithContext(context.Background(), params)
}

// UpdateWithContext updates a given broadcast entry
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) UpdateWithContext(ctx context.Context, params *UpdateBroadcastRequest) (UpdateBroadcastResponse, error) {
	if params.BroadcastId == "" {
		return UpdateBroadcastResponse{}, errors.New("[ERROR]: BroadcastId cannot be empty")
	}

	path := "broadcasts/" + params.BroadcastId

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, params)
	if err != nil {
		return UpdateBroadcastResponse{}, ErrFailedToCreateBroadcastUpdateRequest
	}

	// Build response recipient obj
	broadcastResp := new(UpdateBroadcastResponse)

	// Send Request
	_, err = s.client.Perform(req, broadcastResp)

	if err != nil {
		return UpdateBroadcastResponse{}, err
	}

	return *broadcastResp, nil
}

func (s *BroadcastsSvcImpl) Update(params *UpdateBroadcastRequest) (UpdateBroadcastResponse, error) {
	return s.UpdateWithContext(context.Background(), params)
}

// GetWithContext Retrieve a single broadcast.
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) GetWithContext(ctx context.Context, broadcastId string) (Broadcast, error) {

	if broadcastId == "" {
		return Broadcast{}, errors.New("[ERROR]: broadcastId cannot be empty")
	}

	path := "broadcasts/" + broadcastId

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return Broadcast{}, errors.New("[ERROR]: Failed to create Broadcast.Get request")
	}

	broadcast := new(Broadcast)

	// Send Request
	_, err = s.client.Perform(req, broadcast)

	if err != nil {
		return Broadcast{}, err
	}

	return *broadcast, nil
}

// Get retrieves a single broadcast.
func (s *BroadcastsSvcImpl) Get(broadcastId string) (Broadcast, error) {
	return s.GetWithContext(context.Background(), broadcastId)
}

// SendWithContext Sends broadcasts to your audience.
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) SendWithContext(ctx context.Context, params *SendBroadcastRequest) (SendBroadcastResponse, error) {
	if params.BroadcastId == "" {
		return SendBroadcastResponse{}, errors.New("[ERROR]: BroadcastId cannot be empty")
	}

	path := "broadcasts/" + params.BroadcastId + "/send"

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, params)
	if err != nil {
		return SendBroadcastResponse{}, ErrFailedToCreateBroadcastSendRequest
	}

	// Build response recipient obj
	broadcastResp := new(SendBroadcastResponse)

	// Send Request
	_, err = s.client.Perform(req, broadcastResp)

	if err != nil {
		return SendBroadcastResponse{}, err
	}

	return *broadcastResp, nil
}

// Send sends broadcasts to your audience.
func (s *BroadcastsSvcImpl) Send(params *SendBroadcastRequest) (SendBroadcastResponse, error) {
	return s.SendWithContext(context.Background(), params)
}

func (s *BroadcastsSvcImpl) CancelWithContext(ctx context.Context, broadcastId string) (CancelBroadcastResponse, error) {
	if broadcastId == "" {
		return CancelBroadcastResponse{}, errors.New("[ERROR]: broadcastId cannot be empty")
	}

	path := "broadcasts/" + broadcastId + "/cancel"

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return CancelBroadcastResponse{}, ErrFailedToCreateBroadcastCancelRequest
	}

	resp := new(CancelBroadcastResponse)

	// Send Request
	_, err = s.client.Perform(req, resp)

	if err != nil {
		return CancelBroadcastResponse{}, err
	}

	return *resp, nil
}

func (s *BroadcastsSvcImpl) Cancel(broadcastId string) (CancelBroadcastResponse, error) {
	return s.CancelWithContext(context.Background(), broadcastId)
}

// DuplicateWithContext duplicates a broadcast by id, creating a new draft named after the source with " (copy)" appended
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) DuplicateWithContext(ctx context.Context, broadcastId string) (DuplicateBroadcastResponse, error) {
	if broadcastId == "" {
		return DuplicateBroadcastResponse{}, errors.New("[ERROR]: broadcastId cannot be empty")
	}

	path := "broadcasts/" + broadcastId + "/duplicate"

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return DuplicateBroadcastResponse{}, ErrFailedToCreateBroadcastDuplicateRequest
	}

	resp := new(DuplicateBroadcastResponse)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return DuplicateBroadcastResponse{}, err
	}

	return *resp, nil
}

// Duplicate duplicates a broadcast by id
func (s *BroadcastsSvcImpl) Duplicate(broadcastId string) (DuplicateBroadcastResponse, error) {
	return s.DuplicateWithContext(context.Background(), broadcastId)
}

// RemoveWithContext removes a given broadcast by id
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) RemoveWithContext(ctx context.Context, broadcastId string) (RemoveBroadcastResponse, error) {
	path := "broadcasts/" + broadcastId

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return RemoveBroadcastResponse{}, errors.New("[ERROR]: Failed to create Broadcast.Remove request")
	}

	resp := new(RemoveBroadcastResponse)

	// Send Request
	_, err = s.client.Perform(req, resp)

	if err != nil {
		return RemoveBroadcastResponse{}, err
	}

	return *resp, nil
}

// Remove removes a given broadcast entry by id
func (s *BroadcastsSvcImpl) Remove(broadcastId string) (RemoveBroadcastResponse, error) {
	return s.RemoveWithContext(context.Background(), broadcastId)
}

// ListWithOptions returns the list of all broadcasts with pagination options
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) ListWithOptions(ctx context.Context, options *ListOptions) (ListBroadcastsResponse, error) {
	path := "broadcasts" + buildPaginationQuery(options)

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ListBroadcastsResponse{}, errors.New("[ERROR]: Failed to create Broadcasts.List request")
	}

	broadcasts := new(ListBroadcastsResponse)

	// Send Request
	_, err = s.client.Perform(req, broadcasts)

	if err != nil {
		return ListBroadcastsResponse{}, err
	}

	return *broadcasts, nil
}

// ListWithFilters returns broadcasts filtered by status, segment or a name
// search, with pagination options.
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) ListWithFilters(ctx context.Context, options *ListBroadcastsOptions) (ListBroadcastsResponse, error) {
	path := "broadcasts"
	if options != nil {
		query := paginationValues(options.Limit, options.After, options.Before)
		if options.Status != "" {
			query.Set("status", options.Status)
		}
		if options.SegmentId != "" {
			query.Set("segment_id", options.SegmentId)
		}
		if options.Search != "" {
			query.Set("search", options.Search)
		}
		if len(query) > 0 {
			path += "?" + query.Encode()
		}
	}
	resp := new(ListBroadcastsResponse)
	if err := s.do(ctx, http.MethodGet, path, nil, resp, "List"); err != nil {
		return ListBroadcastsResponse{}, err
	}
	return *resp, nil
}

// ListWithContext returns the list of all broadcasts
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) ListWithContext(ctx context.Context) (ListBroadcastsResponse, error) {
	return s.ListWithOptions(ctx, nil)
}

// List returns the list of all broadcasts
func (s *BroadcastsSvcImpl) List() (ListBroadcastsResponse, error) {
	return s.ListWithContext(context.Background())
}

// RecipientsWithContext returns a broadcast's recipients for a given event type, such as who
// opened, clicked, or bounced.
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) RecipientsWithContext(ctx context.Context, broadcastId string, options *ListBroadcastRecipientsOptions) (ListBroadcastRecipientsResponse, error) {
	if broadcastId == "" {
		return ListBroadcastRecipientsResponse{}, errors.New("[ERROR]: broadcastId cannot be empty")
	}

	if options == nil || options.Type == "" {
		return ListBroadcastRecipientsResponse{}, errors.New("[ERROR]: Type cannot be empty")
	}

	query := make(url.Values)
	query.Set("type", options.Type)
	if options.Email != "" {
		query.Set("email", options.Email)
	}
	if options.BounceType != "" {
		query.Set("bounce_type", options.BounceType)
	}
	if options.Limit != nil {
		query.Set("limit", fmt.Sprintf("%d", *options.Limit))
	}
	if options.After != nil {
		query.Set("after", *options.After)
	}
	if options.Before != nil {
		query.Set("before", *options.Before)
	}

	path := "broadcasts/" + broadcastId + "/recipients?" + query.Encode()

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ListBroadcastRecipientsResponse{}, ErrFailedToCreateBroadcastRecipientsRequest
	}

	recipients := new(ListBroadcastRecipientsResponse)

	// Send Request
	_, err = s.client.Perform(req, recipients)

	if err != nil {
		return ListBroadcastRecipientsResponse{}, err
	}

	return *recipients, nil
}

// Recipients returns a broadcast's recipients for a given event type, such as who opened,
// clicked, or bounced.
func (s *BroadcastsSvcImpl) Recipients(broadcastId string, options *ListBroadcastRecipientsOptions) (ListBroadcastRecipientsResponse, error) {
	return s.RecipientsWithContext(context.Background(), broadcastId, options)
}

// ClickedLinksWithOptions returns the clicked links for a broadcast with pagination options
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) ClickedLinksWithOptions(ctx context.Context, broadcastId string, options *ListOptions) (ListBroadcastClickedLinksResponse, error) {
	if broadcastId == "" {
		return ListBroadcastClickedLinksResponse{}, errors.New("[ERROR]: broadcastId cannot be empty")
	}

	path := "broadcasts/" + broadcastId + "/clicked-links" + buildPaginationQuery(options)

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ListBroadcastClickedLinksResponse{}, errors.New("[ERROR]: Failed to create Broadcasts.ClickedLinks request")
	}

	clickedLinks := new(ListBroadcastClickedLinksResponse)

	// Send Request
	_, err = s.client.Perform(req, clickedLinks)

	if err != nil {
		return ListBroadcastClickedLinksResponse{}, err
	}

	return *clickedLinks, nil
}

// ClickedLinksWithContext returns the clicked links for a broadcast
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) ClickedLinksWithContext(ctx context.Context, broadcastId string) (ListBroadcastClickedLinksResponse, error) {
	return s.ClickedLinksWithOptions(ctx, broadcastId, nil)
}

// ClickedLinks returns the clicked links for a broadcast
func (s *BroadcastsSvcImpl) ClickedLinks(broadcastId string) (ListBroadcastClickedLinksResponse, error) {
	return s.ClickedLinksWithContext(context.Background(), broadcastId)
}

// ListBroadcastsOptions filters and pages Broadcasts.ListWithFilters.
type ListBroadcastsOptions struct {
	Limit  *int
	After  *string
	Before *string

	// Status keeps broadcasts in this status: draft, scheduled, queued,
	// sent, canceled or failed.
	Status string

	// SegmentId keeps broadcasts that send to this segment.
	SegmentId string

	// Search keeps broadcasts whose name contains this text.
	Search string
}

// EstimateBroadcastRequest is the audience to estimate, with the same fields
// as CreateBroadcastRequest. Nothing is saved.
type EstimateBroadcastRequest struct {
	SegmentId         string   `json:"segment_id,omitempty"`
	SegmentIds        []string `json:"segment_ids,omitempty"`
	ExcludeSegmentIds []string `json:"exclude_segment_ids,omitempty"`
	Audience          string   `json:"audience,omitempty"`
	TopicId           string   `json:"topic_id,omitempty"`
}

// BroadcastEstimate is how many contacts an audience reaches now, and how
// many it skips because email verification found them invalid.
type BroadcastEstimate struct {
	Object     string                   `json:"object"`
	Recipients int                      `json:"recipients"`
	Skipped    BroadcastSkippedContacts `json:"skipped"`
}

type BroadcastSkippedContacts struct {
	Invalid int `json:"invalid"`
}

// SendBroadcastTestRequest sends a broadcast's current content to 1 to 5
// addresses, with the subject prefixed "[Test]".
type SendBroadcastTestRequest struct {
	To []string `json:"to"`
}

type SendBroadcastTestResponse struct {
	Object string `json:"object"`
	Sent   int    `json:"sent"`
}

type BroadcastStatsProgress struct {
	Queued int `json:"queued"`
	Total  int `json:"total"`
}

// BroadcastStats are a broadcast's delivery and engagement counts and rates
// (rates are fractions, 0 to 1).
type BroadcastStats struct {
	Object          string                   `json:"object"`
	Status          string                   `json:"status"`
	Recipients      int                      `json:"recipients"`
	Sent            int                      `json:"sent"`
	Delivered       int                      `json:"delivered"`
	Bounced         int                      `json:"bounced"`
	Complained      int                      `json:"complained"`
	Suppressed      int                      `json:"suppressed"`
	Failed          int                      `json:"failed"`
	Opened          int                      `json:"opened"`
	UniqueOpened    int                      `json:"unique_opened"`
	Clicked         int                      `json:"clicked"`
	UniqueClicked   int                      `json:"unique_clicked"`
	Unsubscribed    int                      `json:"unsubscribed"`
	Skipped         BroadcastSkippedContacts `json:"skipped"`
	DeliveryRate    float64                  `json:"delivery_rate"`
	OpenRate        float64                  `json:"open_rate"`
	ClickRate       float64                  `json:"click_rate"`
	BounceRate      float64                  `json:"bounce_rate"`
	UnsubscribeRate float64                  `json:"unsubscribe_rate"`
	Progress        BroadcastStatsProgress   `json:"progress"`
}

// BroadcastVersion is a saved version of a broadcast. ListVersions returns
// the summary fields; GetVersion, UpdateVersion and RestoreVersion also fill
// the content and audience.
type BroadcastVersion struct {
	Object        string  `json:"object"`
	Id            string  `json:"id"`
	BroadcastId   string  `json:"broadcast_id"`
	VersionNumber int     `json:"version_number"`
	Name          *string `json:"name"`
	Comment       *string `json:"comment"`
	BroadcastName string  `json:"broadcast_name"`
	Subject       string  `json:"subject"`
	RestoredFrom  *int    `json:"restored_from"`
	CreatedAt     string  `json:"created_at"`

	// Current is true for the version the broadcast holds now.
	Current bool `json:"current"`

	From              string   `json:"from"`
	ReplyTo           []string `json:"reply_to"`
	PreviewText       string   `json:"preview_text"`
	Html              string   `json:"html"`
	Text              string   `json:"text"`
	TemplateEngine    string   `json:"template_engine"`
	Tracking          bool     `json:"tracking"`
	TopicId           *string  `json:"topic_id"`
	SegmentIds        []string `json:"segment_ids"`
	ExcludeSegmentIds []string `json:"exclude_segment_ids"`
	Audience          string   `json:"audience"`
	UnsubscribeScope  string   `json:"unsubscribe_scope"`
}

type ListBroadcastVersionsResponse struct {
	Object  string             `json:"object"`
	HasMore bool               `json:"has_more"`
	Data    []BroadcastVersion `json:"data"`
}

// UpdateBroadcastVersionRequest labels a version; its content is immutable.
// Nil fields are left as they are.
type UpdateBroadcastVersionRequest struct {
	Name    *string `json:"name,omitempty"`
	Comment *string `json:"comment,omitempty"`
}

// do sends one request and decodes the response into out.
func (s *BroadcastsSvcImpl) do(ctx context.Context, method, path string, body, out any, name string) error {
	req, err := s.client.NewRequest(ctx, method, path, body)
	if err != nil {
		return errors.New("[ERROR]: Failed to create Broadcasts." + name + " request")
	}
	_, err = s.client.Perform(req, out)
	return err
}

// EstimateWithContext counts the contacts an audience reaches now, before
// anything is saved.
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) EstimateWithContext(ctx context.Context, params *EstimateBroadcastRequest) (BroadcastEstimate, error) {
	if params == nil {
		params = &EstimateBroadcastRequest{}
	}
	resp := new(BroadcastEstimate)
	if err := s.do(ctx, http.MethodPost, "broadcasts/estimate", params, resp, "Estimate"); err != nil {
		return BroadcastEstimate{}, err
	}
	return *resp, nil
}

// Estimate counts the contacts an audience reaches now.
func (s *BroadcastsSvcImpl) Estimate(params *EstimateBroadcastRequest) (BroadcastEstimate, error) {
	return s.EstimateWithContext(context.Background(), params)
}

// SendTestWithContext sends a broadcast's current content to 1 to 5 test
// addresses.
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) SendTestWithContext(ctx context.Context, broadcastId string, params *SendBroadcastTestRequest) (SendBroadcastTestResponse, error) {
	if broadcastId == "" {
		return SendBroadcastTestResponse{}, errors.New("[ERROR]: broadcastId cannot be empty")
	}
	if params == nil || len(params.To) == 0 || len(params.To) > 5 {
		return SendBroadcastTestResponse{}, errors.New("[ERROR]: To must have 1 to 5 addresses")
	}
	resp := new(SendBroadcastTestResponse)
	if err := s.do(ctx, http.MethodPost, "broadcasts/"+url.PathEscape(broadcastId)+"/test", params, resp, "SendTest"); err != nil {
		return SendBroadcastTestResponse{}, err
	}
	return *resp, nil
}

// SendTest sends a broadcast's current content to 1 to 5 test addresses.
func (s *BroadcastsSvcImpl) SendTest(broadcastId string, params *SendBroadcastTestRequest) (SendBroadcastTestResponse, error) {
	return s.SendTestWithContext(context.Background(), broadcastId, params)
}

// StatsWithContext returns a broadcast's delivery and engagement stats.
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) StatsWithContext(ctx context.Context, broadcastId string) (BroadcastStats, error) {
	if broadcastId == "" {
		return BroadcastStats{}, errors.New("[ERROR]: broadcastId cannot be empty")
	}
	resp := new(BroadcastStats)
	if err := s.do(ctx, http.MethodGet, "broadcasts/"+url.PathEscape(broadcastId)+"/stats", nil, resp, "Stats"); err != nil {
		return BroadcastStats{}, err
	}
	return *resp, nil
}

// Stats returns a broadcast's delivery and engagement stats.
func (s *BroadcastsSvcImpl) Stats(broadcastId string) (BroadcastStats, error) {
	return s.StatsWithContext(context.Background(), broadcastId)
}

// ListVersionsWithOptions lists a broadcast's saved versions, newest first.
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) ListVersionsWithOptions(ctx context.Context, broadcastId string, options *ListOptions) (ListBroadcastVersionsResponse, error) {
	if broadcastId == "" {
		return ListBroadcastVersionsResponse{}, errors.New("[ERROR]: broadcastId cannot be empty")
	}
	resp := new(ListBroadcastVersionsResponse)
	if err := s.do(ctx, http.MethodGet, "broadcasts/"+url.PathEscape(broadcastId)+"/versions"+buildPaginationQuery(options), nil, resp, "ListVersions"); err != nil {
		return ListBroadcastVersionsResponse{}, err
	}
	return *resp, nil
}

// ListVersions lists a broadcast's saved versions, newest first.
func (s *BroadcastsSvcImpl) ListVersions(broadcastId string) (ListBroadcastVersionsResponse, error) {
	return s.ListVersionsWithOptions(context.Background(), broadcastId, nil)
}

func versionPath(broadcastId, version string) (string, error) {
	if broadcastId == "" {
		return "", errors.New("[ERROR]: broadcastId cannot be empty")
	}
	if version == "" {
		return "", errors.New("[ERROR]: version cannot be empty")
	}
	return "broadcasts/" + url.PathEscape(broadcastId) + "/versions/" + url.PathEscape(version), nil
}

// GetVersionWithContext returns one saved version of a broadcast with its
// content and audience. version is a version id or its version number ("3").
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) GetVersionWithContext(ctx context.Context, broadcastId, version string) (BroadcastVersion, error) {
	path, err := versionPath(broadcastId, version)
	if err != nil {
		return BroadcastVersion{}, err
	}
	resp := new(BroadcastVersion)
	if err := s.do(ctx, http.MethodGet, path, nil, resp, "GetVersion"); err != nil {
		return BroadcastVersion{}, err
	}
	return *resp, nil
}

// GetVersion returns one saved version of a broadcast.
func (s *BroadcastsSvcImpl) GetVersion(broadcastId, version string) (BroadcastVersion, error) {
	return s.GetVersionWithContext(context.Background(), broadcastId, version)
}

// UpdateVersionWithContext sets a version's name and comment.
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) UpdateVersionWithContext(ctx context.Context, broadcastId, version string, params *UpdateBroadcastVersionRequest) (BroadcastVersion, error) {
	path, err := versionPath(broadcastId, version)
	if err != nil {
		return BroadcastVersion{}, err
	}
	if params == nil {
		params = &UpdateBroadcastVersionRequest{}
	}
	resp := new(BroadcastVersion)
	if err := s.do(ctx, http.MethodPatch, path, params, resp, "UpdateVersion"); err != nil {
		return BroadcastVersion{}, err
	}
	return *resp, nil
}

// UpdateVersion sets a version's name and comment.
func (s *BroadcastsSvcImpl) UpdateVersion(broadcastId, version string, params *UpdateBroadcastVersionRequest) (BroadcastVersion, error) {
	return s.UpdateVersionWithContext(context.Background(), broadcastId, version, params)
}

// RestoreVersionWithContext copies a saved version back into a draft or
// scheduled broadcast and returns the new version it records.
// https://developer.send.io/docs/api
func (s *BroadcastsSvcImpl) RestoreVersionWithContext(ctx context.Context, broadcastId, version string) (BroadcastVersion, error) {
	path, err := versionPath(broadcastId, version)
	if err != nil {
		return BroadcastVersion{}, err
	}
	resp := new(BroadcastVersion)
	if err := s.do(ctx, http.MethodPost, path+"/restore", nil, resp, "RestoreVersion"); err != nil {
		return BroadcastVersion{}, err
	}
	return *resp, nil
}

// RestoreVersion copies a saved version back into a draft or scheduled broadcast.
func (s *BroadcastsSvcImpl) RestoreVersion(broadcastId, version string) (BroadcastVersion, error) {
	return s.RestoreVersionWithContext(context.Background(), broadcastId, version)
}
