package storage

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrColumnExists = errors.New("column exists")
	ErrCardExists   = errors.New("column exists")
)
