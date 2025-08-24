
# Service Account Usage

This document explains the concept and usage of Service Accounts for authenticating automated, non-human clients with the LLM Gateway.

## What is a Service Account?

A Service Account is a non-human client that needs to interact with the gateway. Examples include:
-   A CI/CD pipeline that uses an LLM for code analysis.
-   A data processing job that uses an LLM for text enrichment.
-   A backend service that needs to call an LLM as part of its business logic.

These clients cannot use a traditional user login flow (i.e., a browser-based redirect). Instead, they use a direct, machine-to-machine authentication method.

## Authentication Flow: OIDC Client Credentials

The LLM Gateway uses the **OAuth 2.0 Client Credentials Grant** flow for service accounts. This is a standard, secure method for machine-to-machine (M2M) authentication.

The flow works as follows:

1.  **Issuing Credentials**: A gateway administrator creates a "client" in the OIDC provider (e.g., Keycloak). This client is configured to use the `client_credentials` grant type. The administrator is issued a `Client ID` and a `Client Secret`. These are the long-lived credentials for the service account.

2.  **Storing Credentials**: The end-user (the owner of the service) must store the `Client ID` and `Client Secret` securely. In a Kubernetes environment, they should be stored as **Kubernetes Secrets** and exposed to the application's pods as environment variables or mounted files. **They should never be hardcoded in source code or container images.**

3.  **Requesting a Token**: The service account's application code uses the `Client ID` and `Client Secret` to make a direct `POST` request to the OIDC provider's token endpoint.

4.  **Receiving a Token**: The OIDC provider validates the credentials and, if they are correct, issues a short-lived JSON Web Token (JWT). This JWT is the service account's proof of identity for the gateway.

5.  **Accessing the Gateway**: The service account includes this JWT in the `Authorization: Bearer <JWT>` header of every request it makes to the LLM Gateway.

6.  **Token Expiration & Refresh**: Because the JWT is short-lived (e.g., 5-60 minutes), the application is responsible for automatically refreshing it. When the token is close to expiring, the application must repeat the token request (step 3) to get a new one. The Python example provided in the [End-User Authentication Guide](./8_end_user_authentication.md) handles this logic automatically.

## Authorization

Once authenticated, the gateway needs to decide what the service account is allowed to do. This is handled through **authorization**.

The JWT issued by the OIDC provider can contain claims, such as `groups` or `roles`, that describe the service account's permissions. For example, a JWT might contain:

```json
{
  "sub": "service-account-data-pipeline",
  "groups": ["data-enrichment-team", "read-only-models"],
  "iss": "https://keycloak.example.com/auth/realms/my-realm",
  ...
}
```

The gateway can be configured with policies that use this information. For example:
-   Allow any service account in the `data-enrichment-team` group to access the `gpt-4-turbo` model.
-   Deny access to fine-tuning endpoints for any client in the `read-only-models` group.

This allows for fine-grained control over which services can access which LLM resources.
