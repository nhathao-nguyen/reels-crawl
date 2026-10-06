# 02 - Phạm Vi Dự Án & Danh Sách Tính Năng (Scope & MVP)

> **Mục tiêu:** Tránh hội chứng "tham tính năng" (Feature Creep) khiến dự án làm mãi không xong. Phải phân tách rõ cái gì BẮT BUỘC có để chạy được (MVP) và cái gì CÓ THÌ TỐT (làm sau).

---

### 1. Bảng phân loại tính năng

Hãy liệt kê tất cả các ý tưởng tính năng bạn nghĩ ra và gắn nhóm cho chúng:

| Mã | Tên tính năng | Mô tả ngắn gọn | Mức độ ưu tiên (`MVP` / `Phase 2` / `Bỏ qua`) |
| :---: | :--- | :--- | :---: |
| **F-01** | Nhập liên kết (URL input) | Cho phép dán 1 link hoặc file danh sách nhiều link | **MVP** |
| **F-02** | Xử lý & Thu thập (Crawl Engine) | Tải nội dung/video và trích xuất thông tin cơ bản | **MVP** |
| **F-03** | Lưu trữ kết quả | Lưu vào thư mục máy tính hoặc xuất file CSV/Excel | **MVP** |
| **F-04** | Giao diện đồ họa Desktop App (Wails v2) | Giao diện trực quan bấm nút cho nhân sự nội bộ (song song với CLI) | **MVP** |
| **F-05** | Hẹn giờ chạy tự động (Cron/Scheduler) | Định kỳ chạy vào khung giờ cố định | Phase 2 |
| **F-06** | Thông báo kết quả qua Telegram/Discord | Gửi thông báo khi hoàn thành hoặc có lỗi | Phase 2 |
| **F-07** | Quản lý Proxy / Xoay IP | Tránh bị block / rate limit khi cào số lượng lớn | Phase 2 |

---

### 2. Định nghĩa MVP (Phiên bản chạy được đầu tiên)
*Được đóng gói thành 1 file thực thi duy nhất (`.exe`) chạy ở **2 chế độ (Dual-Mode)**: Mặc định mở giao diện **Desktop App (Wails v2)** cho nhân sự; hoặc chạy cờ lệnh **CLI** qua Terminal:*
1. **Giao diện Desktop (Wails v2) & CLI:** Form nhập link, danh sách task trực quan, chọn template CapCut, cấu hình API/Whisper model và nút kích hoạt xuất video 1 chạm.
2. **Nhập liên kết (URL input):** Nhập 1 link Reel hoặc đọc file danh sách link `links.txt` / profile URL.
3. **Xử lý & Thu thập (Crawl Engine):** Thu thập Reel video không watermark, bóc tách caption và nội dung liên kết.
4. **Lưu trữ dữ liệu (Data storage):** Lưu cấu trúc thư mục `downloads/` và SQLite chống trùng lặp.
5. **Tạo phụ đề tự động (Auto Subtitles):** Nhận diện giọng nói qua Groq API siêu tốc hoặc tải Whisper/Faster-Whisper model về máy để chạy Offline theo nhu cầu người dùng.
6. **Tự động hóa CapCut (CapCut Automation):** Tự động clone project template, nạp phụ đề/video mới, và hỗ trợ nút bấm / lệnh xuất hàng loạt (`--export`).

---

### 3. Những thứ KHÔNG LÀM (Out of Scope - Tạm thời chưa đụng đến)
*Ghi rõ những tính năng phức tạp để không bị phân tâm trong quá trình code:*
* *Ví dụ: Chưa làm hệ thống thanh toán, chưa hỗ trợ đa ngôn ngữ, chưa làm ứng dụng di động.*
- Giao diện website công khai (chỉ làm Desktop App nội bộ)
- Hẹn giờ chạy tự động (Cron/Scheduler)
- Thông báo kết quả qua Telegram/Discord
- Quản lý Proxy / Xoay IP tự động
