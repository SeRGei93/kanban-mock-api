package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3" // for init driver
	"kanban-app/internal/storage"
)

type Storage struct {
	db *sql.DB
}

type Column struct {
	id   int64
	name string
	sort int64
}

type Card struct {
	id       int64
	name     string
	content  string
	sort     int64
	columnId int64
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = migrate(db)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) AddColumn(name string, sort int64) (int64, error) {
	const op = "storage.sqlite.AddColumn"

	stmt, err := s.db.Prepare(`INSERT INTO columns (name, sort) VALUES (?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(name, sort)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrColumnExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) AddCard(name string, content string, sort int64, columnId int64) (int64, error) {
	const op = "storage.sqlite.AddCard"

	stmt, err := s.db.Prepare(`INSERT INTO cards (name, content, sort, columnId) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(name, content, sort, columnId)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrCardExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetColumns() ([]Column, error) {
	const op = "storage.sqlite.GetColumns"

	rows, err := s.db.Query("SELECT * FROM columns")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var items = make([]Column, 0)
	for rows.Next() {
		var c Column
		err := rows.Scan(&c.id, &c.name, &c.sort)
		if err != nil {
			return nil, err
		}

		items = append(items, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Storage) GetCards() ([]Card, error) {
	const op = "storage.sqlite.GetCards"

	rows, err := s.db.Query("SELECT * FROM items")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var items = make([]Card, 0)
	for rows.Next() {
		var c Card
		err := rows.Scan(&c.id, &c.name, &c.content, &c.sort, &c.columnId)
		if err != nil {
			return nil, err
		}

		items = append(items, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func migrate(db *sql.DB) error {
	const op = "storage.sqlite.migrate"
	queries := []string{
		`CREATE TABLE IF NOT EXISTS columns (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			sort INTEGER DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS cards (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			content TEXT DEFAULT '',
			sort INTEGER DEFAULT 0,
			column_id INTEGER,
			FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE ON UPDATE CASCADE
		);`,
	}

	for _, query := range queries {
		stmt, err := db.Prepare(query)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		_, err = stmt.Exec()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
