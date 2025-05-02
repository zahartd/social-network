import os
import time
import uuid

import pytest
import requests
from helpers.utils import auth_headers, make_request
from kafka import KafkaConsumer, TopicPartition


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

@pytest.fixture
def unique_user():
    unique = str(uuid.uuid4())[:8]
    user_data = {
        "login": f"user_{unique}",
        "firstname": f"First_{unique}",
        "surname": f"Last_{unique}",
        "email": f"user_{unique}@example.com",
        "password": "TestPass123"
    }
    return user_data

@pytest.fixture
def register_user(api_gateway_url, unique_user):
    url = api_gateway_url + "/user"
    resp = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
    assert resp.status_code == 201, f"Регистрация не удалась: {resp.text}"
    unique_user["id"] = resp.json()["user"]["id"]
    return unique_user

@pytest.fixture
def login_user(api_gateway_url, register_user):
    login = register_user["login"]
    password = register_user["password"]
    url = api_gateway_url + "/user/login"
    params = {"login": login, "password": password}
    resp = make_request("GET", url, params=params)
    assert resp.status_code == 200, f"Логин не удался: {resp.text}"
    data = resp.json()
    token = data.get("token")
    assert token
    return token, register_user

@pytest.fixture
def user_factory(api_gateway_url):
    def _create():
        uniq = str(uuid.uuid4())[:8]
        user_data = {
            "login":    f"user_{uniq}",
            "firstname":f"First_{uniq}",
            "surname":  f"Last_{uniq}",
            "email":    f"user_{uniq}@example.com",
            "password": "TestPass123"
        }
        resp = make_request(
            "POST",
            api_gateway_url + "/user",
            headers={"Content-Type": "application/json"},
            data=user_data
        )
        assert resp.status_code == 201, f"Регистрация упала: {resp.text}"
        resp = make_request(
            "GET",
            api_gateway_url + "/user/login",
            params={"login": user_data["login"], "password": user_data["password"]}
        )
        assert resp.status_code == 200, f"Логин не удался: {resp.text}"
        data = resp.json()
        token = data.get("token")
        assert token
        return token, user_data
    return _create


@pytest.fixture
def created_post(api_gateway_url, login_user):
    token, user_data = login_user
    url = api_gateway_url + "/posts"
    payload = {
        "title": "Мой первый пост",
        "description": "Это содержимое поста, созданного для тестирования.",
        "is_private": False,
        "tags": ["тест", "golang", "api"]
    }
    headers = {**auth_headers(token), "Content-Type": "application/json"}
    resp = make_request("POST", url, data=payload, headers=headers)
    assert resp.status_code == 201, f"Ошибка создания поста: {resp.text}"
    post = resp.json()
    return post, token, user_data

@pytest.fixture(scope="session", autouse=True)
def ensure_api_is_ready(api_gateway_url):
    wait_for_service(api_gateway_url)

@pytest.fixture(scope="module")
def kafka_consumer():
    broker = os.environ.get("KAFKA_BROKER_URL", "kafka:9092")

    consumer = KafkaConsumer(
        bootstrap_servers=[broker],
        auto_offset_reset='latest',
        enable_auto_commit=False,
        consumer_timeout_ms=5000,
        key_deserializer=lambda k: k.decode() if k else None,
        value_deserializer=lambda v: v.decode(),
    )

    topics = ['user-registrations','post-views','post-likes','post-comments']
    tps = [TopicPartition(t, 0) for t in topics]
    consumer.assign(tps)

    consumer.seek_to_end(*tps)
    yield consumer
    consumer.close()

@pytest.fixture(autouse=True)
def fast_forward_consumer(kafka_consumer):
    tps = kafka_consumer.assignment()
    kafka_consumer.seek_to_end(*tps)
    yield
