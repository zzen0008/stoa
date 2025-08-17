# LLM Gateway - Design Document (v2.1)

### **1. Overview & Core Principles**

This document outlines the architecture for a high-performance, reliable, and extensible LLM Gateway. It acts as a unified entry point for multiple downstream LLM providers, inspired by the robust design of systems like Bifrost.

**Core Principles:**

*   **Maintainability & Simplicity:** The highest priority. The code should be easy to understand, modify, and support. We will favor clarity over cleverness.
*   **Modularity:** The system will be composed of decoupled components with clear responsibilities, particularly the separation of core logic from the transport layer.
*   **Reliability & Fault Tolerance:** The gateway must be resilient to downstream provider failures through features like retries and fallbacks.
*   **Extensibility:** A clean, plugin-based architecture to allow for the easy addition of new features like authentication, logging, and rate limiting.

### **2. Implementation Philosophy**

To uphold our core principles, we will adhere to the following implementation guidelines:

1.  **Standard Library First:** We will use Go's standard libraries, specifically `net/http` for the web server and `encoding/json` for JSON processing. This ensures stability, portability, and zero learning curve for new contributors.
2.  **Minimal Dependencies:** Third-party libraries will be chosen carefully and only when they provide a significant advantage in terms of simplicity or modularity (e.g., a well-regarded router like `chi` or `gin-gonic`). We will avoid dependencies that offer pure performance micro-optimizations (e.g., `sonic`, `fasthttp`).
3.  **Lean Internal Schemas:** We will use internal, provider-agnostic structs to facilitate modularity. However, these schemas will be kept lean, containing only the fields essential for the gateway's logic (e.g., `model`, `stream`). Provider-specific parameters will be passed through more opaquely to avoid tight coupling.
4.  **Strategic Optimization:** Performance optimizations like `sync.Pool` will be used surgically, not aggressively. They will be applied only to clear, high-impact bottlenecks identified through profiling, such as the pooling of large buffers for request/response bodies.

### **3. High-Level Architecture**

The gateway will be architected with a clean separation between the **Core Engine** and the **Transport Layer**.

```mermaid
graph TB
    subgraph "Transport Layer"
        HTTP_API[HTTP Transport (net/http)]
    end

    subgraph "Core Engine (Go Package)"
        Router[Request Router]
        Middleware[Middleware/Plugin Pipeline]
        ProviderMgr[Provider Manager]
        ConfigMgr[Configuration Manager]
    end

    subgraph "Downstream Providers"
        Provider1[Provider A (e.g., OpenAI)]
        Provider2[Provider B (e.g., vLLM)]
        ProviderN[Provider N (e.g., Gemini)]
    end

    HTTP_API --> Router
    Router --> Middleware
    Middleware --> ProviderMgr
    ProviderMgr --> ConfigMgr

    ProviderMgr --> Provider1
    ProviderMgr --> Provider2
    ProviderMgr --> ProviderN
```

*   **Core Engine:** A self-contained Go package (`/internal/core`) with no knowledge of HTTP. It contains all business logic: routing, provider communication, fallbacks, and the plugin pipeline.
*   **Transport Layer:** An HTTP server (`/internal/transport`) built on the standard `net/http` library. It uses the Core Engine package to handle incoming requests.

### **4. Provider Management & Routing**

**Provider Isolation:**
Each downstream provider will be managed in isolation, with its own dedicated configuration and `http.Client`. This prevents a failure or slowdown in one provider from impacting others.

**Model Routing & Namespacing:**
Routing decisions will be made based on a namespaced model name from the request body (e.g., `"model": "openai/gpt-4o"`). The `provider` part dictates the routing, and the `model_name` is sent to the downstream API.

### **5. Configuration (`config.yaml`)**

The configuration will support provider-specific settings and fallback chains.

```yaml
server:
  host: "0.0.0.0"
  port: 8080

logging:
  level: "info"

# Defines routing strategies and provider groups
strategies:
  - name: "main_models"
    # Fallback chain: try openai, then vllm
    providers: ["openai", "vllm"] 

# List of all available downstream LLM providers
providers:
  - name: "openai"
    enabled: true
    target_url: "https://api.openai.com"
    api_key: "${OPENAI_API_KEY}"
    timeout: "30s"
    max_retries: 3

  - name: "vllm"
    enabled: true
    target_url: "http://localhost:8000"
    api_key: null
    timeout: "60s"
    max_retries: 1
```

### **6. API Endpoints**

**a) `POST /v1/chat/completions`**

*   **Function:** The primary intelligent routing endpoint.
*   **Request Flow:**
    1.  The `net/http` server receives the request.
    2.  The request body is parsed using the standard `encoding/json` library to extract the namespaced model.
    3.  The **Router** identifies the appropriate strategy and fallback chain.
    4.  The request enters the **Middleware Pipeline**.
    5.  The **Provider Manager** attempts to proxy the request to the first provider in the chain, rewriting the `model` field in the body to the simple `model_name`.
    6.  **On Failure:** The manager automatically retries with the next provider in the chain.
    7.  **On Success:** The streaming response is passed back to the client.

**b) `GET /v1/models`**

*   **Function:** Returns a unified list of all models from all *enabled* providers.
*   **Implementation:**
    1.  On startup and periodically, the gateway queries the `/v1/models` endpoint of each enabled provider.
    2.  Model IDs are automatically prefixed with their provider name (e.g., `id: "openai/gpt-4o"`).
    3.  The aggregated list is cached in memory for this endpoint to serve.

**c) `GET /v1/info`**

*   **Function:** A simple, unauthenticated health check endpoint.
*   **Response:** `{"status": "ok", "version": "0.1.0"}`

### **7. Extensibility: Middleware Pipeline**

We will use the standard `net/http` middleware pattern (chaining `http.Handler`s) to handle cross-cutting concerns. This is where future features will be implemented:
*   **Authentication (Keycloak)**
*   **Rate Limiting**
*   **Observability (ECS Logging)**

### **8. Project Structure**

```
/llm-gateway
|-- /cmd/gateway/main.go        # Application entry point
|-- /internal/
|   |-- core/                   # Core engine: routing, providers, business logic
|   |   |-- provider/           # Provider-specific clients and logic
|   |   |-- router/             # Request routing and strategy logic
|   |   |-- middleware/         # Core middleware logic (if any)
|   |-- transport/              # HTTP transport layer (net/http)
|   |   |-- handlers/           # HTTP handlers for API endpoints
|   |   |-- middleware/         # HTTP-specific middleware (logging, auth)
|   |-- config/                 # Configuration loading and parsing
|-- /pkg/                       # Shared libraries (if any)
|-- go.mod
|-- go.sum
|-- config.example.yaml
|-- design.md
```

### **9. Next Steps (Implementation Plan)**

1.  **Setup Project Structure:** Create the directories and initialize the Go module.
2.  **Implement Configuration:** Create structs and loading logic for `config.yaml`.
3.  **Implement Provider Manager:** Build logic to manage `http.Client`s for each provider.
4.  **Implement `/v1/models` Aggregation:** Implement fetching and caching of the namespaced model list.
5.  **Implement Core Router:** Build the routing logic that selects the provider strategy.
6.  **Implement `/v1/chat/completions`:** Wire up the core routing to the reverse proxy and fallback logic.
7.  **Build Transport Layer:** Create the `net/http` server and handlers.
8.  **Write `main.go`:** Tie everything together.
