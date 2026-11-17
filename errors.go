package sendio

import (
	"errors"
	"fmt"
)

// MissingRequiredFieldsError is used when a required field is missing before making an API request
type MissingRequiredFieldsError struct {
	message string
}

func (e *MissingRequiredFieldsError) Error() string {
	return fmt.Sprintf("%s", e.message)
}

// APIError is returned when the send.io API answers with an error status
// (other than 429, which returns *RateLimitError). Read it with errors.As:
//
//	var apiErr *sendio.APIError
//	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound { ... }
type APIError struct {
	// StatusCode is the HTTP status of the response.
	StatusCode int `json:"statusCode"`
	// Name is the API's error name, e.g. "validation_error" or "not_found".
	// Empty when the response had no JSON body.
	Name string `json:"name"`
	// Message is the API's error message, or the HTTP status text when the
	// response had no JSON body.
	Message string `json:"message"`
}

// Error returns "[ERROR]: <message>", the same text earlier versions returned.
func (e *APIError) Error() string {
	if e.Message == "" {
		return "[ERROR]: Unknown Error"
	}
	return "[ERROR]: " + e.Message
}

// API error names a caller may want to handle by name: compare with
// APIError.Name, or RateLimitError.Name for the 429s. The API documents the
// full list at https://developer.send.io/docs/api#errors.
const (
	// ErrorNameInsufficientCredits (402): the credit balance does not cover
	// the send. Nothing was sent; buy credits, then retry.
	ErrorNameInsufficientCredits = "insufficient_credits"
	// ErrorNameRateLimitExceeded (429): too many requests per second.
	ErrorNameRateLimitExceeded = "rate_limit_exceeded"
	// ErrorNameDailyQuotaExceeded (429): the account's daily sending limit.
	ErrorNameDailyQuotaExceeded = "daily_quota_exceeded"
	// ErrorNameMonthlyQuotaExceeded (429): the account's monthly sending limit.
	ErrorNameMonthlyQuotaExceeded = "monthly_quota_exceeded"
	// ErrorNameSubaccountDailyQuotaExceeded (429): a subaccount's own daily limit.
	ErrorNameSubaccountDailyQuotaExceeded = "subaccount_daily_quota_exceeded"
	// ErrorNameSubaccountMonthlyQuotaExceeded (429): a subaccount's own monthly limit.
	ErrorNameSubaccountMonthlyQuotaExceeded = "subaccount_monthly_quota_exceeded"
)

// ErrRateLimit is a sentinel error for rate limit detection with errors.Is
var ErrRateLimit = errors.New("rate limit exceeded")

// RateLimitError represents a rate limit error with metadata from response headers
type RateLimitError struct {
	// Name is the API's error name: rate_limit_exceeded, or a quota name such
	// as daily_quota_exceeded or subaccount_monthly_quota_exceeded.
	Name string

	// Message is the error message from the API
	Message string

	// Limit is the maximum number of requests allowed in the current window (raw header value)
	Limit string

	// Remaining is the number of requests remaining in the current window (raw header value)
	Remaining string

	// Reset is the time when the rate limit will reset in seconds (raw header value)
	Reset string

	// RetryAfter is the recommended wait time before retrying in seconds (raw header value)
	RetryAfter string
}

// Error implements the error interface
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s (limit: %s, remaining: %s, reset: %s, retry after: %s)",
		e.Message, e.Limit, e.Remaining, e.Reset, e.RetryAfter)
}

// Is implements errors.Is support for detecting rate limit errors
func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimit
}

// BroadcastsSvc errors
var (
	ErrFailedToCreateBroadcastUpdateRequest     = errors.New("[ERROR]: Failed to create Broadcasts.Update request")
	ErrFailedToCreateBroadcastSendRequest       = errors.New("[ERROR]: Failed to create Broadcasts.Send request")
	ErrFailedToCreateBroadcastCreateRequest     = errors.New("[ERROR]: Failed to create Broadcasts.Create request")
	ErrFailedToCreateBroadcastCancelRequest     = errors.New("[ERROR]: Failed to create Broadcasts.Cancel request")
	ErrFailedToCreateBroadcastDuplicateRequest  = errors.New("[ERROR]: Failed to create Broadcasts.Duplicate request")
	ErrFailedToCreateBroadcastRecipientsRequest = errors.New("[ERROR]: Failed to create Broadcasts.Recipients request")
)

