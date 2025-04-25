import json
import logging
import uuid
import pytest
import requests

LOGGER = logging.getLogger(__name__)

def make_request(method, url, params=None, data=None, headers=None):
    if data is not None:
        data = json.dumps(data)
    LOGGER.info(f"Request: {method} {url} params: {params} data: {data} headers: {headers}")
    resp = requests.request(method, url, params=params, data=data, headers=headers)
    LOGGER.info(f"Response: {resp.status_code} {resp.text}")
    return resp

def auth_headers(token):
    return {"Authorization": f"Bearer {token}"}

@pytest.fixture
def unique_user():
    unique = str(uuid.uuid4())[:8]
    user_data = {
        "login": f"user_{unique}",
        "firstname": f"First_{unique}",
        "surname": f"Last_{unique}",
        "email": f"user_{unique}@example.com",
        "password": "TestPass123"
    }
    return user_data

@pytest.fixture
def register_user(api_gateway_url, unique_user):
    url = api_gateway_url + "/user"
    resp = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
    assert resp.status_code == 201, f"Регистрация не удалась: {resp.text}"
    unique_user["id"] = resp.json()["user"]["id"]
    return unique_user

@pytest.fixture
def login_user(api_gateway_url, register_user):
    login = register_user["login"]
    password = register_user["password"]
    url = api_gateway_url + "/user/login"
    params = {"login": login, "password": password}
    resp = make_request("GET", url, params=params)
    assert resp.status_code == 200, f"Логин не удался: {resp.text}"
    data = resp.json()
    token = data.get("token")
    assert token
    return token, register_user

# --- Тесты для API пользователя ---

class TestUserAPI:
    def test_signup_and_login(self, api_gateway_url, unique_user):
        # Регистрация
        url = api_gateway_url + "/user"
        resp = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
        assert resp.status_code == 201, f"Ошибка регистрации: {resp.text}"
        data = resp.json()
        user_data = data.get("user")
        assert user_data
        assert user_data.get("login") == unique_user["login"]

        # Логин
        url = api_gateway_url + "/user/login"
        params = {"login": unique_user["login"], "password": unique_user["password"]}
        resp = make_request("GET", url, params=params)
        assert resp.status_code == 200, f"Ошибка логина: {resp.text}"
        data = resp.json()
        token = data.get("token")
        assert token

    def test_duplicate_signup(self, api_gateway_url, unique_user):
        url = api_gateway_url + "/user"
        # Первая регистрация
        resp = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
        assert resp.status_code == 201
        # Повторная регистрация должна завершиться ошибкой
        resp2 = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
        assert resp2.status_code != 201, "Дублирование регистрации прошло успешно, хотя должно быть отказано"

    def test_logout(self, api_gateway_url, login_user):
        token, user_data = login_user
        url = api_gateway_url + "/user/logout"
        headers = auth_headers(token)
        resp = make_request("GET", url, headers=headers)
        # Допустим, успешный выход возвращает 200 или 204
        assert resp.status_code in [200, 204]

    def test_get_user_profile(self, api_gateway_url, login_user):
        token, user_data = login_user
        identifier = user_data["login"]
        url = api_gateway_url + f"/user/{identifier}"
        headers = auth_headers(token)
        resp = make_request("GET", url, headers=headers)
        assert resp.status_code == 200, f"Ошибка получения профиля: {resp.text}"
        profile = resp.json()
        # Если JWT принадлежит пользователю, ожидаем полный профиль с email
        assert profile.get("email") == user_data["email"]

    def test_update_user_profile(self, api_gateway_url, login_user):
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
        # Проверка обновлённого профиля
        resp_get = make_request("GET", url, headers=auth_headers(token))
        profile = resp_get.json()
        assert profile.get("email") == update_payload["email"]

    def test_delete_user(self, api_gateway_url, login_user):
        token, user_data = login_user
        identifier = user_data["login"]
        url = api_gateway_url + f"/user/{identifier}"
        headers = auth_headers(token)
        resp = make_request("DELETE", url, headers=headers)
        assert resp.status_code in [200, 204], f"Ошибка удаления пользователя: {resp.text}"
        # После удаления профиль должен отсутствовать
        # new_token, new_user_data = login_user
        # new_identifier = new_user_data["login"]
        # new_url = api_gateway_url + f"/user/{new_identifier}"
        # new_headers = auth_headers(new_token)
        # resp_get = make_request("GET", new_url, headers=new_headers)
        # assert resp_get.status_code == 404

    def test_get_profile_with_invalid_token(self, api_gateway_url, unique_user):
        # Регистрируем пользователя
        url = api_gateway_url + "/user"
        resp = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
        assert resp.status_code == 201
        identifier = unique_user["login"]
        # Запрос с невалидным токеном
        url = api_gateway_url + f"/user/{identifier}"
        headers = auth_headers("invalidtoken")
        resp = make_request("GET", url, headers=headers)
        assert resp.status_code in [401, 403]

# --- Тесты для API постов ---

class TestPostAPI:
    @pytest.fixture
    def created_post(self, api_gateway_url, login_user):
        token, user_data = login_user
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
        return post, token, user_data

    def test_create_post(self, created_post):
        post, token, user_data = created_post
        assert post.get("id") is not None
        assert post.get("title") == "Мой первый пост"

    def test_get_post_by_id(self, api_gateway_url, created_post):
        post, token, user_data = created_post
        post_id = post.get("id")
        url = api_gateway_url + f"/posts/{post_id}"
        headers = auth_headers(token)
        resp = make_request("GET", url, headers=headers)
        assert resp.status_code == 200, f"Ошибка получения поста: {resp.text}"
        post_data = resp.json()
        assert post_data.get("id") == post_id

    def test_update_post(self, api_gateway_url, created_post):
        post, token, user_data = created_post
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

    def test_delete_post(self, api_gateway_url, created_post):
        post, token, user_data = created_post
        post_id = post.get("id")
        url = api_gateway_url + f"/posts/{post_id}"
        headers = auth_headers(token)
        resp = make_request("DELETE", url, headers=headers)
        assert resp.status_code in [200, 204], f"Ошибка удаления поста: {resp.text}"
        # После удаления, повторный запрос должен вернуть 404
        resp_get = make_request("GET", url, headers=headers)
        assert resp_get.status_code == 404

    def test_list_my_posts(self, api_gateway_url, login_user):
        token, user_data = login_user
        # Создаём несколько постов
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

    def test_list_all_public_posts(self, api_gateway_url, login_user):
        token, user_data = login_user
        url = api_gateway_url + "/posts/list/public"
        params = {"page": 1, "page_size": 15}
        headers = auth_headers(token)
        resp = make_request("GET", url, params=params, headers=headers)
        assert resp.status_code == 200, f"Ошибка получения публичных постов: {resp.text}"
        list_data = resp.json()
        assert "posts" in list_data

    def test_list_public_posts_by_user(self, api_gateway_url, login_user, created_post):
        post, token, user_data = created_post
        target_user_id = user_data["id"]
        token, user_data = login_user
        # В качестве идентификатора используем login пользователя
        identifier = user_data["id"]
        url = api_gateway_url + f"/posts/list/public/{target_user_id}"
        params = {"page": 1, "page_size": 4}
        headers = auth_headers(token)
        resp = make_request("GET", url, params=params, headers=headers)
        assert resp.status_code == 200, f"Ошибка получения публичных постов пользователя: {resp.text}"
        list_data = resp.json()
        assert "posts" in list_data
