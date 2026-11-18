package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"id.azz/pow/pkg/logger"
	_ "modernc.org/sqlite"
)

type sqliteDB struct {
	db *sql.DB
}

func InitDB(dbPath string) (*sqliteDB, error) {
	// Initialize the database connection and create necessary tables
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		criteria TEXT NOT NULL,
		active BOOLEAN DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS daily_audits (
		audit_date TEXT PRIMARY KEY,
		result TEXT NOT NULL,
		what_you_did TEXT NOT NULL,
		raw_evidence TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	logger.Println("Database initialized successfully.")

	return &sqliteDB{db: db}, nil
}

func (s *sqliteDB) Close() {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			logger.Println("Error closing database:", err)
		} else {
			logger.Println("Database connection closed.")
		}
	}
}

func (s *sqliteDB) GetActiveRulesText(ctx context.Context) (string, error) {
	rows, err := s.db.Query("SELECT title, criteria FROM rules WHERE active = 1")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return "", rows.Err()
	}

	var rulesText strings.Builder
	for rows.Next() {
		var title, criteria string
		if err := rows.Scan(&title, &criteria); err != nil {
			return "", err
		}
		fmt.Fprintf(&rulesText, "- %s: %s\n", title, criteria)
	}

	return rulesText.String(), nil
}
