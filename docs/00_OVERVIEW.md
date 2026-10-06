# 00 - Tổng Quan Dự Án (Project Overview)

> **Mục tiêu:** Làm rõ lý do dự án tồn tại và bài toán cốt lõi cần giải quyết. Không cần viết dài, chỉ cần đúng trọng tâm.
  - Cần có một công cụ hỗ trợ để việc lấy video và nội dung nhanh gọn, chính xác.
---

### 1. Vấn đề thực tế (Problem Statement)
*Hiện tại bạn (hoặc team nội bộ) đang gặp khó khăn gì khi chưa có công cụ này?*
  - Hiện tại việc tải/tạo video Reels/TikTok và lưu metadata phải làm thủ công bằng tay từng link, mất quá nhiều thời gian và hay bị sót thông tin.
* **Vấn đề cụ thể:**
  - tải từng video, caption, content site quá lâu
  - edit video bằng capcut khó khăn. xuất từng video bằng tay hiệu xuất kém
  - làm cả 200 300 video mỗi ngày
* **Tần suất lặp lại:** (Mỗi ngày / Mỗi tuần / Thỉnh thoảng?)
  - mỗi ngày

---

### 2. Giải pháp mong đợi (Solution Vision)
*Dự án này sẽ làm gì để giải quyết vấn đề trên?*
  - Công cụ này sẽ có thể tải video reels trên facebook về máy bằng link profile hoặc nhiều link profile facebook. 
  - Tải video tiktok và reels về máy không có watermark.
  - Tự động tạo project capcut theo project template được cấu hình sẵn trong capcut.
  - Tự động export video
  - Có thể tạo phụ đề cho video đã tải. 
  - có thể clone 100% hiệu ứng, cấu hình, text trong project template được tạo sẵn.

* **Giải pháp tóm gọn trong 1-2 câu:**
  - Hệ thống tự động hóa trọn gói (End-to-End Pipeline) viết bằng Go đóng gói `.exe` hỗ trợ song song **Desktop App giao diện đồ họa (Wails v2)** cho nhân sự và **CLI** cho tự động hóa: từ khâu cào video Facebook/TikTok không watermark + bóc tách link/content, tạo phụ đề tự động (Groq / Whisper on-demand), đến khâu nạp vào CapCut Template và tự động xuất (export) video hàng loạt cho 10 nhân sự nội bộ.

---

### 3. Người hưởng lợi & Đối tượng sử dụng
*Dự án này phục vụ cho ai?*
- [x] Tôi và một vài đồng nghiệp trong team nội bộ (Internal team, số lượng ~ 10 người).

---

### 4. Tiêu chí hoàn thành (Definition of Done / Success Criteria)
*Khi nào thì bạn coi là phiên bản đầu tiên đã thành công và dùng được?*


* **Tiêu chí 1:** Tải thành công video từ 1 link profile facebook bất kỳ không dính watermark, lưu kèm file json chứa metadata trong dưới 10 giây.
* **Tiêu chí 2:** Lấy được đầy đủ caption, link được gán vào caption (hoặc trong bình luận của chính tác giả kênh).
* **Tiêu chí 3:** Bóc tách sạch nội dung văn bản từ link lấy được (bỏ qua quảng cáo, menu rác).
* **Tiêu chí 4:** Tạo phụ đề cho video tải về chuẩn xác mốc thời gian (timestamp). 
* **Tiêu chí 5:** Tự động tạo project CapCut theo project template dựng sẵn và xuất ra video hoàn chỉnh chuẩn style/font.
* **Tiêu chí 6:** Tỷ lệ xử lý thành công đạt trên 95% với các link hợp lệ; khi gặp lỗi (link chết, mạng lag) có ghi nhận log chi tiết và tự động chuyển sang video tiếp theo mà không làm gián đoạn chương trình.
