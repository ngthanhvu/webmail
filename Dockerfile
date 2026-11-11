# syntax=docker/dockerfile:1

########################
# DEV TARGETS (docker-compose.yml)
########################

FROM golang:1.26-alpine AS backend-dev
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest
ENV GOFLAGS=-mod=mod
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
EXPOSE 8080
CMD ["sh", "-c", "go mod download || true; air -c .air.toml"]

FROM node:20-alpine AS frontend-dev
WORKDIR /app/frontend
EXPOSE 5173
CMD ["sh", "-c", "npm install --no-audit --no-fund && npm run dev -- --host 0.0.0.0 --port 5173"]
