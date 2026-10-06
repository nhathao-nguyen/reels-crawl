# Feature 04: Tự Động Hóa CapCut & Xuất Video (CapCut Template Automation & Auto Export)

> **Mục tiêu:** Đặc tả cơ chế nhân bản project từ Template CapCut dựng sẵn, bơm tài nguyên mới (video, text, phụ đề), tự động kích hoạt xuất file (Export) và các kịch bản xử lý ngoại lệ an toàn.

---

## 1. Cơ chế hoạt động của Template CapCut
CapCut Desktop (Windows) lưu trữ toàn bộ thông tin dự án tại:
`%LocalAppData%\CapCut\User Data\Projects\com.lveditor.draft\[Tên_Project]\`
Trong đó có các file nòng cốt:
* `draft_content.json`: Chứa timeline, hiệu ứng (effects), animation, sticker, filter, track âm thanh, text style.
* `draft_meta_info.json`: Chứa ID dự án (`draft_id`), tên dự án, thời gian sửa đổi, tỷ lệ khung hình.
* `draft_cover.jpg` (hoặc `.png`): Ảnh thumbnail đại diện hiển thị trên giao diện trang chủ của CapCut Desktop.

### Quy trình tạo Project tự động (Pha 1 - Chạy ngầm 100%):
1. **Tìm & Cài đặt Template:**
   - Hệ thống tự động quét tìm thư mục `com.lveditor.draft` trên tất cả các ổ đĩa của Windows (`C:`, `D:`...).
   - Nếu máy chưa có Template, tool tự động kiểm tra và sao chép template mặc định (`template-fb`) từ thư mục ứng dụng vào thư mục CapCut Draft.
2. **Khởi tạo Draft mới:** Khi có video mới tải về, tool tạo một thư mục dự án mới với `draft_id` ngẫu nhiên duy nhất (UUID) và đặt tên hiển thị: `Dự án {platform}_{video_id}`.
3. **Đọc và chỉnh sửa file `draft_content.json`:**
   - Thay thế đường dẫn video mẫu thành đường dẫn `raw_video.mp4` mới tải về.
   - Cập nhật thời lượng video (duration). *Lưu ý:* Đơn vị thời gian trong CapCut là **Microsecond** ($10^{-6}$ giây, tức 1 giây = 1,000,000 đơn vị).
   - **Quy tắc Heuristic nhận diện track phụ đề thoại:** Quét mảng `tracks` có `type == "text"`, ưu tiên chọn track có số lượng đoạn thoại (`segments`) nhiều nhất và có tọa độ trục Y ở nửa dưới màn hình (`transform.y <= 0`).
   - **Giữ nguyên 100% Style/Font/Animation:** Nhân bản đối tượng Style từ segment mẫu của track phụ đề đã tìm được, thay chuỗi chữ theo file phụ đề `.srt` mới và cập nhật mốc thời gian (`start`, `duration`).
   - Đăng ký project mới vào `root_meta_info.json` để CapCut nhận diện ngay lập tức khi khởi động.
4. **Thumbnail dự án:** CapCut sẽ tự động render ảnh bìa hiển thị từ frame của video mới khi mở project.
5. **Đặc điểm vận hành:** Giai đoạn này chạy 100% ngầm (Headless File System Manipulation), cực nhanh (~0.5s - 1s/video) và hoàn toàn không chiếm chuột/bàn phím của người dùng.

---

## 2. Cơ chế Xuất Video Hàng Loạt (Pha 2 - Nút / Lệnh Auto Export)

* **Kích hoạt:** Người dùng bấm nút "⚡ Xuất Video Hàng Loạt" trên Desktop App hoặc gõ lệnh CLI `reels-tool.exe --export` khi rảnh tay hoặc cuối ngày.
* **Cơ chế tiền xử lý an toàn:**
  1. **Dò tìm ứng dụng:** Tự động tra cứu Windows Registry (`App Paths\CapCut.exe`) hoặc thư mục `%LOCALAPPDATA%\CapCut\Apps` (lấy phiên bản mới nhất) để tìm file thực thi `CapCut.exe` mà không cần hardcode đường dẫn.
  2. **Chống lệch màn hình:** Kích hoạt chế độ Per-Monitor DPI Awareness (`SetProcessDpiAwareness`) để đảm bảo phím tắt và thao tác giao diện không bị lệch trên màn hình scale 125% - 150%.
  3. **Đóng popup chắn giao diện:** Tự động phát hiện và đóng các hộp thoại thông báo cập nhật hoặc popup khuyến mãi rác của CapCut (`dismiss_capcut_popups`).
* **Quy tắc quét & xuất video:**
  1. Quét DB SQLite và thư mục dự án CapCut (`com.lveditor.draft\`).
  2. **Bỏ qua tất cả các project có chứa chữ `template` trong tên**.
  3. **Bỏ qua tất cả các project đã xuất thành công** (đã có trạng thái `status = 'exported'` hoặc đã tồn tại file `final_video.mp4`).
  4. Duyệt tuần tự qua từng dự án chưa xuất (`status = 'draft_created'`):
     - Mở dự án trong CapCut Desktop.
     - Sử dụng Windows UI Automation gửi phím tắt `Ctrl + E`.
     - Tự động chọn độ phân giải 1080p, nhấn nút **Export** và lưu video ra thư mục `downloads/{platform}_{video_id}/exported/final_video.mp4`.
     - Giám sát tiến trình render (chờ đạt 100% hoặc file thành phẩm xuất hiện).
     - Cập nhật trạng thái `exported` vào DB SQLite, đóng CapCut và chuyển sang project tiếp theo trong hàng đợi cho đến khi hoàn tất.

---

## 3. Yêu cầu cấu hình từ người dùng
* Đường dẫn thư mục CapCut Draft (`com.lveditor.draft`) (mặc định để trống để tool tự động tìm).
* Tên thư mục Template mẫu chuẩn (mặc định là `template-fb`).
* Thư mục lưu video thành phẩm xuất ra (mặc định `downloads/`).

---

## 4. Xử Lý Ngoại Lệ & Cơ Chế Phục Hồi (Exceptions & Resilience)

Nhằm đảm bảo quy trình chạy liên tục, tự động và không bị dừng đột ngột giữa chừng, hệ thống thiết lập các quy tắc xử lý ngoại lệ nghiêm ngặt:

| Tình huống ngoại lệ | Hành vi hệ thống & Cách xử lý |
| :--- | :--- |
| **1. Máy chưa cài CapCut Desktop** | * Quét Registry & ổ cứng không thấy `CapCut.exe`.<br>* **Xử lý:** Hiển thị thông báo dạng Pop-up Modal / Console cảnh báo: *"Không tìm thấy CapCut Desktop trên máy! Vui lòng cài đặt CapCut từ capcut.com để sử dụng tính năng tạo và xuất video."* Chặn kích hoạt Pha 2 nhưng vẫn cho phép tải video và phụ đề ở Pha 1. |
| **2. Chưa có Template CapCut mẫu** | * Thư mục `com.lveditor.draft` chưa có thư mục template được chỉ định (mặc định là `template-fb`).<br>* **Xử lý:** Nếu trong thư mục cài đặt tool có sẵn template đóng gói thì tự động copy sang; nếu không có, hiển thị thông báo bắt buộc: *"Chưa phát hiện Template chuẩn! Vui lòng mở CapCut tạo 1 dự án mẫu tên 'template-fb' có sẵn text/style mẫu trước khi chạy."* Không tiến hành clone để tránh sinh project hỏng. |
| **3. Máy nhân sự đang mở sẵn CapCut** | * **Ở Pha 1 (Tạo dự án ngầm / JSON Injection):** **Hoàn toàn không ảnh hưởng.** Nhân sự vẫn mở và edit CapCut bình thường vì Pha 1 chỉ ghi file JSON vào ổ đĩa. CapCut sẽ tự động nhận diện draft mới khi người dùng quay về trang chủ (Home).<br>* **Ở Pha 2 (Auto Export):** Để tránh xung đột cửa sổ, mất focus phím tắt `Ctrl + E` hoặc kẹt trạng thái pop-up cũ, tool sẽ: **(1)** Hiển thị thông báo trên màn hình: *"Đang đóng CapCut hiện tại để làm sạch môi trường xuất video hàng loạt..."* **(2)** Tự động ra lệnh đóng/kill sạch các tiến trình `CapCut.exe` đang chạy, sau đó mới khởi động lại từng project một cách độc lập và chuẩn xác. |
| **4. Export bị treo (Timeout > 5 phút)** | * Khi mở project và gửi lệnh Export, nếu sau **300 giây (5 phút)** mà tiến trình render không hoàn thành (hoặc CapCut bị Not Responding):<br>* **Xử lý:** Tool tự động force-kill tiến trình `CapCut.exe`, ghi log lỗi vào SQLite: `error_message = 'Export Timeout (>300s)'`, đánh dấu trạng thái `status = 'failed'`. Tool **không dừng lại** mà lập tức chuyển sang project tiếp theo trong hàng đợi. Cuối mẻ xuất, người dùng có thể bấm nút "Thử lại các video lỗi" trên giao diện. |
| **5. Không xử lý lại các dự án đã xuất thành công** | * Khi bấm nút Xuất Video Hàng Loạt hoặc quét thư mục draft:<br>* **Xử lý:** Tool đối chiếu với SQLite: chỉ lấy các bản ghi có `status = 'draft_created'`. Nếu bản ghi đã là `status = 'exported'` HOẶC file `downloads/{platform}_{video_id}/exported/final_video.mp4` đã tồn tại trên đĩa, tool **tự động bỏ qua (Skip 100%)**, tuyệt đối không mở lại để render trùng lặp gây lãng phí thời gian và tài nguyên máy tính. |

---

## 5. Điều Kiện Nghiệm Thu (Acceptance Criteria)

- [ ] Tự động định vị chính xác thư mục `com.lveditor.draft` trên Windows mà không cần người dùng nhập tay.
- [ ] Nhân bản thành công `template-fb` sang project mới, giữ nguyên 100% hiệu ứng, sticker, animation và font chữ của template gốc.
- [ ] Thay thế đường dẫn video và cập nhật đúng thời lượng (tính theo microsecond $10^{-6}$s).
- [ ] Giải thuật Heuristic phát hiện đúng track phụ đề thoại và chèn chính xác nội dung từ file `.srt`.
- [ ] Dự án mới xuất hiện ngay trên màn hình trang chủ của CapCut Desktop mà không cần khởi động lại máy.
- [ ] Trong Pha 2, tool tự động đóng các tiến trình CapCut cũ đang mở trước khi bắt đầu chu kỳ xuất sạch.
- [ ] Gửi phím tắt `Ctrl + E` và kích hoạt nút Export thành công ở độ phân giải 1080p bằng Windows UI Automation.
- [ ] Tự động ngắt (kill) và chuyển tiếp an toàn khi gặp sự cố treo render quá 5 phút.
- [ ] Bỏ qua 100% các video đã có trạng thái `exported`.
