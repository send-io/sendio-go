package sendio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type AutomationStatus = string

const (
	AutomationStatusEnabled  AutomationStatus = "enabled"
	AutomationStatusDisabled AutomationStatus = "disabled"
)

type AutomationRunStatus = string

const (
	AutomationRunStatusRunning   AutomationRunStatus = "running"
	AutomationRunStatusCompleted AutomationRunStatus = "completed"
	AutomationRunStatusFailed    AutomationRunStatus = "failed"
	AutomationRunStatusCancelled AutomationRunStatus = "cancelled"
)

type AutomationStepType = string

const (
	AutomationStepTypeTrigger       AutomationStepType = "trigger"
	AutomationStepTypeSendEmail     AutomationStepType = "send_email"
	AutomationStepTypeDelay         AutomationStepType = "delay"
	AutomationStepTypeWaitForEvent  AutomationStepType = "wait_for_event"
	AutomationStepTypeCondition     AutomationStepType = "condition"
	AutomationStepTypeContactUpdate AutomationStepType = "contact_update"
	AutomationStepTypeContactDelete AutomationStepType = "contact_delete"
	AutomationStepTypeAddToSegment  AutomationStepType = "add_to_segment"
	AutomationStepTypeWebhook       AutomationStepType = "webhook"
)

type AutomationConnectionType = string

const (
	AutomationConnectionTypeDefault         AutomationConnectionType = "default"
	AutomationConnectionTypeConditionMet    AutomationConnectionType = "condition_met"
	AutomationConnectionTypeConditionNotMet AutomationConnectionType = "condition_not_met"
	AutomationConnectionTypeTimeout         AutomationConnectionType = "timeout"
	AutomationConnectionTypeEventReceived   AutomationConnectionType = "event_received"
)

type AutomationStep struct {
	Key    string             `json:"key"`
	Type   AutomationStepType `json:"type"`
	Config map[string]any     `json:"config"`
}

// WebhookHeader is one header of a webhook ("HTTP request") step. Value may
// hold {{scope.path}} tokens. A Secret header is stored encrypted: reads
// return an empty Value with HasValue true, and sending "" back keeps it.
type WebhookHeader struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Secret   bool   `json:"secret,omitempty"`
	HasValue bool   `json:"has_value,omitempty"`
}

// WebhookFormField is one field of a form body (BodyType "form").
type WebhookFormField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// WebhookStepConfig configures a webhook ("HTTP request") step, which calls
// your own https URL. URL, header values and Body may hold {{event.x}},
// {{contact.email}} and other tokens.
type WebhookStepConfig struct {
	URL string
	// Method is POST (the default), PUT, PATCH, GET or DELETE. A body is
	// sent only for POST, PUT and PATCH.
	Method string
	// Headers: at most 10.
	Headers []WebhookHeader
	// BodyType is json (the default), form, text or none.
	BodyType string
	// Body: json takes any JSON value, form a []WebhookFormField, text a string.
	Body any
	// ContinueOnFailure follows the default connection after the last retry fails.
	ContinueOnFailure bool
	// TimeoutSeconds is 1 to 30; 0 means the default (10).
	TimeoutSeconds int
}

// Map returns the step config as sent to the API, leaving out unset fields.
func (c WebhookStepConfig) Map() map[string]any {
	m := map[string]any{"url": c.URL}
	if c.Method != "" {
		m["method"] = c.Method
	}
	if len(c.Headers) > 0 {
		m["headers"] = c.Headers
	}
	if c.BodyType != "" {
		m["body_type"] = c.BodyType
	}
	if c.Body != nil {
		m["body"] = c.Body
	}
	if c.ContinueOnFailure {
		m["continue_on_failure"] = true
	}
	if c.TimeoutSeconds > 0 {
		m["timeout_seconds"] = c.TimeoutSeconds
	}
	return m
}

// NewWebhookStep builds a webhook ("HTTP request") step.
func NewWebhookStep(key string, config WebhookStepConfig) AutomationStep {
	return AutomationStep{Key: key, Type: AutomationStepTypeWebhook, Config: config.Map()}
}

