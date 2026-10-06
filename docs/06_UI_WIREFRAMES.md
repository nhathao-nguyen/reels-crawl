# 06 - Thiết Kế Giao Diện & Bố Cục Desktop App (UI Wireframes & Mockups)

> **Mục tiêu:** Cung cấp bản phác thảo chi tiết (Wireframe & Layout Breakdown) cho giao diện đồ họa **Desktop App (Wails v2)** để đội ngũ lập trình Frontend có thể hiện thực hóa chuẩn xác 100% về bố cục, tương tác và trạng thái người dùng.

---

## 1. Tổng quan Kiến Trúc Giao Diện (UI Architecture)

Ứng dụng được thiết kế theo phong cách hiện đại (Modern Dark Mode / Glassmorphism), chia làm 2 khu vực chính:
* **Thanh điều hướng bên trái (Sidebar Navigation):** Chuyển đổi giữa các phân hệ:
  * 📥 **Tác vụ & Tiến trình (Dashboard / Tasks)**
  * 🌐 **Kênh theo dõi (Tracked Sources)**
  * ⚙️ **Cài đặt & Cấu hình (Settings)**
* **Khu vực nội dung chính (Main Content Area):** Hiển thị các bảng điều khiển, danh sách tác vụ và thẻ thao tác nhanh.

---

## 📸 Ảnh Thiết Kế Giao Diện Trực Quan (High-Fidelity Mockup)

![Giao diện Reels Automation Studio Desktop Wails v2](C:/Users/nhathao/.gemini/antigravity-ide/brain/da6da18e-2f43-4b73-9055-baab24f7b77f/reels_desktop_ui_1791212693567.jpg)

