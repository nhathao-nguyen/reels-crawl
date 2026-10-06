# 04 - Yêu Cầu Phi Chức Năng & Đánh Giá Rủi Ro (NFR & Risk Assessment)

> **Mục tiêu:** Định lượng các tiêu chuẩn chất lượng (hiệu năng, bảo mật, độ ổn định, tương thích, trải nghiệm) và lập kế hoạch ứng phó rủi ro kỹ thuật cho hệ thống kết hợp **CLI + Desktop App (Wails)**.

---

## PHẦN A: YÊU CẦU PHI CHỨC NĂNG (NON-FUNCTIONAL REQUIREMENTS - NFR)

Hệ thống được thiết kế theo 5 nhóm chỉ số đo lường cụ thể:

### 1. Hiệu năng & Xử lý đồng thời (Performance & Concurrency)
* **Tốc độ xử lý từng video:**
  * **Tải video & metadata:** $\le 10$ giây / video (với đường truyền tối thiểu 30 Mbps).
  * **Tạo phụ đề (STT):**
    * *Chế độ Cloud API (Groq Whisper-large-v3):* $\le 1 - 3$ giây / video dưới 60 giây.
    * *Chế độ Local Offline (Whisper.cpp / Faster-Whisper):* Phụ thuộc phần cứng máy (mô hình `base`/`small` hoàn thành trong $\approx 10 - 20$ giây / video).
  * **Tạo CapCut Draft (JSON Injection):** $\le 1$ giây / video.
* **Tiêu hao phần cứng trên Windows:**
  * Dung lượng RAM trạng thái chờ (Idle): $\le 90$ MB.
  * Dung lượng RAM tối đa khi cào hoặc render giao diện Wails: $\le 650$ MB (chạy mượt trên máy văn phòng 8GB RAM).
  * Mức tải CPU: Không quá 70% CPU toàn hệ thống trong các tác vụ cào dữ liệu và tạo draft ngầm.
* **Kiểm soát luồng (Concurrency Control):**
  * Giới hạn mặc định tải đồng thời 2-3 tác vụ để tránh nghẽn băng thông mạng nội bộ khi 10 người cùng sử dụng.

### 2. Bảo mật & Quản lý dữ liệu người dùng (Security & Privacy)
* **Bảo vệ Cookie & Tài khoản mạng xã hội:**
  * File `cookies.txt` và chuỗi session Facebook/TikTok được lưu cục bộ trên máy cá nhân của nhân sự, tuyệt đối **không gửi lên bất kỳ server trung gian nào**.
* **Bảo mật API Key:**
  * Khóa `groq_api_key` hoặc OpenAI Key lưu trong file `config.json` chỉ được đọc trực tiếp bởi ứng dụng để gọi API chính thức, không ghi log ra màn hình console hay file log.
* **Cô lập CSDL nội bộ:**
  * File SQLite (`data.db`) lưu cục bộ không mở port kết nối mạng ra ngoài, đảm bảo an toàn trong môi trường mạng LAN nội bộ.

### 3. Độ tin cậy & Cơ chế tự khôi phục (Reliability & Resilience)
* **Khôi phục sự cố (Crash Recovery):**
  * Trường hợp ứng dụng bị tắt đột ngột (mất điện, kill process), khi mở lại app sẽ đọc lại DB SQLite để tiếp tục các video có trạng thái dở dang (`status = 'downloaded'` nhưng chưa tạo draft).
* **Đóng tiến trình an toàn (Graceful Shutdown):**
  * Khi người dùng nhấn nút Hủy tác vụ hoặc đóng cửa sổ app / `Ctrl + C`, hệ thống phải chủ động kill các tiến trình con (`yt-dlp.exe`, `ffmpeg.exe`, `chromedp`, `whisper-cli.exe`), ngăn ngừa tiến trình mồ côi (Zombie Processes) gây đơ máy.
* **Ngưỡng chặn treo (Circuit Breaker):**
  * Khâu cào link bài viết/chapter có timeout 10 giây/request và giới hạn tối đa 20 chapters để chống kẹt vòng lặp vô tận.

### 4. Tính tương thích & Môi trường (Compatibility & Portability)
* **Hệ điều hành:**
  * Tương thích hoàn toàn với **Windows 10 (Build 1809 trở lên) và Windows 11 (64-bit)**.
  * Tận dụng Microsoft Edge WebView2 Runtime (đã tích hợp sẵn trên Windows 10/11) thông qua kiến trúc Wails.
* **Hỗ trợ phóng to màn hình (Per-Monitor DPI Awareness):**
  * Giao diện Wails và module tự động hóa xuất video (Windows UI Automation) thích ứng chuẩn xác ở mọi tỉ lệ scale màn hình laptop: `100%`, `125%`, `150%`, không bị lệch tọa độ chuột.
* **Tương thích CapCut Desktop:**
  * Tương thích với các phiên bản CapCut Desktop quốc tế (bản 3.x, 4.x trở lên).

### 5. Trải nghiệm người dùng & Triển khai (Usability & UX)
* **Mô hình nhị phân kép (Dual-Mode Execution):**
  * Double-click chạy file `.exe`: Khởi động **Desktop App (Wails GUI)** với giao diện hiện đại, dễ thao tác cho số đông nhân sự.
  * Chạy qua dòng lệnh (ví dụ `reels-tool.exe --cli --url "..."` hoặc `reels-tool.exe --export`): Chạy ở chế độ **Headless CLI**, thuận tiện cho tự động hóa hoặc lập trình viên.
