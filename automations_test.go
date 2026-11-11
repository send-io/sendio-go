package sendio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAutomation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"object":"automation","id":"aut_123"}`)
	})

	resp, err := client.Automations.Create(&CreateAutomationRequest{
		Name: "Welcome Flow",
		Steps: []AutomationStep{
			{Key: "trigger_1", Type: AutomationStepTypeTrigger, Config: map[string]any{"event_name": "user.created"}},
			{Key: "send_1", Type: AutomationStepTypeSendEmail, Config: map[string]any{"template": map[string]any{"id": "tpl_abc"}}},
		},
		Connections: []AutomationConnection{
			{From: "trigger_1", To: "send_1"},
		},
	})
	if err != nil {
		t.Fatalf("Automations.Create returned error: %v", err)
	}
	assert.Equal(t, "automation", resp.Object)
	assert.Equal(t, "aut_123", resp.Id)
}

func TestCreateAutomationWithWebhookStep(t *testing.T) {
	setup()
	defer teardown()

	var sent map[string]any
	mux.HandleFunc("/automations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"object":"automation","id":"aut_123"}`)
	})

	_, err := client.Automations.Create(&CreateAutomationRequest{
		Name: "Sync to CRM",
		Steps: []AutomationStep{
			{Key: "trigger", Type: AutomationStepTypeTrigger, Config: map[string]any{"event_name": "user.created"}},
			NewWebhookStep("crm", WebhookStepConfig{
				URL:               "https://api.example.com/leads?src={{event.source}}",
				Method:            "POST",
				Headers:           []WebhookHeader{{Name: "Authorization", Value: "Bearer sk_test", Secret: true}},
				BodyType:          "json",
				Body:              map[string]any{"email": "{{contact.email}}", "n": 1},
				ContinueOnFailure: true,
				TimeoutSeconds:    15,
			}),
			NewWebhookStep("ping", WebhookStepConfig{URL: "https://hooks.example.com/ping", BodyType: "none"}),
		},
		Connections: []AutomationConnection{{From: "trigger", To: "crm"}, {From: "crm", To: "ping"}},
	})
	if err != nil {
		t.Fatalf("Automations.Create returned error: %v", err)
	}
	steps := sent["steps"].([]any)
	assert.Equal(t, map[string]any{
		"key":  "crm",
		"type": "webhook",
		"config": map[string]any{
			"url":                 "https://api.example.com/leads?src={{event.source}}",
			"method":              "POST",
			"headers":             []any{map[string]any{"name": "Authorization", "value": "Bearer sk_test", "secret": true}},
			"body_type":           "json",
			"body":                map[string]any{"email": "{{contact.email}}", "n": float64(1)},
			"continue_on_failure": true,
			"timeout_seconds":     float64(15),
		},
	}, steps[1])
	assert.Equal(t, map[string]any{"url": "https://hooks.example.com/ping", "body_type": "none"}, steps[2].(map[string]any)["config"])
}

func TestGetAutomation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations/aut_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"object": "automation",
			"id": "aut_123",
			"name": "Welcome Flow",
			"status": "enabled",
			"created_at": "2026-04-01 00:00:00+00",
			"updated_at": "2026-04-01 00:00:00+00",
			"steps": [
				{"key": "trigger_1", "type": "trigger", "config": {"event_name": "user.created"}},
				{"key": "send_1", "type": "send_email", "config": {"template": {"id": "tpl_abc"}}}
			],
			"connections": [
				{"from": "trigger_1", "to": "send_1"}
			]
		}`)
	})

	resp, err := client.Automations.Get("aut_123")
	if err != nil {
		t.Fatalf("Automations.Get returned error: %v", err)
	}
	assert.Equal(t, "automation", resp.Object)
	assert.Equal(t, "aut_123", resp.Id)
	assert.Equal(t, "Welcome Flow", resp.Name)
	assert.Equal(t, AutomationStatusEnabled, resp.Status)
	assert.Equal(t, 2, len(resp.Steps))
	assert.Equal(t, "trigger_1", resp.Steps[0].Key)
	assert.Equal(t, AutomationStepTypeTrigger, resp.Steps[0].Type)
	assert.Equal(t, "send_1", resp.Steps[1].Key)
	assert.Equal(t, 1, len(resp.Connections))
}

func TestListAutomations(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"object": "list",
			"has_more": false,
			"data": [
				{"id": "aut_123", "name": "Welcome Flow", "status": "enabled", "created_at": "2026-04-01 00:00:00+00", "updated_at": "2026-04-01 00:00:00+00"},
				{"id": "aut_456", "name": "Onboarding", "status": "disabled", "created_at": "2026-04-02 00:00:00+00", "updated_at": "2026-04-02 00:00:00+00"}
			]
		}`)
	})

	resp, err := client.Automations.List()
	if err != nil {
		t.Fatalf("Automations.List returned error: %v", err)
	}
	assert.Equal(t, "list", resp.Object)
	assert.Equal(t, false, resp.HasMore)
	assert.Equal(t, 2, len(resp.Data))
	assert.Equal(t, "aut_123", resp.Data[0].Id)
	assert.Equal(t, "Welcome Flow", resp.Data[0].Name)
	assert.Equal(t, AutomationStatusEnabled, resp.Data[0].Status)
	assert.Equal(t, "aut_456", resp.Data[1].Id)
	assert.Equal(t, AutomationStatusDisabled, resp.Data[1].Status)
}

func TestListAutomationsWithOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "enabled", r.URL.Query().Get("status"))
		assert.Equal(t, "5", r.URL.Query().Get("limit"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"object": "list",
			"has_more": true,
			"data": [
				{"id": "aut_123", "name": "Welcome Flow", "status": "enabled", "created_at": "2026-04-01 00:00:00+00", "updated_at": "2026-04-01 00:00:00+00"}
			]
		}`)
	})

	status := AutomationStatusEnabled
	limit := 5
	resp, err := client.Automations.ListWithOptions(context.Background(), &ListAutomationsOptions{
		Status: &status,
		Limit:  &limit,
	})
	if err != nil {
		t.Fatalf("Automations.ListWithOptions returned error: %v", err)
	}
	assert.Equal(t, true, resp.HasMore)
	assert.Equal(t, 1, len(resp.Data))
	assert.Equal(t, "aut_123", resp.Data[0].Id)
}

func TestUpdateAutomation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations/aut_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"object":"automation","id":"aut_123"}`)
	})

	resp, err := client.Automations.Update("aut_123", &UpdateAutomationRequest{
		Status: AutomationStatusEnabled,
	})
	if err != nil {
		t.Fatalf("Automations.Update returned error: %v", err)
	}
	assert.Equal(t, "automation", resp.Object)
	assert.Equal(t, "aut_123", resp.Id)
}

func TestRemoveAutomation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations/aut_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"object":"automation","id":"aut_123","deleted":true}`)
	})

	resp, err := client.Automations.Remove("aut_123")
	if err != nil {
		t.Fatalf("Automations.Remove returned error: %v", err)
	}
	assert.Equal(t, "automation", resp.Object)
	assert.Equal(t, "aut_123", resp.Id)
	assert.Equal(t, true, resp.Deleted)
}

func TestDuplicateAutomation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations/aut_123/duplicate", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"object":"automation","id":"aut_456"}`)
	})

	resp, err := client.Automations.Duplicate("aut_123")
	if err != nil {
		t.Fatalf("Automations.Duplicate returned error: %v", err)
	}
	assert.Equal(t, "automation", resp.Object)
	assert.Equal(t, "aut_456", resp.Id)
}

func TestStopAutomation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations/aut_123/stop", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"object":"automation","id":"aut_123","status":"disabled"}`)
	})

	resp, err := client.Automations.Stop("aut_123")
	if err != nil {
		t.Fatalf("Automations.Stop returned error: %v", err)
	}
	assert.Equal(t, "automation", resp.Object)
	assert.Equal(t, "aut_123", resp.Id)
	assert.Equal(t, "disabled", resp.Status)
}

