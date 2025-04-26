import time

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