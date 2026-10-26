package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anarchis12/nodalibre/v2/internal/provider"
)

type ChatGPT struct {
	BaseURL string
	HTTP    *http.Client
}

func NewChatGPT(httpClient *http.Client) *ChatGPT {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ChatGPT{BaseURL: "https://api.openai.com/v1", HTTP: httpClient}
}

func (c *ChatGPT) ID() string { return "chatgpt" }

func (c *ChatGPT) ListModels(ctx context.Context, account provider.Account) ([]provider.Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+"/models", nil)
	if err != nil {
		return nil, err
	}
	setHeaders(req, account.AccessToken)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list ChatGPT models: %w", err)
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var payload struct {
		Models []struct {
			Slug        string `json:"slug"`
			DisplayName string `json:"display_name"`
			Visibility  string `json:"visibility"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode ChatGPT models: %w", err)
	}
	models := make([]provider.Model, 0, len(payload.Models))
	for _, item := range payload.Models {
		if item.Visibility != "list" || item.Slug == "" {
			continue
		}
		name := item.DisplayName
		if name == "" {
			name = item.Slug
		}
		models = append(models, provider.Model{ID: item.Slug, DisplayName: name, Capabilities: []provider.Capability{provider.CapabilityText}})
	}
	return models, nil
}

func (c *ChatGPT) Stream(ctx context.Context, account provider.Account, chat provider.ChatRequest, sink provider.EventSink) error {
	instructions := strings.TrimSpace(chat.Instructions)
	input := make([]provider.Message, 0, len(chat.Messages))
	for _, message := range chat.Messages {
		// ChatGPT plan usage rejects explicit system input items. Preserve their
		// content through the supported top-level instructions field instead.
		if message.Role == "system" {
			if instructions != "" {
				instructions += "\n\n"
			}
			instructions += message.Content
			continue
		}
		input = append(input, message)
	}
	payload := struct {
		Model        string             `json:"model"`
		Input        []provider.Message `json:"input"`
		Instructions string             `json:"instructions,omitempty"`
		Store        bool               `json:"store"`
		Stream       bool               `json:"stream"`
	}{Model: chat.Model, Input: input, Instructions: instructions, Store: false, Stream: true}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/responses", bytes.NewReader(body))
	if err != nil {
		return err
	}
	setHeaders(req, account.AccessToken)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("stream ChatGPT response: %w", err)
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return err
	}
	completed := false
	if err := provider.ReadSSE(resp.Body, func(data []byte) error {
		var event struct {
			Type     string `json:"type"`
			Delta    string `json:"delta"`
			Response struct {
				Error *struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			} `json:"response"`
		}
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		switch event.Type {
		case "response.output_text.delta":
			return sink(provider.Event{Type: provider.EventTextDelta, Text: event.Delta})
		case "response.completed":
			completed = true
			if event.Response.Usage.InputTokens != 0 || event.Response.Usage.OutputTokens != 0 {
				return sink(provider.Event{Type: provider.EventUsage, InputTokens: event.Response.Usage.InputTokens, OutputTokens: event.Response.Usage.OutputTokens})
			}
		case "response.failed":
			if event.Response.Error != nil {
				return fmt.Errorf("ChatGPT response failed (%s): %s", event.Response.Error.Code, event.Response.Error.Message)
			}
			return fmt.Errorf("ChatGPT response failed")
		case "response.incomplete":
			return fmt.Errorf("ChatGPT response was incomplete")
		}
		return nil
	}); err != nil {
		return err
	}
	if !completed {
		return fmt.Errorf("ChatGPT stream ended without response.completed")
	}
	return sink(provider.Event{Type: provider.EventCompleted})
}

func setHeaders(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	return fmt.Errorf("OpenAI API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
}
