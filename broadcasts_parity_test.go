package sendio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// serveJSON serves path with method, records the raw body and query it
// receives, and answers with resp.
func serveJSON(t *testing.T, path, method, resp string, body *string, query *string) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, method)
		if body != nil {
			b, _ := io.ReadAll(r.Body)
			*body = string(b)
		}
		if query != nil {
			*query = r.URL.RawQuery
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, resp)
	})
}

func TestUpdateBroadcastIdNotInBody(t *testing.T) {
	setup()
	defer teardown()
	var body string
	serveJSON(t, "/broadcasts/b1", http.MethodPatch, `{"object":"broadcast","id":"b1"}`, &body, nil)

	_, err := client.Broadcasts.Update(&UpdateBroadcastRequest{BroadcastId: "b1", Name: "Renamed"})
	assert.NoError(t, err)
	// A sent broadcast accepts only name: the id must never be in the body.
	assert.JSONEq(t, `{"name":"Renamed"}`, body)
}

func TestSendBroadcastIdNotInBody(t *testing.T) {
	setup()
	defer teardown()
	var body string
	serveJSON(t, "/broadcasts/b1/send", http.MethodPost, `{"id":"b1"}`, &body, nil)

	_, err := client.Broadcasts.Send(&SendBroadcastRequest{BroadcastId: "b1", ScheduledAt: "in 1 hour"})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"scheduled_at":"in 1 hour"}`, body)
}

func TestUpdateBroadcastTrackingFalseIsSent(t *testing.T) {
	setup()
	defer teardown()
	var body string
	serveJSON(t, "/broadcasts/b1", http.MethodPatch, `{"object":"broadcast","id":"b1"}`, &body, nil)

	_, err := client.Broadcasts.Update(&UpdateBroadcastRequest{
		BroadcastId: "b1", Tracking: Bool(false), Audience: "all", UnsubscribeScope: "list",
		ExcludeSegmentIds: []string{"x1"}, TemplateEngine: "simple",
	})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"tracking":false,"audience":"all","unsubscribe_scope":"list","exclude_segment_ids":["x1"],"template_engine":"simple"}`, body)
}

func TestListBroadcastsWithFilters(t *testing.T) {
	setup()
	defer teardown()
	var query string
	serveJSON(t, "/broadcasts", http.MethodGet, `{"object":"list","has_more":true,"data":[{"id":"b1","name":null,"status":"draft","segment_ids":[],"audience":"all"}]}`, nil, &query)

	limit, after := 10, "b0"
	resp, err := client.Broadcasts.ListWithFilters(context.Background(), &ListBroadcastsOptions{Limit: &limit, After: &after, Status: "draft", SegmentId: "s1", Search: "news"})
	assert.NoError(t, err)
	assert.Equal(t, "after=b0&limit=10&search=news&segment_id=s1&status=draft", query)
	assert.True(t, resp.HasMore)
	assert.Equal(t, "", resp.Data[0].Name, "a null name reads as empty")
	assert.Equal(t, "all", resp.Data[0].Audience)

	_, err = client.Broadcasts.ListWithFilters(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "", query)
}

func TestListBroadcastsUnchanged(t *testing.T) {
	setup()
	defer teardown()
	var query string
	serveJSON(t, "/broadcasts", http.MethodGet, `{"object":"list","has_more":false,"data":[]}`, nil, &query)

	limit := 5
	_, err := client.Broadcasts.ListWithOptions(context.Background(), &ListOptions{Limit: &limit})
	assert.NoError(t, err)
	assert.Equal(t, "limit=5", query)
}

