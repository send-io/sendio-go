package sendio

import (
	"context"
	"net/http"
)

// Mailbox kinds
const (
	MailboxKindNormal = "normal"
	MailboxKindAgent  = "agent"
)

// Mailbox IMAP/SMTP states
const (
	MailboxImapSmtpReady       = "ready"
	MailboxImapSmtpPending     = "pending"
	MailboxImapSmtpRetrying    = "retrying"
	MailboxImapSmtpUnavailable = "unavailable"
)

// Mailbox health values
const (
	MailboxHealthNoData  = "no_data"
	MailboxHealthHealthy = "healthy"
	MailboxHealthAtRisk  = "at_risk"
	MailboxHealthPoor    = "poor"
)

// Mailbox is an email account with an IMAP/SMTP login.
// Sent, Delivered, Bounced and Complained cover the last 30 days; SentTotal is all time.
type Mailbox struct {
	Object          string  `json:"object"`
	Id              string  `json:"id"`
	Address         string  `json:"address"`
	Kind            string  `json:"kind"`
	AgentInboxId    *string `json:"agent_inbox_id"`
	ImapSmtp        string  `json:"imap_smtp"`
	DomainId        *string `json:"domain_id"`
	Domain          *string `json:"domain"`
	WarmupEnabled   bool    `json:"warmup_enabled"`
	WarmupChangedAt *string `json:"warmup_changed_at"`
	CreatedAt       string  `json:"created_at"`
	Sent            int     `json:"sent"`
	SentTotal       int     `json:"sent_total"`
	Delivered       int     `json:"delivered"`
	Bounced         int     `json:"bounced"`
	Complained      int     `json:"complained"`
	LastSentAt      *string `json:"last_sent_at"`
	Health          string  `json:"health"`
	// CanSend says whether POST /emails accepts this address as From now
	// (the API's own rule); SendReason says why not, in a short line.
	CanSend    bool    `json:"can_send"`
	SendReason *string `json:"send_reason"`
	// Credentials is only set by Create, when IMAP/SMTP is ready. It is shown once.
	Credentials *MailboxCredentials `json:"credentials,omitempty"`
}

// MailboxCredentials is the IMAP/SMTP login of a mailbox.
// The password is also an API key for that mailbox.
type MailboxCredentials struct {
	Username string        `json:"username"`
	Password string        `json:"password"`
	Imap     MailboxServer `json:"imap"`
	Smtp     MailboxServer `json:"smtp"`
}

// MailboxServer is an IMAP or SMTP server endpoint.
type MailboxServer struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Security string `json:"security"`
}

