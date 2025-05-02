import json
import logging
import time

import requests

LOGGER = logging.getLogger(__name__)

def auth_headers(token):
    return {"Authorization": f"Bearer {token}"}

def make_request(method, url, params=None, data=None, headers=None):
    if data is not None:
        data = json.dumps(data)
    LOGGER.info(f"Request: {method} {url} params: {params} data: {data} headers: {headers}")
    resp = requests.request(method, url, params=params, data=data, headers=headers)
    LOGGER.info(f"Response: {resp.status_code} {resp.text}")
    return resp

def wait_for_kafka(consumer, *, topic, predicate, timeout_sec=5, step_ms=200):
    deadline = time.time() + timeout_sec
    while time.time() < deadline:
        batches = consumer.poll(timeout_ms=step_ms)
        for tp, messages in batches.items():
            if tp.topic != topic:
                continue
            for msg in messages:
                try:
                    if predicate(msg):
                        return True
                finally:
                    pass
    return False