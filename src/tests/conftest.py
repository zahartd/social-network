import os
import time
import pytest
import requests
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


@pytest.fixture(scope="session", autouse=True)
def ensure_api_is_ready(api_gateway_url):
    wait_for_service(api_gateway_url)


@pytest.fixture(scope="module")
def kafka_consumer():
    broker = os.environ.get("KAFKA_BROKER_URL", "kafka:9092")
    # создаём консьюмера, но НЕ подписываемся методом subscribe()
    consumer = KafkaConsumer(
        bootstrap_servers=[broker],
        auto_offset_reset='latest',
        enable_auto_commit=False,
        consumer_timeout_ms=5000,
        key_deserializer=lambda k: k.decode() if k else None,
        value_deserializer=lambda v: v.decode(),
    )
    # Явно назначаем каждому топику партицию 0
    topics = ['user-registrations','post-views','post-likes','post-comments']
    tps = [TopicPartition(t, 0) for t in topics]
    consumer.assign(tps)
    # перематываем сразу на конец
    consumer.seek_to_end(*tps)
    yield consumer
    consumer.close()

@pytest.fixture(autouse=True)
def fast_forward_consumer(kafka_consumer):
    """
    Перед каждым тестом — снова перемотать консьюмера
    на конец, чтобы читать только новые сообщения.
    """
    tps = kafka_consumer.assignment()
    kafka_consumer.seek_to_end(*tps)
    yield