func TestListAutomationRuns(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations/aut_123/runs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"object": "list",
			"has_more": false,
			"data": [
				{"id": "run_1", "status": "completed", "started_at": "2026-04-01 00:00:00+00", "completed_at": "2026-04-01 00:01:00+00", "created_at": "2026-04-01 00:00:00+00"},
				{"id": "run_2", "status": "running", "started_at": "2026-04-02 00:00:00+00", "completed_at": null, "created_at": "2026-04-02 00:00:00+00"}
			]
		}`)
	})

	resp, err := client.Automations.ListRuns("aut_123")
	if err != nil {
		t.Fatalf("Automations.ListRuns returned error: %v", err)
	}
	assert.Equal(t, "list", resp.Object)
	assert.Equal(t, false, resp.HasMore)
	assert.Equal(t, 2, len(resp.Data))
	assert.Equal(t, "run_1", resp.Data[0].Id)
	assert.Equal(t, AutomationRunStatusCompleted, resp.Data[0].Status)
	assert.NotNil(t, resp.Data[0].StartedAt)
	assert.NotNil(t, resp.Data[0].CompletedAt)
	assert.Equal(t, "run_2", resp.Data[1].Id)
	assert.Equal(t, AutomationRunStatusRunning, resp.Data[1].Status)
	assert.Nil(t, resp.Data[1].CompletedAt)
}

func TestListAutomationRunsWithOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations/aut_123/runs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "completed,failed", r.URL.Query().Get("status"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"object": "list",
			"has_more": false,
			"data": [
				{"id": "run_1", "status": "completed", "started_at": "2026-04-01 00:00:00+00", "completed_at": "2026-04-01 00:01:00+00", "created_at": "2026-04-01 00:00:00+00"}
			]
		}`)
	})

	limit := 10
	resp, err := client.Automations.ListRunsWithContext(context.Background(), "aut_123", &ListAutomationRunsOptions{
		Status: []AutomationRunStatus{AutomationRunStatusCompleted, AutomationRunStatusFailed},
		Limit:  &limit,
	})
	if err != nil {
		t.Fatalf("Automations.ListRunsWithContext returned error: %v", err)
	}
	assert.Equal(t, 1, len(resp.Data))
	assert.Equal(t, "run_1", resp.Data[0].Id)
}

func TestCreateAutomationWithDelayAndWaitForEvent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"object":"automation","id":"aut_456"}`)
	})

	resp, err := client.Automations.Create(&CreateAutomationRequest{
		Name: "Onboarding Flow",
		Steps: []AutomationStep{
			{Key: "trigger_1", Type: AutomationStepTypeTrigger, Config: map[string]any{
				"event_name": "user.created",
			}},
			{Key: "delay_1", Type: AutomationStepTypeDelay, Config: map[string]any{
				"duration": "30 minutes",
			}},
			{Key: "wait_1", Type: AutomationStepTypeWaitForEvent, Config: map[string]any{
				"event_name": "user.verified",
				"timeout":    "1 hour",
			}},
			{Key: "send_1", Type: AutomationStepTypeSendEmail, Config: map[string]any{
				"template": map[string]any{"id": "tpl_abc"},
			}},
		},
		Connections: []AutomationConnection{
			{From: "trigger_1", To: "delay_1"},
			{From: "delay_1", To: "wait_1"},
			{From: "wait_1", To: "send_1", Type: AutomationConnectionTypeEventReceived},
			{From: "wait_1", To: "send_1", Type: AutomationConnectionTypeTimeout},
		},
	})
	if err != nil {
		t.Fatalf("Automations.Create returned error: %v", err)
	}
	assert.Equal(t, "automation", resp.Object)
	assert.Equal(t, "aut_456", resp.Id)
}

func TestGetAutomationStepResponseKeys(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations/aut_456", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"object": "automation",
			"id": "aut_456",
			"name": "Onboarding Flow",
			"status": "disabled",
			"created_at": "2026-04-10 00:00:00+00",
			"updated_at": "2026-04-10 00:00:00+00",
			"steps": [
				{"key": "trigger_1", "type": "trigger", "config": {"event_name": "user.created"}},
				{"key": "delay_1", "type": "delay", "config": {"duration": "30 minutes"}},
				{"key": "wait_1", "type": "wait_for_event", "config": {"event_name": "user.verified", "timeout": "1 hour"}},
				{"key": "send_1", "type": "send_email", "config": {"template": {"id": "tpl_abc"}}}
			],
			"connections": [
				{"from": "trigger_1", "to": "delay_1"},
				{"from": "delay_1", "to": "wait_1"},
				{"from": "wait_1", "to": "send_1", "type": "event_received"},
				{"from": "wait_1", "to": "send_1", "type": "timeout"}
			]
		}`)
	})

	resp, err := client.Automations.Get("aut_456")
	if err != nil {
		t.Fatalf("Automations.Get returned error: %v", err)
	}
	assert.Equal(t, 4, len(resp.Steps))

	assert.Equal(t, "trigger_1", resp.Steps[0].Key)
	assert.Equal(t, AutomationStepTypeTrigger, resp.Steps[0].Type)

	assert.Equal(t, "delay_1", resp.Steps[1].Key)
	assert.Equal(t, AutomationStepTypeDelay, resp.Steps[1].Type)
	assert.Equal(t, "30 minutes", resp.Steps[1].Config["duration"])

	assert.Equal(t, "wait_1", resp.Steps[2].Key)
	assert.Equal(t, AutomationStepTypeWaitForEvent, resp.Steps[2].Type)
	assert.Equal(t, "user.verified", resp.Steps[2].Config["event_name"])
	assert.Equal(t, "1 hour", resp.Steps[2].Config["timeout"])

	assert.Equal(t, "send_1", resp.Steps[3].Key)
	assert.Equal(t, AutomationStepTypeSendEmail, resp.Steps[3].Type)

	assert.Equal(t, 4, len(resp.Connections))
	assert.Equal(t, AutomationConnectionTypeEventReceived, resp.Connections[2].Type)
	assert.Equal(t, AutomationConnectionTypeTimeout, resp.Connections[3].Type)
}

