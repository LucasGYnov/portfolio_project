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

	CREATE TABLE IF NOT EXISTS me (
        id INTEGER PRIMARY KEY,
        title TEXT,
        content TEXT,
        user_id INTEGER,
		post_id INTEGER
    );

    CREATE TABLE IF NOT EXISTS about (
        id INTEGER PRIMARY KEY,
        content TEXT,
        user_id INTEGER,
		post_id INTEGER
    );

    CREATE TABLE IF NOT EXISTS contact (
        id INTEGER PRIMARY KEY,
		instagram TEXT,
		twitter TEXT,
		behance TEXT,
		github TEXT,
		mail TEXT,
		linkedin TEXT,
		image BLOB,
		user_id INTEGER,
		post_id INTEGER
    );

    CREATE TABLE IF NOT EXISTS formation (
        id INTEGER PRIMARY KEY,
		title TEXT,
		content TEXT,
		years TEXT,
		link TEXT,
		image BLOB,
		user_id INTEGER,
		post_id INTEGER
    );

	CREATE TABLE IF NOT EXISTS project (
        id INTEGER PRIMARY KEY,
		title TEXT,
		content TEXT,
		years TEXT,
		link TEXT,
		image BLOB,
		user_id INTEGER,
		post_id INTEGER
    );

    `
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}
	return db, nil
}