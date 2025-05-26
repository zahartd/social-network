import time

import pytest
from helpers.utils import auth_headers, make_request


@pytest.mark.asyncio
async def test_post_stats_and_dynamics_and_top(api_gateway_url, login_user):
    token, user = login_user
    headers = auth_headers(token)
    # 1) создаём пост
    post = make_request(
        "POST", f"{api_gateway_url}/posts",
        headers={**headers, "Content-Type": "application/json"},
        data={
            "title": "Stats Test",
            "description": "Testing stats endpoints",
            "is_private": False,
            "tags": ["stats", "test"]
        }
    ).json()
    post_id = post["id"]

    # 2) генерируем события: 2 просмотра, 1 лайк, 1 отмена лайка, 3 комментария
    for _ in range(2):
        make_request("POST", f"{api_gateway_url}/posts/{post_id}/view", headers=headers)
    make_request("POST", f"{api_gateway_url}/posts/{post_id}/like", headers=headers)
    make_request("DELETE", f"{api_gateway_url}/posts/{post_id}/like", headers=headers)
    for i in range(3):
        make_request(
            "POST", f"{api_gateway_url}/posts/{post_id}/comments",
            headers={**headers, "Content-Type": "application/json"},
            data={"text": f"Comment {i}"}
        )

    # ждём, чтобы статистика успела запуститься
    time.sleep(1)

    # 3) Проверяем общий статистический эндпойнт
    resp = make_request("GET", f"{api_gateway_url}/stats/posts/{post_id}", headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    # просмотры = 2, лайки = 0 (1 like + 1 unlike), комментарии = 3
    assert data["views"] == 2, f"expected 2 views, got {data['views']}"
    assert data["likes"] == 0, f"expected 0 likes, got {data['likes']}"
    assert data["comments"] == 3, f"expected 3 comments, got {data['comments']}"

    # 4) Динамика просмотров
    resp = make_request(
        "GET", f"{api_gateway_url}/stats/posts/{post_id}/dynamics",
        headers=headers,
        params={"metric": "views"}
    )
    assert resp.status_code == 200
    dyn = resp.json()["data"]
    assert isinstance(dyn, list) and dyn, "Expected non-empty dynamics for views"
    # одна запись за сегодня с count=2
    today = time.strftime("%Y-%m-%d")
    assert any(d["date"] == today and d["count"] == 2 for d in dyn)

    # 5) Динамика лайков
    resp = make_request(
        "GET", f"{api_gateway_url}/stats/posts/{post_id}/dynamics",
        headers=headers,
        params={"metric": "likes"}
    )
    assert resp.status_code == 200
    dyn_l = resp.json()["data"]
    # должна быть запись за сегодня, count = (1 like + 0 после unlike) = 1?
    assert any(d["date"] == today and d["count"] == 1 for d in dyn_l)

    # 6) Динамика комментариев
    resp = make_request(
        "GET", f"{api_gateway_url}/stats/posts/{post_id}/dynamics",
        headers=headers,
        params={"metric": "comments"}
    )
    assert resp.status_code == 200
    dyn_c = resp.json()["data"]
    assert any(d["date"] == today and d["count"] == 3 for d in dyn_c)

    # 7) Топ-посты по каждому метрику
    for by in ("views", "likes", "comments"):
        resp = make_request(
            "GET", f"{api_gateway_url}/stats/top/posts",
            headers=headers,
            params={"by": by}
        )
        assert resp.status_code == 200
        items = resp.json()["items"]
        assert isinstance(items, list) and items, f"Expected non-empty top posts for {by}"
        # текущий post_id должен быть в списке, хотя для likes count=0, но он по-прежнему может быть среди топ-10
        assert any(item["post_id"] == post_id for item in items)

    # 8) Топ-пользователи по каждому метрику
    user_id = user["id"]
    for by in ("views", "likes", "comments"):
        resp = make_request(
            "GET", f"{api_gateway_url}/stats/top/users",
            headers=headers,
            params={"by": by}
        )
        assert resp.status_code == 200
        items = resp.json()["items"]
        assert isinstance(items, list) and items, f"Expected non-empty top users for {by}"
        # текущий юзер точно будет в топ-10 по просмотрам и комментариям
        assert any(item["user_id"] == user_id for item in items)
