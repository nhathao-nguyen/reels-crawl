package database

import (
	"database/sql"
	"fmt"
	"time"
)

// IsVideoExists checks if a video ID is already tracked in the database.
// This fulfills the deduplication requirement (<10ms lookup).
func (db *DB) IsVideoExists(videoID string) (bool, error) {
	var exists int
	err := db.conn.QueryRow("SELECT 1 FROM videos WHERE video_id = ? LIMIT 1", videoID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to query video existence: %w", err)
	}
	return true, nil
}

// UpsertVideo inserts a new video record or updates status/error/paths if it exists.
func (db *DB) UpsertVideo(v *Video) error {
	query := `
	INSERT INTO videos (video_id, source_id, platform, title, template_name, status, export_path, error_message, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(video_id) DO UPDATE SET
		source_id = COALESCE(excluded.source_id, videos.source_id),
		platform = excluded.platform,
		title = CASE WHEN excluded.title != '' THEN excluded.title ELSE videos.title END,
		template_name = CASE WHEN excluded.template_name != '' THEN excluded.template_name ELSE videos.template_name END,
		status = excluded.status,
		export_path = CASE WHEN excluded.export_path != '' THEN excluded.export_path ELSE videos.export_path END,
		error_message = excluded.error_message,
		updated_at = CURRENT_TIMESTAMP;
	`
	_, err := db.conn.Exec(query, v.VideoID, v.SourceID, v.Platform, v.Title, v.TemplateName, v.Status, v.ExportPath, v.ErrorMessage)
	if err != nil {
		return fmt.Errorf("failed to upsert video %s: %w", v.VideoID, err)
	}
	return nil
}

// UpdateVideoStatus updates the status, export path, and error message of a video.
func (db *DB) UpdateVideoStatus(videoID string, status string, exportPath string, errorMsg string) error {
	query := `
	UPDATE videos
	SET status = ?, export_path = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP
	WHERE video_id = ?;
	`
	_, err := db.conn.Exec(query, status, exportPath, errorMsg, videoID)
	if err != nil {
		return fmt.Errorf("failed to update status for video %s: %w", videoID, err)
	}
	return nil
}

// GetVideo retrieves a single video by its ID.
func (db *DB) GetVideo(videoID string) (*Video, error) {
	query := `
	SELECT video_id, source_id, platform, title, template_name, status, export_path, error_message, created_at, updated_at
	FROM videos
	WHERE video_id = ?;
	`
	var v Video
	var sourceID sql.NullInt64
	var exportPath, errorMsg sql.NullString

	err := db.conn.QueryRow(query, videoID).Scan(
		&v.VideoID,
		&sourceID,
		&v.Platform,
		&v.Title,
		&v.TemplateName,
		&v.Status,
		&exportPath,
		&errorMsg,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get video %s: %w", videoID, err)
	}

	if sourceID.Valid {
		v.SourceID = &sourceID.Int64
	}
	if exportPath.Valid {
		v.ExportPath = exportPath.String
	}
	if errorMsg.Valid {
		v.ErrorMessage = errorMsg.String
	}

	return &v, nil
}

