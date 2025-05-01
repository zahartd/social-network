from helpers.utils import auth_headers, make_request


async def test_replies_listing(api_gateway_url, login_user):
    token, _ = login_user

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