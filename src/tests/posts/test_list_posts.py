from helpers.utils import auth_headers, make_request


async def test_list_my_posts(api_gateway_url, login_user):
    token, _ = login_user
    for i in range(3):
        url = api_gateway_url + "/posts"
        payload = {
            "title": f"Пост {i}",
            "description": f"Описание поста {i}",
            "is_private": False,
            "tags": ["list", "my"]
        }
        headers = {**auth_headers(token), "Content-Type": "application/json"}
        resp = make_request("POST", url, data=payload, headers=headers)
        assert resp.status_code == 201

    url = api_gateway_url + "/posts/list/my"
    params = {"page": 1, "page_size": 3}
    headers = auth_headers(token)
    resp = make_request("GET", url, params=params, headers=headers)
    assert resp.status_code == 200, f"Ошибка получения списка моих постов: {resp.text}"
    list_data = resp.json()
    assert "posts" in list_data
    assert isinstance(list_data["posts"], list)


async def test_list_all_public_posts(api_gateway_url, login_user):
    token, _ = login_user
    url = api_gateway_url + "/posts/list/public"
    params = {"page": 1, "page_size": 15}
    headers = auth_headers(token)
    resp = make_request("GET", url, params=params, headers=headers)
    assert resp.status_code == 200, f"Ошибка получения публичных постов: {resp.text}"
    list_data = resp.json()
    assert "posts" in list_data

async def test_list_public_posts_by_user(api_gateway_url, login_user, created_post):
    _, token, user_data = created_post
    target_user_id = user_data["id"]
    token, user_data = login_user
    url = api_gateway_url + f"/posts/list/public/{target_user_id}"
    params = {"page": 1, "page_size": 4}
    headers = auth_headers(token)
    resp = make_request("GET", url, params=params, headers=headers)
    assert resp.status_code == 200, f"Ошибка получения публичных постов пользователя: {resp.text}"
    list_data = resp.json()
    assert "posts" in list_data
