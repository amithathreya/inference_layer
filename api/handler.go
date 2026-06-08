package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)


type BackendInterface interface {
	ForwardChat(req ChatRequest) (*http.Response, error)
}

type Handler struct {
	Backend BackendInterface
}

func NewHandler(b BackendInterface) *Handler {
	return &Handler{Backend: b}
}

func (h *Handler) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	resp, err := h.Backend.ForwardChat(req)
	if err != nil {
		http.Error(w, "Failed to contact model backend", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if req.Stream {
		h.handleStreaming(w, resp)
	} else {
		h.handleBlocking(w, resp)
	}
}

func (h *Handler) handleBlocking(w http.ResponseWriter, resp *http.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		http.Error(w, "Failed to parse backend response", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(chatResp)
}

func (h *Handler) handleStreaming(w http.ResponseWriter, resp *http.Response) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(resp.StatusCode)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported by network stack", http.StatusInternalServerError)
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Pass each raw SSE chunk line directly back to our client client
		fmt.Fprintf(w, "%s\n\n", line)
		flusher.Flush()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error encountered reading stream buffer: %v", err)
	}
}
