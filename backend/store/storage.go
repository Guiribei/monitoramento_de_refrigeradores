package store

import (
	"context"
	"database/sql"
	"errors"
	_ "modernc.org/sqlite"
	"time"
)

type Store struct {
	db *sql.DB
}

type Snapshot struct {
	FetchedAtMs int64
	RawJSON     []byte
}

func Open(path string) (*Store, error) {
	dsn := "file:" + path + "?mode=rwc&cache=shared&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) initSchema() error {
	const ddl = `
CREATE TABLE IF NOT EXISTS device_latest (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  fetched_at_ms INTEGER NOT NULL,
  raw_json TEXT NOT NULL
);
`
	_, err := s.db.Exec(ddl)
	return err
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) SaveLatest(ctx context.Context, raw []byte, fetchedAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO device_latest (id, fetched_at_ms, raw_json)
VALUES (1, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  fetched_at_ms = excluded.fetched_at_ms,
  raw_json      = excluded.raw_json
`, fetchedAt.UnixMilli(), string(raw))
	return err
}

func (s *Store) GetLatest(ctx context.Context) (*Snapshot, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT fetched_at_ms, raw_json
  FROM device_latest
 WHERE id = 1
`)
	var snap Snapshot
	var raw string
	if err := row.Scan(&snap.FetchedAtMs, &raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	snap.RawJSON = []byte(raw)
	return &snap, nil
}
