package main

import (
	"bufio"
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"reels-crawl/internal/config"
	"reels-crawl/internal/database"
	"reels-crawl/internal/pipeline"
	"reels-crawl/internal/subtitles"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/src
var assets embed.FS

func main() {
	// 1. Check CLI flags for headless/scripting mode
	cliMode := flag.Bool("cli", false, "Chạy ứng dụng qua dòng lệnh (Headless CLI mode)")
	exportOnly := flag.Bool("export", false, "Kích hoạt tự động xuất video hàng loạt cho các dự án CapCut (Pha 2)")
	urlFlag := flag.String("url", "", "URL video Reel đơn lẻ cần cào và tạo dự án")
	fileFlag := flag.String("file", "", "Đường dẫn file .txt chứa danh sách link (mỗi dòng 1 link)")
	tplFlag := flag.String("template", "", "Tên template CapCut áp dụng (ví dụ: template, template-fb)")
	downloadModelFlag := flag.String("download-model", "", "Tải model Whisper Offline về máy (tiny, base, small)")
	flag.Parse()

	if *downloadModelFlag != "" {
		attachParentConsole()
		fmt.Printf("📥 Bắt đầu tải model Whisper '%s' về thư mục models/...\n", *downloadModelFlag)
		err := subtitles.DownloadWhisperModel(*downloadModelFlag, "models", func(pct float64, speed string) {
			fmt.Printf("\r[TIẾN TRÌNH] %5.1f%% (%s)   ", pct, speed)
		})
		if err != nil {
			fmt.Printf("\n❌ Lỗi tải model: %v\n", err)
		} else {
			fmt.Printf("\n🎉 Đã tải và lưu thành công model '%s' tại models/!\n", *downloadModelFlag)
		}
		return
	}

	if *cliMode || *exportOnly || *urlFlag != "" || *fileFlag != "" {
		runCLI(*exportOnly, *urlFlag, *fileFlag, *tplFlag)
		return
	}

	// 2. Desktop GUI Mode (Wails v2)
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "Reels Automation Studio",
		Width:            1280,
		Height:           820,
		MinWidth:         1024,
		MinHeight:        680,
		BackgroundColour: &options.RGBA{R: 11, G: 15, B: 25, A: 1},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			BackdropType:         windows.Mica,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error launching application: %v\n", err)
	}
}

