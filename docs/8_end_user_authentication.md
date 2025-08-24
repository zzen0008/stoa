
# End-User Authentication Guide (Python)

This guide explains how end-users and automated services can connect to the LLM Gateway using the standard `openai` Python library. The gateway handles authentication via OIDC, making the process secure and transparent.

## The Goal

The primary goal is to allow developers to use the official `openai` client with minimal changes. The gateway should handle user authentication and authorization, and then securely manage downstream API keys on behalf of the user.

This is achieved by pointing the `openai` client to the gateway's URL and providing a custom authentication helper that transparently manages time-limited credentials (JWTs).

## How it Works

1.  **Authentication:** The user's application authenticates against the gateway using an OIDC token obtained via the **Client Credentials Flow**. This is ideal for machine-to-machine communication.
2.  **Token Management:** The user's JWT is short-lived for security. The provided code snippet includes a helper class that automatically fetches a new token when the old one expires.
3.  **Standard Client:** The `openai` library is configured to use a custom HTTP transport. This transport intercepts every request to add the `Authorization: Bearer <JWT>` header, handling the token refresh logic in the background.
4.  **Gateway Responsibility:** The gateway receives the request, validates the user's JWT, and (if valid) replaces the header with its own downstream API key before forwarding the request to the LLM provider.

## Prerequisites

The user's application needs the following Python libraries:

```bash
pip install openai httpx "requests-oauthlib"
```

## Python Client Example

This is the complete, self-contained code your application needs. You can copy the `GatewayAuth` class directly into your project.

```python
import os
import time
import logging
import httpx
from openai import OpenAI
from requests_oauthlib import OAuth2Session
from oauthlib.oauth2 import BackendApplicationClient

# --- Configuration ---
# These values are provided to you by the gateway administrator.
OIDC_TOKEN_URL = os.getenv("OIDC_TOKEN_URL") # e.g., "https://keycloak.example.com/auth/realms/my-realm/protocol/openid-connect/token"
GATEWAY_API_URL = os.getenv("GATEWAY_API_URL") # e.g., "http://llm-gateway.my-org.com/v1"
CLIENT_ID = os.getenv("SERVICE_ACCOUNT_CLIENT_ID")
CLIENT_SECRET = os.getenv("SERVICE_ACCOUNT_CLIENT_SECRET")

# Setup basic logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')


# --- Reusable Authentication Helper ---
# This is the small, self-contained utility you should copy into your project.

class GatewayAuth(httpx.Auth):
    """
    An httpx.Auth class that automatically fetches and refreshes
    an OIDC token for the LLM Gateway.
    """
    def __init__(self, token_url: str, client_id: str, client_secret: str):
        if not all([token_url, client_id, client_secret]):
            raise ValueError("token_url, client_id, and client_secret must be provided.")
        self.token_url = token_url
        self.client_id = client_id
        self.client_secret = client_secret

        # The session manages the token and its lifecycle
        client = BackendApplicationClient(client_id=self.client_id)
        self._session = OAuth2Session(client=client)
        self._token = {}

    def _ensure_token(self):
        """Fetches a token if one is not present or is about to expire."""
        # Add a 60-second buffer to prevent using a token right before it expires
        if not self._token or self._token.get("expires_at", 0) < time.time() + 60:
            logging.info("Gateway token expired or not found. Fetching a new one...")
            try:
                self._token = self._session.fetch_token(
                    token_url=self.token_url,
                    client_id=self.client_id,
                    client_secret=self.client_secret
                )
                logging.info("Successfully fetched new gateway token.")
            except Exception as e:
                logging.error(f"Failed to fetch gateway token: {e}")
                raise

    def auth_flow(self, request: httpx.Request) -> httpx.Request:
        """The main entry point for httpx, called on every request."""
        self._ensure_token()
        access_token = self._token.get("access_token")
        request.headers["Authorization"] = f"Bearer {access_token}"
        yield request


# --- Main Application Logic ---

def main():
    """
    Main function demonstrating usage of the standard OpenAI client
    with custom authentication for the gateway.
    """
    logging.info("--- Starting Application ---")

    try:
        # 1. Create the custom auth helper.
        gateway_auth = GatewayAuth(
            token_url=OIDC_TOKEN_URL,
            client_id=CLIENT_ID,
            client_secret=CLIENT_SECRET
        )

        # 2. Create an httpx client that uses our custom auth.
        http_client = httpx.Client(auth=gateway_auth)

        # 3. Configure the standard OpenAI client.
        # - It points to the gateway's URL.
        # - It uses our custom, auth-aware http_client.
        # - 'api_key' is now irrelevant for auth, so we can pass a dummy value.
        client = OpenAI(
            base_url=GATEWAY_API_URL,
            api_key="dummy-key-since-auth-is-handled-by-transport",
            http_client=http_client
        )
        logging.info("OpenAI client configured to use the LLM Gateway.")

        # 4. Use the OpenAI client as usual.
        chat_completion = client.chat.completions.create(
            messages=[{"role": "user", "content": "Say this is a test!"}],
            model="gpt-4-turbo",
        )
        logging.info(f"SUCCESS! Response: {chat_completion.choices[0].message.content}")

    except Exception as e:
        logging.error(f"An application error occurred: {e}")

if __name__ == "__main__":
    main()
```
