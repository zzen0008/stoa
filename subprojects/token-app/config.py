from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    # All variables will be prefixed with TOKEN_APP_
    # e.g. TOKEN_APP_CLIENT_ID
    model_config = SettingsConfigDict(env_prefix="TOKEN_APP_")

    # OIDC settings
    server_metadata_url: str
    client_id: str
    client_secret: str | None = None
    redirect_uri: str
    cookie_secret: str

settings = Settings()
