package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/georgysavva/scany/v2/dbscan"
	"log/slog"
	"strings"
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

type Card struct {
	ActivitiesNo string `json:"activities_no"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	AuthorId     int    `json:"author_id"`
	Marked       time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    time.Time
}

type CardsParam struct {
	AuthorID int
	PaginationParams
}

type PaginationParams struct {
	Page int
	Size int
}

func New(db DB) *Repository {
	return &Repository{db: db}
}

// repository user
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

func (r *Repository) CreateUser(ctx context.Context, data User) error {
	query := `INSERT INTO user (name, email, password_hash) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, data.Name, data.Email, data.PasswordHash)
	if err != nil {
		return err
	}
	return err
}


activities_no VARCHAR(10) NOT NULL PRIMARY KEY,
author_id INT NOT NULL,
title VARCHAR(100) NOT NULL,
content TEXT  NOT NULL,
marked TIMESTAMP NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
deleted_at TIMESTAMP NULL


// card repository
func (r *Repository) CreateCard(ctx context.Context, data Card) error {
	selectQuery := r.SelectQuery("SELECT activities_no FROM card")
	latestActivities := r.Count(ctx,selectQuery)
	activitiesNo := fmt.Sprintf("AC-%04d", latestActivities)

	query := `INSET INTO card (activities_no, author_id, title, content) VALUES (?,?,?,?)`
	_, err := r.db.ExecContext(ctx, query, activitiesNo, data.AuthorId, data.Title, data.Content)
	if err != nil { return err}
	return err
}

func (r *Repository) GetCards(ctx context.Context, param CardsParam) ([]Card, int) {
	if param.Page <= 0 {
		param.Page = 1
	}

	if param.Size <= 0 {
		param.Size = 10
	}

	query := r.SelectQuery("SELECT * FROM card")
	var args []any

	if param.AuthorID > 0 {
		query += " WHERE author_id = ?"
		args = append(args, param.AuthorID)
	}

	total := r.Count(ctx, query, args...)
	query = r.paginationQuery(query, param.PaginationParams)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0
	}

	var res []Card
	dbscan.ScanAll(&res, rows)
	return res, total
}

// common func
func (r *Repository) SelectQuery(query string) string {
	if r.ForUpdate {
		query += " FOR UPDATE"
	}

	return query
}

func (r *Repository) paginationQuery(query string, param PaginationParams) string {
	limit := param.Size
	offset := (param.Page - 1) * param.Size

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}
	return query
}

func (r *Repository) Count(ctx context.Context, query string, args ...any) int {
	query = strings.Replace(query, " * ", " COUNT(*) ", 1)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return 0
	}

	var total int
	dbscan.ScanOne(&total, rows)
	return total
}
