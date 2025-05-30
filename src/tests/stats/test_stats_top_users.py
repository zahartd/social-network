import time

import pytest
from helpers.utils import auth_headers, make_request


@pytest.mark.asyncio
async def test_top_users_ordering(api_gateway_url, user_factory):
    token1, _ = user_factory()
    token2, _ = user_factory()
    h1, h2 = auth_headers(token1), auth_headers(token2)

    p1 = make_request(
        "POST", f"{api_gateway_url}/posts",
        headers={**h1, "Content-Type": "application/json"},
        data={"title": "X", "description": "P1", "is_private": False, "tags": []}
    ).json()
    p2 = make_request(
        "POST", f"{api_gateway_url}/posts",
        headers={**h2, "Content-Type": "application/json"},
        data={"title": "Y", "description": "P2", "is_private": False, "tags": []}
    ).json()

    uid_field = "user_id" if "user_id" in p1 else "userId"
    u1, u2 = p1[uid_field], p2[uid_field]
    assert u1 and u2

    for _ in range(3):
        make_request("POST", f"{api_gateway_url}/posts/{p1['id']}/view", headers=h1)
    for _ in range(5):
        make_request("POST", f"{api_gateway_url}/posts/{p2['id']}/view", headers=h2)

    for _ in range(2):
        make_request("POST", f"{api_gateway_url}/posts/{p1['id']}/like", headers=h1)
    for _ in range(1):
        make_request("POST", f"{api_gateway_url}/posts/{p2['id']}/like", headers=h2)

    for _ in range(1):
        make_request(
            "POST", f"{api_gateway_url}/posts/{p1['id']}/comments",
            headers={**h1, "Content-Type": "application/json"}, data={"text": "C1"}
        )
    for _ in range(4):
        make_request(
            "POST", f"{api_gateway_url}/posts/{p2['id']}/comments",
            headers={**h2, "Content-Type": "application/json"}, data={"text": "C2"}
        )

    time.sleep(2)

    compare_map = {
        "views":    lambda i1, i2: i2 < i1,
        "likes":    lambda i1, i2: i1 < i2,
        "comments": lambda i1, i2: i2 < i1,
    }

    for by, cmp_fn in compare_map.items():
        resp = make_request(
            "GET", f"{api_gateway_url}/stats/top/users",
            headers=h1, params={"by": by}
        )
        assert resp.status_code == 200, f"Top users for '{by}' failed: {resp.text}"
        items = resp.json().get("items", [])

        positions = {it["user_id"]: idx for idx, it in enumerate(items)}

        if u1 in positions and u2 in positions:
            i1, i2 = positions[u1], positions[u2]
            assert cmp_fn(i1, i2), (
                f"For metric '{by}', expected compare({i1},{i2}) to be True, "
                f"got positions u1={i1}, u2={i2}"
            )
