FROM golang:1.24.0-alpine AS builder
WORKDIR /workspace
COPY go.work go.work.sum ./
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
COPY services/post-service/go.mod services/post-service/go.sum ./services/post-service/
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
RUN go work sync
COPY . .
RUN go build -o /app/user-service ./services/user-service/cmd

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/user-service .
COPY ./services/user-service/migrations ./migrations
CMD ["./user-service"]
