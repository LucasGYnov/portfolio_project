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
        content TEXT,
        exp_id INTEGER,
		FOREIGN KEY (exp_id) REFERENCES utilisateurs (id)
    );

    CREATE TABLE IF NOT EXISTS contact (
        id INTEGER PRIMARY KEY,
		numero TEXT,
		email TEXT,
		postal TEXT,
		contact_id INTEGER,
		FOREIGN KEY (contact_id) REFERENCES utilisateurs (id)
    );

    CREATE TABLE IF NOT EXISTS formation (
        id INTEGER PRIMARY KEY,
		title TEXT,
		years TEXT,
		formation_id INTEGER,
		FOREIGN KEY (formation_id) REFERENCES utilisateurs (id)
    );

    CREATE TABLE IF NOT EXISTS tech (
        id INTEGER PRIMARY KEY,
		title TEXT,
		content TEXT,
		tech_id INTEGER,
		FOREIGN KEY (tech_id) REFERENCES utilisateurs (id)
    );

    `
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}
	return db, nil
}