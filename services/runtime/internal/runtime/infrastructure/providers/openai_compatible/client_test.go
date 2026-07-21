package openai_compatible

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

func TestClient_Execute_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("auth: %s", got)
		}
		body, _ := io.ReadAll(r.Body)
		var req chatCompletionRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "gpt-4o-mini" || req.Stream {
			t.Fatalf("req: %+v", req)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"choices":[{"message":{"role":"assistant","content":"{\"status\":\"ok\"}"}}]
		}`))
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL+"/v1", "test-key", srv.Client())
	if err != nil {
		t.Fatal(err)
	}

	req, err := entities.NewRequest("gpt-4o-mini", []entities.Message{
		{Role: "user", Content: "Return JSON"},
	}, false)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Succeeded() {
		t.Fatalf("resp: %+v", resp)
	}
	if resp.Body != `{"status":"ok"}` {
		t.Fatalf("body: %q", resp.Body)
	}
	if resp.LatencyMs < 0 {
		t.Fatalf("latency: %d", resp.LatencyMs)
	}
}

func TestClient_Execute_parsesToolCallsAndUsage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"choices":[{
				"message":{
					"role":"assistant",
					"content":"",
					"tool_calls":[
						{"id":"c1","type":"function","function":{"name":"search","arguments":"{}"}}
					]
				}
			}],
			"usage":{"prompt_tokens":12,"completion_tokens":3,"total_tokens":15}
		}`))
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL+"/v1", "k", srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	req, err := entities.NewRequest("m", []entities.Message{{Role: "user", Content: "x"}}, false)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Succeeded() {
		t.Fatalf("resp: %+v", resp)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Name != "search" {
		t.Fatalf("tool_calls: %+v", resp.ToolCalls)
	}
	if resp.PromptTokens != 12 || resp.CompletionTokens != 3 {
		t.Fatalf("tokens: prompt=%d completion=%d", resp.PromptTokens, resp.CompletionTokens)
	}
}

func TestClient_Execute_503_retryable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL+"/v1", "k", srv.Client())
	req, _ := entities.NewRequest("m", []entities.Message{{Role: "user", Content: "x"}}, false)

	resp, err := client.Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 503 || resp.Succeeded() {
		t.Fatalf("resp: %+v", resp)
	}
}

func TestClient_Execute_400_not_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL+"/v1", "k", srv.Client())
	req, _ := entities.NewRequest("m", []entities.Message{{Role: "user", Content: "x"}}, false)

	resp, err := client.Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Succeeded() || resp.HTTPStatus != 400 {
		t.Fatalf("resp: %+v", resp)
	}
}

func TestClient_Execute_rejects_stream(t *testing.T) {
	client, _ := NewClient("http://example.com/v1", "k", nil)
	req, _ := entities.NewRequest("m", []entities.Message{{Role: "user", Content: "x"}}, true)

	_, err := client.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("expected stream error")
	}
}
