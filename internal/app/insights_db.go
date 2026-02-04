package app

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func OpenInsightsDB(dbPath string) (*sql.DB, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, nil
	}

	db, err := sql.Open("sqlite3", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func DefaultInsightsDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/bujo-companion/claude-insights.db"
}
