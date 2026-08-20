FROM golang:1.26.3-alpine AS builder
WORKDIR /build/marks-service

COPY marks-service/go.mod marks-service/go.sum ./
RUN go mod download -x

COPY marks-service ./
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags="-s -w" -o marks-service

FROM alpine:3.18
WORKDIR /marks-service

COPY --from=builder /build/marks-service/marks-service ./
CMD ["./marks-service"]
