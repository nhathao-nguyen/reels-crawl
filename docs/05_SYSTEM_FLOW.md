# 05 - Sơ Đồ Luồng Tổng Thể Hệ Thống (System Flowchart & Sequence Diagram)

> **Mục tiêu:** Trực quan hóa toàn bộ luồng hoạt động (End-to-End Pipeline) của hệ thống bằng sơ đồ chuẩn Mermaid, giúp lập trình viên và các bên liên quan nắm bắt kiến trúc xử lý 2 pha trong 30 giây.

---

## 1. Sơ đồ luồng hoạt động tổng quan (Overall Flowchart)

```mermaid
flowchart TD
    Start([Khởi động Tool: Wails Desktop App / CLI]) --> ModeCheck{Chế độ chạy?}
    
    ModeCheck -->|Mặc định / GUI| OpenGUI[Hiển thị giao diện Wails v2]
    ModeCheck -->|--cli hoặc --export| ExecCLI[Thực thi qua Command Line]
    
    OpenGUI --> ActionSelect{Người dùng chọn thao tác}
    ExecCLI --> ActionSelect
    
    %% --- PHA 1: THU THẬP & TẠO PROJECT NGẦM ---
    ActionSelect -->|Nạp link đơn lẻ / File danh sách / Quét Page| P1_Start[BẮT ĐẦU PHA 1: XỬ LÝ NGẦM]
    
    subgraph Phase1 [PHA 1: THU THẬP & TẠO DRAFT CAPCUT NGẦM]
        P1_Start --> CheckDB{Đã tồn tại trong SQLite?}
        CheckDB -->|Đã có video_id| SkipVideo[Bỏ qua - Skip video]
        
        CheckDB -->|Chưa có| FetchInfo[Quét GraphQL / DOM lấy Metadata & Direct URL]
        FetchInfo --> YtDlpDownload[yt-dlp tải raw_video.mp4 không watermark]
        
        YtDlpDownload --> LinkExtract{Có link web trong caption/comment?}
        LinkExtract -->|Có| UnshortenLink[Unshorten HEAD Request -> Bóc tách Trafilatura -> content/]
        LinkExtract -->|Không| SkipWeb[Bỏ qua cào web]
        
        UnshortenLink --> ExtractAudio[FFmpeg tách audio.wav 16kHz mono]
        SkipWeb --> ExtractAudio
        
        ExtractAudio --> STTMode{Chọn engine STT?}
        STTMode -->|Cloud API| GroqSTT[Gửi Groq Whisper API ~1-2s -> subtitles.srt]
        STTMode -->|Local Offline| LocalSTT[Chạy faster-whisper / whisper.cpp -> subtitles.srt]
        
        GroqSTT --> CapCutInject[Đọc template-fb -> Inject video + .srt vào draft_content.json]
        LocalSTT --> CapCutInject
        
        CapCutInject --> SaveDB[Ghi SQLite: status = 'draft_created']
    end
    
    SaveDB --> NextVideo{Còn video tiếp theo?}
    SkipVideo --> NextVideo
    NextVideo -->|Còn link| P1_Start
    NextVideo -->|Hết danh sách| P1_Done([Hoàn tất Pha 1 - Sẵn sàng xuất])
    
    %% --- PHA 2: AUTO EXPORT HÀNG LOẠT ---
    ActionSelect -->|Bấm 'Xuất Video' hoặc --export| P2_Start[BẮT ĐẦU PHA 2: XUẤT HÀNG LOẠT]
    
    subgraph Phase2 [PHA 2: TỰ ĐỘNG HÓA XUẤT CAPCUT - UI AUTOMATION]
        P2_Start --> PreCheckCapCut{CapCut có cài trên máy?}
        PreCheckCapCut -->|Không| ErrorNoCapCut[Báo lỗi: Yêu cầu cài CapCut Desktop]
        PreCheckCapCut -->|Có| CheckRunning{CapCut đang mở sẵn?}
        
        CheckRunning -->|Đang mở| KillOldCapcut[Cảnh báo & Kill CapCut.exe cũ để làm sạch môi trường]
        CheckRunning -->|Không mở| FilterProjects[Quét com.lveditor.draft: Bỏ qua 'template*' và 'exported']
        KillOldCapcut --> FilterProjects
        
        FilterProjects --> LoopProjects{Còn project 'draft_created'?}
        LoopProjects -->|Hết| P2_Done([Hoàn thành toàn bộ video!])
        
        LoopProjects -->|Còn| LaunchProject[Mở CapCut với project hiện tại]
        LaunchProject --> ClosePopups[Đóng pop-up thông báo / update rác]
        ClosePopups --> SendShortcut[Gửi phím tắt Ctrl + E mở Export Dialog]
        SendShortcut --> ClickExport[Chọn 1080p -> Bấm Export -> Giám sát render]
        
        ClickExport --> ExportMonitor{Kết quả render?}
        ExportMonitor -->|Thành công 100%| MarkExported[Lưu final_video.mp4 -> Update SQLite: 'exported']
        ExportMonitor -->|Treo quá 5 phút / Lỗi| HandleHang[Kill CapCut.exe -> Log 'timeout' -> Thử lại 1 lần / Bỏ qua]
        
        MarkExported --> CloseCapcut[Đóng CapCut -> Chuyển project kế tiếp]
        HandleHang --> CloseCapcut
        CloseCapcut --> LoopProjects
    end
```

