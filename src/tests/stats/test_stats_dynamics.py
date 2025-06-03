import time

import pytest
from helpers.utils import auth_headers, make_request


@pytest.mark.asyncio
async def test_post_dynamics(api_gateway_url, login_user):
    token, _ = login_user
    headers = auth_headers(token)

    post = make_request(
        "POST", f"{api_gateway_url}/posts",
        headers={**headers, "Content-Type": "application/json"},
        data={"title":"Dynamics","description":"Dyn","is_private":False,"tags":[]}
    ).json()
    post_id = post["id"]

    for _ in range(3):
        make_request("POST", f"{api_gateway_url}/posts/{post_id}/view", headers=headers)
    make_request("POST", f"{api_gateway_url}/posts/{post_id}/like", headers=headers)
    for i in range(2):
        make_request(
            "POST", f"{api_gateway_url}/posts/{post_id}/comments",
            headers={**headers, "Content-Type": "application/json"},
            data={"text": f"C{i}"}
        )

    time.sleep(2)

    today = time.strftime("%Y-%m-%d")
    for metric, expected in (("views", 3), ("likes", 1), ("comments", 2)):
        resp = make_request(
            "GET", f"{api_gateway_url}/stats/posts/{post_id}/dynamics",
            headers=headers, params={"metric": metric}
        )
        assert resp.status_code == 200
        data = resp.json()["data"]
        counts = {d["date"]: d["count"] for d in data}
        assert counts.get(today, 0) == expected, f"{metric} динамика: ожидали {expected}, получили {counts.get(today)}"
