from helpers.utils import auth_headers, make_request


async def test_signup_and_login(api_gateway_url, unique_user):
    url = api_gateway_url + "/user"
    resp = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
    assert resp.status_code == 201, f"Ошибка регистрации: {resp.text}"
    data = resp.json()
    user_data = data.get("user")
    assert user_data
    assert user_data.get("login") == unique_user["login"]

    url = api_gateway_url + "/user/login"
    params = {"login": unique_user["login"], "password": unique_user["password"]}
    resp = make_request("GET", url, params=params)
    assert resp.status_code == 200, f"Ошибка логина: {resp.text}"
    data = resp.json()
    token = data.get("token")
    assert token

async def test_duplicate_signup(api_gateway_url, unique_user):
    url = api_gateway_url + "/user"
    resp = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
    assert resp.status_code == 201
    resp2 = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
    assert resp2.status_code != 201, "Дублирование регистрации прошло успешно, хотя должно быть отказано"

async def test_logout(api_gateway_url, login_user):
    token, user_data = login_user
    url = api_gateway_url + "/user/logout"
    headers = auth_headers(token)
    resp = make_request("GET", url, headers=headers)
    assert resp.status_code in [200, 204]
