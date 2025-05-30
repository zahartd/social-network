import time

import pytest
from helpers.utils import auth_headers, make_request


@pytest.mark.asyncio
async def test_top_posts_ordering(api_gateway_url, login_user):
    token, _ = login_user
    headers = auth_headers(token)

    p1 = make_request(
        "POST", f"{api_gateway_url}/posts",
        headers={**headers, "Content-Type": "application/json"},
        data={"title": "P1", "description": "Post1", "is_private": False, "tags": []}
    ).json()
    p1_id = p1["id"]
    assert p1_id, "Failed to create P1"

    p2 = make_request(
        "POST", f"{api_gateway_url}/posts",
        headers={**headers, "Content-Type": "application/json"},
        data={"title": "P2", "description": "Post2", "is_private": False, "tags": []}
    ).json()
    p2_id = p2["id"]
    assert p2_id, "Failed to create P2"

    for _ in range(5):
        make_request("POST", f"{api_gateway_url}/posts/{p1_id}/view", headers=headers)
    for _ in range(2):
        make_request("POST", f"{api_gateway_url}/posts/{p2_id}/view", headers=headers)

    for _ in range(3):
        make_request("POST", f"{api_gateway_url}/posts/{p1_id}/like", headers=headers)
    for _ in range(4):
        make_request("POST", f"{api_gateway_url}/posts/{p2_id}/like", headers=headers)

    for _ in range(1):
        make_request(
            "POST", f"{api_gateway_url}/posts/{p1_id}/comments",
            headers={**headers, "Content-Type": "application/json"},
            data={"text": "A"}
        )
    for _ in range(2):
        make_request(
            "POST", f"{api_gateway_url}/posts/{p2_id}/comments",
            headers={**headers, "Content-Type": "application/json"},
            data={"text": "B"}
        )

    time.sleep(2)

    ordering = {
        "views":    lambda i1, i2: i1 < i2,
        "likes":    lambda i1, i2: i2 < i1,
        "comments": lambda i1, i2: i2 < i1,
    }

    for metric, compare in ordering.items():
        resp = make_request(
            "GET", f"{api_gateway_url}/stats/top/posts",
            headers=headers, params={"by": metric}
        )
        assert resp.status_code == 200, f"Top posts for '{metric}' failed: {resp.text}"
        items = resp.json().get("items", [])

        positions = {it["post_id"]: idx for idx, it in enumerate(items)}

        if p1_id in positions and p2_id in positions:
            i1, i2 = positions[p1_id], positions[p2_id]
            assert compare(i1, i2), (
                f"For metric '{metric}', expected compare({i1},{i2}) to be True, "
                f"but got positions P1={i1}, P2={i2}"
            )
