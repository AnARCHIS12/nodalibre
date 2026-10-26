package mistral

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

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{BaseURL: "https://api.mistral.ai/v1", HTTP: httpClient}
}

func (c *Client) ID() string { return "mistral" }

func (c *Client) ListModels(ctx context.Context, account provider.Account) ([]provider.Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+"/models", nil)
	if err != nil {
		return nil, err
	}
	setHeaders(req, account.AccessToken)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list Mistral models: %w", err)
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode Mistral models: %w", err)
	}
	models := make([]provider.Model, 0, len(payload.Data))
	for _, item := range payload.Data {
		models = append(models, provider.Model{ID: item.ID, DisplayName: item.ID, Capabilities: capabilities(item.ID)})
	}
	return models, nil
}

func (c *Client) Stream(ctx context.Context, account provider.Account, chat provider.ChatRequest, sink provider.EventSink) error {
	payload := struct {
		Model    string             `json:"model"`
		Messages []provider.Message `json:"messages"`
		Stream   bool               `json:"stream"`
	}{Model: chat.Model, Messages: chat.Messages, Stream: true}
	if chat.Instructions != "" {
		payload.Messages = append([]provider.Message{{Role: "system", Content: chat.Instructions}}, payload.Messages...)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	setHeaders(req, account.AccessToken)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("stream Mistral response: %w", err)
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return err
	}
	if err := provider.ReadSSE(resp.Body, func(data []byte) error {
		var event struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		for _, choice := range event.Choices {
			if choice.Delta.Content != "" {
				if err := sink(provider.Event{Type: provider.EventTextDelta, Text: choice.Delta.Content}); err != nil {
					return err
				}
			}
		}
		if event.Usage.PromptTokens != 0 || event.Usage.CompletionTokens != 0 {
			return sink(provider.Event{Type: provider.EventUsage, InputTokens: event.Usage.PromptTokens, OutputTokens: event.Usage.CompletionTokens})
		}
		return nil
	}); err != nil {
		return err
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
	return fmt.Errorf("Mistral API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
}

func capabilities(model string) []provider.Capability {
	caps := []provider.Capability{provider.CapabilityText}
	if strings.Contains(model, "mistral-large") || strings.Contains(model, "mistral-medium") || strings.Contains(model, "mistral-small") || strings.Contains(model, "ministral") {
		caps = append(caps, provider.CapabilityVision, provider.CapabilityFiles)
	}
	return caps
}
