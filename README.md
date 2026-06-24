# trongcon-api

## Chạy local

1. PostgreSQL (tạo database trùng `DB_NAME` trong `.env`, ví dụ `trongcon`).
2. Copy `.env.example` → `.env` và chỉnh nếu cần.
3. Chạy API:

```bash
go mod tidy
go run ./cmd/api
```

## Swagger UI (runtime)

Sau khi server chạy, mở trình duyệt:

- **Swagger UI:** `http://localhost:8080/swagger`
- **JSON:** `http://localhost:8080/api.json`

OpenAPI được tạo runtime từ route đã register, nên thêm route mới sẽ tự xuất hiện trong Swagger mà không cần generate docs.

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
