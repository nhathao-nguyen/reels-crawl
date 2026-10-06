package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the SQL connection pool and exposes repository operations.
type DB struct {
	conn *sql.DB
}

// Open initializes or connects to the SQLite database at dbPath and runs schema migrations.
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// SQLite connection string with busy timeout and WAL pragmas
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
	}

	// Set connection pool limits
	conn.SetMaxOpenConns(1) // modernc SQLite works best with 1 writer connection
	conn.SetMaxIdleConns(1)

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn returns the underlying *sql.DB
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// migrate creates necessary tables and indexes if they do not already exist.
func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tracked_sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		platform TEXT NOT NULL,
		source_url TEXT UNIQUE NOT NULL,
		author_name TEXT,
		last_checked_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS videos (
		video_id TEXT PRIMARY KEY,
		source_id INTEGER NULL REFERENCES tracked_sources(id) ON DELETE SET NULL,
		platform TEXT NOT NULL,
		title TEXT DEFAULT '',
		template_name TEXT DEFAULT '',
		status TEXT NOT NULL,
		export_path TEXT,
		error_message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_videos_source_id ON videos(source_id);
	CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
	CREATE INDEX IF NOT EXISTS idx_tracked_sources_platform ON tracked_sources(platform);
	`

	if _, err := db.conn.Exec(schema); err != nil {
		return err
	}

	// Ensure template_name column exists for databases created before this change
	_, _ = db.conn.Exec("ALTER TABLE videos ADD COLUMN template_name TEXT DEFAULT '';")
	return nil
}
