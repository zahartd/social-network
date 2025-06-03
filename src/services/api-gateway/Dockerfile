FROM golang:1.24.3-alpine AS builder
WORKDIR /workspace
COPY . .
RUN go work sync
RUN go mod download
RUN go build -o /workspace/api-gateway ./services/api-gateway/cmd

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates curl
COPY --from=builder /workspace/api-gateway .
CMD ["./api-gateway"]
