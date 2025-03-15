package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3" // for init driver
	"kanban-app/internal/config"
	"kanban-app/internal/storage"
)

type Storage struct {
	db *sql.DB
}

type Column struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Sort int64  `json:"sort"`
}

type Card struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	Sort     int64  `json:"sort"`
	ColumnId int64  `json:"columnId"`
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if config.MigrateFlag {
		err = migrate(db)
		if err != nil {
			return nil, err
		}

	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveColumn(name string, sort int64) (int64, error) {
	const op = "storage.sqlite.SaveColumn"

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

func (s *Storage) SaveCard(name string, content string, sort int64, columnId int64) (int64, error) {
	const op = "storage.sqlite.SaveCard"

	stmt, err := s.db.Prepare(`INSERT INTO cards (name, content, sort, column_id) VALUES (?, ?, ?, ?)`)
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
		err := rows.Scan(&c.Id, &c.Name, &c.Sort)
		if err != nil {
			return nil, err
		}

		items = append(items, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to get columns: %w", op, err)
	}

	return items, nil
}

func (s *Storage) GetCards() ([]Card, error) {
	const op = "storage.sqlite.GetCards"

	rows, err := s.db.Query("SELECT * FROM cards")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var items = make([]Card, 0)
	for rows.Next() {
		var c Card
		err := rows.Scan(&c.Id, &c.Name, &c.Content, &c.Sort, &c.ColumnId)
		if err != nil {
			return nil, err
		}

		items = append(items, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to get cards: %w", op, err)
	}

	return items, nil
}

func (s *Storage) FindCard(id int64) (*Card, error) {
	const op = "storage.sqlite.FindCard"

	stmt, err := s.db.Prepare(`SELECT * from cards WHERE id = ?`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	var c Card
	err = stmt.QueryRow(id).Scan(&c.Id, &c.Name, &c.Content, &c.Sort, &c.ColumnId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: failed to find card: %w", op, err)
	}

	return &c, nil
}

func (s *Storage) RemoveCard(id int64) error {
	const op = "storage.sqlite.RemoveCard"

	stmt, err := s.db.Prepare(`DELETE FROM cards WHERE id = ?`)
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateCard(card *Card) error {
	const op = "storage.sqlite.UpdateCard"
	stmt, err := s.db.Prepare(`UPDATE cards set name = ?, content = ?, sort = ?, column_id = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(card.Name, card.Content, card.Sort, card.ColumnId, card.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func migrate(db *sql.DB) error {
	const op = "storage.sqlite.migrate"
	queries := []string{
		`DROP TABLE cards`,
		`DROP TABLE columns`,
		`CREATE TABLE columns (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			sort INTEGER DEFAULT 0
		);`,
		`CREATE TABLE cards (
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

	err := addColumns(db)
	if err != nil {
		return err
	}

	fmt.Println("Migration complete")

	return nil
}

func addColumns(db *sql.DB) error {
	const op = "storage.sqlite.addColumns"

	columns := `[
	  { "name": "Todo", "sort": 10 },
	  { "name": "Block", "sort": 20 },
	  { "name": "Development", "sort": 30 },
	  { "name": "Review", "sort": 40 },
	  { "name": "In test", "sort": 50 },
	  { "name": "Done", "sort": 60 }
	]`

	type col struct {
		Name string `json:"name"`
		Sort int    `json:"sort"`
	}

	var cols = make([]col, 6)

	err := json.Unmarshal([]byte(columns), &cols)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, c := range cols {
		stmt, err := db.Prepare("INSERT INTO columns (name, sort) VALUES (?, ?)")
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		_, err = stmt.Exec(c.Name, c.Sort)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
