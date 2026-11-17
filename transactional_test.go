package sendio

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const txID = "tx_2hX9Yq1bVv0kq3W4m8sZJt1aB7c"

const txJSON = `{"object":"transactional_email","id":"` + txID + `","name":"Sign-in code","template_id":"34a080c9-b17d-4187-ad80-5af20266e535",
	"template_name":"Sign-in code","template_alias":"signin","template_deleted":false,"published_version_id":"6a1f4c2e-8b3d-4e5f-9a7b-0c1d2e3f4a5b",
	"published_version_number":3,"from":"Acme <security@acme.com>","subject":"Your code is {{CODE}}","preview_text":null,"reply_to":[],
	"track_opens":false,"track_clicks":true,"created_via":"api","created_at":"2026-11-16T09:00:00Z","updated_at":"2026-11-16T09:05:00Z",
	"variables":[{"key":"CODE","type":"string","fallback_value":null}]}`

func TestCreateTransactionalEmail(t *testing.T) {
	setup()
	defer teardown()
	var body, idem string
	mux.HandleFunc("/transactional", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		idem = r.Header.Get("Idempotency-Key")
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(txJSON))
	})
	tx, err := client.Transactional.CreateWithOptions(context.Background(), &CreateTransactionalEmailRequest{
		Name: "Sign-in code", TemplateId: "signin", From: "Acme <security@acme.com>", Subject: "Your code is {{CODE}}", TrackClicks: Bool(true),
	}, &TransactionalEmailOptions{IdempotencyKey: "tx-1"})
	assert.NoError(t, err)
	assert.Equal(t, "tx-1", idem)
	assert.JSONEq(t, `{"name":"Sign-in code","template_id":"signin","from":"Acme <security@acme.com>","subject":"Your code is {{CODE}}","track_clicks":true}`, body)
	assert.Equal(t, txID, tx.Id)
	assert.Equal(t, 3, *tx.PublishedVersionNumber)
	assert.Nil(t, tx.PreviewText)
	assert.True(t, tx.TrackClicks)
	assert.Equal(t, "CODE", tx.Variables[0].Key)
}

func TestListTransactionalEmailsFilters(t *testing.T) {
	setup()
	defer teardown()
	var query string
	serveJSON(t, "/transactional", http.MethodGet, `{"object":"list","has_more":false,"data":[`+txJSON+`]}`, nil, &query)
	limit, after := 10, txID
	l, err := client.Transactional.List(&ListTransactionalEmailsOptions{Limit: &limit, After: &after, Search: "code", TemplateId: "signin"})
	assert.NoError(t, err)
	assert.Equal(t, "after="+txID+"&limit=10&search=code&template_id=signin", query)
	assert.Len(t, l.Data, 1)
}

