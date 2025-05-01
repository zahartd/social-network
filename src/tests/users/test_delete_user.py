from helpers.utils import auth_headers, make_request


async def test_delete_user(api_gateway_url, user_factory):
    token1, user1 = user_factory()
    token2, user2 = user_factory()
    assert user1 != user2

    url = api_gateway_url + f"/user/{user1["login"]}"
    resp = make_request("DELETE", url, headers=auth_headers(token1))
    assert resp.status_code in [200, 204], f"Ошибка удаления пользователя: {resp.text}"

    resp_get = make_request("GET", url, headers=auth_headers(token2))
    assert resp_get.status_code == 404