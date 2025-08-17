# 7. Tech Choice: OIDC Authentication Middleware

This document outlines the decision-making process for selecting a Go library to implement OpenID Connect (OIDC) authentication middleware, specifically for validating JWTs issued by a provider like Keycloak.

### 1. The Goal

The primary objective is to secure API endpoints by validating a JWT Bearer Token sent in the `Authorization` header. This validation must be secure, reliable, and align with our project's principle of maintainability.

A proper OIDC validation process involves several steps:
1.  **Fetch Provider Configuration:** Discover the OIDC provider's endpoints (like the JWKS URI) from its issuer URL.
2.  **Fetch Public Keys (JWKS):** Download the JSON Web Key Set, which contains the public keys used to sign the JWTs.
3.  **Handle Key Rotation:** The provider's signing keys can and will rotate. The client must be able to handle this gracefully, fetching new keys as needed.
4.  **Verify Token Signature:** Use the correct public key to verify the cryptographic signature of the JWT.
5.  **Validate Standard Claims:** Check standard OIDC claims, such as:
    *   `iss` (Issuer): Ensure the token was issued by the expected authority.
    *   `aud` (Audience): Ensure the token is intended for our application.
    *   `exp` (Expiration): Ensure the token has not expired.

### 2. Library Candidates

Two primary candidates were considered for this task:

1.  **`golang-jwt/jwt`**: A popular, general-purpose library for parsing and validating JWTs.
2.  **`coreos/go-oidc`**: A specialized client library for the OpenID Connect protocol.

### 3. Comparison and Rationale

While `golang-jwt/jwt` is an excellent library for handling the JWT format itself, it is not aware of the broader OIDC protocol. `coreos/go-oidc`, on the other hand, is designed specifically for this protocol.

| Feature | `golang-jwt/jwt` (General Purpose) | `coreos/go-oidc` (OIDC Specialist) |
| :--- | :--- | :--- |
| **Scope** | Handles JWT parsing, validation, and creation. | Handles the entire OIDC discovery and validation workflow. |
| **Provider Discovery** | ❌ **Manual Implementation Required.** You would need to write your own code to fetch the provider's configuration. | ✅ **Built-in.** Automatically discovers endpoints from the issuer URL. |
| **JWKS Management** | ❌ **Manual Implementation Required.** You would need to write your own code to fetch, parse, cache, and rotate the public keys. | ✅ **Built-in.** Manages the entire lifecycle of fetching and caching the provider's public keys. |
| **OIDC Claim Validation** | ⚠️ **Partial.** The library provides helpers for standard claims like `exp`, but you must manually verify the `iss` and `aud` against the provider's configuration. | ✅ **Built-in.** Performs a comprehensive validation of all required OIDC claims out-of-the-box. |
| **Simplicity** | **Lower.** Requires significant boilerplate code to handle the OIDC protocol details, which is complex and security-critical. | **Higher.** Provides a simple, high-level API (`verifier.Verify(token)`) that abstracts away the underlying complexity. |

### 4. Conclusion

For validating JWTs from an OIDC provider like Keycloak, **`coreos/go-oidc` is the superior choice.**

It abstracts away the complex and error-prone details of the OIDC protocol, particularly the critical process of public key management. By handling discovery, JWKS fetching, and key rotation automatically, it provides a solution that is not only **simpler to implement** but also **more secure and maintainable** in the long run. It allows us to focus on our business logic instead of reinventing a security-critical OIDC client.
