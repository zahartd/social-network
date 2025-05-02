from helpers.utils import auth_headers, make_request


async def test_create_post(api_gateway_url, login_user):
    token, _ = login_user
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
    assert post.get("id") is not None
    assert post.get("title") == "Мой первый пост"
