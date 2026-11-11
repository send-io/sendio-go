package sendio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// captureBody serves path, records the JSON body it receives and answers with id.
func captureBody(t *testing.T, path, method string, body *map[string]any) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, method)
		if err := json.NewDecoder(r.Body).Decode(body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id": "49a3999c-0ce1-4ea6-ab68-afcd6dc2e794"}`)
	})
}

// newBroadcastKeys are the fields this change adds: an older request must send none of them.
var newBroadcastKeys = []string{"preview_text", "topic_id", "segment_ids", "exclude_segment_ids", "audience", "unsubscribe_scope", "tracking", "template_engine"}

func TestCreateBroadcastPreviewTextAndTopic(t *testing.T) {
	setup()
	defer teardown()
	var body map[string]any
	captureBody(t, "/broadcasts", http.MethodPost, &body)

	_, err := client.Broadcasts.Create(&CreateBroadcastRequest{
		SegmentId:   "78261eea-8f8b-4381-83c6-79fa7120f1cf",
		From:        "hi@example.com",
		Subject:     "Hello",
		Html:        "<p>Hi</p>",
		PreviewText: "What is new this month",
		TopicId:     "b6d24b8e-af0b-4c3c-be0c-359bbd97381e",
	})
	assert.NoError(t, err)
	assert.Equal(t, "What is new this month", body["preview_text"])
	assert.Equal(t, "b6d24b8e-af0b-4c3c-be0c-359bbd97381e", body["topic_id"])
	assert.Equal(t, "78261eea-8f8b-4381-83c6-79fa7120f1cf", body["segment_id"])
}

func TestCreateBroadcastExtensions(t *testing.T) {
	setup()
	defer teardown()
	var body map[string]any
	captureBody(t, "/broadcasts", http.MethodPost, &body)

	_, err := client.Broadcasts.Create(&CreateBroadcastRequest{
		From:              "hi@example.com",
		Subject:           "Hello",
		Html:              "<p>Hi</p>",
		SegmentIds:        []string{"s1", "s2"},
		ExcludeSegmentIds: []string{"vip"},
		UnsubscribeScope:  "list",
		Tracking:          Bool(false),
		TemplateEngine:    "handlebars",
	})
	assert.NoError(t, err)
	assert.Equal(t, []any{"s1", "s2"}, body["segment_ids"])
	assert.Equal(t, []any{"vip"}, body["exclude_segment_ids"])
	assert.Equal(t, "list", body["unsubscribe_scope"])
	assert.Equal(t, false, body["tracking"], "tracking off must be sent, not dropped by omitempty")
	assert.Equal(t, "handlebars", body["template_engine"])
	assert.NotContains(t, body, "audience")
}

func TestCreateBroadcastAllContacts(t *testing.T) {
	setup()
	defer teardown()
	var body map[string]any
	captureBody(t, "/broadcasts", http.MethodPost, &body)

	_, err := client.Broadcasts.Create(&CreateBroadcastRequest{From: "hi@example.com", Subject: "Hello", Text: "Hi", Audience: "all"})
	assert.NoError(t, err)
	assert.Equal(t, "all", body["audience"])
	assert.NotContains(t, body, "segment_id")
	assert.NotContains(t, body, "segment_ids")
}

func TestCreateBroadcastOmitsUnsetNewFields(t *testing.T) {
	setup()
	defer teardown()
	var body map[string]any
	captureBody(t, "/broadcasts", http.MethodPost, &body)

	// A request written before these fields existed sends exactly what it did.
	_, err := client.Broadcasts.Create(&CreateBroadcastRequest{AudienceId: "a1", From: "hi@example.com", Subject: "Hello", Html: "<p>Hi</p>"})
	assert.NoError(t, err)
	for _, k := range newBroadcastKeys {
		assert.NotContains(t, body, k)
	}
}

func TestCreateBroadcastNeedsAnAudience(t *testing.T) {
	setup()
	defer teardown()
	// Neither a segment nor "all": still refused before any request.
	_, err := client.Broadcasts.Create(&CreateBroadcastRequest{From: "hi@example.com", Subject: "Hello", Audience: "segments"})
	assert.EqualError(t, err, `[ERROR]: Either SegmentId, AudienceId or SegmentIds must be provided, or Audience must be "all"`)
	_, err = client.Broadcasts.Create(&CreateBroadcastRequest{From: "hi@example.com", Subject: "Hello", SegmentIds: []string{}})
	assert.Error(t, err)
}

func TestUpdateBroadcastPreviewTextTopicAndTracking(t *testing.T) {
	setup()
	defer teardown()
	var body map[string]any
	captureBody(t, "/broadcasts/559ac32e-9ef5-46fb-82a1-b76b840c0f7b", http.MethodPatch, &body)

	_, err := client.Broadcasts.Update(&UpdateBroadcastRequest{
		BroadcastId: "559ac32e-9ef5-46fb-82a1-b76b840c0f7b",
		PreviewText: "New preheader",
		TopicId:     "b6d24b8e-af0b-4c3c-be0c-359bbd97381e",
		Tracking:    Bool(true),
	})
	assert.NoError(t, err)
	assert.Equal(t, "New preheader", body["preview_text"])
	assert.Equal(t, "b6d24b8e-af0b-4c3c-be0c-359bbd97381e", body["topic_id"])
	assert.Equal(t, true, body["tracking"])
	assert.NotContains(t, body, "subject", "an update sends only the fields that are set")
}

func TestUpdateBroadcastOmitsUnsetNewFields(t *testing.T) {
	setup()
	defer teardown()
	var body map[string]any
	captureBody(t, "/broadcasts/559ac32e-9ef5-46fb-82a1-b76b840c0f7b", http.MethodPatch, &body)

	_, err := client.Broadcasts.Update(&UpdateBroadcastRequest{BroadcastId: "559ac32e-9ef5-46fb-82a1-b76b840c0f7b", Name: "Renamed"})
	assert.NoError(t, err)
	assert.Equal(t, "Renamed", body["name"])
	for _, k := range newBroadcastKeys {
		assert.NotContains(t, body, k)
	}
}

func TestGetBroadcastTopicAndExtensions(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/broadcasts/b1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"broadcast","id":"b1","preview_text":"Pre","topic_id":"t1",
			"segment_ids":["s1","s2"],"exclude_segment_ids":["vip"],"audience":"segments",
			"unsubscribe_scope":"list","tracking":false,"template_engine":"handlebars"}`)
	})
	mux.HandleFunc("/broadcasts/b2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"broadcast","id":"b2","topic_id":null,"audience":"all","tracking":true}`)
	})

	b, err := client.Broadcasts.Get("b1")
	assert.NoError(t, err)
	assert.Equal(t, "Pre", b.PreviewText)
	if assert.NotNil(t, b.TopicId) {
		assert.Equal(t, "t1", *b.TopicId)
	}
	assert.Equal(t, []string{"s1", "s2"}, b.SegmentIds)
	assert.Equal(t, []string{"vip"}, b.ExcludeSegmentIds)
	assert.Equal(t, "segments", b.Audience)
	assert.Equal(t, "list", b.UnsubscribeScope)
	assert.False(t, b.Tracking)
	assert.Equal(t, "handlebars", b.TemplateEngine)

	b2, err := client.Broadcasts.Get("b2")
	assert.NoError(t, err)
	assert.Nil(t, b2.TopicId, "no topic is nil, not an empty string")
	assert.Equal(t, "all", b2.Audience)
	assert.True(t, b2.Tracking)
}
