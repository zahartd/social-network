from helpers.utils import make_request, wait_for_kafka


async def test_user_registration_emits_event(api_gateway_url, unique_user, kafka_consumer):
    resp = make_request(
        "POST",
        f"{api_gateway_url}/user",
        headers={"Content-Type": "application/json"},
        data=unique_user,
    )
    assert resp.status_code == 201
    user_id = resp.json()["user"]["id"]

    ok = wait_for_kafka(
        kafka_consumer,
        topic="user-registrations",
        predicate=lambda m: user_id in m.value,
    )
    assert ok, "Событие user-registrations не найдено"