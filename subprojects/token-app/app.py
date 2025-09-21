import streamlit as st
from streamlit_oauth import OAuth2Component
import os
from dotenv import load_dotenv

# Load environment variables, though Docker Compose will provide them
load_dotenv()

# --- Configuration ---
AUTHORIZE_URL = os.environ.get('AUTHORIZE_URL')
TOKEN_URL = os.environ.get('TOKEN_URL')
REFRESH_TOKEN_URL = os.environ.get('REFRESH_TOKEN_URL')
REVOKE_TOKEN_URL = os.environ.get('REVOKE_TOKEN_URL')
CLIENT_ID = os.environ.get('CLIENT_ID')
CLIENT_SECRET = os.environ.get('CLIENT_SECRET')
REDIRECT_URI = os.environ.get('REDIRECT_URI')
SCOPE = os.environ.get('SCOPE')

st.set_page_config(layout="wide")
st.title("OIDC Token Fetcher")

# Create OAuth2Component instance
oauth2 = OAuth2Component(
    client_id=CLIENT_ID,
    client_secret=CLIENT_SECRET,
    authorize_endpoint=AUTHORIZE_URL,
    token_endpoint=TOKEN_URL,
    refresh_token_endpoint=REFRESH_TOKEN_URL,
    revoke_token_endpoint=REVOKE_TOKEN_URL
)

# Check if a token exists in the session state
if 'token' not in st.session_state:
    # If not, show the authorization button
    result = oauth2.authorize_button(
        name="Login with Keycloak",
        redirect_uri=REDIRECT_URI,
        scope=SCOPE,
        icon="https://www.keycloak.org/resources/images/keycloak_icon_512px.svg"
    )
    
    if result and 'token' in result:
        # If authorization is successful, save the token in the session state
        st.session_state.token = result.get('token')
        st.rerun()
else:
    # If a token exists, display it
    token = st.session_state['token']
    
    st.success("You are logged in!")
    
    access_token = token.get("access_token", "Token not found in response.")
    
    st.text_area(
        "Your Access Token",
        access_token,
        height=250,
    )
    with st.expander("Full Token Response Details"):
        st.json(token)

    if st.button("Logout"):
        # Revoke the token and clear the session state
        oauth2.revoke_token(token)
        del st.session_state['token']
        st.rerun()