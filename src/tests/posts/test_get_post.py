from helpers.utils import auth_headers, make_request


async def test_get_post_by_id(api_gateway_url, created_post):
    post, token, _ = created_post
    post_id = post.get("id")
    url = api_gateway_url + f"/posts/{post_id}"
    headers = auth_headers(token)
    resp = make_request("GET", url, headers=headers)
    assert resp.status_code == 200, f"Ошибка получения поста: {resp.text}"
    post_data = resp.json()
    assert post_data.get("id") == post_id