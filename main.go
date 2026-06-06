package main

import (
	"log"
	"net/http"

	"github.com/amithathreya/inference_layer/api"
	"github.com/amithathreya/inference_layer/backend"
)

func main() {
	llamaBackend := backend.NewLlamaBackend("http://localhost:8080")
	serverHandler := api.NewHandler(llamaBackend)

	// Expose standard path layouts
	http.HandleFunc("/v1/chat/completions", serverHandler.HandleChatCompletions)

	log.Println("Go Inference Gateway listening on :9090...")
	if err := http.ListenAndServe(":9090", nil); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
