#!/bin/bash

set -e

# Variables
KEYCLOAK_EXTERNAL_URL="http://localhost:8180"
KEYCLOAK_INTERNAL_URL="http://localhost:8080"
ADMIN_USER="admin"
ADMIN_PASSWORD="admin"
REALM_NAME="llm-test-realm"
CLIENT_NAME="llm-gateway-client"
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

# Create token-app client if it doesn't exist
echo "Ensuring client 'token-app' exists..."
TOKEN_APP_CLIENT_ID=$(run_kcadm get clients -r $REALM_NAME -q clientId=token-app --fields id --format csv --noquotes 2>/dev/null || true)
if [ -z "$TOKEN_APP_CLIENT_ID" ]; then
    run_kcadm create clients -r $REALM_NAME \
        -s clientId=token-app \
        -s enabled=true \
        -s publicClient=true \
        -s 'redirectUris=["http://localhost:8501/oauth2callback"]' \
        -s directAccessGrantsEnabled=false
    echo "Client 'token-app' created."
fi

# Create client if it doesn't exist
echo "Ensuring client '$CLIENT_NAME' exists..."
CLIENT_ID=$(run_kcadm get clients -r $REALM_NAME -q clientId=$CLIENT_NAME --fields id --format csv --noquotes 2>/dev/null || true)
if [ -z "$CLIENT_ID" ]; then
    CLIENT_ID=$(run_kcadm create clients -r $REALM_NAME \
        -s clientId=$CLIENT_NAME \
        -s enabled=true \
        -s publicClient=false \
        -s directAccessGrantsEnabled=true \
        -s serviceAccountsEnabled=true \
        -s secret=my-secret -i)
    
    echo "Client created with ID: $CLIENT_ID. Adding protocol mappers..."
    # Add audience mapper
    run_kcadm create clients/$CLIENT_ID/protocol-mappers/models -r $REALM_NAME \
        -s name=audience-mapper \
        -s protocol=openid-connect \
        -s protocolMapper=oidc-audience-mapper \
        -s 'config."id.token.claim"="false"' \
        -s 'config."access.token.claim"="true"' \
        -s 'config."included.client.audience"="'$CLIENT_NAME'"'

    # Add groups mapper directly to the client
    run_kcadm create clients/$CLIENT_ID/protocol-mappers/models -r $REALM_NAME \
        -s name=groups-mapper \
        -s protocol=openid-connect \
        -s protocolMapper=oidc-group-membership-mapper \
        -s 'config."claim.name"="groups"' \
        -s 'config."full.path"="false"' \
        -s 'config."id.token.claim"="true"' \
        -s 'config."access.token.claim"="true"' \
        -s 'config."userinfo.token.claim"="true"'
fi


# Delete user if exists to ensure a clean state
echo "Deleting user 'testuser' if it exists..."
USER_ID=$(run_kcadm get users -r $REALM_NAME -q username=testuser --fields id --format csv --noquotes 2>/dev/null || true)
if [ -n "$USER_ID" ]; then
    run_kcadm delete users/$USER_ID -r $REALM_NAME
    echo "Deleted existing user."
fi

# Create user
echo "Creating user 'testuser'வுகளை..."
USER_ID=$(run_kcadm create users -r $REALM_NAME \
    -s username=testuser \
    -s enabled=true \
    -s emailVerified=true \
    -s 'credentials=[{"type":"password","value":"test","temporary":false}]' -i)
echo "User 'testuser' created with ID: $USER_ID"

# Create group if it doesn't exist
echo "Ensuring group 'testgroup' exists..."
GROUP_ID=$(run_kcadm get groups -r $REALM_NAME -q name=testgroup --fields id --format csv --noquotes 2>/dev/null || true)
if [ -z "$GROUP_ID" ]; then
    GROUP_ID=$(run_kcadm create groups -r $REALM_NAME -s name=testgroup -i)
    echo "Group 'testgroup' created with ID: $GROUP_ID"
fi

# Assign user to group
echo "Assigning user $USER_ID to group $GROUP_ID..."
run_kcadm update "users/$USER_ID/groups/$GROUP_ID" -r $REALM_NAME -n

echo "Keycloak entity setup complete!"
