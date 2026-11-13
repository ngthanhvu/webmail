# syntax=docker/dockerfile:1
# Dev: hot-reload Go bằng air, mount source qua volume

FROM golang:1.26-alpine
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest
ENV GOFLAGS=-mod=mod
WORKDIR /app/backend
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 8080
CMD ["sh", "-c", "go mod download || true; air -c .air.toml"]