---

## 2. Sơ đồ tuần tự tương tác (Sequence Diagram) - Pha 1: Crawl & Inject

```mermaid
sequenceDiagram
    autonumber
    actor User as Nhân sự (User)
    participant UI as Wails Desktop GUI
    participant Core as Engine Điều Phối (Go Core)
    participant DB as SQLite (data.db)
    participant YTDLP as yt-dlp.exe
    participant Web as Web Scraper (Trafilatura)
    participant FFmpeg as bin/ffmpeg.exe
    participant STT as Groq API / Whisper Local
    participant CapCutFS as CapCut Draft Storage

    User->>UI: Nhập URL Profile / File danh sách link + Bấm "Bắt đầu"
    UI->>Core: StartPipeline(sources, config)
    
    loop Duyệt từng video phát hiện được
        Core->>DB: CheckExist(video_id)
        alt Đã tồn tại trong DB
            DB-->>Core: Record exists
            Core->>UI: UpdateProgress(video_id, "Skipped - Đã có")
        else Chưa có trong DB
            Core->>DB: InsertVideo(video_id, source_id, status='downloading')
            Core->>YTDLP: DownloadVideo(url, target_dir)
            YTDLP-->>Core: Trả về raw_video.mp4 sạch
            
            opt Có link trong caption/comment
                Core->>Web: Unshorten & ExtractContent(article_url)
                Web-->>Core: Lưu content/ vào thư mục video
            end
            
            Core->>FFmpeg: Trích xuất audio 16kHz mono WAV
            FFmpeg-->>Core: File audio.wav
            
            Core->>STT: Transcribe(audio.wav)
            STT-->>Core: File phụ đề subtitles.srt
            
            Core->>CapCutFS: Sao chép template-fb thành project mới (UUID)
            Core->>CapCutFS: Cập nhật draft_content.json (video path, duration, texts)
            CapCutFS-->>Core: Đăng ký draft vào root_meta_info.json
            
            Core->>DB: UpdateStatus(video_id, status='draft_created')
            Core->>UI: UpdateProgress(video_id, "Hoàn tất Draft", 100%)
        end
    end
    Core->>UI: Thông báo hoàn tất Pha 1
    UI->>User: Hiển thị nút "Kích hoạt Auto Export"
```

---

## 3. Sơ đồ tuần tự tương tác (Sequence Diagram) - Pha 2: Auto Export

```mermaid
sequenceDiagram
    autonumber
    actor User as Nhân sự (User)
    participant UI as Wails Desktop GUI
    participant Auto as Auto Export Worker (Go)
    participant WinProc as Windows Process Manager
    participant CapCutApp as CapCut Desktop UI
    participant WinUI as Windows UI Automation API
    participant DB as SQLite (data.db)

    User->>UI: Bấm nút "Xuất Video Hàng Loạt"
    UI->>Auto: StartBatchExport()
    
    Auto->>WinProc: CheckProcess("CapCut.exe")
    alt CapCut đang mở sẵn
        Auto->>UI: Hiển thị cảnh báo: "Đang đóng CapCut để làm sạch môi trường"
        Auto->>WinProc: KillProcess("CapCut.exe")
    end
    
    Auto->>DB: GetPendingExportVideos()
    DB-->>Auto: Danh sách video_ids (chỉ lấy status = 'draft_created')
    
    loop Với từng dự án chưa xuất
        Auto->>WinProc: StartProcess("CapCut.exe", draft_id)
        Auto->>WinUI: WaitForWindow("CapCut", timeout=15s)
        Auto->>WinUI: DetectAndDismissPopups()
        
        Auto->>WinUI: SendKeys("Ctrl+E")
        Auto->>WinUI: SetExportResolution("1080p")
        Auto->>WinUI: ClickButton("Export")
        
        loop Giám sát Render (Polling mỗi 2 giây, timeout 300s)
            Auto->>WinUI: CheckRenderProgress()
            alt Render đạt 100% / Xuất hiện file .mp4
                WinUI-->>Auto: Render Complete
                Auto->>DB: UpdateStatus(video_id, status='exported', export_path)
                Auto->>UI: UpdateExportProgress(video_id, "Thành công")
            else Vượt quá 300 giây (Treo)
                Auto->>WinProc: ForceKill("CapCut.exe")
                Auto->>DB: UpdateStatus(video_id, status='failed', error='Export Timeout')
                Auto->>UI: UpdateExportProgress(video_id, "Lỗi treo - Đã bỏ qua")
            end
        end
        
        Auto->>WinProc: CloseProcess("CapCut.exe")
    end
    
    Auto->>UI: Báo cáo tổng kết (Số lượng xuất thành công / thất bại)
    UI->>User: Thông báo chuông/popup hoàn tất toàn bộ quá trình
```
