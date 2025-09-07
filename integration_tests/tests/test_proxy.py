import openai
import pytest

def test_proxy_authentication(test_user_token):
    """
    A simple test to check that the proxy is up and authenticating requests.
    """
    client = openai.OpenAI(
        base_url="http://gateway:8080/v1",
        api_key=test_user_token["access_token"],
    )

    try:
        response = client.chat.completions.create(
            model="mockllm/mock-gpt-3.5-turbo",
            messages=[
                {"role": "system", "content": "You are a helpful assistant."},
                {"role": "user", "content": "Hello!"},
            ],
        )
        print(response)
        assert response is not None
    except openai.APIStatusError as e:
        pytest.fail(f"API request failed with status code {e.status_code}: {e.response}")