func TestUpdateTransactionalEmailSendsOnlyWhatIsSet(t *testing.T) {
	setup()
	defer teardown()
	var body string
	serveJSON(t, "/transactional/"+txID, http.MethodPatch, txJSON, &body, nil)
	empty := []string{}
	_, err := client.Transactional.Update(txID, &UpdateTransactionalEmailRequest{Subject: "New", ReplyTo: &empty, TrackOpens: Bool(false), ExpectedUpdatedAt: "2026-11-16T09:05:00Z"})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"subject":"New","reply_to":[],"track_opens":false,"expected_updated_at":"2026-11-16T09:05:00Z"}`, body)
}

func TestTransactionalEmailGetRemoveTestStats(t *testing.T) {
	setup()
	defer teardown()
	var testBody, statsQuery string
	mux.HandleFunc("/transactional/"+txID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete {
			w.Write([]byte(`{"object":"transactional_email","id":"` + txID + `","deleted":true}`))
			return
		}
		testMethod(t, r, http.MethodGet)
		w.Write([]byte(txJSON))
	})
	serveJSON(t, "/transactional/"+txID+"/test", http.MethodPost, `{"id":"49a3999c-0ce1-4ea6-ab68-afcd6dc2e794"}`, &testBody, nil)
	serveJSON(t, "/transactional/"+txID+"/stats", http.MethodGet, `{"object":"transactional_email_stats","transactional_id":"`+txID+`","totals":{"sent":3}}`, nil, &statsQuery)

	g, err := client.Transactional.Get(txID)
	assert.NoError(t, err)
	assert.Equal(t, "signin", g.TemplateAlias)
	sent, err := client.Transactional.SendTest(txID, &SendTestTransactionalEmailRequest{To: []string{"me@acme.com"}, Variables: map[string]any{"CODE": "123456"}})
	assert.NoError(t, err)
	assert.Equal(t, "49a3999c-0ce1-4ea6-ab68-afcd6dc2e794", sent.Id)
	assert.JSONEq(t, `{"to":["me@acme.com"],"variables":{"CODE":"123456"}}`, testBody)
	st, err := client.Transactional.Stats(txID, &TransactionalEmailStatsOptions{Granularity: "hourly", Timezone: "Europe/Paris"})
	assert.NoError(t, err)
	assert.Equal(t, "granularity=hourly&timezone=Europe%2FParis", statsQuery)
	assert.Equal(t, float64(3), st.Totals["sent"])
	rm, err := client.Transactional.Remove(txID)
	assert.NoError(t, err)
	assert.True(t, rm.Deleted)
}

// A transactional send carries the id and the variables, and leaves from
// and subject to the transactional email; other sends are unchanged.
func TestSendEmailTransactionalId(t *testing.T) {
	b, err := json.Marshal(SendEmailRequest{TransactionalId: txID, To: []string{"ada@example.com"}, Variables: map[string]any{"CODE": "123456"}, TrackOpens: Bool(false)})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"transactional_id":"`+txID+`","to":["ada@example.com"],"variables":{"CODE":"123456"},"track_opens":false}`, string(b))

	b, err = json.Marshal(SendEmailRequest{From: "a@acme.com", To: []string{"ada@example.com"}, Html: "<p>x</p>"})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"from":"a@acme.com","to":["ada@example.com"],"subject":"","html":"<p>x</p>"}`, string(b))

	var e Email
	assert.NoError(t, json.Unmarshal([]byte(`{"id":"e1","transactional_id":"`+txID+`"}`), &e))
	assert.Equal(t, txID, e.TransactionalId)
}

func TestBroadcastTemplateLink(t *testing.T) {
	setup()
	defer teardown()
	var create, update, detach map[string]any
	captureBody(t, "/broadcasts", http.MethodPost, &create)
	_, err := client.Broadcasts.Create(&CreateBroadcastRequest{Name: "Codes", SegmentId: "s1", TemplateId: "signin", TemplateVersionId: "6a1f4c2e-8b3d-4e5f-9a7b-0c1d2e3f4a5b"})
	assert.NoError(t, err)
	assert.Equal(t, "signin", create["template_id"])
	assert.Equal(t, "6a1f4c2e-8b3d-4e5f-9a7b-0c1d2e3f4a5b", create["template_version_id"])

	b, _ := json.Marshal(UpdateBroadcastRequest{BroadcastId: "b1", TemplateVersionId: "v2"})
	assert.NoError(t, json.Unmarshal(b, &update))
	assert.Equal(t, map[string]any{"template_version_id": "v2"}, update)
	b, _ = json.Marshal(UpdateBroadcastRequest{BroadcastId: "b1", DetachTemplate: true})
	assert.NoError(t, json.Unmarshal(b, &detach))
	assert.Equal(t, map[string]any{"template_id": nil}, detach)

	var got Broadcast
	assert.NoError(t, json.Unmarshal([]byte(`{"id":"b1","template_id":"t1","template_version_id":"v1","template_version_number":2,"template_deleted":false,"unfillable_variables":["CODE"]}`), &got))
	assert.Equal(t, "v1", *got.TemplateVersionId)
	assert.Equal(t, 2, *got.TemplateVersionNumber)
	assert.Equal(t, []string{"CODE"}, got.UnfillableVariables)
}