// runCLI provides full command line operations when launched with --cli or --export.
func runCLI(exportOnly bool, singleURL string, filePath string, templateName string) {
	attachParentConsole()

	fmt.Println("==================================================")
	fmt.Println("🚀 Reels Automation Studio - CLI Pipeline Mode")
	fmt.Println("==================================================")

	cfg, err := config.Load("config.json")
	if err != nil {
		defaultCfg := config.Default()
		cfg = &defaultCfg
	}

	if templateName != "" {
		cfg.TemplateName = templateName
	}

	db, err := database.Open("data.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Lỗi khởi tạo SQLite: %v\n", err)
		return
	}
	defer db.Close()

	pipe := pipeline.NewService(db, cfg)

	if exportOnly {
		fmt.Println("\n⚡ [BẮT ĐẦU PHA 2: XUẤT VIDEO HÀNG LOẠT]")
		fmt.Println("--------------------------------------------------")
		success, failed, err := pipe.StartBatchExport(
			context.Background(),
			func(videoID string, percent float64, message string, status string) {
				fmt.Printf("[%5.1f%%] [%s] %s\n", percent, status, message)
			},
		)
		if err != nil {
			fmt.Printf("❌ Lỗi xuất video: %v\n", err)
		} else {
			fmt.Printf("\n🎉 Hoàn thành xuất video! Thành công: %d | Thất bại: %d\n", success, failed)
		}
		return
	}

	// Phase 1: Collect URLs
	var targetURLs []string
	if singleURL != "" {
		targetURLs = append(targetURLs, singleURL)
	}

	if filePath != "" {
		if f, err := os.Open(filePath); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				line = strings.TrimPrefix(line, "\ufeff")
				if line != "" && !strings.HasPrefix(line, "#") {
					targetURLs = append(targetURLs, line)
				}
			}
			f.Close()
		} else {
			fmt.Printf("⚠️ Không thể đọc file %s: %v\n", filePath, err)
		}
	}

	stats, err := db.GetPipelineStats()
	if err == nil {
		fmt.Printf("✓ Thống kê tác vụ: Tổng %d | Chờ %d | Tải gốc %d | Draft %d | Đã xuất %d | Lỗi %d\n",
			stats.Total, stats.Pending, stats.Downloaded, stats.DraftCreated, stats.Exported, stats.Failed)
	}
	sources, _ := db.ListTrackedSources()
	fmt.Printf("✓ Kênh theo dõi: %d nguồn\n", len(sources))

	// If no URLs specified via flags, check pending items in DB
	if len(targetURLs) == 0 {
		pendingVideos, _ := db.ListVideos("pending", 50, 0)
		for _, v := range pendingVideos {
			targetURLs = append(targetURLs, v.VideoID)
		}
	}

	if len(targetURLs) == 0 {
		fmt.Println("\nℹ️ Không có URL nào được cung cấp. Hãy truyền --url <link> hoặc --file <links.txt>")
		fmt.Println("👉 Hoặc chạy với cờ --export để kích hoạt Pha 2 xuất video.")
		return
	}

	fmt.Printf("\n📥 [BẮT ĐẦU PHA 1: CÀO & TẠO CAPCUT DRAFT CHO %d VIDEO]\n", len(targetURLs))
	fmt.Printf("🎯 Template áp dụng: %s\n", cfg.TemplateName)
	fmt.Println("--------------------------------------------------")

	for i, u := range targetURLs {
		fmt.Printf("\n[%d/%d] 🔄 Bắt đầu xử lý: %s\n", i+1, len(targetURLs), u)
		_, err := pipe.ProcessSingleURL(
			context.Background(),
			u,
			cfg.TemplateName,
			cfg.SkipDuplicates,
			cfg.ExtractCommentLinks,
			func(event pipeline.ProgressEvent) {
				var icon string
				switch event.Step {
				case "check_db":
					icon = "🔍"
				case "downloading":
					icon = "⬇️"
				case "scraping":
					icon = "📄"
				case "transcribing":
					icon = "🎙️"
				case "capcut_inject":
					icon = "🎬"
				case "complete":
					icon = "✅"
				case "failed":
					icon = "❌"
				default:
					icon = "ℹ️"
				}
				fmt.Printf("  ├── %s [%5.1f%%] %s\n", icon, event.Percent, event.Message)
			},
		)
		if err != nil {
			fmt.Printf("  └── ❌ Thất bại: %v\n", err)
		} else {
			fmt.Println("  └── 🎉 Xử lý thành công!")
		}
	}

	fmt.Println("\n==================================================")
	fmt.Println("🎉 TOÀN BỘ DANH SÁCH ĐÃ ĐƯỢC XỬ LÝ XONG!")
	fmt.Println("👉 Chạy lại với cờ --export để kích hoạt Pha 2 xuất video thành phẩm.")
	fmt.Println("==================================================")
}

func attachParentConsole() {
	mod := syscall.NewLazyDLL("kernel32.dll")
	proc := mod.NewProc("AttachConsole")
	const ATTACH_PARENT_PROCESS = ^uintptr(0) // (DWORD)-1
	r1, _, _ := proc.Call(ATTACH_PARENT_PROCESS)
	if r1 != 0 {
		procCP := mod.NewProc("SetConsoleOutputCP")
		procCP.Call(uintptr(65001))

		conout, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
		if err == nil {
			os.Stdout = conout
			os.Stderr = conout
		}
	}
}
