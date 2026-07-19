# trongcon-api

## Chạy local

1. PostgreSQL — có thể dùng Docker:

```bash
docker compose up -d
```

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
| GET | `/api/v1/exercises` | Danh sách bài tập (public, chỉ `active`) |
| GET | `/api/v1/exercises/:slug` | Chi tiết bài tập theo slug (public) |
| GET | `/api/v1/muscles` | Danh sách nhóm cơ (public) |
| GET | `/api/v1/muscles/:id` | Chi tiết nhóm cơ (public) |
| GET | `/api/v1/equipments` | Danh sách dụng cụ (public) |
| GET | `/api/v1/equipments/:id` | Chi tiết dụng cụ (public) |
| POST | `/api/v1/tools/tdee` | Tính TDEE (public) |
| POST | `/api/v1/tools/macros` | Tính macros (public) |
| POST | `/api/v1/tools/one-rep-max` | Tính 1RM (public) |
| GET | `/api/v1/workouts` | Danh sách buổi tập mẫu (public) |
| GET | `/api/v1/workouts/:id` | Chi tiết buổi tập (public) |
| GET | `/api/v1/routines` | Danh sách chương trình công khai (`is_public=true`) |
| GET | `/api/v1/routines/:id` | Chi tiết chương trình công khai |
| GET | `/api/v1/articles` | Danh sách bài viết (category `active`) |
| GET | `/api/v1/articles/:slug` | Chi tiết bài viết theo slug |
| GET | `/api/v1/categories` | Danh sách danh mục (`status=active`) |
| GET | `/api/v1/categories/:id` | Chi tiết danh mục active |
| GET | `/api/v1/foods` | Danh sách thực phẩm (public) |
| GET | `/api/v1/foods/:id` | Chi tiết thực phẩm (public) |
| GET | `/api/v1/meal-plans` | Meal plans công khai (`is_public=true`) |
| GET | `/api/v1/meal-plans/:id` | Chi tiết meal plan công khai |
| GET | `/api/v1/admin/stats/overview` | Thống kê tổng quan (super) |
| POST | `/api/v1/admin/login` | Đăng nhập admin (chỉ role `super`) → JWT |
| POST | `/api/v1/user/signup` | Đăng ký user (role `user`) → JWT |
| POST | `/api/v1/user/login` | Đăng nhập user → JWT |
| POST | `/api/v1/admin/users` | Tạo user — cần `Authorization: Bearer` (super) |
| GET | `/api/v1/admin/users` | Danh sách |
| GET | `/api/v1/admin/users/:id` | Chi tiết |
| PUT | `/api/v1/admin/users/:id` | Cập nhật |
| DELETE | `/api/v1/admin/users/:id` | Xóa (soft delete) |

Biến môi trường: `JWT_SECRET`, `JWT_EXPIRE_HOURS` (xem `.env.example`).
