package provider

import "context"

type Capability string

const (
	CapabilityText       Capability = "text"
	CapabilityVision     Capability = "vision"
	CapabilityFiles      Capability = "files"
	CapabilityWebSearch  Capability = "web_search"
	CapabilityImage      Capability = "image_generation"
	CapabilitySpeech     Capability = "speech"
	CapabilityTranscribe Capability = "transcription"
	CapabilityOCR        Capability = "ocr"
)

type Account struct {
	ID          string
	AccessToken string
}

type Model struct {
	ID           string       `json:"id"`
	DisplayName  string       `json:"display_name"`
	Capabilities []Capability `json:"capabilities,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model        string
	Messages     []Message
	Instructions string
}

type EventType string

const (
	EventTextDelta EventType = "text_delta"
	EventUsage     EventType = "usage"
	EventCompleted EventType = "completed"
)

type Event struct {
	Type         EventType
	Text         string
	InputTokens  int
	OutputTokens int
}

type EventSink func(Event) error

type ChatProvider interface {
	ID() string
	ListModels(context.Context, Account) ([]Model, error)
	Stream(context.Context, Account, ChatRequest, EventSink) error
}
