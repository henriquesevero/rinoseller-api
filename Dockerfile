# Build compartilhado por todos os targets de execução (runtime e lambda) —
# o binário compilado é idêntico; o que muda entre Lambda, ECS, EC2 e App
# Runner é só a camada final do container, nunca o código Go.
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

COPY . .
RUN swag init -g cmd/server/main.go
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

# ── Target "lambda": AWS Lambda (container image) ───────────────────────────
# Mesma imagem/binário do target runtime, com o AWS Lambda Web Adapter
# adicionado como extension — ele traduz eventos de invocação do Lambda em
# requisições HTTP normais contra o servidor Go, que continua sem nenhuma
# dependência do SDK/eventos do Lambda.
# docker build --target lambda -t rinoseller-api:lambda .
FROM alpine:latest AS lambda

COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter
ENV AWS_LWA_PORT=8080

WORKDIR /app
COPY --from=builder /out/server .

EXPOSE 8080
CMD ["./server"]

# ── Target "runtime": EC2 / ECS / Fargate / App Runner / Railway ────────────
# Estágio final por padrão (sem precisar de --target) — builders como o do
# Railway usam o último estágio do Dockerfile quando nenhum target é informado.
# docker build --target runtime -t rinoseller-api:runtime .
FROM alpine:latest AS runtime

RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /out/server .
USER app

EXPOSE 8080
CMD ["./server"]
