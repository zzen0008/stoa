# Toy Examples for LLM Gateway Concepts

This directory contains a series of simple, self-contained, and runnable Go programs. Each program demonstrates a key architectural concept that will be used to build the LLM Gateway.

To run any example, simply use `go run <filename>.go`.

### Examples

1.  `1_basic_reverse_proxy.go`: Shows the simplest possible reverse proxy using `net/http/httputil`. This is the foundation of our gateway.

2.  `2_proxy_with_request_modification.go`: Demonstrates how to modify headers before forwarding a request (e.g., to add an API key). This is essential for authenticating with downstream providers.

3.  `3_proxy_with_body_inspection.go`: Shows how to safely read the request body for routing decisions and then reconstruct it for the downstream service. This is the core of our intelligent routing.

4.  `4_proxy_with_fallback.go`: A simple example of trying a primary backend and then a secondary one if the first fails. This demonstrates our reliability pattern.

5.  `5_proxy_with_middleware.go`: Illustrates how to wrap the proxy handler with a simple logging middleware, which is how we will add features like authentication and observability.

### Design Documents

*   `6_schema_design.md`: Outlines the design for our internal, provider-agnostic Go structs, prioritizing simplicity and maintainability.
