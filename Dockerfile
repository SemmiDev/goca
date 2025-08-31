# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary (static)
RUN CGO_ENABLED=0 GOOS=linux go build -o main -ldflags '-w -s' ./cmd/api

# Stage 2: Final Image
FROM alpine:latest

WORKDIR /app

# Install timezone data only in final image
RUN apk add --no-cache tzdata

# Set timezone
ENV TZ=Asia/Jakarta
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy binary
COPY --from=builder /app/main .

# Run as non-root user
RUN adduser -D appuser
USER appuser

EXPOSE 8080

CMD ["/app/main"]
