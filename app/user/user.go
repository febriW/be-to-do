package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/febriW/be-to-do/repository"
	"github.com/febriW/be-to-do/server"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

var (
	ErrAlreadyRegistered = errors.New("account already registered")
	ErrInvalidLogin      = errors.New("invalid login")
	ErrNotFound          = errors.New("not found")
)

type User struct {
	Name     string
	Email    string
	Password string
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Register(ctx context.Context, name, email, password string) error {
	err := s.execTx(ctx, func(r *repository.Repository) error {
		u := r.CheckUser(ctx, email)
		if u != nil {
			return fmt.Errorf("email %s already exists", email, ErrAlreadyRegistered)
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("error when registered: %w", err)
		}

		return r.CreateUser(ctx, repository.User{
			Name:         name,
			Email:        email,
			PasswordHash: string(passwordHash),
		})
	})

	return err
}

func (s *Service) HandleRegister() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			server.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		err = s.Register(r.Context(), input.Name, input.Email, input.Password)
		if err != nil {
			server.ErrorResponse(w, http.StatusUnprocessableEntity, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Service) execTx(ctx context.Context, fn func(*repository.Repository) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	repo := repository.New(tx)
	err = fn(repo)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
