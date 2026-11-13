# webmail frontend (Vue 3 + TypeScript)

Webmail client: IMAP đọc thư, SMTP gửi thư. Viết lại từ bản JavaScript cũ, giờ dùng TypeScript đầy đủ (vue-tsc kiểm tra cả template).

## Stack

- Bun (engine + package manager, lockfile `bun.lock`)
- Vue 3.5 `<script setup lang="ts">`
- TypeScript strict (`vue-tsc -b` chạy trước khi build)
- Vite 8 (dev proxy `/api` → backend)
- Pinia 3 (stores setup-style)
- Vue Router 4 (`RouteMeta` typed)
- Tailwind CSS 4 — cấu hình ngay trong `src/style.css` bằng `@theme` (không có `tailwind.config`), design token `oc-*` qua CSS variables (dark mode class)

## Cấu trúc

```
src/
  api.ts              # client API có kiểu (khớp JSON của backend Go)
  types.ts            # domain types: Mailbox, MessageSummary, Message, Draft, SendPayload...
  utils/format.ts     # format ngày/thư/tên hộp thư
  stores/auth.ts      # phiên đăng nhập
  stores/theme.ts     # dark/light
  router/index.ts     # route + guard auth
  components/         # Sidebar, MessageList, MessageDetail, ComposeModal, ConfirmDialog, MdiIcon
  views/              # LoginView, MailView
```

## Lệnh

```sh
bun install       # dùng Bun làm engine (lockfile: bun.lock)
bun run dev       # dev server, proxy /api tới VITE_PROXY_TARGET (mặc định http://localhost:8080)
bun run typecheck # vue-tsc
bun run build     # vue-tsc -b && vite build -> dist/
```

> Node vẫn chạy được (`npm run dev`...) vì không có dependency native, nhưng repo này chuẩn hoá theo Bun: lockfile, Docker image (`oven/bun`) đều dùng bun.

## Ghi chú

- Alias import `@/` trỏ tới `src/`.
- Theme áp trước khi render (script inline trong `index.html`) để tránh flash.
- Thư HTML render trong `iframe sandbox=""`, đính kèm tải qua `/api/messages/:uid/attachment`.
