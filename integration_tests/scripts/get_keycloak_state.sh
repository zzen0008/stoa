#!/bin/bash

set -e

# Get the directory where the script is located
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

REALM_NAME="llm-test-realm"
KCADM="/opt/keycloak/bin/kcadm.sh"

# Function to run kcadm commands in the keycloak container
run_kcadm() {
    docker compose -f "$SCRIPT_DIR/../docker-compose.yml" exec keycloak $KCADM "$@"
}

# Login to the admin CLI
run_kcadm config credentials --server http://localhost:8080 --realm master --user admin --password admin

# Get realm info
echo "--- Realm: $REALM_NAME ---"
run_kcadm get realms/$REALM_NAME

echo "--- Clients ---"
run_kcadm get clients -r $REALM_NAME

echo "--- Users ---"
run_kcadm get users -r $REALM_NAME

echo "--- Groups ---"
run_kcadm get groups -r $REALM_NAME