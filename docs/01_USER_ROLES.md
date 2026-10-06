# 01 - Vai Trò Người Dùng & Phân Quyền (User Roles & Permissions)

> **Mục tiêu:** Xác định xem công cụ có cần bảo mật, phân quyền hay cơ chế đăng nhập không. Với công cụ cá nhân/nội bộ, càng đơn giản càng tốt.

---

### 1. Cơ chế truy cập
*Hệ thống có cần đăng nhập (Authentication) không?*
- [x] **Không cần đăng nhập:** Mở lên là dùng ngay (Thích hợp cho CLI tool, Desktop app cá nhân, chạy trên localhost/mạng LAN nội bộ).

---

### 2. Danh sách vai trò (Roles)
*Điền các vai trò tham gia vào hệ thống:*

| Vai trò | Mô tả ai là người giữ vai trò này | Hành động được phép làm | Hành động bị cấm |
| :--- | :--- | :--- | :--- |
| **Admin (Chính bạn)** | Quản trị toàn bộ | Cấu hình hệ thống, chạy tác vụ, xóa dữ liệu | Không có |
| **Member (Đồng nghiệp)** | Người dùng nội bộ khác | Chỉ kích hoạt chạy, xem kết quả, tải file | Không được xóa cấu hình / tài khoản |

*(Nếu dự án chỉ có 1 mình bạn dùng, bạn có thể bỏ qua bảng trên).*
