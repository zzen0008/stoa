# 6. Schema Design

This document outlines the design for our gateway's internal schemas. The design prioritizes **simplicity, maintainability, and loose coupling** over attempting to create a comprehensive "superset" of all possible provider features.

### Design Philosophy

1.  **Lean & Focused:** Our internal structs will only contain fields that the gateway's core logic *actually needs* to function (e.g., the model name for routing).
2.  **Passthrough for Flexibility:** All other provider-specific parameters (`temperature`, `top_p`, `max_tokens`, etc.) will be passed through to the downstream API without being formally defined in our core structs. This prevents our gateway from needing code changes every time a provider adds a new feature.
3.  **Standard Library Only:** We will rely only on Go's standard `encoding/json` library. We will avoid complex custom marshaling logic for "union types" (fields that can be one of multiple types).
4.  **Separate Schemas per Concern:** We will define separate, explicit structs for each type of request (e.g., `ChatRequest`). We will not use a single, generic input struct.

---

### Proposed Go Schemas

The following Go structs will serve as our internal, provider-agnostic "common language".

```go
// In file: internal/core/schema.go

package core

// ChatRequest is our gateway's internal, provider-agnostic representation
// of an incoming chat completion request. It contains only the fields
// that the gateway's core logic needs to function.
type ChatRequest struct {
	// The namespaced model name (e.g., "openai/gpt-4o").
	// The router uses this to select the correct provider.
	Model string `json:"model"`

	// The list of messages in the conversation.
	Messages []Message `json:"messages"`

	// A boolean indicating if the request is for a streaming response.
	// The proxy needs this to handle the response correctly.
	Stream bool `json:"stream"`

	// --- Passthrough Fields ---
	// The following fields are not directly used by the gateway's core logic,
	// but they need to be passed through to the downstream provider.
	// We use a map to avoid tightly coupling our schema to every possible
	// provider parameter. This makes our gateway more maintainable.

	// Optional parameters like temperature, top_p, max_tokens, etc.
	// These will be sent directly to the downstream provider.
	Params map[string]any `json:"params,omitempty"`

	// Optional tools for the model to use.
	Tools []any `json:"tools,omitempty"`
}

// Message represents a single message in a chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	// We are deliberately keeping this simple. We will not support
	// complex content types (like images or tool calls in the content block)
	// in our initial design to maintain simplicity.
}

// ChatResponse is our gateway's internal representation of a response.
// For now, this struct is minimal. The proxy will primarily stream the raw
// response body back to the client. We will add fields here later as
// needed for logging and metrics (e.g., Usage).
type ChatResponse struct {
	// We can add fields here later, such as:
	// ID      string  `json:"id"`
	// Model   string  `json:"model"`
	// Usage   Usage   `json:"usage"`
}

// Usage represents the token usage for a request.
// This is useful for logging and metrics.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
```

### Rationale for Key Decisions

| Decision | Why It Aligns With Our Goals |
| :--- | :--- |
| **Lean `ChatRequest` Struct** | **(Simplicity)** The struct is small and easy to understand. It only contains what's necessary for the gateway's logic. |
| **`Params` map for Passthrough** | **(Maintainability & Modularity)** The gateway is decoupled from provider-specific features. When OpenAI adds a new parameter, we don't need to update and redeploy the gateway. It just works. |
| **Simple `Message.Content`** | **(Simplicity)** Avoids complex and error-prone custom JSON handling for multi-part content, focusing on the 95% use case of text-only chat. |
| **Minimal `ChatResponse`** | **(Simplicity)** Follows the YAGNI ("You Ain't Gonna Need It") principle. We avoid parsing the entire response into a complex struct when the primary goal is to stream it. We will only add fields as required by other features. |
