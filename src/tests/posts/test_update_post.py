from helpers.utils import auth_headers, make_request


async def test_update_post(api_gateway_url, created_post):
    post, token, _ = created_post
    post_id = post.get("id")
    url = api_gateway_url + f"/posts/{post_id}"
    headers = {**auth_headers(token), "Content-Type": "application/json"}
    update_payload = {
        "title": "Обновленный заголовок поста",
        "description": "Это обновленное описание. Пост теперь приватный!",
        "is_private": True,
        "tags": ["тест", "обновление", "приватность"]
    }
    resp = make_request("PUT", url, data=update_payload, headers=headers)
    assert resp.status_code == 200, f"Ошибка обновления поста: {resp.text}"
    updated_post = resp.json()
    assert updated_post.get("title") == update_payload["title"]