type AutomationStepResponse struct {
	Key    string             `json:"key"`
	Type   AutomationStepType `json:"type"`
	Config map[string]any     `json:"config"`
}

type AutomationConnection struct {
	From string                   `json:"from"`
	To   string                   `json:"to"`
	Type AutomationConnectionType `json:"type,omitempty"`
}

type CreateAutomationRequest struct {
	Name        string                 `json:"name"`
	Status      AutomationStatus       `json:"status,omitempty"`
	Steps       []AutomationStep       `json:"steps"`
	Connections []AutomationConnection `json:"connections"`
}

type CreateAutomationResponse struct {
	Object string `json:"object"`
	Id     string `json:"id"`
}

type UpdateAutomationRequest struct {
	Name        string                 `json:"name,omitempty"`
	Status      AutomationStatus       `json:"status,omitempty"`
	Steps       []AutomationStep       `json:"steps,omitempty"`
	Connections []AutomationConnection `json:"connections,omitempty"`
}

type UpdateAutomationResponse struct {
	Object string `json:"object"`
	Id     string `json:"id"`
}

type DeleteAutomationResponse struct {
	Object  string `json:"object"`
	Id      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

type DuplicateAutomationResponse struct {
	Object string `json:"object"`
	Id     string `json:"id"`
}

type StopAutomationResponse struct {
	Object string `json:"object"`
	Id     string `json:"id"`
	Status string `json:"status"`
}

type AutomationListItem struct {
	Id        string           `json:"id"`
	Name      string           `json:"name"`
	Status    AutomationStatus `json:"status"`
	CreatedAt string           `json:"created_at"`
	UpdatedAt string           `json:"updated_at"`
}

type ListAutomationsResponse struct {
	Object  string               `json:"object"`
	HasMore bool                 `json:"has_more"`
	Data    []AutomationListItem `json:"data"`
}

type Automation struct {
	Object      string                   `json:"object"`
	Id          string                   `json:"id"`
	Name        string                   `json:"name"`
	Status      AutomationStatus         `json:"status"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
	Steps       []AutomationStepResponse `json:"steps"`
	Connections []AutomationConnection   `json:"connections"`
}

type AutomationRunListItem struct {
	Id          string              `json:"id"`
	Status      AutomationRunStatus `json:"status"`
	StartedAt   *string             `json:"started_at"`
	CompletedAt *string             `json:"completed_at"`
	CreatedAt   string              `json:"created_at"`
}

type ListAutomationRunsResponse struct {
	Object  string                  `json:"object"`
	HasMore bool                    `json:"has_more"`
	Data    []AutomationRunListItem `json:"data"`
}

type AutomationRunStep struct {
	Key         string             `json:"key"`
	Type        AutomationStepType `json:"type"`
	Status      string             `json:"status"`
	StartedAt   *string            `json:"started_at"`
	CompletedAt *string            `json:"completed_at"`
	Output      any                `json:"output"`
	Error       any                `json:"error"`
	CreatedAt   string             `json:"created_at"`
}

type AutomationRun struct {
	Object      string              `json:"object"`
	Id          string              `json:"id"`
	Status      AutomationRunStatus `json:"status"`
	StartedAt   *string             `json:"started_at"`
	CompletedAt *string             `json:"completed_at"`
	CreatedAt   string              `json:"created_at"`
	Steps       []AutomationRunStep `json:"steps"`
}

type AutomationTemplateCategory = string

const (
	AutomationTemplateCategoryOnboarding   AutomationTemplateCategory = "onboarding"
	AutomationTemplateCategoryBilling      AutomationTemplateCategory = "billing"
	AutomationTemplateCategoryEcommerce    AutomationTemplateCategory = "ecommerce"
	AutomationTemplateCategoryEngagement   AutomationTemplateCategory = "engagement"
	AutomationTemplateCategoryLeadCapture  AutomationTemplateCategory = "lead_capture"
	AutomationTemplateCategoryAppointments AutomationTemplateCategory = "appointments"
	AutomationTemplateCategoryEvents       AutomationTemplateCategory = "events"
	AutomationTemplateCategoryEducation    AutomationTemplateCategory = "education"
	AutomationTemplateCategoryAccount      AutomationTemplateCategory = "account"
	AutomationTemplateCategoryNonprofit    AutomationTemplateCategory = "nonprofit"
	AutomationTemplateCategoryInternal     AutomationTemplateCategory = "internal"
)

type AutomationTemplateKind = string

const (
	AutomationTemplateKindMarketing     AutomationTemplateKind = "marketing"
	AutomationTemplateKindTransactional AutomationTemplateKind = "transactional"
)

type AutomationTemplateParamType = string

const (
	AutomationTemplateParamTypeDuration AutomationTemplateParamType = "duration"
	AutomationTemplateParamTypeURL      AutomationTemplateParamType = "url"
)

type AutomationTemplateEventMode = string

const (
	// AutomationTemplateEventModeCreate uses the event, creating it if your team does not have it (the default).
	AutomationTemplateEventModeCreate AutomationTemplateEventMode = "create"
	// AutomationTemplateEventModeExisting maps the template's event to one your app already sends.
	AutomationTemplateEventModeExisting AutomationTemplateEventMode = "existing"
)

// AutomationTemplateEvent is an event a template uses. Exists says whether
// your team already has it. Schema and Adds are only returned by Get: Adds
// holds the fields setup would add to your team's event.
type AutomationTemplateEvent struct {
	Ref     string            `json:"ref"`
	Name    string            `json:"name"`
	Purpose string            `json:"purpose"`
	Exists  bool              `json:"exists"`
	Schema  map[string]string `json:"schema,omitempty"`
	Adds    map[string]string `json:"adds,omitempty"`
}

// AutomationTemplateEmail is one of a template's emails, filled with your
// team's brand as setup will create it.
type AutomationTemplateEmail struct {
	Ref       string           `json:"ref"`
	StepLabel string           `json:"step_label"`
	StepKey   string           `json:"step_key"`
	Subject   string           `json:"subject"`
	Preview   string           `json:"preview"`
	Html      string           `json:"html"`
	Text      string           `json:"text"`
	Variables []map[string]any `json:"variables"`
}

// AutomationTemplateParam is a value setup asks for. A duration param has a
// Default; a url param has none and must be given.
type AutomationTemplateParam struct {
	Ref         string                      `json:"ref"`
	Type        AutomationTemplateParamType `json:"type"`
	Label       string                      `json:"label"`
	Default     string                      `json:"default"`
	Placeholder string                      `json:"placeholder,omitempty"`
	Hint        string                      `json:"hint,omitempty"`
}

// AutomationTemplateListItem is a template as returned by Automations.Templates.List.
// Its Steps carry only Key and Type.
type AutomationTemplateListItem struct {
	Object      string                     `json:"object"`
	Slug        string                     `json:"slug"`
	Version     int                        `json:"version"`
	Name        string                     `json:"name"`
	Category    AutomationTemplateCategory `json:"category"`
	Kind        AutomationTemplateKind     `json:"kind"`
	Outcome     string                     `json:"outcome"`
	Events      []AutomationTemplateEvent  `json:"events"`
	EmailCount  int                        `json:"email_count"`
	Steps       []AutomationStepResponse   `json:"steps"`
	Connections []AutomationConnection     `json:"connections"`
	InUseCount  int                        `json:"in_use_count"`
}

type ListAutomationTemplatesResponse struct {
	Object     string                       `json:"object"`
	Categories []AutomationTemplateCategory `json:"categories"`
	Data       []AutomationTemplateListItem `json:"data"`
}

// AutomationTemplate is one template in full, as returned by Automations.Templates.Get.
type AutomationTemplate struct {
	Object      string                     `json:"object"`
	Slug        string                     `json:"slug"`
	Version     int                        `json:"version"`
	Name        string                     `json:"name"`
	Category    AutomationTemplateCategory `json:"category"`
	Kind        AutomationTemplateKind     `json:"kind"`
	Outcome     string                     `json:"outcome"`
	HowItWorks  string                     `json:"how_it_works"`
	Events      []AutomationTemplateEvent  `json:"events"`
	EmailCount  int                        `json:"email_count"`
	Emails      []AutomationTemplateEmail  `json:"emails"`
	Params      []AutomationTemplateParam  `json:"params"`
	Steps       []AutomationStepResponse   `json:"steps"`
	Connections []AutomationConnection     `json:"connections"`
	Settings    map[string]any             `json:"settings,omitempty"`
	InUseCount  int                        `json:"in_use_count"`
}

// ListAutomationTemplatesOptions contains filter parameters for listing automation templates
type ListAutomationTemplatesOptions struct {
	Category *AutomationTemplateCategory
}

// AutomationTemplateEventChoice picks the event used for one of a template's event refs.
type AutomationTemplateEventChoice struct {
	Mode AutomationTemplateEventMode `json:"mode,omitempty"`
	Name string                      `json:"name"`
}

// AutomationTemplateSender is the sender of a template's emails. From must be
// on one of your verified sending domains.
type AutomationTemplateSender struct {
	From    string `json:"from"`
	ReplyTo string `json:"reply_to,omitempty"`
}

type CreateAutomationFromTemplateRequest struct {
	Slug string `json:"slug"`
	// Version is the version you reviewed; a newer one returns 409 template_version_changed.
	Version int `json:"version,omitempty"`
	// Name defaults to the template's name.
	Name string `json:"name,omitempty"`
	// Events maps an event ref to the event to use.
	Events map[string]AutomationTemplateEventChoice `json:"events,omitempty"`
	Sender AutomationTemplateSender                 `json:"sender"`
	// TopicId is required for a marketing template.
	TopicId string `json:"topic_id,omitempty"`
	// Params maps a param ref to its value, for example "2 days" or a https link.
	Params map[string]string `json:"params,omitempty"`
}

// CreateAutomationFromTemplateOptions holds request options for
// Automations.CreateFromTemplateWithOptions. A retry with the same
// IdempotencyKey returns the first answer instead of a second automation.
type CreateAutomationFromTemplateOptions struct {
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

func (o CreateAutomationFromTemplateOptions) GetIdempotencyKey() string {
	return o.IdempotencyKey
}

type AutomationSourceTemplate struct {
	Slug    string `json:"slug"`
	Version int    `json:"version"`
	Name    string `json:"name"`
}

// AutomationEmail is a workflow email owned by an automation.
type AutomationEmail struct {
	Object  string `json:"object"`
	Id      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	StepKey string `json:"step_key"`
}

type CreateAutomationFromTemplateResponse struct {
	Object         string                   `json:"object"`
	Id             string                   `json:"id"`
	Status         AutomationStatus         `json:"status"`
	SourceTemplate AutomationSourceTemplate `json:"source_template"`
	Emails         []AutomationEmail        `json:"emails"`
	EventsCreated  []string                 `json:"events_created"`
	EventsReused   []string                 `json:"events_reused"`
}

// ListAutomationsOptions contains pagination and filter parameters for listing automations
type ListAutomationsOptions struct {
	Status *AutomationStatus
	Limit  *int
	After  *string
	Before *string
}

// ListAutomationRunsOptions contains pagination and filter parameters for listing automation runs
type ListAutomationRunsOptions struct {
	Status []AutomationRunStatus
	Limit  *int
	After  *string
	Before *string
}

type AutomationsSvc interface {
	CreateWithContext(ctx context.Context, params *CreateAutomationRequest) (CreateAutomationResponse, error)
	Create(params *CreateAutomationRequest) (CreateAutomationResponse, error)
	GetWithContext(ctx context.Context, automationId string) (Automation, error)
	Get(automationId string) (Automation, error)
	ListWithContext(ctx context.Context) (ListAutomationsResponse, error)
	List() (ListAutomationsResponse, error)
	ListWithOptions(ctx context.Context, options *ListAutomationsOptions) (ListAutomationsResponse, error)
	UpdateWithContext(ctx context.Context, automationId string, params *UpdateAutomationRequest) (UpdateAutomationResponse, error)
	Update(automationId string, params *UpdateAutomationRequest) (UpdateAutomationResponse, error)
	RemoveWithContext(ctx context.Context, automationId string) (DeleteAutomationResponse, error)
	Remove(automationId string) (DeleteAutomationResponse, error)
	DuplicateWithContext(ctx context.Context, automationId string) (DuplicateAutomationResponse, error)
	Duplicate(automationId string) (DuplicateAutomationResponse, error)
	StopWithContext(ctx context.Context, automationId string) (StopAutomationResponse, error)
	Stop(automationId string) (StopAutomationResponse, error)
	ListRunsWithContext(ctx context.Context, automationId string, options *ListAutomationRunsOptions) (ListAutomationRunsResponse, error)
	ListRuns(automationId string) (ListAutomationRunsResponse, error)
	GetRunWithContext(ctx context.Context, automationId string, runId string) (AutomationRun, error)
	GetRun(automationId string, runId string) (AutomationRun, error)
	CreateFromTemplateWithOptions(ctx context.Context, params *CreateAutomationFromTemplateRequest, options *CreateAutomationFromTemplateOptions) (CreateAutomationFromTemplateResponse, error)
	CreateFromTemplateWithContext(ctx context.Context, params *CreateAutomationFromTemplateRequest) (CreateAutomationFromTemplateResponse, error)
	CreateFromTemplate(params *CreateAutomationFromTemplateRequest) (CreateAutomationFromTemplateResponse, error)
}

type AutomationsSvcImpl struct {
	client    *Client
	Templates AutomationTemplatesSvc
}

type AutomationTemplatesSvc interface {
	ListWithOptions(ctx context.Context, options *ListAutomationTemplatesOptions) (ListAutomationTemplatesResponse, error)
	ListWithContext(ctx context.Context) (ListAutomationTemplatesResponse, error)
	List() (ListAutomationTemplatesResponse, error)
	GetWithContext(ctx context.Context, slug string) (AutomationTemplate, error)
	Get(slug string) (AutomationTemplate, error)
}

type AutomationTemplatesSvcImpl struct {
	client *Client
}

func buildAutomationsQuery(options *ListAutomationsOptions) string {
	if options == nil {
		return ""
	}
	query := make(url.Values)
	if options.Status != nil {
		query.Set("status", *options.Status)
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
	if len(query) > 0 {
		return "?" + query.Encode()
	}
	return ""
}

func buildAutomationRunsQuery(options *ListAutomationRunsOptions) string {
	if options == nil {
		return ""
	}
	query := make(url.Values)
	if len(options.Status) > 0 {
		statuses := make([]string, len(options.Status))
		for i, s := range options.Status {
			statuses[i] = s
		}
		query.Set("status", strings.Join(statuses, ","))
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
	if len(query) > 0 {
		return "?" + query.Encode()
	}
	return ""
}

// CreateWithContext creates a new Automation
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) CreateWithContext(ctx context.Context, params *CreateAutomationRequest) (CreateAutomationResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, "automations", params)
	if err != nil {
		return CreateAutomationResponse{}, ErrFailedToCreateAutomationCreateRequest
	}

	resp := new(CreateAutomationResponse)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return CreateAutomationResponse{}, err
	}

	return *resp, nil
}

// Create creates a new Automation
func (s *AutomationsSvcImpl) Create(params *CreateAutomationRequest) (CreateAutomationResponse, error) {
	return s.CreateWithContext(context.Background(), params)
}

// GetWithContext retrieves a single Automation by ID
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) GetWithContext(ctx context.Context, automationId string) (Automation, error) {
	path := "automations/" + automationId

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return Automation{}, ErrFailedToCreateAutomationGetRequest
	}

	resp := new(Automation)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return Automation{}, err
	}

	return *resp, nil
}

// Get retrieves a single Automation by ID
func (s *AutomationsSvcImpl) Get(automationId string) (Automation, error) {
	return s.GetWithContext(context.Background(), automationId)
}

// ListWithOptions retrieves a list of Automations with pagination and filter options
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) ListWithOptions(ctx context.Context, options *ListAutomationsOptions) (ListAutomationsResponse, error) {
	path := "automations" + buildAutomationsQuery(options)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ListAutomationsResponse{}, ErrFailedToCreateAutomationListRequest
	}

	resp := new(ListAutomationsResponse)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return ListAutomationsResponse{}, err
	}

	return *resp, nil
}

// ListWithContext retrieves a list of Automations
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) ListWithContext(ctx context.Context) (ListAutomationsResponse, error) {
	return s.ListWithOptions(ctx, nil)
}

// List retrieves a list of Automations
func (s *AutomationsSvcImpl) List() (ListAutomationsResponse, error) {
	return s.ListWithContext(context.Background())
}

// UpdateWithContext updates an existing Automation
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) UpdateWithContext(ctx context.Context, automationId string, params *UpdateAutomationRequest) (UpdateAutomationResponse, error) {
	path := "automations/" + automationId

	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, params)
	if err != nil {
		return UpdateAutomationResponse{}, ErrFailedToCreateAutomationUpdateRequest
	}

	resp := new(UpdateAutomationResponse)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return UpdateAutomationResponse{}, err
	}

	return *resp, nil
}

// Update updates an existing Automation
func (s *AutomationsSvcImpl) Update(automationId string, params *UpdateAutomationRequest) (UpdateAutomationResponse, error) {
	return s.UpdateWithContext(context.Background(), automationId, params)
}

// RemoveWithContext deletes an Automation by ID
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) RemoveWithContext(ctx context.Context, automationId string) (DeleteAutomationResponse, error) {
	path := "automations/" + automationId

	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return DeleteAutomationResponse{}, ErrFailedToCreateAutomationRemoveRequest
	}

	resp := new(DeleteAutomationResponse)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return DeleteAutomationResponse{}, err
	}

	return *resp, nil
}

// Remove deletes an Automation by ID
func (s *AutomationsSvcImpl) Remove(automationId string) (DeleteAutomationResponse, error) {
	return s.RemoveWithContext(context.Background(), automationId)
}

// DuplicateWithContext duplicates an Automation by ID
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) DuplicateWithContext(ctx context.Context, automationId string) (DuplicateAutomationResponse, error) {
	path := "automations/" + automationId + "/duplicate"

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return DuplicateAutomationResponse{}, ErrFailedToCreateAutomationDuplicateRequest
	}

	resp := new(DuplicateAutomationResponse)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return DuplicateAutomationResponse{}, err
	}

	return *resp, nil
}

// Duplicate duplicates an Automation by ID
func (s *AutomationsSvcImpl) Duplicate(automationId string) (DuplicateAutomationResponse, error) {
	return s.DuplicateWithContext(context.Background(), automationId)
}

// StopWithContext stops a running Automation by ID
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) StopWithContext(ctx context.Context, automationId string) (StopAutomationResponse, error) {
	path := "automations/" + automationId + "/stop"

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return StopAutomationResponse{}, ErrFailedToCreateAutomationStopRequest
	}

	resp := new(StopAutomationResponse)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return StopAutomationResponse{}, err
	}

	return *resp, nil
}

// Stop stops a running Automation by ID
func (s *AutomationsSvcImpl) Stop(automationId string) (StopAutomationResponse, error) {
	return s.StopWithContext(context.Background(), automationId)
}

// ListRunsWithContext retrieves a list of runs for an Automation
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) ListRunsWithContext(ctx context.Context, automationId string, options *ListAutomationRunsOptions) (ListAutomationRunsResponse, error) {
	path := "automations/" + automationId + "/runs" + buildAutomationRunsQuery(options)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ListAutomationRunsResponse{}, ErrFailedToCreateAutomationListRunsRequest
	}

	resp := new(ListAutomationRunsResponse)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return ListAutomationRunsResponse{}, err
	}

	return *resp, nil
}

// ListRuns retrieves a list of runs for an Automation
func (s *AutomationsSvcImpl) ListRuns(automationId string) (ListAutomationRunsResponse, error) {
	return s.ListRunsWithContext(context.Background(), automationId, nil)
}

// GetRunWithContext retrieves a single run for an Automation
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) GetRunWithContext(ctx context.Context, automationId string, runId string) (AutomationRun, error) {
	path := "automations/" + automationId + "/runs/" + runId

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return AutomationRun{}, ErrFailedToCreateAutomationGetRunRequest
	}

	resp := new(AutomationRun)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return AutomationRun{}, err
	}

	return *resp, nil
}

// GetRun retrieves a single run for an Automation
func (s *AutomationsSvcImpl) GetRun(automationId string, runId string) (AutomationRun, error) {
	return s.GetRunWithContext(context.Background(), automationId, runId)
}

// CreateFromTemplateWithOptions sets up an automation template in one go: the
// events your team does not have, the workflow's emails and the automation,
// which is created disabled. Pass an IdempotencyKey to make retries safe.
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) CreateFromTemplateWithOptions(ctx context.Context, params *CreateAutomationFromTemplateRequest, options *CreateAutomationFromTemplateOptions) (CreateAutomationFromTemplateResponse, error) {
	var opts Options
	if options != nil {
		opts = options
	}
	req, err := s.client.NewRequestWithOptions(ctx, http.MethodPost, "automations/from-template", params, opts)
	if err != nil {
		return CreateAutomationFromTemplateResponse{}, ErrFailedToCreateAutomationCreateFromTemplateRequest
	}

	resp := new(CreateAutomationFromTemplateResponse)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return CreateAutomationFromTemplateResponse{}, err
	}

	return *resp, nil
}

// CreateFromTemplateWithContext creates an Automation from an automation template
// https://developer.send.io/docs/api
func (s *AutomationsSvcImpl) CreateFromTemplateWithContext(ctx context.Context, params *CreateAutomationFromTemplateRequest) (CreateAutomationFromTemplateResponse, error) {
	return s.CreateFromTemplateWithOptions(ctx, params, nil)
}

// CreateFromTemplate creates an Automation from an automation template
func (s *AutomationsSvcImpl) CreateFromTemplate(params *CreateAutomationFromTemplateRequest) (CreateAutomationFromTemplateResponse, error) {
	return s.CreateFromTemplateWithContext(context.Background(), params)
}

func buildAutomationTemplatesQuery(options *ListAutomationTemplatesOptions) string {
	if options == nil || options.Category == nil {
		return ""
	}
	query := make(url.Values)
	query.Set("category", *options.Category)
	return "?" + query.Encode()
}

// ListWithOptions retrieves the automation templates, optionally of one category.
// The list is not paginated.
// https://developer.send.io/docs/api
func (s *AutomationTemplatesSvcImpl) ListWithOptions(ctx context.Context, options *ListAutomationTemplatesOptions) (ListAutomationTemplatesResponse, error) {
	path := "automation-templates" + buildAutomationTemplatesQuery(options)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ListAutomationTemplatesResponse{}, ErrFailedToCreateAutomationTemplatesListRequest
	}

	resp := new(ListAutomationTemplatesResponse)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return ListAutomationTemplatesResponse{}, err
	}

	return *resp, nil
}

// ListWithContext retrieves every automation template
// https://developer.send.io/docs/api
func (s *AutomationTemplatesSvcImpl) ListWithContext(ctx context.Context) (ListAutomationTemplatesResponse, error) {
	return s.ListWithOptions(ctx, nil)
}

// List retrieves every automation template
func (s *AutomationTemplatesSvcImpl) List() (ListAutomationTemplatesResponse, error) {
	return s.ListWithContext(context.Background())
}

// GetWithContext retrieves one automation template in full by slug: its
// steps, emails (with your team's brand), params and events.
// https://developer.send.io/docs/api
func (s *AutomationTemplatesSvcImpl) GetWithContext(ctx context.Context, slug string) (AutomationTemplate, error) {
	path := "automation-templates/" + slug

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return AutomationTemplate{}, ErrFailedToCreateAutomationTemplatesGetRequest
	}

	resp := new(AutomationTemplate)
	_, err = s.client.Perform(req, resp)
	if err != nil {
		return AutomationTemplate{}, err
	}

	return *resp, nil
}

// Get retrieves one automation template in full by slug
func (s *AutomationTemplatesSvcImpl) Get(slug string) (AutomationTemplate, error) {
	return s.GetWithContext(context.Background(), slug)
}
