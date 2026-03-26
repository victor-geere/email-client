package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func staticToken(ctx context.Context) (string, error) {
	return "test-token", nil
}

func TestResolveFolderID_FindsInbox(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"value": []map[string]string{
				{"id": "folder-1", "displayName": "Drafts"},
				{"id": "folder-2", "displayName": "Inbox"},
				{"id": "folder-3", "displayName": "Sent Items"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, staticToken)
	client.HTTPClient = server.Client()

	id, err := client.ResolveFolderID(context.Background(), "Inbox")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "folder-2" {
		t.Errorf("expected folder-2, got %s", id)
	}
}

func TestResolveFolderID_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"value": []map[string]string{
				{"id": "folder-1", "displayName": "Inbox"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, staticToken)
	client.HTTPClient = server.Client()

	_, err := client.ResolveFolderID(context.Background(), "Archive")
	if err == nil {
		t.Fatal("expected error for missing folder")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func makeGraphMessage(id, convID, subject, fromName, fromAddr, body string) map[string]interface{} {
	return map[string]interface{}{
		"id":               id,
		"conversationId":   convID,
		"subject":          subject,
		"receivedDateTime": "2025-01-14T14:00:00Z",
		"from": map[string]interface{}{
			"emailAddress": map[string]string{
				"name":    fromName,
				"address": fromAddr,
			},
		},
		"toRecipients": []map[string]interface{}{
			{"emailAddress": map[string]string{"name": "Recipient", "address": "to@example.com"}},
		},
		"ccRecipients": []interface{}{},
		"body": map[string]string{
			"contentType": "text",
			"content":     body,
		},
		"bodyPreview": body[:min(len(body), 50)],
	}
}

func TestFetchMessages_PaginatesCorrectly(t *testing.T) {
	callCount := 0
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"value": []interface{}{
					makeGraphMessage("msg-1", "conv-1", "Hello", "Alice", "alice@example.com", "First message"),
				},
				"@odata.nextLink": serverURL + "/page2",
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"value": []interface{}{
				makeGraphMessage("msg-2", "conv-1", "Re: Hello", "Bob", "bob@example.com", "Second message"),
			},
		})
	}))
	defer server.Close()
	serverURL = server.URL

	client := NewClient(server.URL, staticToken)
	client.HTTPClient = server.Client()

	msgs, err := client.FetchMessages(context.Background(), "inbox", FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].ID != "msg-1" || msgs[1].ID != "msg-2" {
		t.Errorf("unexpected message IDs: %s, %s", msgs[0].ID, msgs[1].ID)
	}
}

func TestFetchMessages_Retries429(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"value": []interface{}{
				makeGraphMessage("msg-1", "conv-1", "Hello", "Alice", "alice@example.com", "Body"),
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, staticToken)
	client.HTTPClient = server.Client()

	msgs, err := client.FetchMessages(context.Background(), "inbox", FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

func TestFetchMessages_Handles401WithRefresh(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"value": []interface{}{
				makeGraphMessage("msg-1", "conv-1", "Hello", "Alice", "alice@example.com", "Body"),
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, staticToken)
	client.HTTPClient = server.Client()

	msgs, err := client.FetchMessages(context.Background(), "inbox", FetchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
