import json
import logging
import uuid
import pytest
import requests

from kafka_utils import wait_for_kafka

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

@pytest.fixture
def user_factory(api_gateway_url):
    def _create():
        uniq = str(uuid.uuid4())[:8]
        user_data = {
            "login":    f"user_{uniq}",
            "firstname":f"First_{uniq}",
            "surname":  f"Last_{uniq}",
            "email":    f"user_{uniq}@example.com",
            "password": "TestPass123"
        }
        resp = make_request(
            "POST",
            api_gateway_url + "/user",
            headers={"Content-Type": "application/json"},
            data=user_data
        )
        assert resp.status_code == 201, f"Регистрация упала: {resp.text}"
        resp = make_request(
            "GET",
            api_gateway_url + "/user/login",
            params={"login": user_data["login"], "password": user_data["password"]}
        )
        assert resp.status_code == 200, f"Логин не удался: {resp.text}"
        data = resp.json()
        token = data.get("token")
        assert token
        return token, user_data
    return _create

class TestUserAPI:
    def test_signup_and_login(self, api_gateway_url, unique_user):
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

    def test_duplicate_signup(self, api_gateway_url, unique_user):
        url = api_gateway_url + "/user"
        resp = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
        assert resp.status_code == 201
        resp2 = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
        assert resp2.status_code != 201, "Дублирование регистрации прошло успешно, хотя должно быть отказано"

    def test_logout(self, api_gateway_url, login_user):
        token, user_data = login_user
        url = api_gateway_url + "/user/logout"
        headers = auth_headers(token)
        resp = make_request("GET", url, headers=headers)
        assert resp.status_code in [200, 204]

    def test_get_user_profile(self, api_gateway_url, login_user):
        token, user_data = login_user
        identifier = user_data["login"]
        url = api_gateway_url + f"/user/{identifier}"
        headers = auth_headers(token)
        resp = make_request("GET", url, headers=headers)
        assert resp.status_code == 200, f"Ошибка получения профиля: {resp.text}"
        profile = resp.json()
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
        resp_get = make_request("GET", url, headers=auth_headers(token))
        profile = resp_get.json()
        assert profile.get("email") == update_payload["email"]

    def test_delete_user(self, api_gateway_url, user_factory):
        token1, user1 = user_factory()
        token2, user2 = user_factory()
        assert user1 != user2

        url = api_gateway_url + f"/user/{user1["login"]}"
        resp = make_request("DELETE", url, headers=auth_headers(token1))
        assert resp.status_code in [200, 204], f"Ошибка удаления пользователя: {resp.text}"

        resp_get = make_request("GET", url, headers=auth_headers(token2))
        assert resp_get.status_code == 404

    def test_get_profile_with_invalid_token(self, api_gateway_url, unique_user):
        url = api_gateway_url + "/user"
        resp = make_request("POST", url, data=unique_user, headers={"Content-Type": "application/json"})
        assert resp.status_code == 201
        identifier = unique_user["login"]
        url = api_gateway_url + f"/user/{identifier}"
        headers = auth_headers("invalidtoken")
        resp = make_request("GET", url, headers=headers)
        assert resp.status_code in [401, 403]

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
        resp_get = make_request("GET", url, headers=headers)
        assert resp_get.status_code == 404

    def test_list_my_posts(self, api_gateway_url, login_user):
        token, user_data = login_user
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
        url = api_gateway_url + f"/posts/list/public/{target_user_id}"
        params = {"page": 1, "page_size": 4}
        headers = auth_headers(token)
        resp = make_request("GET", url, params=params, headers=headers)
        assert resp.status_code == 200, f"Ошибка получения публичных постов пользователя: {resp.text}"
        list_data = resp.json()
        assert "posts" in list_data