// ApiKeySvc errors
var (
	ErrFailedToCreateApiKeysCreateRequest = errors.New("[ERROR]: Failed to create ApiKeys.Create request")
	ErrFailedToCreateApiKeysListRequest   = errors.New("[ERROR]: Failed to create ApiKeys.List request")
	ErrFailedToCreateApiKeysRemoveRequest = errors.New("[ERROR]: Failed to create ApiKeys.Remove request")
	ErrFailedToCreateApiKeysUpdateRequest = errors.New("[ERROR]: Failed to create ApiKeys.Update request")
)

// EmailsSvc errors
var (
	ErrFailedToCreateUpdateEmailRequest           = errors.New("[ERROR]: Failed to create UpdateEmail request")
	ErrFailedToCreateEmailsSendRequest            = errors.New("[ERROR]: Failed to create SendEmail request")
	ErrFailedToCreateEmailsGetRequest             = errors.New("[ERROR]: Failed to create GetEmail request")
	ErrFailedToCreateEmailsListRequest            = errors.New("[ERROR]: Failed to create ListEmails request")
	ErrFailedToCreateEmailsGetAttachmentRequest   = errors.New("[ERROR]: Failed to create Emails.GetAttachment request")
	ErrFailedToCreateEmailsListAttachmentsRequest = errors.New("[ERROR]: Failed to create Emails.ListAttachments request")
	ErrFailedToCreateEmailsShareRequest           = errors.New("[ERROR]: Failed to create Emails.Share request")
	ErrFailedToCreateEmailsMetricsRequest         = errors.New("[ERROR]: Failed to create Emails.Metrics request")
)

// TemplatesSvc errors
var (
	ErrFailedToCreateTemplateCreateRequest    = errors.New("[ERROR]: Failed to create Templates.Create request")
	ErrFailedToCreateTemplateGetRequest       = errors.New("[ERROR]: Failed to create Templates.Get request")
	ErrFailedToCreateTemplateListRequest      = errors.New("[ERROR]: Failed to create Templates.List request")
	ErrFailedToCreateTemplateUpdateRequest    = errors.New("[ERROR]: Failed to create Templates.Update request")
	ErrFailedToCreateTemplatePublishRequest   = errors.New("[ERROR]: Failed to create Templates.Publish request")
	ErrFailedToCreateTemplateDuplicateRequest = errors.New("[ERROR]: Failed to create Templates.Duplicate request")
	ErrFailedToCreateTemplateRemoveRequest    = errors.New("[ERROR]: Failed to create Templates.Remove request")
)

// ReceivingSvc errors
var (
	ErrFailedToCreateReceivingGetRequest             = errors.New("[ERROR]: Failed to create Receiving.Get request")
	ErrFailedToCreateReceivingListRequest            = errors.New("[ERROR]: Failed to create Receiving.List request")
	ErrFailedToCreateReceivingGetAttachmentRequest   = errors.New("[ERROR]: Failed to create Receiving.GetAttachment request")
	ErrFailedToCreateReceivingListAttachmentsRequest = errors.New("[ERROR]: Failed to create Receiving.ListAttachments request")
)

// TopicsSvc errors
var (
	ErrFailedToCreateTopicCreateRequest = errors.New("[ERROR]: Failed to create Topics.Create request")
	ErrFailedToCreateTopicGetRequest    = errors.New("[ERROR]: Failed to create Topics.Get request")
	ErrFailedToCreateTopicListRequest   = errors.New("[ERROR]: Failed to create Topics.List request")
	ErrFailedToCreateTopicUpdateRequest = errors.New("[ERROR]: Failed to create Topics.Update request")
	ErrFailedToCreateTopicRemoveRequest = errors.New("[ERROR]: Failed to create Topics.Remove request")
)

// LogsSvc errors
var (
	ErrFailedToCreateLogsGetRequest  = errors.New("[ERROR]: Failed to create Logs.Get request")
	ErrFailedToCreateLogsListRequest = errors.New("[ERROR]: Failed to create Logs.List request")
)

