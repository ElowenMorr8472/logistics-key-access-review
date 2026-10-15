package logistics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReviewUsesOneCredentialForKeyAuditAndPDF(t *testing.T) {
	var calls int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("missing shared bearer credential")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/account/keys/list":
			if r.Method != http.MethodGet {
				t.Fatalf("keys method = %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"ok":true,"data":{"items":[{"key_id":"key_1","name":"dispatch","scopes":["logistics.shipment_events.read","pdf.generate"]}]},"error":null,"metadata":{}}`))
		case "/v1/pdf/generate":
			if r.Method != http.MethodPost || r.Header.Get("Idempotency-Key") == "" {
				t.Fatalf("pdf request lacks method or idempotency key")
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(body["markdown"].(string), "shipment-42") {
				t.Fatalf("review does not identify shipment")
			}
			_, _ = w.Write([]byte(`{"ok":true,"data":{"document_id":"doc_1"},"error":null,"metadata":{}}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodPost, "/reviews", strings.NewReader(`{"shipment_id":"shipment-42","events":[{"type":"delivered","occurred_at":"2026-10-01T08:00:00Z"}],"pod_files":[{"name":"signed.pdf","media_type":"application/pdf"}],"exceptions":[]}`))
	recorder := httptest.NewRecorder()
	NewReviewHandler(NewClientWithBaseURL("test-key", upstream.URL)).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated || calls != 2 {
		t.Fatalf("status = %d, calls = %d, body = %s", recorder.Code, calls, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"proof-of-delivery files"`) {
		t.Fatalf("expected missing POD access in %s", recorder.Body.String())
	}
}
