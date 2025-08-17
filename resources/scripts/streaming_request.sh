#!/bin/bash
#
# Example of a streaming request to the LLM Gateway.
# The -N flag disables buffering, which is important for streaming.
# The response will arrive in chunks as the model generates it.
#

curl -i -N -X POST http://localhost:8080/v1/chat/completions \
-H "Content-Type: application/json" \
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
