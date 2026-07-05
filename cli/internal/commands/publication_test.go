package commands

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicationEventsCommandPrintsLifecycleEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/publications/pub_1/events" {
			t.Fatalf("path = %s, want /api/v1/publications/pub_1/events", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":"evt_1","publication_id":"pub_1","rendition_id":"rend_1","type":"published","status":"succeeded","message":"rendition published","metadata":{"platform":"x"},"created_at":"2026-07-04T21:00:00Z"}
		]`))
	}))
	defer srv.Close()

	out, err := executeRootCaptureStdout(t, "--instance", srv.URL, "--token", "op_cli_test", "publication", "events", "pub_1")

	if err != nil {
		t.Fatalf("publication events returned error: %v", err)
	}
	for _, want := range []string{"published", "succeeded", "rendition published", "rend_1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output %q missing %q", out, want)
		}
	}
}

func TestPublicationCommentsCommandHidesUnsupportedActions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/renditions/rend_1/comments" {
			t.Fatalf("path = %s, want /api/v1/renditions/rend_1/comments", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"comments":[
			{"id":"comment_1","rendition_id":"rend_1","provider_comment_id":"provider_1","author_name":"Rita","text":"Nice launch","hidden":false,"can_reply":true,"can_hide":false,"can_delete":false}
		]}`))
	}))
	defer srv.Close()

	out, err := executeRootCaptureStdout(t, "--instance", srv.URL, "--token", "op_cli_test", "publication", "comments", "rend_1")

	if err != nil {
		t.Fatalf("publication comments returned error: %v", err)
	}
	if !strings.Contains(out, "Nice launch") || !strings.Contains(out, "reply") {
		t.Fatalf("output %q missing comment text or supported action", out)
	}
	if strings.Contains(out, "hide") || strings.Contains(out, "delete") {
		t.Fatalf("output %q should not include unsupported actions", out)
	}
}
