import time

import pytest
from helpers.utils import auth_headers, make_request


@pytest.mark.asyncio
async def test_basic_counters(api_gateway_url, login_user):
    token, user = login_user
    headers = auth_headers(token)

    post = make_request(
        "POST", f"{api_gateway_url}/posts",
        headers={**headers, "Content-Type": "application/json"},
        data={
            "title": "Basic Stats",
            "description": "Checking basic counts",
            "is_private": False,
            "tags": ["stats", "basic"]
        }
    ).json()
    post_id = post["id"]

    make_request("POST", f"{api_gateway_url}/posts/{post_id}/view", headers=headers)
    for _ in range(2):
        make_request("POST", f"{api_gateway_url}/posts/{post_id}/like", headers=headers)
    make_request("DELETE", f"{api_gateway_url}/posts/{post_id}/like", headers=headers)
    for i in range(2):
        make_request(
            "POST", f"{api_gateway_url}/posts/{post_id}/comments",
            headers={**headers, "Content-Type": "application/json"},
            data={"text": f"Comment {i}"}
        )

    time.sleep(2)

    resp = make_request("GET", f"{api_gateway_url}/stats/posts/{post_id}", headers=headers)
    assert resp.status_code == 200
    stats = resp.json()
    assert stats["views"] == 1
    assert stats["likes"] == 1
    assert stats["comments"] == 2
