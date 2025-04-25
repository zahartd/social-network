import os
import time
import pytest
import requests

@pytest.fixture(scope="session")
def api_gateway_url():
    return os.environ.get("API_GATEWAY_URL", "http://api-gateway:8080")

def wait_for_service(url, timeout=30):
    start = time.time()
    while time.time() - start < timeout:
        try:
            resp = requests.get(url + "/ping")
            if resp.status_code == 200:
                return
        except requests.RequestException:
            pass
        time.sleep(1)
    pytest.exit("API Gateway не запустился вовремя", returncode=1)

@pytest.fixture(scope="session", autouse=True)
def ensure_api_is_ready(api_gateway_url):
    wait_for_service(api_gateway_url)
