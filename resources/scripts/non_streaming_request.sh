#!/bin/bash
#
# Example of a non-streaming request to the LLM Gateway.
# This request asks for a complete story and waits for the full response.
#

# Fetch the JWT token first
#TOKEN=$(./resources/scripts/get_token.sh | tail -n 1)

if [ -z "$TOKEN" ]; then
    echo "Failed to get token. Exiting."
    exit 1
fi

curl -i -X POST http://localhost:8080/v1/chat/completions \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d '{"model": "mockllm/mock-gpt-4", "messages": [{"role": "user", "content": "Write a short story about a robot who discovers music."}], "stream": false}'

