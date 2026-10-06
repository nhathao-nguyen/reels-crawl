# Hướng Dẫn & Cấu Trúc Tài Liệu Yêu Cầu Dự Án (BA Kit)

Bộ tài liệu này được thiết kế theo dạng **Lean / Agile Spec** tinh gọn, chuyên biệt cho dự án nội bộ và sử dụng cá nhân. Mục tiêu là giúp bạn làm rõ luồng tư duy, tránh lan man, và có đầy đủ thông tin để bắt tay vào code hoặc nhờ AI sinh code chính xác.

---

## 📁 Cấu Trúc Tài Liệu (Thư mục `docs/`)

| File | Mục đích | Trạng thái |
| :--- | :--- | :---: |
| [00_OVERVIEW.md](file:///d:/nhathao/Vibe/reels-crawl/docs/00_OVERVIEW.md) | Bài toán, mục tiêu, lý do làm và tiêu chí hoàn thành | Đã hoàn thiện |
| [01_USER_ROLES.md](file:///d:/nhathao/Vibe/reels-crawl/docs/01_USER_ROLES.md) | Ai dùng hệ thống và quyền hạn của từng người | Đã hoàn thiện |
| [02_SCOPE_MVP.md](file:///d:/nhathao/Vibe/reels-crawl/docs/02_SCOPE_MVP.md) | Bảng phân loại tính năng: Làm ngay (MVP) vs Làm sau | Đã hoàn thiện |
| [03_TECH_CONSTRAINTS.md](file:///d:/nhathao/Vibe/reels-crawl/docs/03_TECH_CONSTRAINTS.md) | Nền tảng, công nghệ, môi trường chạy, API, bảo mật | Đã hoàn thiện |
| [04_NFR_AND_RISKS.md](file:///d:/nhathao/Vibe/reels-crawl/docs/04_NFR_AND_RISKS.md) | Yêu cầu phi chức năng (5 nhóm) & Ma trận rủi ro (R-01 -> R-06) | Đã hoàn thiện |
| [05_SYSTEM_FLOW.md](file:///d:/nhathao/Vibe/reels-crawl/docs/05_SYSTEM_FLOW.md) | Sơ đồ luồng tổng quan & Sơ đồ tuần tự (Flowchart & Sequence) | Mới bổ sung |
| [06_UI_WIREFRAMES.md](file:///d:/nhathao/Vibe/reels-crawl/docs/06_UI_WIREFRAMES.md) | Thiết kế bố cục & Wireframe Desktop App (Wails v2) | Mới bổ sung |
| [features/feat-01-core-process.md](file:///d:/nhathao/Vibe/reels-crawl/docs/features/feat-01-core-process.md) | Đặc tả tính năng cốt lõi (Input -> Xử lý -> Output -> AC) | Đã cập nhật AC |
| [features/feat-02-data-storage.md](file:///d:/nhathao/Vibe/reels-crawl/docs/features/feat-02-data-storage.md) | Quản lý dữ liệu, Schema SQLite, ERD quan hệ & Dedup | Đã cập nhật ERD |
| [features/feat-03-subtitles-and-content.md](file:///d:/nhathao/Vibe/reels-crawl/docs/features/feat-03-subtitles-and-content.md) | Nhận diện giọng nói (STT) & Cào nội dung link website | Đã hoàn thiện |
| [features/feat-04-capcut-automation.md](file:///d:/nhathao/Vibe/reels-crawl/docs/features/feat-04-capcut-automation.md) | Tự động hóa Template CapCut, Auto Export & Ngoại lệ | Đã cập nhật Ngoại lệ |
| [features/feat-template.md](file:///d:/nhathao/Vibe/reels-crawl/docs/features/feat-template.md) | File mẫu chuẩn để nhân bản khi phát sinh tính năng mới | File mẫu |

---

## 🚀 Trình tự thực hiện đề xuất

1. **Bước 1:** Xem tổng quan bài toán và mục tiêu tại [00_OVERVIEW.md](file:///d:/nhathao/Vibe/reels-crawl/docs/00_OVERVIEW.md).
2. **Bước 2:** Nắm rõ kiến trúc luồng dữ liệu tại [05_SYSTEM_FLOW.md](file:///d:/nhathao/Vibe/reels-crawl/docs/05_SYSTEM_FLOW.md) và giao diện tại [06_UI_WIREFRAMES.md](file:///d:/nhathao/Vibe/reels-crawl/docs/06_UI_WIREFRAMES.md).
3. **Bước 3:** Đọc chi tiết các quy tắc nghiệp vụ trong thư mục `docs/features/` (đặc biệt là xử lý ngoại lệ và điều kiện nghiệm thu).
4. **Bước 4:** Bắt tay vào khởi tạo project (Wails v2 + Go Backend) và triển khai theo từng chặng MVP.
