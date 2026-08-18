FROM golang:1.26.3-alpine AS builder
WORKDIR /build/users-service

COPY users-service/go.mod users-service/go.sum ./
RUN go mod download -x

COPY users-service ./
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags="-s -w" -o users-service

FROM alpine:3.18
WORKDIR /users-service

COPY --from=builder /build/users-service/users-service ./
CMD ["./users-service"]
