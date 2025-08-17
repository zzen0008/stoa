#!/bin/bash

# This script fetches a JWT from the Keycloak instance running in Docker.

# Keycloak server details
KEYCLOAK_URL="http://localhost:8180"
REALM="llm-gateway-realm"
CLIENT_ID="llm-gateway-client"
CLIENT_SECRET="**********" # IMPORTANT: Replace with the actual secret from the Keycloak UI

# User credentials
USERNAME="testuser"
PASSWORD="test"

# The token endpoint URL
TOKEN_URL="${KEYCLOAK_URL}/realms/${REALM}/protocol/openid-connect/token"

# Use curl to get the token
TOKEN_RESPONSE=$(curl -s -X POST "${TOKEN_URL}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=${CLIENT_ID}" \
  -d "client_secret=${CLIENT_SECRET}" \
  -d "username=${USERNAME}" \
  -d "password=${PASSWORD}")

# Extract the access token from the JSON response
ACCESS_TOKEN=$(echo "${TOKEN_RESPONSE}" | jq -r .access_token)

if [ -z "${ACCESS_TOKEN}" ] || [ "${ACCESS_TOKEN}" == "null" ]; then
  echo "Error: Could not fetch access token."
  echo "Response: ${TOKEN_RESPONSE}"
  exit 1
fi

echo "Successfully fetched token:"
echo "${ACCESS_TOKEN}"
