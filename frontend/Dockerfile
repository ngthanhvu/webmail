# syntax=docker/dockerfile:1
# Dev: Vite dev server, mount source qua volume

FROM oven/bun:1-alpine
WORKDIR /app
EXPOSE 5173
CMD ["sh", "-c", "bun install && bun run dev -- --host 0.0.0.0 --port 5173"]
