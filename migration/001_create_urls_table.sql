CREATE TABLE IF NOT EXISTS urls (
                                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                                    short_code TEXT NOT NULL UNIQUE,
                                    long_url TEXT NOT NULL,
                                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
