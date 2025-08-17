# LLM Gateway

This project implements a high-performance, extensible LLM Gateway that acts as a unified entry point for multiple downstream LLM providers. It is designed to be modular, reliable, and easy to maintain.

## Core Features

- **Provider Routing:** Intelligently routes requests to different downstream LLM providers based on the model name in the request.
- **Fallback Chains:** Supports configurable fallback strategies, automatically retrying requests with a secondary provider if the primary one fails.
- **Unified API:** Exposes a consistent `/v1/chat/completions` and `/v1/models` API, aggregating models from all configured providers.
- **Extensible:** Built with a middleware-friendly architecture to easily add features like authentication, rate limiting, and logging.

## Project Structure

The project is organized into two main subprojects:

-   `subprojects/gateway`: The core LLM Gateway application.
-   `subprojects/mockLLM`: A dummy LLM API server for testing and development. It provides mock endpoints for `/v1/models` and `/v1/chat/completions` and returns Lorem Ipsum text.

## Getting Started

The entire application stack (gateway and mock server) is managed via Docker and Docker Compose.

### Prerequisites

-   Docker
-   Docker Compose

### Running the Application

To build and start the services, run the following command from the project root:

```bash
make up
```

This will:
1.  Build the Docker images for both the `gateway` and `mockLLM` services.
2.  Start the containers in detached mode.

The gateway will be available at `http://localhost:8080`.

### Other Commands

-   **Stop the services:**
    ```bash
    make down
    ```

-   **View logs:**
    ```bash
    make logs
    ```

## Configuration

The gateway's behavior is controlled by `subprojects/gateway/config.yaml`. By default, it is configured to route requests to the `mockLLM` service.
