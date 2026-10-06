package database

import "time"

// TrackedSource represents a Facebook Page or TikTok profile monitored for new reels.
type TrackedSource struct {
	ID            int64      `json:"id"`
	Platform      string     `json:"platform"`
	SourceURL     string     `json:"source_url"`
	AuthorName    string     `json:"author_name"`
	LastCheckedAt *time.Time `json:"last_checked_at"`
	VideoCount    int        `json:"video_count,omitempty"`
}

// Video represents a downloaded/processed reel task.
type Video struct {
	VideoID      string     `json:"video_id"`
	SourceID     *int64     `json:"source_id"`
	Platform     string     `json:"platform"`
	Title        string     `json:"title"`
	TemplateName string     `json:"template_name"`
	Status       string     `json:"status"` // pending | downloaded | draft_created | exported | failed
	ExportPath   string     `json:"export_path"`
	ErrorMessage string     `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// PipelineStats summarizes video counts across different processing statuses.
type PipelineStats struct {
	Total        int `json:"total"`
	Pending      int `json:"pending"`
	Downloaded   int `json:"downloaded"`
	DraftCreated int `json:"draft_created"`
	Exported     int `json:"exported"`
	Failed       int `json:"failed"`
}
