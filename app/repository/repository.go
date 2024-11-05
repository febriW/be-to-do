package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/georgysavva/scany/v2/dbscan"
	"log/slog"
	"time"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type Repository struct {
	db        DB
	ForUpdate bool
}

type User struct {
	ID           int    `db:"id"`
	Name         string `db:"name"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func New(db DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CheckUser(ctx context.Context, email string) *User {
	query := r.SelectQuery(`SELECT * FROM user WHERE email = ? LIMIT 1`)
	rows, err := r.db.QueryContext(ctx, query, email)
	fmt.Println(err)
	if err != nil {
		slog.Error("failed to query user", "email", email, "err", err)
		return nil
	}

	var res User
	err = dbscan.ScanOne(&res, rows)
	if err != nil {
		slog.Error("failed to scan user", "email", email, "err", err)
		return nil
	}

	return &res
}

func (r *Repository) SelectQuery(query string) string {
	if r.ForUpdate {
		query += " FOR UPDATE"
	}

	return query
}

func (r *Repository) CreateUser(ctx context.Context, data User) error {
	query := `INSERT INTO user (name, email, password_hash) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, data.Name, data.Email, data.PasswordHash)
	if err != nil {
		return err
	}
	return err
}
