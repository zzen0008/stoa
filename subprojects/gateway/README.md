# LLM Gateway

This project is a high-performance, extensible LLM Gateway designed to act as a unified entry point for multiple downstream LLM providers. It provides intelligent routing, provider fallbacks, and a unified API, simplifying the process of interacting with various language models.

## How It Works

The gateway is built with a clean separation between its core logic and the transport layer, following the principles outlined in the main `design.md`.

### Core Concepts

1.  **Namespaced Models**: The gateway uses a namespaced model format (`provider_name/model_name`) in API requests. This tells the gateway which provider to route the request to. For example, a request for `"model": "openai/gpt-4o"` will be routed to the `openai` provider.

2.  **Configuration-Driven Routing**: All routing logic, provider details, and fallback strategies are defined in a central `config.yaml` file. This makes it easy to add new providers or change routing behavior without modifying the code.

3.  **Provider Fallbacks**: The `config.yaml` allows you to define strategies with fallback chains. If a request to the primary provider fails, the gateway will automatically retry the request with the next provider in the chain.

4.  **Unified Model List**: The gateway exposes a `/v1/models` endpoint that returns a single, aggregated list of all available models from all enabled providers. It automatically fetches this information in the background and prefixes the model IDs with their provider name (e.g., `openai/gpt-4o`).

## Getting Started

### Prerequisites

*   Go (version 1.21 or later)
*   A configured downstream LLM provider (e.g., OpenAI, Llama.cpp server)

### 1. Configuration

Before running the application, you need to set up your configuration file.

1.  Copy the example configuration:
    ```bash
    cp config.example.yaml config.yaml
    ```
2.  Edit `config.yaml` to add your provider details, including target URLs and API keys. API keys can be loaded from environment variables (e.g., `api_key: "${OPENAI_API_KEY}"`).

### 2. Build and Run

A `Makefile` is provided to simplify the build and run process.

*   **To build the application:**
    ```bash
    make build
    ```
    This will compile the source code and create an executable at `./cmd/gateway/gateway`.

*   **To run the application:**
    ```bash
    make run
    ```
    This will first build the application and then start the server. The server will run in the foreground and listen on the host and port specified in your `config.yaml`.

*   **To clean up build artifacts:**
    ```bash
    make clean
    ```

## API Usage

The gateway exposes an OpenAI-compatible API.

### Example Requests

Example `curl` requests are available in the `/resources/scripts/` directory at the root of the main project.

*   **Non-Streaming Request:**
    ```bash
    ./resources/scripts/non_streaming_request.sh
    ```

*   **Streaming Request:**
    ```bash
    ./resources/scripts/streaming_request.sh
    ```

## Middleware

The gateway features a two-part middleware system designed for extensibility, allowing for custom logic to be executed both before and after a request is proxied to a downstream provider. This design explicitly supports streaming responses.

### 1. Transport Middleware (Pre-Forwarding)

This middleware acts on the initial incoming HTTP request before any core gateway logic is executed. It uses the standard Go `func(http.Handler) http.Handler` pattern.

*   **Purpose**: Ideal for cross-cutting concerns that don't depend on the downstream provider's response.
    *   Request Logging
    *   Authentication & Authorization
    *   Rate Limiting
    *   Request Header Manipulation
*   **Location**: `internal/transport/middleware/`
*   **How it Works**: A chain of middleware is applied to the main HTTP router in `cmd/gateway/main.go`. Each middleware in the chain wraps the next, allowing them to execute logic before and after the request is handled.

### 2. Core Middleware (Post-Forwarding & Streaming-Aware)

This middleware acts on the HTTP response received from the downstream provider *before* it is streamed back to the original client. It is specifically designed to handle streaming bodies correctly.

*   **Purpose**: Ideal for logic that needs to inspect, modify, or act upon the provider's response.
    *   Logging the final completion payload to a service like Elasticsearch.
    *   Modifying response headers.
    *   Transforming the response body (e.g., redacting sensitive information).
*   **Location**: `internal/core/middleware/`
*   **How it Works**:
    1.  The core proxy's `ModifyResponse` hook is used to trigger the middleware.
    2.  The middleware function receives the `*http.Response` and can inspect headers. It returns an `OnCompletionFunc`.
    3.  To handle streams, the original response body is wrapped in a custom `StreamInterceptor`. This interceptor passes data directly to the client without delay.
    4.  As the data is streamed, the interceptor buffers it internally. When the stream ends (`io.EOF`), it triggers the `OnCompletionFunc` with the complete buffered body, allowing for safe post-stream processing without blocking the client.

### Adding New Middleware

To add new middleware, follow the pattern in `cmd/gateway/main.go`:

1.  **For Transport Middleware**:
    *   Create your middleware function in the `internal/transport/middleware` package.
    *   In `main.go`, add it to the `transportmw.Chain(...)`.

2.  **For Core Middleware**:
    *   Create your `ResponseMiddleware` function in the `internal/core/middleware` package.
    *   In `main.go`, pass your function (or a chain of functions) when creating the `core.NewProxy`.
