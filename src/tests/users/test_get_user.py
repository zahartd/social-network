from helpers.utils import auth_headers, make_request


async def test_get_user_profile(api_gateway_url, login_user):
    token, user_data = login_user
    identifier = user_data["login"]
    url = api_gateway_url + f"/user/{identifier}"
    headers = auth_headers(token)
    resp = make_request("GET", url, headers=headers)
    assert resp.status_code == 200, f"Ошибка получения профиля: {resp.text}"
    profile = resp.json()
    assert profile.get("email") == user_data["email"]

async def test_get_profile_with_invalid_token(api_gateway_url, unique_user):
    url = api_gateway_url + "/user"
    resp = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
    assert resp.status_code == 201
    identifier = unique_user["login"]
    url = api_gateway_url + f"/user/{identifier}"
    headers = auth_headers("invalidtoken")
    resp = make_request("GET", url, headers=headers)
    assert resp.status_code in [401, 403]
