# Stage 1: Builder
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go.mod và go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ source code (bao gồm cmd/, internal/, v.v.)
COPY . .

# Build binary
RUN go build -o blog-api ./cmd/api

# Stage 2: Runtime
FROM debian:bookworm-slim AS runner

WORKDIR /app
COPY --from=builder /app/blog-api .

EXPOSE 8080
CMD ["./blog-api"]
