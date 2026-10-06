# Feature 03: Tạo Phụ Đề & Bóc Tách Nội Dung Trang Web (Subtitles & Web Scraping)

> **Mục tiêu:** Đặc tả chi tiết khâu chuyển giọng nói thành phụ đề (Speech-to-Text) và khâu quét đường link trong caption/comment để cào bài viết website.

---

### PHẦN A: BÓC TÁCH LIÊN KẾT & NỘI DUNG (Content Site Scraping)

#### 1. Cơ chế tìm và giải mã liên kết
* Hệ thống quét qua chuỗi `caption` và danh sách bình luận (lọc theo bình luận của chủ kênh / tác giả).
* Trích xuất các liên kết dạng `http://` hoặc `https://` (hoặc link rút gọn như bit.ly, tinyurl...).
* **Giải mã link rút gọn (Unshorten):** Sử dụng HTTP Client gửi request `HEAD` (hoặc fallback `GET` với `follow_redirects=True`) để lấy URL đích thực tế cuối cùng trước khi cào dữ liệu.

#### 2. Bóc tách nội dung website (Web Scraping Engine)
* **Các dạng nội dung & Quy tắc nhận diện tự động (Smart Heuristic):**
  1. **Bước 1 (Làm sạch link):** Tự động loại bỏ các tham số rác và popup quảng cáo như `#google_vignette`, `?fbclid=...`, `utm_...`
  2. **Bước 2 (Giải thuật trích xuất văn bản sạch - Boilerplate Removal):**
     - Áp dụng giải thuật bóc tách bài viết theo chuẩn **Trafilatura / Readability** (như thư viện `go-readability` trong Go).
     - Tự động tách phần thân bài viết (main body content), loại bỏ triệt để thanh menu, sidebar, footer, quảng cáo, nút chia sẻ mạng xã hội.
  3. **Bước 3 (Xử lý Chapter & Ngưỡng an toàn):**
     - **Dò tìm Chapter:** Kiểm tra URL có chứa `/chapter-\d+`, `/chap-\d+`, `/part-\d+` hoặc tìm nút điều hướng `Next Chapter` / `Chương tiếp`.
     - **Nếu là Chapter:** Tool duyệt tuần tự từ Chapter 1 sang các Chapter tiếp theo. Lưu từng file `content/chapters/chapter_01.txt`, `chapter_02.txt`... và file tổng hợp `content/chapters/full_article.txt`.
     - **Ngưỡng an toàn chống treo (Circuit Breaker):** Giới hạn tối đa `max_chapters_per_story` (mặc định 20 chapters) và timeout mỗi request 10 giây để ngăn chặn bẫy vòng lặp vô tận.
     - **Nếu là One-page:** Trích xuất toàn bộ văn bản chính và lưu vào `content/full_article.txt`.

---

### PHẦN B: TỰ ĐỘNG TẠO PHỤ ĐỀ (Speech-to-Text / Auto Subtitles)

#### 1. Tiền xử lý âm thanh (FFmpeg Audio Normalization)
* Trước khi đưa vào Whisper, hệ thống dùng FFmpeg tách audio từ `raw_video.mp4` và chuẩn hóa về định dạng tối ưu nhất cho mô hình:
  `ffmpeg -i raw_video.mp4 -vn -ar 16000 -ac 1 -c:a pcm_s16le audio.wav`
  *(Tần số 16kHz mono giúp Whisper nhận diện nhanh hơn gấp nhiều lần so với nạp trực tiếp file video)*.

#### 2. Các chế độ nhận diện giọng nói
Hệ thống hỗ trợ cả 2 chế độ (chuyển đổi linh hoạt qua giao diện Desktop App hoặc file `config.json`):

1. **Chế độ 1: Cloud API (Groq Audio API / OpenAI Whisper API) [Mặc định ban đầu]:**
   - Gửi file audio lên cloud để nhận diện siêu tốc (chỉ mất ~0.5 - 2 giây/video với Groq Whisper-large-v3-turbo).
   - Tối ưu cho máy tính cấu hình văn phòng, không cần tải model nặng về máy, ứng dụng khởi động nhẹ nhàng (~20MB).

2. **Chế độ 2: Local Offline (Whisper.cpp / Faster-Whisper) [Tùy chọn tải về theo nhu cầu]:**
   - Chạy hoàn toàn offline trên CPU/GPU của máy cá nhân qua file thực thi nhúng `whisper-cli.exe` hoặc thư viện tối ưu `faster-whisper`.
   - Hoàn toàn miễn phí, không phụ thuộc vào internet hay quota API.
   - **Cơ chế Tải Model Theo Nhu Cầu (On-Demand Model Downloader):**
     - Ban đầu ứng dụng **không đóng gói sẵn các file model nặng hàng trăm MB** để tránh phình to bộ cài.
     - Trên màn hình Desktop App (hoặc lệnh CLI `reels-tool.exe --download-model`), hệ thống cung cấp danh sách model để người dùng tự chọn tải:
       - `tiny` (~75 MB): Tải nhanh, nhận diện siêu nhẹ.
       - `base` (~145 MB): Khuyên dùng cho máy văn phòng thông thường, độ chính xác tốt.
       - `small` (~460 MB): Độ chính xác cao, khuyên dùng nếu máy có card đồ họa.
     - Quá trình tải có thanh phần trăm tiến trình (Progress Bar), hỗ trợ resume nếu mất mạng và tự động lưu vào thư mục `models/`.

#### 3. Định dạng đầu ra của phụ đề
* File `.srt` chuẩn mốc thời gian (Timestamp: `00:00:01,200 --> 00:00:03,500`).
* Nạp trực tiếp vào mảng phụ đề `texts` trong file cấu hình dự án của CapCut (`draft_content.json`).

---

### PHẦN C: ĐIỀU KIỆN NGHIỆM THU (ACCEPTANCE CRITERIA)
- [ ] Tách được audio 16kHz mono WAV từ video gốc không bị lỗi codec.
- [ ] Chế độ Groq API trả về kết quả phụ đề `.srt` trong dưới 3 giây.
- [ ] Khi chọn Whisper Local lần đầu, hệ thống hiển thị hộp thoại tải model rõ ràng, không bị crash hoặc treo đơ app.
- [ ] File `.srt` sinh ra có mốc thời gian khớp với tiếng nói trong video, hỗ trợ font tiếng Việt không bị lỗi font/dấu.