func TestGetAutomationRun(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automations/aut_123/runs/run_1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"object": "automation_run",
			"id": "run_1",
			"status": "completed",
			"started_at": "2026-04-01 00:00:00+00",
			"completed_at": "2026-04-01 00:01:00+00",
			"created_at": "2026-04-01 00:00:00+00",
			"steps": [
				{"key": "trigger_1", "type": "trigger", "status": "completed", "started_at": "2026-04-01 00:00:00+00", "completed_at": "2026-04-01 00:00:01+00", "output": null, "error": null, "created_at": "2026-04-01 00:00:00+00"},
				{"key": "send_1", "type": "send_email", "status": "completed", "started_at": "2026-04-01 00:00:01+00", "completed_at": "2026-04-01 00:01:00+00", "output": null, "error": null, "created_at": "2026-04-01 00:00:01+00"}
			]
		}`)
	})

	resp, err := client.Automations.GetRun("aut_123", "run_1")
	if err != nil {
		t.Fatalf("Automations.GetRun returned error: %v", err)
	}
	assert.Equal(t, "automation_run", resp.Object)
	assert.Equal(t, "run_1", resp.Id)
	assert.Equal(t, AutomationRunStatusCompleted, resp.Status)
	assert.NotNil(t, resp.StartedAt)
	assert.NotNil(t, resp.CompletedAt)
	assert.Equal(t, 2, len(resp.Steps))
	assert.Equal(t, "trigger_1", resp.Steps[0].Key)
	assert.Equal(t, AutomationStepTypeTrigger, resp.Steps[0].Type)
	assert.Equal(t, "completed", resp.Steps[0].Status)
	assert.Equal(t, "send_1", resp.Steps[1].Key)
	assert.Equal(t, AutomationStepTypeSendEmail, resp.Steps[1].Type)
}

func TestListAutomationTemplates(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automation-templates", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "", r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"object": "list",
			"categories": ["onboarding", "billing"],
			"data": [{
				"object": "automation_template",
				"slug": "email-verification-reminder",
				"version": 1,
				"name": "Email verification reminder",
				"category": "onboarding",
				"kind": "transactional",
				"outcome": "Remind users who haven't verified their email, at most twice.",
				"events": [
					{"ref": "trigger", "name": "user.created", "purpose": "Starts the workflow.", "exists": true},
					{"ref": "verified", "name": "user.email.verified", "purpose": "Ends the workflow.", "exists": false}
				],
				"email_count": 2,
				"steps": [{"key": "trigger", "type": "trigger"}, {"key": "wait_1", "type": "wait_for_event"}],
				"connections": [{"from": "trigger", "to": "wait_1", "type": "default"}],
				"in_use_count": 1
			}]
		}`)
	})

	resp, err := client.Automations.Templates.List()
	if err != nil {
		t.Fatalf("Automations.Templates.List returned error: %v", err)
	}
	assert.Equal(t, "list", resp.Object)
	assert.Equal(t, []string{"onboarding", "billing"}, resp.Categories)
	assert.Len(t, resp.Data, 1)
	tpl := resp.Data[0]
	assert.Equal(t, "email-verification-reminder", tpl.Slug)
	assert.Equal(t, 1, tpl.Version)
	assert.Equal(t, AutomationTemplateCategoryOnboarding, tpl.Category)
	assert.Equal(t, AutomationTemplateKindTransactional, tpl.Kind)
	assert.Equal(t, 2, tpl.EmailCount)
	assert.Equal(t, 1, tpl.InUseCount)
	assert.True(t, tpl.Events[0].Exists)
	assert.False(t, tpl.Events[1].Exists)
	assert.Equal(t, "user.email.verified", tpl.Events[1].Name)
	assert.Equal(t, AutomationStepTypeWaitForEvent, tpl.Steps[1].Type)
	assert.Equal(t, AutomationConnectionTypeDefault, tpl.Connections[0].Type)
}

