from helpers.utils import auth_headers, make_request, wait_for_kafka


async def test_view_post_emits_event(api_gateway_url, login_user, kafka_consumer):
    token, _ = login_user
    post_id = make_request(
        "POST", f"{api_gateway_url}/posts",
        headers={**auth_headers(token),"Content-Type":"application/json"},
        data={"title":"t","description":"d","is_private":False,"tags":[]}
    ).json()["id"]

    make_request("POST", f"{api_gateway_url}/posts/{post_id}/view",
                    headers=auth_headers(token))

    ok = wait_for_kafka(
        kafka_consumer,
        topic="post-views",
        predicate=lambda m: post_id in m.value
    )
    assert ok, "Событие post-views не найдено"


async def test_like_and_unlike_emits_event(api_gateway_url, login_user, kafka_consumer):
    token, _ = login_user
    post_id = make_request(
        "POST", f"{api_gateway_url}/posts",
        headers={**auth_headers(token),"Content-Type":"application/json"},
        data={"title":"t","description":"d","is_private":False,"tags":[]}
    ).json()["id"]

    make_request("POST",    f"{api_gateway_url}/posts/{post_id}/like", headers=auth_headers(token))
    make_request("DELETE",  f"{api_gateway_url}/posts/{post_id}/like", headers=auth_headers(token))

    ok = wait_for_kafka(
        kafka_consumer,
        topic="post-likes",
        predicate=lambda m: post_id in m.value
    )
    assert ok, "Событие post-likes не найдено"


async def test_comment_and_list_emits_event(api_gateway_url, login_user, kafka_consumer):
    token, _ = login_user
    post_id = make_request(
        "POST", f"{api_gateway_url}/posts",
        headers={**auth_headers(token),"Content-Type":"application/json"},
        data={"title":"t","description":"d","is_private":False,"tags":[]}
    ).json()["id"]

    comment_text = "hello kafka"
    resp = make_request(
        "POST", f"{api_gateway_url}/posts/{post_id}/comments",
        headers={**auth_headers(token),"Content-Type":"application/json"},
        data={"text": comment_text}
    )
    comment_id = resp.json()["comment"]["id"]

    ok = wait_for_kafka(
        kafka_consumer,
        topic="post-comments",
        predicate=lambda m: comment_id in m.value and comment_text in m.value
    )
    assert ok, "Событие post-comments не найдено"