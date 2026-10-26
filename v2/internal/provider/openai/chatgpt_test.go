package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/anarchis12/nodalibre/v2/internal/provider"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func TestChatGPTListsOnlyVisibleModels(t *testing.T) {
	httpClient := testClient(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "Bearer oauth-token" {
			t.Fatalf("unexpected authorization header: %q", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"models":[{"slug":"gpt-visible","display_name":"GPT Visible","visibility":"list"},{"slug":"gpt-hidden","visibility":"hidden"}]}`)),
		}, nil
	})
	c := NewChatGPT(httpClient)
	models, err := c.ListModels(context.Background(), provider.Account{AccessToken: "oauth-token"})
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 || models[0].ID != "gpt-visible" {
		t.Fatalf("unexpected models: %#v", models)
	}
}

func TestChatGPTRequiresCompletedEvent(t *testing.T) {
	httpClient := testClient(func(r *http.Request) (*http.Response, error) {
		var payload struct {
			Input        []provider.Message `json:"input"`
			Instructions string             `json:"instructions"`
			Store        bool               `json:"store"`
			Stream       bool               `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Store || !payload.Stream {
			t.Fatalf("ChatGPT plan requests require store=false and stream=true: %#v", payload)
		}
		if len(payload.Input) != 1 || payload.Input[0].Role != "user" || payload.Instructions != "Réponds brièvement." {
			t.Fatalf("system message was not normalized: %#v", payload)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader("data: {\"type\":\"response.output_text.delta\",\"delta\":\"bonjour\"}\n\n")),
		}, nil
	})
	c := NewChatGPT(httpClient)
	err := c.Stream(context.Background(), provider.Account{AccessToken: "oauth-token"}, provider.ChatRequest{Model: "gpt-visible", Messages: []provider.Message{{Role: "system", Content: "Réponds brièvement."}, {Role: "user", Content: "Salut"}}}, func(provider.Event) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "without response.completed") {
		t.Fatalf("expected missing completed error, got %v", err)
	}
}
