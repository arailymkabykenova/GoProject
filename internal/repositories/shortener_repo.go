package repositories

import (
	"database/sql"
	"errors"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var ErrNotFound = errors.New("record not found")

type ShortenerRepository interface {
	InitSchema() error
	SaveMapping(shortCode, longURL string) (int64, error)
	FindByShortCode(shortCode string) (string, error)
	FindByLongURL(longURL string) (string, error)
	UpdateLongURL(shortCode, newLongURL string) error
	DeleteMapping(shortCode string) error
}

type SQLiteShortenerRepo struct {
	db *sql.DB
}

func ConnectDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	log.Println("Database connection established.")
	return db, nil
}

func NewSQLiteShortenerRepo(db *sql.DB) *SQLiteShortenerRepo {
	return &SQLiteShortenerRepo{db: db}
}

func (r *SQLiteShortenerRepo) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		short_code TEXT NOT NULL UNIQUE,
		long_url TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
	CREATE INDEX IF NOT EXISTS idx_long_url ON urls(long_url);
	`
	_, err := r.db.Exec(schema)
	if err != nil {
		log.Printf("Error initializing schema: %v", err)
		return err
	}
	log.Println("Database schema initialized successfully.")
	return nil
}

func (r *SQLiteShortenerRepo) SaveMapping(shortCode, longURL string) (int64, error) {
	stmt, err := r.db.Prepare("INSERT INTO urls(short_code, long_url, created_at) VALUES(?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(shortCode, longURL, time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *SQLiteShortenerRepo) FindByShortCode(shortCode string) (string, error) {
	var longURL string
	err := r.db.QueryRow("SELECT long_url FROM urls WHERE short_code = ?", shortCode).Scan(&longURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return longURL, nil
}

func (r *SQLiteShortenerRepo) FindByLongURL(longURL string) (string, error) {
	var shortCode string
	err := r.db.QueryRow("SELECT short_code FROM urls WHERE long_url = ? LIMIT 1", longURL).Scan(&shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return shortCode, nil
}

func (r *SQLiteShortenerRepo) UpdateLongURL(shortCode, newLongURL string) error {
	stmt, err := r.db.Prepare("UPDATE urls SET long_url = ? WHERE short_code = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(newLongURL, shortCode)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *SQLiteShortenerRepo) DeleteMapping(shortCode string) error {
	stmt, err := r.db.Prepare("DELETE FROM urls WHERE short_code = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(shortCode)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *SQLiteShortenerRepo) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