func TestGetBroadcastAllFields(t *testing.T) {
	setup()
	defer teardown()
	serveJSON(t, "/broadcasts/b1", http.MethodGet, `{"object":"broadcast","id":"b1","name":null,"segment_id":null,"audience_id":null,
		"version_number":3,"updated_at":"2026-10-01 10:00:00+00","failed_reason":null,"hold_reason":"bulk_cap","hold_until":"2026-10-02 00:00:00+00"}`, nil, nil)
	serveJSON(t, "/broadcasts/b2", http.MethodGet, `{"object":"broadcast","id":"b2","version_number":null}`, nil, nil)

	b, err := client.Broadcasts.Get("b1")
	assert.NoError(t, err)
	assert.Equal(t, "", b.Name)
	assert.Equal(t, "", b.SegmentId)
	if assert.NotNil(t, b.VersionNumber) {
		assert.Equal(t, 3, *b.VersionNumber)
	}
	assert.Equal(t, "2026-10-01 10:00:00+00", b.UpdatedAt)
	assert.Equal(t, "bulk_cap", b.HoldReason)
	assert.Equal(t, "2026-10-02 00:00:00+00", b.HoldUntil)

	b2, err := client.Broadcasts.Get("b2")
	assert.NoError(t, err)
	assert.Nil(t, b2.VersionNumber)
}

func TestEstimateBroadcast(t *testing.T) {
	setup()
	defer teardown()
	var body string
	serveJSON(t, "/broadcasts/estimate", http.MethodPost, `{"object":"broadcast_estimate","recipients":42,"skipped":{"invalid":3}}`, &body, nil)

	est, err := client.Broadcasts.Estimate(&EstimateBroadcastRequest{SegmentIds: []string{"s1", "s2"}, ExcludeSegmentIds: []string{"x"}, TopicId: "t1"})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"segment_ids":["s1","s2"],"exclude_segment_ids":["x"],"topic_id":"t1"}`, body)
	assert.Equal(t, BroadcastEstimate{Object: "broadcast_estimate", Recipients: 42, Skipped: BroadcastSkippedContacts{Invalid: 3}}, est)

	_, err = client.Broadcasts.Estimate(&EstimateBroadcastRequest{Audience: "all"})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"audience":"all"}`, body)

	_, err = client.Broadcasts.Estimate(&EstimateBroadcastRequest{SegmentId: "s1"})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"segment_id":"s1"}`, body)
}

func TestSendBroadcastTest(t *testing.T) {
	setup()
	defer teardown()
	var body string
	serveJSON(t, "/broadcasts/b1/test", http.MethodPost, `{"object":"broadcast_test","sent":2}`, &body, nil)

	resp, err := client.Broadcasts.SendTest("b1", &SendBroadcastTestRequest{To: []string{"a@example.com", "b@example.com"}})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"to":["a@example.com","b@example.com"]}`, body)
	assert.Equal(t, 2, resp.Sent)
	assert.Equal(t, "broadcast_test", resp.Object)

	_, err = client.Broadcasts.SendTest("b1", &SendBroadcastTestRequest{})
	assert.EqualError(t, err, "[ERROR]: To must have 1 to 5 addresses")
	_, err = client.Broadcasts.SendTest("b1", &SendBroadcastTestRequest{To: []string{"1", "2", "3", "4", "5", "6"}})
	assert.Error(t, err)
	_, err = client.Broadcasts.SendTest("", &SendBroadcastTestRequest{To: []string{"a@example.com"}})
	assert.EqualError(t, err, "[ERROR]: broadcastId cannot be empty")
}

func TestBroadcastStats(t *testing.T) {
	setup()
	defer teardown()
	serveJSON(t, "/broadcasts/b1/stats", http.MethodGet, `{"object":"broadcast_stats","status":"sent","recipients":100,"sent":98,"delivered":95,
		"bounced":2,"complained":1,"suppressed":2,"failed":0,"opened":60,"unique_opened":50,"clicked":20,"unique_clicked":10,"unsubscribed":1,
		"skipped":{"invalid":4},"delivery_rate":0.9694,"open_rate":0.5263,"click_rate":0.1053,"bounce_rate":0.0204,"unsubscribe_rate":0.0105,
		"progress":{"queued":98,"total":100}}`, nil, nil)

	st, err := client.Broadcasts.Stats("b1")
	assert.NoError(t, err)
	assert.Equal(t, BroadcastStats{
		Object: "broadcast_stats", Status: "sent", Recipients: 100, Sent: 98, Delivered: 95, Bounced: 2, Complained: 1, Suppressed: 2,
		Opened: 60, UniqueOpened: 50, Clicked: 20, UniqueClicked: 10, Unsubscribed: 1, Skipped: BroadcastSkippedContacts{Invalid: 4},
		DeliveryRate: 0.9694, OpenRate: 0.5263, ClickRate: 0.1053, BounceRate: 0.0204, UnsubscribeRate: 0.0105,
		Progress: BroadcastStatsProgress{Queued: 98, Total: 100},
	}, st)

	_, err = client.Broadcasts.Stats("")
	assert.Error(t, err)
}

