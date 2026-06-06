package backend

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/amithathreya/inference_layer/api"
)

type LlamaBackend struct {
	TargetURL string
	Client    *http.Client
}

func NewLlamaBackend(targetURL string) *LlamaBackend {
	return &LlamaBackend{
		TargetURL: targetURL,
		Client:    &http.Client{},
	}
}

// ForwardChat forwards the incoming request structure to the native llama.cpp server.
func (b *LlamaBackend) ForwardChat(req api.ChatRequest) (*http.Response, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	endpoint := b.TargetURL + "/v1/chat/completions"
	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	return b.Client.Do(httpReq)
}
