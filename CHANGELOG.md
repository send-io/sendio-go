# Changelog

## 1.2.0 (2026-11-18)

- New `Transactional` resource (`/transactional`): `Create` / `CreateWithOptions` (Idempotency-Key), `List` (`Search`, `TemplateId`, pagination), `Get` (with the published version's `Variables`), `Update` (`ExpectedUpdatedAt` for a 409 on a stale save), `Remove`, `SendTest` and `Stats`, each with a `...WithContext` variant. `TrackOpens` and `TrackClicks` are off by default.
- `SendEmailRequest.TransactionalId` and `Variables` send a transactional email by id; `from` and `subject` are then left out of the body unless set. `SendEmailRequest.TrackOpens` / `TrackClicks` override tracking for one email. `Email.TransactionalId` reads it back.
- `CreateBroadcastRequest` and `UpdateBroadcastRequest` take `TemplateId` and `TemplateVersionId` (the exact published version to copy); `UpdateBroadcastRequest.DetachTemplate` sends `template_id: null`. `Broadcasts.Create` no longer requires `From` and `Subject` when `TemplateId` is set. `Broadcast` reads the template link fields and `UnfillableVariables`.
- Deprecated: sending a template with `SendEmailRequest.Template` (template id). Send a transactional email with `TransactionalId` instead; the old way still works for now.
- `Mailbox` reads `CanSend` and `SendReason` (whether the address is accepted as `From` on `Emails.Send`, and why not).
- `Webhook` and `WebhookInList` read `DisabledReason` and `FailingSince`; `Webhook` also reads `SigningSecretHidden`.
- New automation template categories: `AutomationTemplateCategoryAppointments`, `Events`, `Education`, `Account` and `Nonprofit`.
- User-Agent `sendio-go:1.2.0`.

## 1.1.0 (2026-11-11)

First public release.

- Automation templates: `Automations.Templates.List` / `ListWithContext` / `ListWithOptions` (filter by `Category`, for example `sendio.AutomationTemplateCategoryOnboarding`) and `Automations.Templates.Get(slug)` / `GetWithContext` read the prebuilt workflows; `Automations.CreateFromTemplate` / `CreateFromTemplateWithContext` / `CreateFromTemplateWithOptions` set one up (`POST /automations/from-template`) with `Sender`, `Events`, `Params`, `TopicId`, and an `Idempotency-Key` from `CreateAutomationFromTemplateOptions.IdempotencyKey`. `Client.Automations` is now a `*AutomationsSvcImpl` (it still implements `AutomationsSvc`) so it can carry `Templates`.
- Removed `Client.DomainClaims` (`Create`, `Get`, `Verify`) and the `DomainClaim` types: the API has no domain-claim routes, so every call failed with a 404. Claim a domain by deleting it from the team that owns it, then adding it to yours.
- `UpdateDomainRequest.Capabilities` turns receiving on or off for a domain (`&sendio.DomainCapabilities{Receiving: sendio.DomainCapabilityStatusDisabled}`); `"disabled"` is sent as-is. `CreateDomainRequest.Capabilities` (receiving at create) is now honoured by the API. Turning receiving on with `Update` asks for the domain's MX record (publish it to start receiving email); sending can't be disabled.
- New `Mailboxes` resource: `List`, `Create` (by `Address`, `LocalPart` + `DomainId`, or a free mailbox by `Username`), `Update` (`WarmupEnabled`, sent even when false), `ResetPassword` and `Remove`, each with a `...WithContext` variant. `Create` and `ResetPassword` return the IMAP/SMTP `MailboxCredentials`.
- `CreateBroadcastRequest` and `UpdateBroadcastRequest` set `PreviewText` and `TopicId`, plus the send.io extensions `SegmentIds`, `ExcludeSegmentIds`, `Audience` (`"all"` sends to every contact), `UnsubscribeScope`, `Tracking` (`*bool`: `sendio.Bool(false)` turns it off) and `TemplateEngine`. Unset fields are not sent, so existing requests send exactly what they did.
- `Broadcasts.Create` accepts `SegmentIds` or `Audience: "all"` in place of `SegmentId`; its error for a request with no audience now names them.
- `Broadcast` (from `Get` and `List`) reads `TopicId` (nil when there is none) and the same extensions.
- `Broadcasts.Update` and `Broadcasts.Send` no longer put `broadcast_id` in the JSON body; the id goes in the path only. The API refuses any key but `name` when renaming a broadcast that is sending or sent, so renaming one failed with a 422.
- New: `Broadcasts.ListWithFilters` (`Status`, `SegmentId`, `Search` plus pagination), `Estimate` (audience size before saving), `SendTest` (1 to 5 addresses), `Stats`, and the versions API: `ListVersions`/`ListVersionsWithOptions`, `GetVersion`, `UpdateVersion` (name and comment) and `RestoreVersion`.
- `Broadcast` also reads `VersionNumber`, `UpdatedAt`, `FailedReason`, `HoldReason` and `HoldUntil`.
- API errors are returned as `*sendio.APIError` with `StatusCode`, `Name` and `Message` (read them with `errors.As`). `err.Error()` is unchanged (`[ERROR]: <message>`); 429 still returns `*sendio.RateLimitError`. `InvalidRequestError` and `DefaultError` are deprecated.
- `SendEmailRequest.ReplyToAddresses` sends several reply-to addresses (`reply_to` as an array, after `ReplyTo` if it is set). A request with only `ReplyTo` is sent as before.
- User-Agent `sendio-go:1.1.0`.

## 1.0.0 (2026-10-08)

- First release of the send.io Go SDK (`github.com/send-io/sendio-go`).
- Default base URL `https://api.send.io/`, overridable with `SENDIO_BASE_URL` (read at client creation) or `Client.SetBaseURL`.
- `NewClient("")` reads `SENDIO_API_KEY`.
- User-Agent `sendio-go:1.0.0`.
