#!/bin/bash

set -e

# Variables
KEYCLOAK_EXTERNAL_URL="http://localhost:8180"
KEYCLOAK_INTERNAL_URL="http://localhost:8080"
ADMIN_USER="admin"
ADMIN_PASSWORD="admin"
REALM_NAME="llm-test-realm"
KCADM="/opt/keycloak/bin/kcadm.sh"

# Wait for Keycloak to be ready
echo "Waiting for Keycloak to start..."
until $(curl --output /dev/null --silent --head --fail "$KEYCLOAK_EXTERNAL_URL/admin/master/console/"); do
    printf '.'
    sleep 1
done
echo "Keycloak is ready!"

# Function to run kcadm commands in the keycloak container
run_kcadm() {
    docker compose exec keycloak $KCADM "$@"
}

# Login to the admin CLI
run_kcadm config credentials --server $KEYCLOAK_INTERNAL_URL --realm master --user $ADMIN_USER --password $ADMIN_PASSWORD

# Create client if it doesn't exist
echo "Ensuring client 'llm-gateway-client' exists..."
CLIENT_ID=$(run_kcadm get clients -r $REALM_NAME -q clientId=llm-gateway-client --fields id --format csv --noquotes 2>/dev/null || true)
if [ -z "$CLIENT_ID" ]; then
    CLIENT_ID=$(run_kcadm create clients -r $REALM_NAME \
        -s clientId=llm-gateway-client \
        -s enabled=true \
        -s publicClient=false \
        -s directAccessGrantsEnabled=true \
        -s serviceAccountsEnabled=true \
        -s secret=my-secret -i)
    
    echo "Client created with ID: $CLIENT_ID. Adding audience mapper..."
    run_kcadm create clients/$CLIENT_ID/protocol-mappers/models -r $REALM_NAME \
        -s name=audience-mapper \
        -s protocol=openid-connect \
        -s protocolMapper=oidc-audience-mapper \
        -s 'config."id.token.claim"="false"' \
        -s 'config."access.token.claim"="true"' \
        -s 'config."included.client.audience"="llm-gateway-client"'
fi

# Delete user if exists to ensure a clean state
echo "Deleting user 'testuser' if it exists..."
USER_ID=$(run_kcadm get users -r $REALM_NAME -q username=testuser --fields id --format csv --noquotes 2>/dev/null || true)
if [ -n "$USER_ID" ]; then
    run_kcadm delete users/$USER_ID -r $REALM_NAME
    echo "Deleted existing user."
fi

# Create user. Required actions are disabled on the realm itself.
echo "Creating user 'testuser'வுகளை..."
run_kcadm create users -r $REALM_NAME \
    -s username=testuser \
    -s enabled=true \
    -s emailVerified=true \
    -s 'credentials=[{"type":"password","value":"test","temporary":false}]'

echo "Keycloak entity setup complete!"