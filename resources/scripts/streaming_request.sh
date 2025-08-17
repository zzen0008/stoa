#!/bin/bash
#
# Example of a streaming request to the LLM Gateway.
# The -N flag disables buffering, which is important for streaming.
# The response will arrive in chunks as the model generates it.
#

# Fetch the JWT token first
TOKEN=$(./resources/scripts/get_token.sh | tail -n 1)

if [ -z "$TOKEN" ]; then
    echo "Failed to get token. Exiting."
    exit 1
fi

curl -i -N -X POST http://localhost:8080/v1/chat/completions \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d '{
  "model": "mockllm/mock-gpt-4",
  "messages": [
    {
      "role": "user",
      "content": "Tell me a long joke about a programmer."
    }
  ],
  "stream": true
}'

