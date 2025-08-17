package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMaxTokens = 100
	loremIpsum       = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
)

// ModelResponse defines the structure for the /v1/models endpoint
type ModelResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// Model defines the structure of a single model
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ChatCompletionRequest defines the structure for chat completion requests
type ChatCompletionRequest struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Stream    bool   `json:"stream"`
}

// ChatCompletionResponse defines the structure for non-streaming chat completion responses
type ChatCompletionResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   UsageData `json:"usage"`
}

// Choice defines a single choice in a chat completion response
type Choice struct {
	Index        int         `json:"index"`
	Message      Message     `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Message defines the content of a message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// UsageData defines the token usage statistics
type UsageData struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChoice defines a single choice in a streaming response
type StreamChoice struct {
	Index int         `json:"index"`
	Delta StreamDelta `json:"delta"`
}

// StreamDelta defines the content delta in a streaming response
type StreamDelta struct {
	Content string `json:"content"`
}

// StreamCompletionResponse defines the structure for streaming chat completion responses
type StreamCompletionResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// generateLorem generates a lorem ipsum string of a specific token count
func generateLorem(tokens int) string {
	words := strings.Fields(loremIpsum)
	if tokens > len(words) {
		// Repeat the words if requested tokens are more than the base text
		repeatedWords := make([]string, 0, tokens)
		for len(repeatedWords) < tokens {
			repeatedWords = append(repeatedWords, words...)
		}
		return strings.Join(repeatedWords[:tokens], " ")
	}
	return strings.Join(words[:tokens], " ")
}

// modelsHandler handles requests to the /v1/models endpoint
func modelsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request for %s from %s", r.URL.Path, r.RemoteAddr)
	models := ModelResponse{
		Object: "list",
		Data: []Model{
			{
				ID:      "mock-gpt-4",
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "mock-owner",
			},
			{
				ID:      "mock-gpt-3.5-turbo",
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "mock-owner",
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

// chatCompletionsHandler handles requests to the /v1/chat/completions endpoint
func chatCompletionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request for %s from %s", r.URL.Path, r.RemoteAddr)

	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("Request details: Model=%s, MaxTokens=%d, Stream=%v", req.Model, req.MaxTokens, req.Stream)

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	if req.Stream {
		streamResponse(w, req.Model, maxTokens)
	} else {
		nonStreamResponse(w, req.Model, maxTokens)
	}
}

// nonStreamResponse sends a complete chat completion response
func nonStreamResponse(w http.ResponseWriter, model string, maxTokens int) {
	content := generateLorem(maxTokens)
	response := ChatCompletionResponse{
		ID:      "chatcmpl-" + strconv.FormatInt(time.Now().Unix(), 10),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: UsageData{
			PromptTokens:     10, // Dummy value
			CompletionTokens: maxTokens,
			TotalTokens:      maxTokens + 10,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// streamResponse sends a streaming chat completion response
func streamResponse(w http.ResponseWriter, model string, maxTokens int) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	words := strings.Fields(generateLorem(maxTokens))
	for _, word := range words {
		response := StreamCompletionResponse{
			ID:      "chatcmpl-" + strconv.FormatInt(time.Now().UnixNano(), 10),
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   model,
			Choices: []StreamChoice{
				{
					Index: 0,
					Delta: StreamDelta{
						Content: word + " ",
					},
				},
			},
		}

		jsonBytes, _ := json.Marshal(response)
		fmt.Fprintf(w, "data: %s\n\n", jsonBytes)
		flusher.Flush()
		time.Sleep(50 * time.Millisecond) // Simulate delay
	}

	// Send the final done message
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", modelsHandler)
	mux.HandleFunc("/v1/chat/completions", chatCompletionsHandler)

	port := "8081"
	log.Printf("Starting mock LLM server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
