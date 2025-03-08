# Инструкции

## cURL

## Регистрация

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

## Логин

```bash
curl -G http://localhost:8080/user/login \
  --data-urlencode "login=john_doe" \
  --data-urlencode "password=Password123"
```

## Выход

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


## Сборка и запуск

```bash
docker-compose down --rmi all --volumes --remove-orphans
docker compose up --build
docker compose up migrate
