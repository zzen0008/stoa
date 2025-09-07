# Integration Tests

This directory contains the integration tests for the LLM Gateway. The tests are designed to run in a self-contained Docker environment, managed by Docker Compose.

## Prerequisites

- Docker
- Docker Compose

## Environment Setup

1.  **Start the Test Environment:**
    This command will build the Docker images and start the services defined in `docker-compose.yml` in detached mode.

    ```bash
    make up
    ```

2.  **Initialize Keycloak:**
    This command runs a script to configure the Keycloak instance for the tests. It creates a new realm, a client for the gateway, a test user, and a service account.

    ```bash
    make init
    ```

3.  **Verify Environment Setup:**
    After initialization, you can verify that the Keycloak user is configured correctly and can obtain a token. This is a crucial debugging step before running the full test suite.

    ```bash
    make verify
    ```
    If this command fails, there is an issue with the Keycloak configuration that needs to be resolved.

## Running the Tests

Once the environment is up, initialized, and verified, you can run the integration tests:

```bash
make tests
```

This will start the `tests` service, which runs `pytest` against the test files in the `tests/` directory.

## Debugging

Several scripts are provided in the `scripts/` directory to help with debugging the test environment.

### View Keycloak State

To inspect the current configuration of the Keycloak realm, including clients, users, and groups, run:

```bash
./scripts/get_keycloak_state.sh
```

### Decode a JWT

To decode a JWT and view its header and payload, use the `decode_jwt.sh` script. This is useful for inspecting the tokens issued by Keycloak.

```bash
./scripts/decode_jwt.sh <your_jwt_token>
```

### View Proxy State

To view the logs from the gateway service in real-time, run:

```bash
./scripts/get_proxy_state.sh
```

## Cleanup

1.  **Stop and Remove Containers:**
    To stop and remove the containers, networks, and volumes created by Docker Compose, run:

    ```bash
    make down
    ```

2.  **Full Cleanup (including images):**
    To perform a more thorough cleanup, including removing the Docker images built for the tests, run:

    ```bash
    make clean
    ```
