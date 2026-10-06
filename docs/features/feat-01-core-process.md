# Feature 01: Quy Trình Xử Lý Cốt Lõi (Core Process / Workflow)

> **Mục tiêu:** Mô tả chi tiết luồng chính (Happy Path) và các trường hợp lỗi từ lúc người dùng cung cấp dữ liệu đầu vào cho đến khi nhận được kết quả.

---

### 1. Đầu vào (Inputs)
*Hệ thống cần nhận dữ liệu gì từ người dùng để bắt đầu chạy?*
* **Loại dữ liệu:**
  - [x] Một đường link đơn lẻ (URL).
  - [x] Danh sách link (Nhập nhiều dòng hoặc upload file `.txt` / `.csv`).
* **Các tùy chọn kèm theo (Options):**
  - [x] Số lượng tối đa cần lấy.
  - [x] Chỉ lấy video không dính watermark / Logo.  

---

### 2. Các bước xử lý của hệ thống (Step-by-step Execution)
Hệ thống chia làm **2 Pha độc lập** để tối ưu hiệu suất và không chiếm màn hình người dùng:

#### PHA 1: THU THẬP & TẠO PROJECT NGẦM (Lệnh chạy chính)
1. **Bước 1 (Check trùng lặp):** Kiểm tra DB SQLite & thư mục `downloads/{platform}_{video_id}/`. Nếu video đã tải rồi thì **Skip**, chuyển ngay sang link tiếp theo.
2. **Bước 2 (Crawl & Download):** 
   - *Khâu bóc tách:* Sử dụng browser session (có cookie `c_user`) để cuộn trang Profile/Page, bắt các gói tin GraphQL nội bộ để lấy danh sách URL Reel, caption và bình luận của tác giả.
   - *Khâu tải video:* Chuyển URL Reel cho `yt-dlp` (kèm cookie) tải file `raw_video.mp4` gốc sạch không dính watermark.
3. **Bước 3 (Web Scraping):** Tự động phát hiện link trong caption/comment; theo dõi chuyển hướng (unshorten link đích), dùng giải thuật bóc tách văn bản sạch (loại bỏ ads, nav) lưu vào `content/`.
4. **Bước 4 (Auto Subtitles):** Trích xuất audio 16kHz mono qua FFmpeg và chạy Whisper (Local hoặc API) để xuất ra file `subtitles.srt`.
5. **Bước 5 (CapCut Injector):** Nhân bản Template CapCut, nạp video mới, tự động phát hiện track phụ đề theo quy tắc Heuristic để giữ 100% style/font và bơm text phụ đề mới vào `draft_content.json`.
6. **Bước 6 (Hoàn tất):** Cập nhật trạng thái `draft_created` vào SQLite, hiển thị thanh tiến trình trực quan trên màn hình (Live Progress) và chuyển sang video kế tiếp.

#### PHA 2: XUẤT VIDEO HÀNG LOẠT (Nút / Lệnh Auto Export)
* **Kích hoạt:** Bấm nút hoặc chạy lệnh `reels-tool.exe --export`.
* **Cơ chế lọc & xuất an toàn:** 
  1. Tự động thiết lập Per-Monitor DPI Awareness để chống lệch tọa độ trên Windows.
  2. Quét toàn bộ thư mục dự án CapCut (`com.lveditor.draft\`), **bỏ qua tất cả các project có chứa chữ `template` trong tên**.
  3. Mở lần lượt từng project còn lại, tự động phát hiện và đóng các hộp thoại popup/update chắn màn hình của CapCut.
  4. Gửi phím tắt `Ctrl + E`, chọn độ phân giải và nhấn **Export** ra thư mục `downloads/{platform}_{video_id}/exported/final_video.mp4`.
  5. Giám sát khi render xong (100%), cập nhật trạng thái `exported` vào SQLite và đóng CapCut để chuyển sang project tiếp theo.

---

### 3. Đầu ra mong muốn (Outputs)
* **Thư mục lưu trữ theo từng video:** Cấu trúc `downloads/{platform}_{video_id}/`
  - `raw_video.mp4`: Video gốc đã tải (không watermark).
  - `subtitles.srt`: File phụ đề đồng bộ mốc thời gian.
  - `metadata.json`: Chứa toàn bộ thông tin cào được (caption, tác giả, views, link bài viết).
  - `content/`: Chứa file `.txt` bài viết (one-page hoặc danh sách chapter).
  - `exported/final_video.mp4`: Video thành phẩm sau khi xuất từ CapCut.
* **Thông tin kèm theo (Metadata) cần lưu:**
  - [x] Tiêu đề / Caption / Mô tả đầy đủ
  - [x] Tác giả / Tên kênh / Link profile
  - [x] Các liên kết (URL) tìm thấy trong caption hoặc comment của tác giả
  - [x] Nội dung text cào được từ các liên kết website trên
  - [x] Số lượt thích (Likes), bình luận (Comments), chia sẻ (Shares), lượt xem (Views)
  - [x] Thời gian đăng bài và thời gian crawl

---

### 4. Xử lý ngoại lệ & Lỗi (Edge cases / Error handling)
*Nếu gặp các trường hợp sau thì hệ thống ứng xử thế nào?*
* **Link bị chết / Bị xóa / Riêng tư (Private):**
  > [Ghi nhận vào log lỗi, bỏ qua và tiếp tục chạy link kế tiếp]
* **Bị chặn IP / Yêu cầu Captcha / Rate limit:**
  > [Tự động tạm dừng X giây, đổi proxy hoặc báo dừng cho người dùng]
* **Mất kết nối mạng giữa chừng:**
  > [Có cơ chế thử lại (Retry 3 lần) trước khi đánh dấu thất bại]

---

### 5. Điều Kiện Nghiệm Thu (Acceptance Criteria)
- [ ] Nhận diện và xử lý thành công cả 2 hình thức nhập link: link đơn lẻ hoặc danh sách nhiều dòng từ file `.txt`.
- [ ] Tải video Reels từ Facebook và TikTok sạch 100% không dính watermark qua `yt-dlp`.
- [ ] Trích xuất đầy đủ và chính xác metadata lưu vào `metadata.json` (caption, author, views, time).
- [ ] Phát hiện liên kết trong caption/comment, unshorten thành công và bóc tách sạch văn bản (loại bỏ ads, menu header/footer).
- [ ] Tạo phụ đề `subtitles.srt` đồng bộ timestamp với giọng nói trong video qua Groq API hoặc Whisper Local.
- [ ] Tự động nhân bản Template CapCut, nạp video và inject text phụ đề vào `draft_content.json` trong dưới 1 giây.
- [ ] Bỏ qua ngay lập tức các video đã tải trước đó (Deduplication trong < 10ms).
- [ ] Ở Pha 2, kích hoạt lệnh `--export` hoặc bấm nút xuất hàng loạt tạo ra file `exported/final_video.mp4` hoàn chỉnh chuẩn độ phân giải 1080p.
- [ ] Khi gặp link lỗi hoặc chết mạng, hệ thống tự động retry tối đa 3 lần, ghi log và tiếp tục xử lý các video khác mà không làm crash chương trình.
