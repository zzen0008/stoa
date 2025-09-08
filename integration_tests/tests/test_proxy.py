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

def test_rate_limit_exceeded_for_user_group(test_user_token):
    """
    Tests that the rate limit is enforced for a user in a specific group.
    The user belongs to 'testgroup' which has a limit of 3 requests per minute.
    """
    client = openai.OpenAI(
        base_url="http://gateway:8080/v1",
        api_key=test_user_token["access_token"],
    )

    # The rate limit is set to 3 in the config.yaml for 'testgroup'
    allowed_requests = 3

    # Make successful requests up to the limit
    for i in range(allowed_requests):
        try:
            print(f"Making successful request #{i+1}")
            response = client.chat.completions.create(
                model="mockllm/mock-gpt-3.5-turbo",
                messages=[{"role": "user", "content": "Hello!"}],
            )
            print(f"Request #{i+1} succeeded")
        except openai.APIStatusError as e:
            pytest.fail(f"Request #{i+1} failed unexpectedly with status {e.status_code}: {e.response}")

    # The next request should be rate-limited
    try:
        print(f"Making request #{allowed_requests + 1}, expecting it to fail.")
        response = client.chat.completions.create(
            model="mockllm/mock-gpt-3.5-turbo",
            messages=[{"role": "user", "content": "Hello!"}],
        )
        pytest.fail(f"Request #{allowed_requests + 1} succeeded when it should have been rate-limited")
    except openai.APIStatusError as excinfo:
        assert excinfo.status_code == 429
        assert "Too Many Requests" in str(excinfo)
