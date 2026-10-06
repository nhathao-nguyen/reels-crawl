# 03 - Ràng Buộc Kỹ Thuật & Môi Trường Chạy (Technical Constraints)

> **Mục tiêu:** Xác định công nghệ và môi trường triển khai để khi bắt tay vào code không bị đổi kiến trúc giữa chừng.

---

### 1. Hình thức ứng dụng (Application Form)
*Bạn muốn công cụ chạy dưới dạng nào?*
- [x] **Desktop App (Wails v2):** Đóng gói thành file `.exe` duy nhất có giao diện đồ họa hiện đại (HTML/CSS/JS frontend + Go backend), sử dụng WebView2 có sẵn trên Windows 10/11.
- [x] **CLI (Command Line Tool):** Chạy lệnh qua terminal / command prompt khi truyền cờ `--cli` hoặc `--export` (thuận tiện chạy ngầm, script).

---

### 2. Môi trường triển khai & Hệ điều hành
*Công cụ sẽ chạy ở đâu?*
- [x] Máy cá nhân của bạn và nhân sự nội bộ (Hệ điều hành: Windows 10 / Windows 11 64-bit).

---

### 3. Ngôn ngữ & Công nghệ ưa thích (Tech Stack)
*Bạn dự định viết bằng gì (hoặc muốn AI viết bằng gì)?*
* **Ngôn ngữ chính:** Go (Golang)
* **Giao diện Desktop:** **Wails v2** (Go + Frontend hiện đại, siêu nhẹ ~15-20MB, native Windows executable).
* **Thư viện & Công cụ phụ trợ (Đi kèm trong thư mục `bin/`):**
    - `bin/yt-dlp.exe`: Tải video chất lượng cao, không watermark từ TikTok / Facebook / Reels sau khi đã có URL reel cụ thể.
    - `Headless Browser / GraphQL Interceptor (Chromedp / Playwright session)`: Quét Profile/Page với cookie `c_user`, bắt gói tin GraphQL nội bộ để bóc tách URL video Reels, caption và comment tác giả.
    - `Article Extractor Engine`: Sử dụng giải thuật trích xuất văn bản sạch (`go-trafilatura`) và bóc tách link rút gọn (HTTP HEAD redirect tracker).
    - `bin/ffmpeg.exe`: Xử lý cắt/tách audio (convert sang 16kHz mono WAV) để tối ưu tốc độ Whisper.
    - `Speech-to-Text Engine (Linh hoạt 2 chế độ):`
      - **Chế độ Cloud API:** Groq Audio API (Whisper-large-v3, siêu tốc 0.5 - 2s/video, miễn phí/giá rẻ) hoặc OpenAI Whisper API.
      - **Chế độ Local Offline:** `whisper.cpp` / `faster-whisper` chạy trực tiếp trên CPU/GPU máy tính. Hỗ trợ **cơ chế người dùng tự chọn tải model về máy theo nhu cầu** (On-Demand Downloader cho các bản `tiny`, `base`, `small`, lưu vào thư mục `models/`).
    - `Windows UI Automation / RobotGo`: Tự động tương tác GUI CapCut trong Pha 2 (Batch Export), hỗ trợ Per-Monitor DPI Awareness và tự động đóng popup thông báo/cập nhật chắn màn hình.
* **Lưu trữ & Quản lý dữ liệu:**
  - [x] File cấu trúc thư mục (`downloads/{platform}_{video_id}/`).
  - [x] **Cơ sở dữ liệu SQLite nhẹ (Dùng thư viện Pure-Go `modernc.org/sqlite`, không cần cài CGO/gcc):**
    - Quản lý danh sách Page / Kênh / Profile theo dõi.
    - Lưu lịch sử video đã cào để khi quét lại Page sẽ **tự động phát hiện video mới** đăng tải và tải về, bỏ qua các video cũ.
    - Quản lý trạng thái xử lý từng video: `downloaded` ➔ `draft_created` ➔ `exported` (kèm đường dẫn xuất và thông tin lỗi nếu có).
  - [x] Hiển thị tiến trình trực quan trên cả Desktop GUI (Progress Bar) và Terminal (Live Progress).

---

### 4. Tích hợp bên thứ ba & Yêu cầu đặc biệt
* **Cookie / Tài khoản / Session:**
  - Hỗ trợ nạp file `cookies.txt` hoặc chuỗi cookie `c_user=...; xs=...` (đặt cùng thư mục `.exe` hoặc dán trực tiếp trên màn hình Cài đặt Desktop App) để duy trì phiên đăng nhập Facebook/TikTok.
* **Cấu trúc mẫu file cấu hình người dùng (`config.json`):**
  ```json
  {
    "app_mode": "gui",
    "capcut_draft_path": "",
    "template_name": "template",
    "whisper_provider": "groq_api",
    "groq_api_key": "",
    "local_whisper_model": "base",
    "export_resolution": "1080p",
    "download_dir": "downloads",
    "max_chapters_per_story": 20,
    "request_delay_seconds": 2
  }
  ```
  *(Nếu `capcut_draft_path` để trống, hệ thống sẽ tự động quét Registry và các ổ đĩa trên Windows để tìm thư mục `com.lveditor.draft`)*.
* **Cơ chế nhận diện Template CapCut động (User-driven Templates):**
  - Hệ thống tự động quét toàn bộ thư mục dự án CapCut (`com.lveditor.draft`).
  - **Quy tắc:** Bất kỳ thư mục dự án nào có chứa từ khóa `"template"` trong tên (ví dụ: `template`, `template-fb`, `review-template`, `my_template` - không phân biệt hoa thường) đều được tự động nhận diện và đưa vào danh sách để người dùng tùy ý lựa chọn khi cào video hoặc thiết lập mặc định trong Cài đặt.

    
