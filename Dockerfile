FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o blog-api ./cmd/api

FROM debian:bookworm-slim AS runner

WORKDIR /app
COPY --from=builder /app/blog-api .

EXPOSE 8080
CMD ["./blog-api"]
