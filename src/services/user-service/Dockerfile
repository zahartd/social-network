FROM golang:1.24.3-alpine AS builder
WORKDIR /workspace
COPY . .
RUN go work sync
RUN go mod download
RUN go build -o /workspace/user-service ./services/user-service/cmd

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /workspace/user-service .
COPY ./services/user-service/migrations ./migrations
CMD ["./user-service"]
