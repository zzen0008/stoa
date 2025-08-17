# Mock LLM Server

This is a simple, dummy LLM API server created for testing the main LLM Gateway. It mimics the basic behavior of an OpenAI-compatible API.

## Features

-   **`/v1/models`:** Returns a static list of mock model names.
-   **`/v1/chat/completions`:**
    -   Accepts a request with a `model`, `max_tokens`, and `stream` field.
    -   Returns a non-streaming or streaming response with Lorem Ipsum text.
    -   The number of words in the response is determined by the `max_tokens` parameter (defaults to 100).
    -   Logs all incoming requests to standard output.

## Usage

The server is configured to run on port `8081`.

### Running the Server

You can run the server directly or using the provided `Makefile`.

**Directly:**

```bash
go run main.go
```

**Using Makefile:**

```bash
make run
```

### Building the Server

To build a binary:

```bash
make build
```

This will create an executable named `mockLLM` in the current directory.