* **Cơ chế tải Model Whisper theo nhu cầu (On-Demand Model Downloader):**
  * File `.exe` ban đầu được giữ ở dung lượng nhẹ (~20MB) bằng cách sử dụng Groq API mặc định.
  * Khi người dùng chuyển sang chế độ **Whisper Local (Offline)** trên giao diện hoặc cấu hình, hệ thống sẽ hiển thị trình quản lý tải model (Download Manager) cho phép người dùng chọn tải các gói model mong muốn (`tiny`, `base`, `small`) với thanh tiến trình tải trực quan. File model sau khi tải được lưu vào thư mục `models/` dùng vĩnh viễn.
* **Phản hồi trạng thái trực quan:**
  * Bảng danh sách task có progress bar %, thẻ badge trạng thái đổi màu (`Đang tải` ➔ `Tạo phụ đề` ➔ `Tạo Draft` ➔ `Đã xuất video`).
  * Có nút bấm mở nhanh: `Mở thư mục video` và `Mở dự án trên CapCut`.

---

## PHẦN B: MA TRẬN ĐÁNH GIÁ RỦI RO & PHƯƠNG ÁN GIẢM THIỂU (RISK ASSESSMENT)

| Mã | Tên rủi ro | Xác suất | Ảnh hưởng | Kịch bản xảy ra | Phương án phòng ngừa & Xử lý (Mitigation) |
| :---: | :--- | :---: | :---: | :--- | :--- |
| **R-01** | **Meta / Facebook chặn Bot & Checkpoint tài khoản** | **Cao** | **Cao** | Cuộn trang liên tục khiến Facebook bắt xác minh Captcha hoặc vô hiệu hóa cookie `c_user`. | 1. Thêm độ trễ ngẫu nhiên (random delay 2-5s) giữa các lần cuộn trang.<br>2. Hướng dẫn nhân sự dùng tài khoản phụ (clone), không dùng tài khoản chính để xuất cookie.<br>3. Thiết kế kiến trúc module độc lập: Nếu link profile bị chặn, nhân sự vẫn có thể dán thủ công danh sách link Reel vào tool để tiếp tục tải bằng `yt-dlp`. |
| **R-02** | **CapCut cập nhật UI làm gián đoạn Auto Export (Phase 2)** | **Cao** | **Trung bình** | CapCut thay đổi vị trí nút bấm, thay đổi phím tắt hoặc xuất hiện popup khuyến mãi mới chắn màn hình. | 1. Sử dụng **Windows UI Automation API** (dò tìm phần tử theo tên/ID) thay vì click tọa độ chuột cố định.<br>2. Cung cấp file cấu hình tọa độ/hotkey trong `config.json` để người dùng tự điều chỉnh nếu CapCut update.<br>3. Phase 1 tạo project draft đã hoàn thành 100%, nếu Phase 2 lỗi thì nhân sự vẫn mở CapCut bấm xuất video thủ công được. |
| **R-03** | **yt-dlp lỗi tải do nền tảng đổi thuật toán CDN** | **Trung bình** | **Cao** | Facebook/TikTok cập nhật player bảo vệ link video khiến `yt-dlp` cũ không bóc tách được stream. | 1. Tích hợp nút **"Cập nhật yt-dlp"** trên giao diện Desktop (chạy ngầm lệnh `yt-dlp.exe -U`).<br>2. Cho phép người dùng truyền thêm cờ tham số tùy biến cho yt-dlp trong cài đặt. |
| **R-04** | **Tràn dung lượng ổ cứng người dùng** | **Cao** | **Trung bình** | 10 nhân sự x 200-300 video/ngày sản sinh hàng chục GB dữ liệu, nhanh chóng làm đầy ổ đĩa `C:`. | 1. Cho phép người dùng cấu hình thư mục tải về sang ổ `D:`, `E:` ngay trên giao diện cài đặt.<br>2. Thêm tính năng cảnh báo dung lượng trống của ổ đĩa trước khi bắt đầu tải mẻ lớn.<br>3. Tùy chọn dọn dẹp: "Tự động xóa file `raw_video.mp4` sau khi video thành phẩm đã được xuất thành công". |
| **R-05** | **Groq API bị nghẽn mạng / Chạm giới hạn Rate-limit** | **Thấp** | **Trung bình** | Hết hạn mức request miễn phí hoặc mạng văn phòng chập chờn khi gửi file audio. | 1. Tự động Retry 3 lần với cơ chế giãn cách thời gian (Exponential Backoff).<br>2. Hỗ trợ chuyển đổi nhanh (Fallback) sang chế độ **Whisper Local (Offline)** đã tải sẵn trên máy mà không làm đứt đoạn công việc. |
| **R-06** | **Tải Model Whisper Offline thất bại hoặc gián đoạn** | **Trung bình** | **Thấp** | File model offline có dung lượng từ 140MB - 500MB, mạng lag có thể gây hỏng file binary model. | 1. Tải có hỗ trợ Resume (tiếp tục tải nếu đứt mạng).<br>2. Kiểm tra mã băm SHA256 của file model sau khi tải xong trước khi đưa vào thư mục `models/`. |
