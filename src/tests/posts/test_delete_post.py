from helpers.utils import auth_headers, make_request


async def test_delete_post(api_gateway_url, created_post):
    post, token, user_data = created_post
    post_id = post.get("id")
    url = api_gateway_url + f"/posts/{post_id}"
    headers = auth_headers(token)
    resp = make_request("DELETE", url, headers=headers)
    assert resp.status_code in [200, 204], f"Ошибка удаления поста: {resp.text}"
    resp_get = make_request("GET", url, headers=headers)
    assert resp_get.status_code == 404