func TestListAutomationTemplatesWithCategory(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automation-templates", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "category=lead_capture", r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"object":"list","categories":["lead_capture"],"data":[]}`)
	})

	category := AutomationTemplateCategoryLeadCapture
	resp, err := client.Automations.Templates.ListWithOptions(context.Background(), &ListAutomationTemplatesOptions{Category: &category})
	if err != nil {
		t.Fatalf("Automations.Templates.ListWithOptions returned error: %v", err)
	}
	assert.Empty(t, resp.Data)
}

func TestGetAutomationTemplate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/automation-templates/email-verification-reminder", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"object": "automation_template",
			"slug": "email-verification-reminder",
			"version": 2,
			"name": "Email verification reminder",
			"category": "onboarding",
			"kind": "transactional",
			"outcome": "Remind users.",
			"how_it_works": "Each signup starts its own run for that user.",
			"events": [{"ref": "verified", "name": "user.email.verified", "purpose": "Ends it.", "schema": {"user_id": "string"}, "exists": false, "adds": {"user_id": "string"}}],
			"email_count": 1,
			"emails": [{"ref": "reminder_1", "step_label": "Reminder 1", "step_key": "email_1", "subject": "Confirm your email", "preview": "One click", "html": "<p>Hi</p>", "text": "Hi", "variables": [{"key": "verify_url"}]}],
			"params": [
				{"ref": "wait_1", "type": "duration", "label": "First reminder after", "default": "24 hours"},
				{"ref": "resend_url", "type": "url", "label": "Link", "default": "", "placeholder": "https://app.example.com/verify-email", "hint": "Your page"}
			],
			"steps": [{"key": "trigger", "type": "trigger", "config": {"event_name": "user.created"}}],
			"connections": [],
			"settings": {"exit_on": "verified"},
			"in_use_count": 0
		}`)
	})

	resp, err := client.Automations.Templates.Get("email-verification-reminder")
	if err != nil {
		t.Fatalf("Automations.Templates.Get returned error: %v", err)
	}
	assert.Equal(t, "automation_template", resp.Object)
	assert.Equal(t, 2, resp.Version)
	assert.Equal(t, "Each signup starts its own run for that user.", resp.HowItWorks)
	assert.Equal(t, map[string]string{"user_id": "string"}, resp.Events[0].Schema)
	assert.Equal(t, map[string]string{"user_id": "string"}, resp.Events[0].Adds)
	assert.Equal(t, "email_1", resp.Emails[0].StepKey)
	assert.Equal(t, "Reminder 1", resp.Emails[0].StepLabel)
	assert.Equal(t, "<p>Hi</p>", resp.Emails[0].Html)
	assert.Equal(t, "verify_url", resp.Emails[0].Variables[0]["key"])
	assert.Equal(t, AutomationTemplateParamTypeDuration, resp.Params[0].Type)
	assert.Equal(t, "24 hours", resp.Params[0].Default)
	assert.Equal(t, AutomationTemplateParamTypeURL, resp.Params[1].Type)
	assert.Equal(t, "https://app.example.com/verify-email", resp.Params[1].Placeholder)
	assert.Equal(t, "user.created", resp.Steps[0].Config["event_name"])
	assert.Equal(t, "verified", resp.Settings["exit_on"])
}

const createFromTemplateResponse = `{
	"object": "automation",
	"id": "aut_123",
	"status": "disabled",
	"source_template": {"slug": "email-verification-reminder", "version": 2, "name": "Email verification reminder"},
	"emails": [{"object": "automation_email", "id": "tpl_1", "name": "Reminder 1", "status": "published", "step_key": "email_1"}],
	"events_created": ["user.email.verified"],
	"events_reused": ["user.created"]
}`

func TestCreateAutomationFromTemplate(t *testing.T) {
	setup()
	defer teardown()

	var sent map[string]any
	mux.HandleFunc("/automations/from-template", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		_, ok := r.Header["Idempotency-Key"]
		assert.False(t, ok, "expected no Idempotency-Key header")
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, createFromTemplateResponse)
	})

	resp, err := client.Automations.CreateFromTemplate(&CreateAutomationFromTemplateRequest{
		Slug:    "email-verification-reminder",
		Version: 2,
		Name:    "Verify reminders",
		Events: map[string]AutomationTemplateEventChoice{
			"verified": {Mode: AutomationTemplateEventModeExisting, Name: "account.confirmed"},
		},
		Sender:  AutomationTemplateSender{From: "Acme <hello@acme.com>", ReplyTo: "help@acme.com"},
		TopicId: "top_1",
		Params:  map[string]string{"wait_1": "24 hours", "resend_url": "https://app.acme.com/verify-email"},
	})
	if err != nil {
		t.Fatalf("Automations.CreateFromTemplate returned error: %v", err)
	}
	assert.Equal(t, map[string]any{
		"slug":     "email-verification-reminder",
		"version":  float64(2),
		"name":     "Verify reminders",
		"events":   map[string]any{"verified": map[string]any{"mode": "existing", "name": "account.confirmed"}},
		"sender":   map[string]any{"from": "Acme <hello@acme.com>", "reply_to": "help@acme.com"},
		"topic_id": "top_1",
		"params":   map[string]any{"wait_1": "24 hours", "resend_url": "https://app.acme.com/verify-email"},
	}, sent)
	assert.Equal(t, "aut_123", resp.Id)
	assert.Equal(t, AutomationStatusDisabled, resp.Status)
	assert.Equal(t, AutomationSourceTemplate{Slug: "email-verification-reminder", Version: 2, Name: "Email verification reminder"}, resp.SourceTemplate)
	assert.Equal(t, "email_1", resp.Emails[0].StepKey)
	assert.Equal(t, "published", resp.Emails[0].Status)
	assert.Equal(t, []string{"user.email.verified"}, resp.EventsCreated)
	assert.Equal(t, []string{"user.created"}, resp.EventsReused)
}

func TestCreateAutomationFromTemplateWithIdempotencyKey(t *testing.T) {
	setup()
	defer teardown()

	var sent map[string]any
	mux.HandleFunc("/automations/from-template", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "setup-verify-1", r.Header.Get("Idempotency-Key"))
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, createFromTemplateResponse)
	})

	resp, err := client.Automations.CreateFromTemplateWithOptions(context.Background(), &CreateAutomationFromTemplateRequest{
		Slug:   "email-verification-reminder",
		Sender: AutomationTemplateSender{From: "hello@acme.com"},
	}, &CreateAutomationFromTemplateOptions{IdempotencyKey: "setup-verify-1"})
	if err != nil {
		t.Fatalf("Automations.CreateFromTemplateWithOptions returned error: %v", err)
	}
	// Unset optional fields are left out of the body.
	assert.Equal(t, map[string]any{
		"slug":   "email-verification-reminder",
		"sender": map[string]any{"from": "hello@acme.com"},
	}, sent)
	assert.Equal(t, "aut_123", resp.Id)
}
