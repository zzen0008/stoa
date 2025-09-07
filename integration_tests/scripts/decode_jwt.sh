#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <jwt>"
    exit 1
fi

JWT=$1

function base64url_decode() {
    local data=$(echo "$1" | tr '_-' '/+' | tr -d '[:space:]')
    local len=${#data}
    local padlen=$(( (4 - (len % 4)) % 4 ))
    if [ $padlen -ne 0 ]; then
        data="$data$(printf '=%.0s' $(seq 1 $padlen))"
    fi
    echo "$data" | base64 -d
}

HEADER=$(echo "$JWT" | cut -d '.' -f 1)
PAYLOAD=$(echo "$JWT" | cut -d '.' -f 2)

echo "--- Header ---"
base64url_decode "$HEADER" | jq .

echo "--- Payload ---"
base64url_decode "$PAYLOAD" | jq .
