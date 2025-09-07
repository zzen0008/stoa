import pytest
from keycloak import KeycloakOpenID

@pytest.fixture(scope="session")
def keycloak_openid_client():
    return KeycloakOpenID(server_url="http://keycloak:8080/",
                          client_id="llm-gateway-client",
                          realm_name="llm-test-realm",
                          client_secret_key="my-secret")

@pytest.fixture(scope="session")
def test_user_token(keycloak_openid_client):
    return keycloak_openid_client.token("testuser", "test")

@pytest.fixture(scope="session")
def service_account_token(keycloak_openid_client):
    return keycloak_openid_client.token(grant_type="client_credentials")
