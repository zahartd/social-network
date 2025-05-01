# Инструкции

## cURL

## Signup

```bash
curl -X POST http://localhost:8080/user \
  -H "Content-Type: application/json" \
  -d '{
        "login": "john_doe",
        "firstname": "John",
        "surname": "Doe",
        "email": "john@example.com",
        "password": "Password123"
      }'
```

## Login

```bash
curl -G http://localhost:8080/user/login \
  --data-urlencode "login=john_doe" \
  --data-urlencode "password=Password123"
```

## Logout

```bash
curl -X GET http://localhost:8080/user/logout \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

## Get user

```bash
curl -X PUT http://localhost:8080/user/john_doe \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
        "email": "newjohn@example.com",
        "firstname": "John",
        "surname": "Doe",
        "phone": "+1234567890",
        "bio": "Updated biography text"
      }'
```

## Delete user and all sessions

```bash
curl -X DELETE http://localhost:8080/user/john_doe \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

## Create new post

```bash
curl -X POST http://localhost:8080/posts \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Мой первый пост",
    "description": "Это содержимое поста, созданного для тестирования.",
    "is_private": false,
    "tags": ["тест", "golang", "api"]
  }'
```

## Get post by id

```bash
curl -X GET http://localhost:8080/posts/$POST_ID \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Update post

```bash
curl -X PUT http://localhost:8080/posts/$POST_ID \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Обновленный заголовок поста",
    "description": "Это обновленное описание. Пост теперь приватный!",
    "is_private": true,
    "tags": ["тест", "обновление", "приватность"]
  }'
```

## Delete post

```bash
curl -X DELETE http://localhost:8080/posts/$POST_ID \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Get list of my post (pagination)

```bash
curl -X GET 'http://localhost:8080/posts/list/my?page=1&page_size=3' \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Get list of all public posts of all users (pagination)

```bash
curl -X GET 'http://localhost:8080/posts/list/public?page=1&page_size=15' \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Get list of all public posts of user with user_id=$USER_ID (pagination)

```bash
curl -X GET "http://localhost:8080/posts/list/public/${USER_ID}?page=1&page_size=4" \
  -H "Authorization: Bearer $JWT_TOKEN"
```
## Mark post as viewed

```bash
curl -X POST http://localhost:8080/posts/$POST_ID/view \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Like / unlike a post

```bash
# like
curl -X POST http://localhost:8080/posts/$POST_ID/like \
  -H "Authorization: Bearer $JWT_TOKEN"

# unlike
curl -X DELETE http://localhost:8080/posts/$POST_ID/like \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Add a comment

```bash
curl -X POST http://localhost:8080/posts/$POST_ID/comments \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text":"First comment"}'
```

## Add a reply to a comment

```bash
curl -X POST http://localhost:8080/posts/$POST_ID/comments \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
        "parent_comment_id": "'"$PARENT_COMMENT_ID"'",
        "text": "Thanks for the feedback!"
      }'
```

## List top-level comments

```bash
curl -X GET 'http://localhost:8080/posts/$POST_ID/comments?page=1&page_size=10' \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## List replies for a comment

```bash
curl -X GET http://localhost:8080/posts/$POST_ID/comments/$PARENT_COMMENT_ID/replies \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Kafka events

Enjoy the API and keep an eye on Kafka topics at http://localhost:8082

## Build and run

```bash
protoc --proto_path=src/proto --go_out=src/gen/go --go_opt=paths=source_relative --go-grpc_out=src/gen/go --go-grpc_opt=paths=source_relative src/proto/post/post.proto
docker compose down --rmi all --volumes --remove-orphans
docker compose up --build
docker compose up migrate
docker system prune -af --volumes
