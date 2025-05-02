from helpers.utils import auth_headers, make_request


async def test_update_user_profile(api_gateway_url, login_user):
    token, user_data = login_user
    identifier = user_data["login"]
    url = api_gateway_url + f"/user/{identifier}"
    headers = {**auth_headers(token), "Content-Type": "application/json"}
    update_payload = {
        "email": "updated_" + user_data["email"],
        "firstname": user_data["firstname"],
        "surname": user_data["surname"],
        "phone": "+1234567890",
        "bio": "Updated biography text"
    }
    resp = make_request("PUT", url, data=update_payload, headers=headers)
    assert resp.status_code in [200, 204], f"Ошибка обновления профиля: {resp.text}"
    resp_get = make_request("GET", url, headers=auth_headers(token))
    profile = resp_get.json()
    assert profile.get("email") == update_payload["email"]