class TestPostKafkaIntegration:
    def test_user_registration_emits_event(
        self, api_gateway_url, unique_user, kafka_consumer
    ):
        # Регистрируем юзера
        resp = make_request(
            "POST",
            f"{api_gateway_url}/user",
            headers={"Content-Type": "application/json"},
            data=unique_user,
        )
        assert resp.status_code == 201
        user_id = resp.json()["user"]["id"]

        # Ждём событие в топике user-registrations
        ok = wait_for_kafka(
            kafka_consumer,
            topic="user-registrations",
            predicate=lambda m: user_id in m.value,
        )
        assert ok, "Событие user-registrations не найдено"

    def test_view_post_emits_event(self, api_gateway_url, login_user, kafka_consumer):
        token, _ = login_user
        post_id = make_request(
            "POST", f"{api_gateway_url}/posts",
            headers={**auth_headers(token),"Content-Type":"application/json"},
            data={"title":"t","description":"d","is_private":False,"tags":[]}
        ).json()["id"]

        make_request("POST", f"{api_gateway_url}/posts/{post_id}/view",
                     headers=auth_headers(token))

        # ждём сообщение в Kafka
        ok = wait_for_kafka(
            kafka_consumer,
            topic="post-views",
            predicate=lambda m: post_id in m.value
        )
        assert ok, "Событие post-views так и не прилетело"

    def test_like_and_unlike_emits_event(self, api_gateway_url, login_user, kafka_consumer):
        token, _ = login_user
        post_id = make_request(
            "POST", f"{api_gateway_url}/posts",
            headers={**auth_headers(token),"Content-Type":"application/json"},
            data={"title":"t","description":"d","is_private":False,"tags":[]}
        ).json()["id"]

        make_request("POST",    f"{api_gateway_url}/posts/{post_id}/like", headers=auth_headers(token))
        make_request("DELETE",  f"{api_gateway_url}/posts/{post_id}/like", headers=auth_headers(token))

        ok = wait_for_kafka(
            kafka_consumer,
            topic="post-likes",
            predicate=lambda m: post_id in m.value
        )
        assert ok, "Событие post-likes не найдено"

    def test_comment_and_list_emits_event(self, api_gateway_url, login_user, kafka_consumer):
        token, _ = login_user
        post_id = make_request(
            "POST", f"{api_gateway_url}/posts",
            headers={**auth_headers(token),"Content-Type":"application/json"},
            data={"title":"t","description":"d","is_private":False,"tags":[]}
        ).json()["id"]

        comment_text = "hello kafka"
        resp = make_request(
            "POST", f"{api_gateway_url}/posts/{post_id}/comments",
            headers={**auth_headers(token),"Content-Type":"application/json"},
            data={"text": comment_text}
        )
        comment_id = resp.json()["comment"]["id"]

        ok = wait_for_kafka(
            kafka_consumer,
            topic="post-comments",
            predicate=lambda m: comment_id in m.value and comment_text in m.value
        )
        assert ok, "Событие post-comments не найдено"
    
    def test_replies_listing(self, api_gateway_url, login_user):
        token, _ = login_user
        # создаём пост + родительский комментарий
        post_id = make_request(
            "POST", f"{api_gateway_url}/posts",
            headers={**auth_headers(token),"Content-Type":"application/json"},
            data={"title":"t","description":"d","is_private":False,"tags":[]}
        ).json()["id"]
        parent_id = make_request(
            "POST", f"{api_gateway_url}/posts/{post_id}/comments",
            headers={**auth_headers(token),"Content-Type":"application/json"},
            data={"text":"parent"}
        ).json()["comment"]["id"]

        # добавляем два ответа
        for txt in ("r1","r2"):
            resp = make_request(
                "POST", f"{api_gateway_url}/posts/{post_id}/comments",
                headers={**auth_headers(token),"Content-Type":"application/json"},
                data={"parent_comment_id":parent_id,"text":txt}
            )
            assert resp.status_code == 201
            comment = resp.json()["comment"]
            assert "parent_comment_id" in comment
            assert comment["parent_comment_id"] == parent_id

        # проверяем ListReplies
        resp = make_request(
            "GET", f"{api_gateway_url}/posts/{post_id}/comments/{parent_id}/replies",
            headers=auth_headers(token)
        )
        assert resp.status_code == 200

        data = resp.json()
        assert "comments" in data
        replies_comments = data["comments"]
        assert replies_comments
        texts = {c["text"] for c in replies_comments}
        assert {"r1","r2"} <= texts
