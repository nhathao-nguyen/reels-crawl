package crawler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.facebook.com/reel/123456789", "facebook"},
		{"https://fb.watch/abcd1234/", "facebook"},
		{"https://www.tiktok.com/@user/video/987654321", "tiktok"},
		{"https://vt.tiktok.com/ZS2abcd/", "tiktok"},
		{"https://www.youtube.com/watch?v=xyz", "youtube"},
	}

	for _, tt := range tests {
		got := DetectPlatform(tt.url)
		if got != tt.expected {
			t.Errorf("DetectPlatform(%s) = %s; want %s", tt.url, got, tt.expected)
		}
	}
}

func TestExtractVideoID(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.facebook.com/reel/991736342092712", "991736342092712"},
		{"https://www.tiktok.com/@creator/video/718293847561928374", "718293847561928374"},
		{"https://vt.tiktok.com/ZS2k9x81/", "ZS2k9x81"},
	}

	for _, tt := range tests {
		got := ExtractVideoID(tt.url)
		if got != tt.expected {
			t.Errorf("ExtractVideoID(%s) = %s; want %s", tt.url, got, tt.expected)
		}
	}
}

func TestExtractLinksFromText(t *testing.T) {
	sample := "Xem chi tiết tại https://example.com/story-chapter-1 hoặc mua hàng https://bit.ly/deal123."
	links := ExtractLinksFromText(sample)
	if len(links) != 2 {
		t.Fatalf("Expected 2 links, got %d", len(links))
	}
	if links[0] != "https://example.com/story-chapter-1" {
		t.Errorf("Unexpected link 0: %s", links[0])
	}
	if links[1] != "https://bit.ly/deal123" {
		t.Errorf("Unexpected link 1: %s", links[1])
	}
}

func TestCleanURL(t *testing.T) {
	dirty := "https://example.com/article?fbclid=IwAR123&utm_source=fb#google_vignette"
	cleaned := CleanURL(dirty)
	expected := "https://example.com/article"
	if cleaned != expected {
		t.Errorf("CleanURL(%s) = %s; want %s", dirty, cleaned, expected)
	}
}

func TestIsProfileURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://www.tiktok.com/@creator", true},
		{"https://www.tiktok.com/@creator/video/123456", false},
		{"https://www.facebook.com/ghienphimhay", true},
		{"https://www.facebook.com/reel/123456", false},
		{"https://www.facebook.com/watch/?v=123456", false},
	}

	for _, tt := range tests {
		got := IsProfileURL(tt.url)
		if got != tt.expected {
			t.Errorf("IsProfileURL(%s) = %v; want %v", tt.url, got, tt.expected)
		}
	}
}

func TestScrapeArticleWithChapters(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if strings.Contains(r.URL.Path, "chap-2") {
			fmt.Fprintln(w, `<html><head><title>Chapter 2</title></head><body>
				<article><p>Day la noi dung cua chuong thu hai voi nhieu tinh tiet hap dan va kich tinh.</p></article>
			</body></html>`)
		} else {
			fmt.Fprintf(w, `<html><head><title>Chapter 1</title></head><body>
				<article><p>Day la noi dung chinh cua chuong thu nhat mo dau cho cau chuyen dai tap.</p></article>
				<a href="%s/chap-2">Chương sau</a>
			</body></html>`, ts.URL)
		}
	}))
	defer ts.Close()

	art, err := ScrapeArticleWithChapters(ts.URL+"/chap-1", 5)
	if err != nil {
		t.Fatalf("ScrapeArticleWithChapters failed: %v", err)
	}

	if art == nil {
		t.Fatalf("Expected non-nil article")
	}

	if !strings.Contains(art.MainText, "chuong thu nhat") {
		t.Errorf("Missing chapter 1 content")
	}
	if !strings.Contains(art.MainText, "chuong thu hai") {
		t.Errorf("Missing chapter 2 content")
	}

	// Test SaveArticleToDirectory
	tmpDir := t.TempDir()
	err = SaveArticleToDirectory(art, tmpDir)
	if err != nil {
		t.Fatalf("SaveArticleToDirectory failed: %v", err)
	}

	summaryFile := filepath.Join(tmpDir, "content", "summary.json")
	if _, err := os.Stat(summaryFile); err != nil {
		t.Errorf("Expected summary.json to exist at %s", summaryFile)
	}

	fullArticleFile := filepath.Join(tmpDir, "content", "full_article.txt")
	if _, err := os.Stat(fullArticleFile); err != nil {
		t.Errorf("Expected full_article.txt to exist at %s", fullArticleFile)
	}
}
