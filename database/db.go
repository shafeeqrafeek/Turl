package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func (db *DB) InsertUrl(url, shortCode string) error {
	_, err := db.Exec("INSERT INTO urls (url, short_code) VALUES (?, ?)", url, shortCode)
	return err
}

func (db *DB) GetUrls() (map[string]string, error) {
	var url, shortCode string
	urlData := map[string]string{}

	results, err := db.Query("SELECT url, short_code FROM urls")
	if err != nil {
		return urlData, err
	}
	defer results.Close()

	for results.Next() {
		err := results.Scan(&url, &shortCode)
		if err != nil {
			return urlData, err
		}
		urlData[shortCode] = url
	}
	return urlData, nil

}

// GetUrl retrieves the url corresponding to a short code
func (db *DB) GetUrl(shortCode string) (string, error) {
	var url string
	err := db.QueryRow("SELECT url FROM urls WHERE short_code = ?", shortCode).Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

func CreateDB() (*DB, error) {
	db, err := sql.Open("sqlite3", "file:turl.db")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (id INTEGER PRIMARY KEY, url TEXT, short_code TEXT)`)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
