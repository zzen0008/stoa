#!/bin/bash

set -e

REALM_NAME="llm-test-realm"
CLIENT_ID="llm-gateway-client"
CLIENT_SECRET="my-secret"
USERNAME="testuser"
PASSWORD="test"
TOKEN_URL="http://localhost:8180/realms/$REALM_NAME/protocol/openid-connect/token"

echo "Attempting to fetch token for user '$USERNAME'வுகளை..."

RESPONSE=$(curl -s -X POST "$TOKEN_URL" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=$USERNAME" \
  -d "password=$PASSWORD" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "grant_type=password")

if echo "$RESPONSE" | grep -q "access_token"; then
  echo "Successfully fetched token for user '$USERNAME'."
  echo "Environment setup is likely correct."
  exit 0
else
  echo "Error: Failed to fetch token for user '$USERNAME'."
  echo "Keycloak response:"
  echo "$RESPONSE" | jq .
  exit 1
fi
