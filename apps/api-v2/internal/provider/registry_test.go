package provider_test

import (
	"net/http"
	"testing"

	"github.com/anarchis12/nodalibre/v2/internal/provider"
	"github.com/anarchis12/nodalibre/v2/internal/provider/mistral"
	"github.com/anarchis12/nodalibre/v2/internal/provider/openai"
)

func TestRegistryKeepsProvidersIndependent(t *testing.T) {
	r := provider.NewRegistry()
	if err := r.Register(mistral.New(http.DefaultClient)); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(openai.NewChatGPT(http.DefaultClient)); err != nil {
		t.Fatal(err)
	}
	ids := r.IDs()
	if len(ids) != 2 || ids[0] != "chatgpt" || ids[1] != "mistral" {
		t.Fatalf("unexpected providers: %#v", ids)
	}
	if _, ok := r.Get("chatgpt"); !ok {
		t.Fatal("ChatGPT provider was not registered")
	}
}
