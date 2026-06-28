FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

COPY . .
RUN swag init -g cmd/server/main.go
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:latest

RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /out/server .
USER app

EXPOSE 8080
CMD ["./server"]
