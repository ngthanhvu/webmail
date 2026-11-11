# Webmail

Webmail tự dựng (kiểu Gmail) đọc/gửi mail qua **IMAP/SMTP** có sẵn.

- **Backend:** Go (`net/http`, `emersion/go-imap`)
- **Frontend:** Vue 3 + Vite + TailwindCSS + Pinia + Vue Router
- **Chạy code:** Docker / docker-compose
  - Dev: `Dockerfile` + `docker-compose.yml` (hot-reload Go bằng `air`, Vite dev server)
  - Prod: `Dockerfile.prod` + `docker-compose.prod.yml` (1 image duy nhất, Go phục vụ cả API lẫn frontend)

## Tính năng

- Đăng nhập bằng tài khoản email thật (backend xác thực qua IMAP, mật khẩu không lưu trên server — chỉ lưu trong cookie đã mã hoá AES-GCM bằng `APP_SECRET`)
- Danh sách hộp thư + số thư chưa đọc (INBOX, Đã gửi, Thùng rác, ... lấy từ server IMAP)
- Danh sách thư: người gửi, tiêu đề, đoạn xem trước, ngày, trạng thái đã đọc / đánh dấu sao, phân trang
- Đọc thư: text hoặc HTML (render trong iframe sandbox), tải tệp đính kèm
- Đánh dấu sao / chưa đọc, xóa (chuyển vào Thùng rác khi có thể)
- Soạn thư mới / Trả lời / Trả lời tất cả / Chuyển tiếp, gửi kèm tệp đính kèm
- Gửi qua SMTP, tự lưu bản sao vào hộp "Đã gửi" (nếu tìm thấy)
- Tìm kiếm nhanh trong hộp thư đang mở

## Cấu trúc

```
webmail/
├── backend/                  # Go API
│   ├── main.go
│   └── internal/
│       ├── config/           # đọc biến môi trường
│       ├── session/          # cookie phiên đăng nhập (AES-GCM)
│       ├── imapx/            # thao tác IMAP (list, fetch, flags, append)
│       ├── mimex/            # parse/build MIME (text, html, đính kèm, charset)
│       ├── smtpx/            # gửi SMTP (starttls/tls/plain, PLAIN/LOGIN auth)
│       └── server/           # HTTP routes + handlers + static files
├── frontend/                 # Vue 3 + Tailwind
│   └── src/
│       ├── views/            # LoginView, MailView
│       ├── components/       # Sidebar, MessageList, MessageDetail, ComposeModal
│       ├── stores/           # pinia auth
│       ├── api.js, router.js, format.js
├── Dockerfile                # dev targets (backend-dev, frontend-dev)
├── docker-compose.yml        # chạy dev
├── Dockerfile.prod           # image prod (multi-stage)
├── docker-compose.prod.yml   # chạy prod
└── .env.example              # mẫu cấu hình IMAP/SMTP
```

## API (tóm tắt)

| Method | Path | Mô tả |
|---|---|---|
| POST | `/api/login` | đăng nhập (body `{username, password}`) |
| POST | `/api/logout` | đăng xuất |
| GET | `/api/me` | thông tin phiên |
| GET | `/api/mailboxes` | danh sách hộp thư + số chưa đọc |
| GET | `/api/messages?mailbox=INBOX&page=0&limit=50` | danh sách thư |
| GET | `/api/messages/{uid}?mailbox=INBOX` | nội dung 1 thư (tự đánh dấu đã đọc) |
| GET | `/api/messages/{uid}/attachment?mailbox=..&section=2` | tải đính kèm |
| POST | `/api/messages/{uid}/flags?mailbox=..` | `{seen?, flagged?, deleted?}` |
| POST | `/api/send` | gửi thư (to/cc/bcc, text/html, attachments) |

## Chạy DEV

```bash
cp .env.example .env
# sửa IMAP_HOST/SMTP_HOST... trong .env

docker compose up --build
```

- UI: http://localhost:5273 (Vite, sửa code tự reload; Vite proxy `/api` -> backend)
- API: http://localhost:8181

> Port mặc định là `8181` (API) và `5273` (UI dev). Đổi trong `docker-compose.yml` nếu trùng.

## Chạy PROD

```bash
cp .env.example .env
# sửa cấu hình, ĐỔI APP_SECRET thành chuỗi ngẫu nhiên dài

docker compose -f docker-compose.prod.yml up --build -d
```

- Truy cập: http://localhost:8181 (Go phục vụ API + frontend tĩnh)

## Cấu hình (.env)

Xem `.env.example`. Các biến chính:

| Biến | Ý nghĩa |
|---|---|
| `IMAP_HOST`, `IMAP_PORT`, `IMAP_MODE` | máy chủ IMAP (993/tls, 143/starttls) |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_MODE` | máy chủ SMTP (587/starttls, 465/tls) |
| `SMTP_USER`/`SMTP_PASS`, `SMTP_FROM` | để trống nếu dùng tài khoản đăng nhập UI |
| `APP_SECRET` | khoá mã hoá cookie phiên — **đổi khi chạy thật** |
| `MAILBOX_SENT`, `MAILBOX_TRASH` | hộp lưu thư đã gửi / thùng rác |

## Lưu ý bảo mật

- Chạy prod nên đặt trước reverse proxy có HTTPS (nginx/caddy/traefik).
- `APP_SECRET` giữ bí tuyệt đối — ai có nó + cookie có thể giải mã mật khẩu trong phiên.
- `IMAP_SKIP_VERIFY`/`SMTP_SKIP_VERIFY` chỉ dùng khi test với self-signed cert.