func TestBroadcastVersions(t *testing.T) {
	setup()
	defer teardown()
	var query, body string
	serveJSON(t, "/broadcasts/b1/versions", http.MethodGet, `{"object":"list","has_more":false,"data":[
		{"object":"broadcast_version","id":"v2","broadcast_id":"b1","version_number":2,"name":null,"comment":null,"broadcast_name":"News",
		 "subject":"Hi","restored_from":1,"created_at":"2026-10-01","current":true}]}`, nil, &query)
	detail := `{"object":"broadcast_version","id":"v1","broadcast_id":"b1","version_number":1,"name":"First","comment":"c","current":false,
		"from":"a@example.com","reply_to":["r@example.com"],"preview_text":"p","html":"<p>x</p>","text":"x","template_engine":"handlebars",
		"tracking":false,"topic_id":null,"segment_ids":["s1"],"exclude_segment_ids":[],"audience":"segments","unsubscribe_scope":"list"}`
	mux.HandleFunc("/broadcasts/b1/versions/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			b, _ := io.ReadAll(r.Body)
			body = string(b)
		} else {
			testMethod(t, r, http.MethodGet)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, detail)
	})
	serveJSON(t, "/broadcasts/b1/versions/1/restore", http.MethodPost, `{"object":"broadcast_version","id":"v3","version_number":3,"restored_from":1,"current":true}`, &body, nil)

	limit, before := 10, "v1"
	list, err := client.Broadcasts.ListVersionsWithOptions(context.Background(), "b1", &ListOptions{Limit: &limit, Before: &before})
	assert.NoError(t, err)
	assert.Equal(t, "before=v1&limit=10", query)
	assert.Len(t, list.Data, 1)
	assert.Nil(t, list.Data[0].Name)
	assert.True(t, list.Data[0].Current)
	if assert.NotNil(t, list.Data[0].RestoredFrom) {
		assert.Equal(t, 1, *list.Data[0].RestoredFrom)
	}
	_, err = client.Broadcasts.ListVersions("b1")
	assert.NoError(t, err)
	assert.Equal(t, "", query)

	v, err := client.Broadcasts.GetVersion("b1", "1")
	assert.NoError(t, err)
	assert.Equal(t, "First", *v.Name)
	assert.Equal(t, "handlebars", v.TemplateEngine)
	assert.False(t, v.Tracking)
	assert.Nil(t, v.TopicId)
	assert.Equal(t, []string{"s1"}, v.SegmentIds)
	assert.Equal(t, "list", v.UnsubscribeScope)

	name := "Launch copy"
	_, err = client.Broadcasts.UpdateVersion("b1", "1", &UpdateBroadcastVersionRequest{Name: &name})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"Launch copy"}`, body)
	empty := ""
	_, err = client.Broadcasts.UpdateVersion("b1", "1", &UpdateBroadcastVersionRequest{Comment: &empty})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"comment":""}`, body, "an empty comment clears it and must be sent")

	body = "unset"
	r, err := client.Broadcasts.RestoreVersion("b1", "1")
	assert.NoError(t, err)
	assert.Equal(t, "", body, "restore sends no body")
	assert.Equal(t, 3, r.VersionNumber)

	_, err = client.Broadcasts.GetVersion("b1", "")
	assert.EqualError(t, err, "[ERROR]: version cannot be empty")
	_, err = client.Broadcasts.RestoreVersion("", "1")
	assert.EqualError(t, err, "[ERROR]: broadcastId cannot be empty")
}

// The structs marshal exactly the keys the API takes.
func TestBroadcastRequestsMarshal(t *testing.T) {
	b, _ := json.Marshal(UpdateBroadcastRequest{BroadcastId: "b1"})
	assert.Equal(t, `{}`, string(b))
	b, _ = json.Marshal(SendBroadcastRequest{BroadcastId: "b1"})
	assert.Equal(t, `{"scheduled_at":""}`, string(b))
}
