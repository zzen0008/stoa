#!/bin/bash
#
# Example of a non-streaming request to the LLM Gateway.
# This request asks for a complete story and waits for the full response.
#

curl -i -X POST http://localhost:8080/v1/chat/completions \
-H "Content-Type: application/json" \
-d '{
  "model": "mockllm/mock-gpt-4",
  "messages": [
    {
      "role": "user",
      "content": "Write a short story about a robot who discovers music."
    }
  ],
  "stream": false
}'
