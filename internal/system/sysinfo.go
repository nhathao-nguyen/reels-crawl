package system

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// GetFreeDiskSpaceGB returns the free disk space in Gigabytes for the specified path on Windows.
func GetFreeDiskSpaceGB(path string) float64 {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = "C:\\"
	}
	vol := filepath.VolumeName(absPath)
	if vol == "" {
		vol = "C:"
	}
	vol = vol + "\\"

	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	pVol, _ := syscall.UTF16PtrFromString(vol)
	_, _, _ = c.Call(
		uintptr(unsafe.Pointer(pVol)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	return float64(freeBytesAvailable) / (1024 * 1024 * 1024)
}

// OpenFolder opens the specified directory in Windows Explorer.
func OpenFolder(dirPath string) error {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		absPath = dirPath
	}
	cmd := exec.Command("explorer.exe", absPath)
	return cmd.Start()
}

// OpenFile opens a file using the Windows default file association.
func OpenFile(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	cmd := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", absPath)
	return cmd.Start()
}

// VerifyGroqAPIKey sends a lightweight test request to Groq API to verify key validity.
func VerifyGroqAPIKey(apiKey string) (bool, string) {
	if apiKey == "" {
		return false, "API Key đang trống"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://api.groq.com/openai/v1/models", nil)
	if err != nil {
		return false, fmt.Sprintf("Không thể khởi tạo request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Lỗi kết nối tới Groq: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, "Kết nối Groq API thành công! (200 OK)"
	}
	return false, fmt.Sprintf("Groq API phản hồi mã lỗi: %d", resp.StatusCode)
}

// FindCapcutTemplates scans the CapCut drafts directory for any project folder containing the word 'template' (case-insensitive).
func FindCapcutTemplates(draftPath string) []string {
	if draftPath == "" {
		return []string{}
	}

	entries, err := os.ReadDir(draftPath)
	if err != nil {
		return []string{}
	}

	var templates []string
	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), "template") {
			templates = append(templates, entry.Name())
		}
	}
	return templates
}
