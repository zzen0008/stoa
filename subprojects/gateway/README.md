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
    /home/zz/projects/llm-gateway/resources/scripts/non_streaming_request.sh
    ```

*   **Streaming Request:**
    ```bash
    /home/zz/projects/llm-gateway/resources/scripts/streaming_request.sh
    ```