// ListVideos lists videos with optional status filtering. Pass statusFilter = "" for all.
func (db *DB) ListVideos(statusFilter string, limit int, offset int) ([]*Video, error) {
	if limit <= 0 {
		limit = 100
	}

	var rows *sql.Rows
	var err error

	if statusFilter != "" && statusFilter != "all" {
		query := `
		SELECT video_id, source_id, platform, title, template_name, status, export_path, error_message, created_at, updated_at
		FROM videos
		WHERE status = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?;
		`
		rows, err = db.conn.Query(query, statusFilter, limit, offset)
	} else {
		query := `
		SELECT video_id, source_id, platform, title, template_name, status, export_path, error_message, created_at, updated_at
		FROM videos
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?;
		`
		rows, err = db.conn.Query(query, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list videos: %w", err)
	}
	defer rows.Close()

	var list []*Video
	for rows.Next() {
		var v Video
		var sourceID sql.NullInt64
		var exportPath, errorMsg sql.NullString

		if err := rows.Scan(
			&v.VideoID,
			&sourceID,
			&v.Platform,
			&v.Title,
			&v.TemplateName,
			&v.Status,
			&exportPath,
			&errorMsg,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan video row: %w", err)
		}

		if sourceID.Valid {
			v.SourceID = &sourceID.Int64
		}
		if exportPath.Valid {
			v.ExportPath = exportPath.String
		}
		if errorMsg.Valid {
			v.ErrorMessage = errorMsg.String
		}

		list = append(list, &v)
	}

	return list, nil
}

// DeleteVideo removes a video record from the database.
func (db *DB) DeleteVideo(videoID string) error {
	_, err := db.conn.Exec("DELETE FROM videos WHERE video_id = ?", videoID)
	return err
}

// ClearCompletedVideos deletes all video records with status 'exported'.
func (db *DB) ClearCompletedVideos() (int64, error) {
	res, err := db.conn.Exec("DELETE FROM videos WHERE status = 'exported'")
	if err != nil {
		return 0, fmt.Errorf("failed to clear completed videos: %w", err)
	}
	return res.RowsAffected()
}

// GetPipelineStats calculates video counts aggregated by status.
func (db *DB) GetPipelineStats() (*PipelineStats, error) {
	query := `
	SELECT
		COUNT(*) as total,
		SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
		SUM(CASE WHEN status = 'downloaded' THEN 1 ELSE 0 END) as downloaded,
		SUM(CASE WHEN status = 'draft_created' THEN 1 ELSE 0 END) as draft_created,
		SUM(CASE WHEN status = 'exported' THEN 1 ELSE 0 END) as exported,
		SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
	FROM videos;
	`
	var stats PipelineStats
	var pending, downloaded, draftCreated, exported, failed sql.NullInt64

	err := db.conn.QueryRow(query).Scan(
		&stats.Total,
		&pending,
		&downloaded,
		&draftCreated,
		&exported,
		&failed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline stats: %w", err)
	}

	if pending.Valid {
		stats.Pending = int(pending.Int64)
	}
	if downloaded.Valid {
		stats.Downloaded = int(downloaded.Int64)
	}
	if draftCreated.Valid {
		stats.DraftCreated = int(draftCreated.Int64)
	}
	if exported.Valid {
		stats.Exported = int(exported.Int64)
	}
	if failed.Valid {
		stats.Failed = int(failed.Int64)
	}

	return &stats, nil
}

// AddTrackedSource inserts a new tracked channel or returns existing record ID.
func (db *DB) AddTrackedSource(source *TrackedSource) (int64, error) {
	query := `
	INSERT INTO tracked_sources (platform, source_url, author_name, last_checked_at)
	VALUES (?, ?, ?, NULL)
	ON CONFLICT(source_url) DO UPDATE SET
		author_name = CASE WHEN excluded.author_name != '' THEN excluded.author_name ELSE tracked_sources.author_name END
	RETURNING id;
	`
	var id int64
	err := db.conn.QueryRow(query, source.Platform, source.SourceURL, source.AuthorName).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert tracked source: %w", err)
	}
	source.ID = id
	return id, nil
}

// ListTrackedSources lists all tracked pages/profiles along with their video count.
func (db *DB) ListTrackedSources() ([]*TrackedSource, error) {
	query := `
	SELECT s.id, s.platform, s.source_url, COALESCE(s.author_name, ''), s.last_checked_at, COUNT(v.video_id) as video_count
	FROM tracked_sources s
	LEFT JOIN videos v ON s.id = v.source_id
	GROUP BY s.id
	ORDER BY s.id DESC;
	`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tracked sources: %w", err)
	}
	defer rows.Close()

	var sources []*TrackedSource
	for rows.Next() {
		var s TrackedSource
		var lastChecked sql.NullTime

		if err := rows.Scan(&s.ID, &s.Platform, &s.SourceURL, &s.AuthorName, &lastChecked, &s.VideoCount); err != nil {
			return nil, fmt.Errorf("failed to scan tracked source: %w", err)
		}

		if lastChecked.Valid {
			t := lastChecked.Time
			s.LastCheckedAt = &t
		}

		sources = append(sources, &s)
	}

	return sources, nil
}

// UpdateTrackedSourceChecked updates the last_checked_at timestamp for a source.
func (db *DB) UpdateTrackedSourceChecked(id int64, t time.Time) error {
	_, err := db.conn.Exec("UPDATE tracked_sources SET last_checked_at = ? WHERE id = ?", t, id)
	return err
}

// DeleteTrackedSource deletes a tracked source from the database.
func (db *DB) DeleteTrackedSource(id int64) error {
	_, err := db.conn.Exec("DELETE FROM tracked_sources WHERE id = ?", id)
	return err
}
