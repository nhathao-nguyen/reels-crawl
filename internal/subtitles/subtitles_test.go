package subtitles

import (
	"strings"
	"testing"
)

func TestFormatSRTTime(t *testing.T) {
	tests := []struct {
		sec  float64
		want string
	}{
		{0.0, "00:00:00,000"},
		{1.25, "00:00:01,250"},
		{65.5, "00:01:05,500"},
		{3661.002, "01:01:01,002"},
	}

	for _, tc := range tests {
		got := FormatSRTTime(tc.sec)
		if got != tc.want {
			t.Errorf("FormatSRTTime(%f) = %s; want %s", tc.sec, got, tc.want)
		}
	}
}

func TestGenerateDynamicSubtitlesFromMeta(t *testing.T) {
	title := "Khám phá phong cảnh tuyệt đẹp"
	caption := "Hôm nay chúng ta cùng đi du lịch trải nghiệm vẻ đẹp thiên nhiên kỳ thú. https://example.com/tour Đừng quên theo dõi kênh nhé!"
	uploader := "dulich_vietnam"
	duration := 12.0

	segs := GenerateDynamicSubtitlesFromMeta(title, caption, uploader, duration)
	if len(segs) == 0 {
		t.Fatalf("Expected non-empty segments, got 0")
	}

	// Verify duration is respected
	first := segs[0]
	last := segs[len(segs)-1]

	if first.Start != 0.0 {
		t.Errorf("First segment start should be 0.0, got %f", first.Start)
	}
	if last.End != duration {
		t.Errorf("Last segment end should be %f, got %f", duration, last.End)
	}

	// Verify no URLs in subtitle text
	for _, s := range segs {
		if strings.Contains(s.Text, "http") {
			t.Errorf("Subtitle text should not contain URL, got: %s", s.Text)
		}
	}
}

func TestListWhisperModels(t *testing.T) {
	models := ListWhisperModels("models")
	if len(models) < 3 {
		t.Errorf("Expected at least 3 models, got %d", len(models))
	}
}

func TestParseSRTContent(t *testing.T) {
	srtText := `1
00:00:01,000 --> 00:00:03,500
Xin chào các bạn

2
00:00:03,500 --> 00:00:06,000
Chào mừng đến với Reels Tool
`

	segs := ParseSRTContent(srtText)
	if len(segs) != 2 {
		t.Fatalf("Expected 2 segments, got %d", len(segs))
	}

	if segs[0].Text != "Xin chào các bạn" {
		t.Errorf("Unexpected text 0: %s", segs[0].Text)
	}
	if segs[0].Start != 1.0 || segs[0].End != 3.5 {
		t.Errorf("Unexpected timing 0: %f -> %f", segs[0].Start, segs[0].End)
	}

	if segs[1].Text != "Chào mừng đến với Reels Tool" {
		t.Errorf("Unexpected text 1: %s", segs[1].Text)
	}
	if segs[1].Start != 3.5 || segs[1].End != 6.0 {
		t.Errorf("Unexpected timing 1: %f -> %f", segs[1].Start, segs[1].End)
	}
}
