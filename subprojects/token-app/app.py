import streamlit as st
from streamlit.runtime.secrets import secrets_singleton

from config import settings

def load_auth_config():
    """
    Injects the OIDC configuration from Pydantic BaseSettings
    into Streamlit's secret store.
    """
    # The structure must match what st.login expects for a generic OIDC provider
    auth_secrets = {
        "connections": {
            "keycloak": {
                "server_metadata_url": settings.server_metadata_url,
                "client_id": settings.client_id,
                "redirect_uri": settings.redirect_uri,
                "client_secret": settings.client_secret,
            }
        }
    }

    # Inject settings into Streamlit’s secret store
    secrets_singleton._secrets = auth_secrets
    for k, v in auth_secrets.items():
        secrets_singleton._maybe_set_environment_variable(k, v)

# This injection function MUST be run before any other Streamlit command.
load_auth_config()

# --- Main App ---

st.set_page_config(layout="wide")


st.title("OIDC Token Fetcher")

# The st.login() method will now read the config we injected
login_button = st.login("keycloak")

if st.user.email:
    st.success(f"Logged in as {st.user.email}")

    token_info = st.session_state.get("token_info")

    if token_info:
        st.text_area(
            "Your Access Token",
            token_info.get("access_token", "Token not found in session state."),
            height=250,
        )
        with st.expander("Full Token Details"):
            st.json(token_info)
    else:
        st.error("Could not find token in session state. Please try logging in again.")

    if st.button("Logout"):
        st.logout()