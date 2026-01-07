FROM golang:1.25.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o rate-limiter ./cmd/api

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/rate-limiter .

EXPOSE 8080

CMD ["./rate-limiter"]