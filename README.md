# trongcon-api

## Chạy local

1. PostgreSQL (tạo database trùng `DB_NAME` trong `.env`, ví dụ `trongcon`).
2. Copy `.env.example` → `.env` và chỉnh nếu cần.
3. Chạy API:

```bash
go mod tidy
go run ./cmd/api
```

## Swagger UI

Sau khi server chạy, mở trình duyệt:

- **Swagger UI:** `http://localhost:8080/swagger/index.html`
- **JSON:** `http://localhost:8080/swagger/doc.json`

Khi đổi comment `@Summary`, struct API hoặc thêm handler, sinh lại tài liệu:

```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

## Endpoint gợi ý

| Method | Path | Mô tả |
|--------|------|--------|
| GET | `/` | Trạng thái API + link docs |
| GET | `/api/v1/health` | Health check |
| POST | `/api/v1/admin/login` | Đăng nhập admin (chỉ role `super`) → JWT |
| POST | `/api/v1/user/signup` | Đăng ký user (role `user`) → JWT |
| POST | `/api/v1/user/login` | Đăng nhập user → JWT |
| POST | `/api/v1/admin/users` | Tạo user — cần `Authorization: Bearer` (super) |
| GET | `/api/v1/admin/users` | Danh sách |
| GET | `/api/v1/admin/users/:id` | Chi tiết |
| PUT | `/api/v1/admin/users/:id` | Cập nhật |
| DELETE | `/api/v1/admin/users/:id` | Xóa (soft delete) |

Biến môi trường: `JWT_SECRET`, `JWT_EXPIRE_HOURS` (xem `.env.example`).