> 💡 **Bản xem trước HTML tương tác trực tiếp:**  
> Bạn có thể mở trực tiếp file [docs/ui-mockup.html](file:///d:/nhathao/Vibe/reels-crawl/docs/ui-mockup.html) bằng bất kỳ trình duyệt nào (Chrome, Edge...) để bấm thử chuyển tab (Dashboard / Kênh theo dõi / Cài đặt) và mở các hộp thoại (Modal Tải Model Whisper / Modal Xuất Hàng Loạt).

---

## 2. Màn hình 1: Dashboard Chính & Bảng Tiến Trình (Main Dashboard)

```text
+----------------------------------------------------------------------------------------------------+
|  [Reels Automation Studio v1.0]                                                    [_] [口] [X]     |
+-------------------+--------------------------------------------------------------------------------+
|                   |  📥 BẢNG ĐIỀU KHIỂN & TIẾN TRÌNH XỬ LÝ (PIPELINE)                              |
|  [Logo] ReelsTool |--------------------------------------------------------------------------------|
|                   |  [+] Nhập liên kết (Profile Facebook / TikTok URL hoặc dán nhiều link):        |
|  * Tác vụ chính   |  +-------------------------------------------------------------+ +----------+  |
|    (Tasks)        |  | https://www.facebook.com/reel/1029384756...                | | + Thêm   |  |
|                   |  +-------------------------------------------------------------+ +----------+  |
|  * Kênh theo dõi  |  [ Chọn file links.txt ]  [✓] Bỏ qua video trùng lặp   Template: [template-fb v] |
|    (Sources)      |                                                                                |
|                   |  +---------------------------+  +-------------------------------------------+  |
|  * Cài đặt        |  | [▶ BẮT ĐẦU CÀO (PHA 1)]   |  | [⚡ XUẤT VIDEO HÀNG LOẠT (PHA 2)]          |  |
|    (Settings)     |  +---------------------------+  +-------------------------------------------+  |
|                   |                                                                                |
|                   |  TIẾN TRÌNH TỔNG QUAN: [████████████████████░░░░░░░░░░] 65% (13/20 hoàn tất)   |
|                   |--------------------------------------------------------------------------------|
|                   |  Lọc trạng thái: [Tất cả (20)] [Chờ xử lý] [Đã tạo Draft (13)] [Đã xuất (10)] |
|                   |                                                                                |
|                   |  DANH SÁCH TÁC VỤ (TASKS QUEUE):                                               |
|                   |  +----+----------+-----------------------+-------------+-----------+--------+  |
|                   |  | ID | Nền tảng | Tiêu đề / ID video    | Trạng thái  | Tiến độ   | Thao   |  |
|                   |  +----+----------+-----------------------+-------------+-----------+--------+  |
|                   |  | 01 | Facebook | fb_98234125 (Reel #1) | [ĐÃ XUẤT]   | [100%]    | [📁][🎬]|  |
|                   |  | 02 | TikTok   | tt_73819204 (Review)  | [TẠO DRAFT] | [100%]    | [📁][✂] |  |
|                   |  | 03 | Facebook | fb_12948192 (Chap 02) | [ĐANG TẢI]  | [ 45%]    | [  -  ]|  |
|                   |  | 04 | Facebook | fb_84719283 (Post #3) | [LỖI LINK]  | [Thử lại] | [⚠️]   |  |
|                   |  +----+----------+-----------------------+-------------+-----------+--------+  |
|                   |                                                                                |
|  Trạng thái:      |  [ℹ] STT Engine: Groq API (Sẵn sàng) | Thư mục: D:/reels-crawl/downloads       |
|  [● Sẵn sàng]     |  [📁 Mở thư mục Downloads]              [🗑 Dọn dẹp task đã xuất thành công]     |
+-------------------+--------------------------------------------------------------------------------+
```

### Chú thích các nút thao tác nhanh (Actions):
* `[📁]`: Mở trực tiếp thư mục `downloads/{platform}_{video_id}/` trên Windows Explorer.
* `[🎬]`: Mở xem ngay video thành phẩm `final_video.mp4` bằng trình xem media mặc định.
* `[✂]`: Mở nhanh project tương ứng trong CapCut Desktop.
* `[⚠️]`: Xem log chi tiết nguyên nhân lỗi (Link chết, mạng lag, timeout...).

---

## 3. Màn hình 2: Quản Lý Nguồn & Kênh Theo Dõi (Tracked Sources)

```text
+-------------------+--------------------------------------------------------------------------------+
|  [Logo] ReelsTool |  🌐 QUẢN LÝ KÊNH & TRANG THEO DÕI ĐỊNH KỲ (TRACKED PROFILES)                   |
|                   |--------------------------------------------------------------------------------|
|  * Tác vụ chính   |  +-------------------------------------------------------------+ +----------+  |
|                   |  | Nhập link Profile/Page Facebook hoặc TikTok...              | | + Theo dõi |
|  * Kênh theo dõi  |  +-------------------------------------------------------------+ +----------+  |
|    (Sources) [●]  |                                                                                |
|                   |  DANH SÁCH KÊNH ĐANG THEO DÕI (12 Kênh):                                       |
|  * Cài đặt        |  +----+----------+------------------------+-----------+------------+--------+  |
|                   |  | ID | Nền tảng | Tên tác giả / Profile  | Video cào | Lần quét   | Thao   |  |
|                   |  +----+----------+------------------------+-----------+------------+--------+  |
|                   |  | 01 | Facebook | fb.com/reviewphimhay   | 145 video | 10 phút tr | [⟳][✕] |  |
|                   |  | 02 | TikTok   | @toptrendingclips      |  89 video | 1 giờ tr   | [⟳][✕] |  |
|                   |  | 03 | Facebook | fb.com/truyennganvibe  |  34 video | Hôm qua    | [⟳][✕] |  |
|                   |  +----+----------+------------------------+-----------+------------+--------+  |
|                   |  [⟳ Quét kiểm tra tất cả kênh]   [ Tải video mới phát hiện ]                   |
+-------------------+--------------------------------------------------------------------------------+
```

---

## 4. Màn hình 3: Cài Đặt Hệ Thống (Settings Modal / View)

```text
+----------------------------------------------------------------------------------------------------+
|  ⚙️ CÀI ĐẶT HỆ THỐNG & CẤU HÌNH LIÊN KẾT                                                           |
+----------------------------------------------------------------------------------------------------+
|                                                                                                    |
|  1. ĐƯỜNG DẪN CAPCUT DESKTOP & TEMPLATE:                                                           |
|     * Thư mục CapCut Draft: [ C:\Users\Admin\AppData\Local\CapCut\User Data\Projects\com.lveditor ] |
|       [ Tự động quét tìm ]  [ Duyệt thư mục... ]                                                    |
|     * Tên Template mẫu:     [ template-fb                      ] (Mặc định trong com.lveditor.draft)|
|                                                                                                    |
|  2. BÓC TÁCH NỘI DUNG & PHỤ ĐỀ (SPEECH-TO-TEXT):                                                   |
|     * Phương thức nhận diện giọng nói:                                                             |
|       (•) Cloud API siêu tốc (Groq Whisper-large-v3-turbo) [Khuyên dùng]                           |
|           Groq API Key: [ gsk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx ] [Kiểm tra kết nối]    |
|       ( ) Local Offline (faster-whisper / whisper.cpp)                                             |
|           Model hiện tại: [ base.bin (Đã tải) v ]   [ Quản lý tải Model Offline... ]               |
|                                                                                                    |
|  3. PHIÊN ĐĂNG NHẬP (COOKIES):                                                                     |
|     * Cookie Facebook/TikTok: [ c_user=1000...; xs=2...                            ] [Dán từ Clip] |
|       Hoặc nạp file: [ cookies.txt ]  [ Tải file cookies ]                                         |
|                                                                                                    |
|  4. LƯU TRỮ & DUNG LƯỢNG:                                                                          |
|     * Thư mục lưu thành phẩm: [ D:\reels-crawl\downloads                         ] [ Thay đổi... ] |
|     * Dung lượng ổ đĩa D: còn trống: 142 GB                                                        |
|     * [✓] Tự động xóa file raw_video.mp4 sau khi đã xuất video thành phẩm thành công               |
|                                                                                                    |
|  +--------------------+                                                       +-----------------+  |
|  |  [ Khôi phục gốc ] |                                                       |   [ LƯU CẤU HÌNH ] |
|  +--------------------+                                                       +-----------------+  |
+----------------------------------------------------------------------------------------------------+
```

---

## 5. Màn hình 4: Hộp Thoại Tải Model Whisper Offline (Download Manager Modal)

```text
+----------------------------------------------------------------------------------+
|  📥 TRÌNH QUẢN LÝ TẢI MODEL WHISPER (OFFLINE MODE)                                |
+----------------------------------------------------------------------------------+
|  Chọn model phù hợp với cấu hình máy để chạy phụ đề không cần internet:          |
|                                                                                  |
|  1. Whisper Tiny  (~75 MB)   - Rất nhẹ, tốc độ cao, độ chính xác cơ bản.          |
|     Trạng thái: [ Đã cài đặt ]                              [ Xóa model ]        |
|                                                                                  |
|  2. Whisper Base  (~145 MB)  - Khuyên dùng cho máy văn phòng thông thường.       |
|     Tiến trình tải: [████████████████░░░░░░░░░░] 62% (90/145 MB) - 4.5 MB/s      |
|     [ Tạm dừng ]  [ Hủy tải ]                                                    |
|                                                                                  |
|  3. Whisper Small (~460 MB)  - Chuẩn xác cao nhất tiếng Việt, cần máy cấu hình khá.|
|     Trạng thái: [ Chưa tải ]                                [ ⬇ Tải về (460MB) ] |
|                                                                                  |
|  [✓] Tự động kiểm tra mã băm SHA256 sau khi hoàn tất để đảm bảo file vẹn toàn.   |
|                                                                                  |
|                                                                    +----------+  |
|                                                                    | [ Đóng ] |  |
|                                                                    +----------+  |
+----------------------------------------------------------------------------------+
```

---

## 6. Hộp Thoại Cảnh Báo Khi Chạy Pha 2 (Auto Export Modal)

```text
+----------------------------------------------------------------------------------+
|  ⚠️ XÁC NHẬN KÍCH HOẠT XUẤT VIDEO HÀNG LOẠT (PHA 2)                              |
+----------------------------------------------------------------------------------+
|                                                                                  |
|  Hệ thống chuẩn bị tự động điều khiển CapCut để xuất 13 video đã tạo Draft.       |
|                                                                                  |
|  LƯU Ý QUAN TRỌNG ĐỂ QUÁ TRÌNH CHẠY KHÔNG BỊ GIÁN ĐOẠN:                          |
|  1. Nếu CapCut đang mở trên máy, ứng dụng sẽ TỰ ĐỘNG ĐÓNG CapCut để làm sạch     |
|     môi trường trước khi khởi động từng project.                                 |
|  2. Trong quá trình render, vui lòng KHÔNG dùng chuột/bàn phím bấm đè lên màn hình|
|     CapCut để tránh lệch focus phím tắt Ctrl + E.                                |
|  3. Thời gian dự kiến: ~45 giây / video.                                          |
|                                                                                  |
|  [✓] Đã hiểu và sẵn sàng                                                         |
|                                                                                  |
|             +-------------------------+         +---------------------+          |
|             |  [ BẮT ĐẦU XUẤT NGAY ]  |         |     [ HỦY BỎ ]      |          |
|             +-------------------------+         +---------------------+          |
+----------------------------------------------------------------------------------+
```