// CreateMailboxRequest is the request payload for creating a mailbox.
// Set Address (full address), or LocalPart and DomainId. The domain must be verified.
// For a free mailbox with no domain, set Username, and optionally Org and DisplayName.
type CreateMailboxRequest struct {
	Address     string `json:"address,omitempty"`
	LocalPart   string `json:"local_part,omitempty"`
	DomainId    string `json:"domain_id,omitempty"`
	Username    string `json:"username,omitempty"`
	Org         string `json:"org,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

// UpdateMailboxRequest is the request payload for updating a mailbox.
// WarmupEnabled is always sent, false included.
type UpdateMailboxRequest struct {
	WarmupEnabled bool `json:"warmup_enabled"`
}

// ListMailboxesResponse is the response from listing mailboxes
type ListMailboxesResponse struct {
	Object  string     `json:"object"`
	HasMore bool       `json:"has_more"`
	Data    []*Mailbox `json:"data"`
}

// ResetMailboxPasswordResponse is the response from resetting a mailbox password
type ResetMailboxPasswordResponse struct {
	Object      string             `json:"object"`
	Id          string             `json:"id"`
	Address     string             `json:"address"`
	Credentials MailboxCredentials `json:"credentials"`
	Note        string             `json:"note"`
}

// RemoveMailboxResponse is the response from removing a mailbox
type RemoveMailboxResponse struct {
	Object  string `json:"object"`
	Id      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// MailboxesSvc handles operations for mailboxes
type MailboxesSvc interface {
	ListWithContext(ctx context.Context) (*ListMailboxesResponse, error)
	List() (*ListMailboxesResponse, error)
	CreateWithContext(ctx context.Context, params *CreateMailboxRequest) (*Mailbox, error)
	Create(params *CreateMailboxRequest) (*Mailbox, error)
	UpdateWithContext(ctx context.Context, mailboxId string, params *UpdateMailboxRequest) (*Mailbox, error)
	Update(mailboxId string, params *UpdateMailboxRequest) (*Mailbox, error)
	ResetPasswordWithContext(ctx context.Context, mailboxId string) (*ResetMailboxPasswordResponse, error)
	ResetPassword(mailboxId string) (*ResetMailboxPasswordResponse, error)
	RemoveWithContext(ctx context.Context, mailboxId string) (*RemoveMailboxResponse, error)
	Remove(mailboxId string) (*RemoveMailboxResponse, error)
}

// MailboxesSvcImpl is the implementation of the MailboxesSvc interface
type MailboxesSvcImpl struct {
	client *Client
}

// ListWithContext lists all mailboxes in the project (not paginated)
// https://developer.send.io/docs/api
func (s *MailboxesSvcImpl) ListWithContext(ctx context.Context) (*ListMailboxesResponse, error) {
	path := "mailboxes"

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, ErrFailedToCreateMailboxesListRequest
	}

	// Build response recipient obj
	mailboxesResp := new(ListMailboxesResponse)

	// Send Request
	_, err = s.client.Perform(req, mailboxesResp)

	if err != nil {
		return nil, err
	}

	return mailboxesResp, nil
}

// List lists all mailboxes in the project (not paginated)
// https://developer.send.io/docs/api
func (s *MailboxesSvcImpl) List() (*ListMailboxesResponse, error) {
	return s.ListWithContext(context.Background())
}

// CreateWithContext creates a mailbox. The domain must be verified; the API
// answers 409 when the address already exists. When IMAP/SMTP is ready, the
// returned Mailbox carries Credentials, which are shown only once.
// https://developer.send.io/docs/api
func (s *MailboxesSvcImpl) CreateWithContext(ctx context.Context, params *CreateMailboxRequest) (*Mailbox, error) {
	path := "mailboxes"

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, params)
	if err != nil {
		return nil, ErrFailedToCreateMailboxesCreateRequest
	}

	// Build response recipient obj
	mailboxResp := new(Mailbox)

	// Send Request
	_, err = s.client.Perform(req, mailboxResp)

	if err != nil {
		return nil, err
	}

	return mailboxResp, nil
}

// Create creates a mailbox. See CreateWithContext.
// https://developer.send.io/docs/api
func (s *MailboxesSvcImpl) Create(params *CreateMailboxRequest) (*Mailbox, error) {
	return s.CreateWithContext(context.Background(), params)
}

// UpdateWithContext turns warmup on or off for a mailbox
// https://developer.send.io/docs/api
func (s *MailboxesSvcImpl) UpdateWithContext(ctx context.Context, mailboxId string, params *UpdateMailboxRequest) (*Mailbox, error) {
	path := "mailboxes/" + mailboxId

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, params)
	if err != nil {
		return nil, ErrFailedToCreateMailboxesUpdateRequest
	}

	// Build response recipient obj
	mailboxResp := new(Mailbox)

	// Send Request
	_, err = s.client.Perform(req, mailboxResp)

	if err != nil {
		return nil, err
	}

	return mailboxResp, nil
}

// Update turns warmup on or off for a mailbox
// https://developer.send.io/docs/api
func (s *MailboxesSvcImpl) Update(mailboxId string, params *UpdateMailboxRequest) (*Mailbox, error) {
	return s.UpdateWithContext(context.Background(), mailboxId, params)
}

// ResetPasswordWithContext sets a new password for a mailbox and returns the
// new credentials. The old password stops working. The password is also an
// API key for that mailbox.
// https://developer.send.io/docs/api
func (s *MailboxesSvcImpl) ResetPasswordWithContext(ctx context.Context, mailboxId string) (*ResetMailboxPasswordResponse, error) {
	path := "mailboxes/" + mailboxId + "/password"

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, ErrFailedToCreateMailboxesResetPasswordRequest
	}

	// Build response recipient obj
	credentialsResp := new(ResetMailboxPasswordResponse)

	// Send Request
	_, err = s.client.Perform(req, credentialsResp)

	if err != nil {
		return nil, err
	}

	return credentialsResp, nil
}

// ResetPassword sets a new password for a mailbox. See ResetPasswordWithContext.
// https://developer.send.io/docs/api
func (s *MailboxesSvcImpl) ResetPassword(mailboxId string) (*ResetMailboxPasswordResponse, error) {
	return s.ResetPasswordWithContext(context.Background(), mailboxId)
}

// RemoveWithContext deletes a mailbox by ID. This removes the email account,
// and the agent inbox on it if there is one.
// https://developer.send.io/docs/api
func (s *MailboxesSvcImpl) RemoveWithContext(ctx context.Context, mailboxId string) (*RemoveMailboxResponse, error) {
	path := "mailboxes/" + mailboxId

	// Prepare request
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, ErrFailedToCreateMailboxesRemoveRequest
	}

	// Build response recipient obj
	mailboxResp := new(RemoveMailboxResponse)

	// Send Request
	_, err = s.client.Perform(req, mailboxResp)

	if err != nil {
		return nil, err
	}

	return mailboxResp, nil
}

// Remove deletes a mailbox by ID. See RemoveWithContext.
// https://developer.send.io/docs/api
func (s *MailboxesSvcImpl) Remove(mailboxId string) (*RemoveMailboxResponse, error) {
	return s.RemoveWithContext(context.Background(), mailboxId)
}