// AutomationsSvc errors
var (
	ErrFailedToCreateAutomationCreateRequest             = errors.New("[ERROR]: Failed to create Automations.Create request")
	ErrFailedToCreateAutomationGetRequest                = errors.New("[ERROR]: Failed to create Automations.Get request")
	ErrFailedToCreateAutomationListRequest               = errors.New("[ERROR]: Failed to create Automations.List request")
	ErrFailedToCreateAutomationUpdateRequest             = errors.New("[ERROR]: Failed to create Automations.Update request")
	ErrFailedToCreateAutomationRemoveRequest             = errors.New("[ERROR]: Failed to create Automations.Remove request")
	ErrFailedToCreateAutomationDuplicateRequest          = errors.New("[ERROR]: Failed to create Automations.Duplicate request")
	ErrFailedToCreateAutomationStopRequest               = errors.New("[ERROR]: Failed to create Automations.Stop request")
	ErrFailedToCreateAutomationListRunsRequest           = errors.New("[ERROR]: Failed to create Automations.ListRuns request")
	ErrFailedToCreateAutomationGetRunRequest             = errors.New("[ERROR]: Failed to create Automations.GetRun request")
	ErrFailedToCreateAutomationCreateFromTemplateRequest = errors.New("[ERROR]: Failed to create Automations.CreateFromTemplate request")
)

// AutomationTemplatesSvc errors
var (
	ErrFailedToCreateAutomationTemplatesListRequest = errors.New("[ERROR]: Failed to create Automations.Templates.List request")
	ErrFailedToCreateAutomationTemplatesGetRequest  = errors.New("[ERROR]: Failed to create Automations.Templates.Get request")
)

// EventsSvc errors
var (
	ErrFailedToCreateEventCreateRequest = errors.New("[ERROR]: Failed to create Events.Create request")
	ErrFailedToCreateEventGetRequest    = errors.New("[ERROR]: Failed to create Events.Get request")
	ErrFailedToCreateEventListRequest   = errors.New("[ERROR]: Failed to create Events.List request")
	ErrFailedToCreateEventUpdateRequest = errors.New("[ERROR]: Failed to create Events.Update request")
	ErrFailedToCreateEventRemoveRequest = errors.New("[ERROR]: Failed to create Events.Remove request")
	ErrFailedToCreateEventSendRequest   = errors.New("[ERROR]: Failed to create Events.Send request")
)

// OAuthGrantsSvc errors
var (
	ErrFailedToCreateOAuthGrantsListRequest   = errors.New("[ERROR]: Failed to create OAuthGrants.List request")
	ErrFailedToCreateOAuthGrantsRevokeRequest = errors.New("[ERROR]: Failed to create OAuthGrants.Revoke request")
)

// UsageSvc errors
var (
	ErrFailedToCreateUsageGetRequest = errors.New("[ERROR]: Failed to create Usage.Get request")
)

// MailboxesSvc errors
var (
	ErrFailedToCreateMailboxesListRequest          = errors.New("[ERROR]: Failed to create Mailboxes.List request")
	ErrFailedToCreateMailboxesCreateRequest        = errors.New("[ERROR]: Failed to create Mailboxes.Create request")
	ErrFailedToCreateMailboxesUpdateRequest        = errors.New("[ERROR]: Failed to create Mailboxes.Update request")
	ErrFailedToCreateMailboxesResetPasswordRequest = errors.New("[ERROR]: Failed to create Mailboxes.ResetPassword request")
	ErrFailedToCreateMailboxesRemoveRequest        = errors.New("[ERROR]: Failed to create Mailboxes.Remove request")
)

// TransactionalSvc errors
var (
	ErrFailedToCreateTransactionalCreateRequest   = errors.New("[ERROR]: Failed to create Transactional.Create request")
	ErrFailedToCreateTransactionalListRequest     = errors.New("[ERROR]: Failed to create Transactional.List request")
	ErrFailedToCreateTransactionalGetRequest      = errors.New("[ERROR]: Failed to create Transactional.Get request")
	ErrFailedToCreateTransactionalUpdateRequest   = errors.New("[ERROR]: Failed to create Transactional.Update request")
	ErrFailedToCreateTransactionalRemoveRequest   = errors.New("[ERROR]: Failed to create Transactional.Remove request")
	ErrFailedToCreateTransactionalSendTestRequest = errors.New("[ERROR]: Failed to create Transactional.SendTest request")
	ErrFailedToCreateTransactionalStatsRequest    = errors.New("[ERROR]: Failed to create Transactional.Stats request")
)
