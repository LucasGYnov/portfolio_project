package Data

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./Data/Data.db")
	if err != nil {
		return nil, err
	}

	// Créez la table utilisateurs si elle n'existe pas
	createTable := `
    CREATE TABLE IF NOT EXISTS utilisateurs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        password TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS experience (
        id INTEGER PRIMARY KEY,
        title TEXT,
        content TEXT
    );

    CREATE TABLE IF NOT EXISTS contact (
        id INTEGER PRIMARY KEY,
        numero TEXT,
        email TEXT,
        postal TEXT
    );

    CREATE TABLE IF NOT EXISTS formation (
        id INTEGER PRIMARY KEY,
        title TEXT,
        years TEXT
    );

    CREATE TABLE IF NOT EXISTS tech (
        id INTEGER PRIMARY KEY,
        title TEXT,
        content TEXT
    );

    `
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}
	return db, nil
}
