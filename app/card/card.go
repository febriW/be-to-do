package card

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/febriW/be-to-do/repository"
	"github.com/febriW/be-to-do/server"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrNotAuthorized = errors.New("not authorized")
	ErrCantUpdate    = errors.New("can't update data that's already marked")
)

type Card struct {
	ActivitiesNo string `json:"activities_no"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	AuthorId     int    `json:"author_id"`
	MarkedStatus string `json:"marked_status"`
	Marked       string `json:"marked"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	DeletedAt    string `json:"deleted_at"`
}

type CardParamCreate struct {
	AuthorID     int    `json:"author_id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	MarkedStatus string `json:"marked_status"`
	Marked       string `json:"marked"`
}

type CardParamUpdate struct {
	AuthorID     int    `json:"author_id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	MarkedStatus string `json:"marked_status"`
	Marked       string `json:"marked"`
	ActivitiesNo string `json:"activities_no"`
}

type CardsParam struct {
	AuthorID int
	PaginationParam
}

type PaginationParam struct {
	Page int
	Size int
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

//func (s *Service) DeleteCard(ctx context.Context, AuthorID string, ActivitiesNO string) error {
//	err := s.execTx(ctx, func(r *repository.Repository) error {
//
//	})
//}

func (s *Service) UpdateCard(ctx context.Context, params CardParamUpdate) error {
	err := s.execTx(ctx, func(r *repository.Repository) error {
		if params.AuthorID <= 0 {
			return fmt.Errorf("card with author id %s %w", strconv.Itoa(params.AuthorID), ErrNotFound)
		}

		if params.ActivitiesNo == "" {
			return fmt.Errorf("Card activities no %s %w", params.ActivitiesNo, ErrNotFound)
		}

		c := r.CheckCard(ctx, params.ActivitiesNo, params.AuthorID)
		if c == nil {
			return fmt.Errorf("card %s from author id %v %w", params.ActivitiesNo, params.AuthorID, ErrNotFound)
		}

		if c.Marked != nil {
			return fmt.Errorf("Card number %s %w", params.ActivitiesNo, ErrCantUpdate)
		}

		var markedTime *time.Time
		if params.Marked != "" {
			parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", params.Marked)
			if parseErr != nil {
				return fmt.Errorf("invalid date format for Marked: %w", parseErr)
			}
			markedTime = &parsedTime
		} else {
			markedTime = nil
		}

		var markedStatus *string
		if params.MarkedStatus != "" {
			markedStatus = &params.MarkedStatus
		} else {
			markedStatus = nil
		}

		return r.UpdateCard(ctx, repository.Card{
			ActivitiesNo: params.ActivitiesNo,
			AuthorID:     params.AuthorID,
			Title:        params.Title,
			Content:      params.Content,
			Marked:       markedTime,
			MarkedStatus: markedStatus,
		})
	})

	return err
}

func (s *Service) HandleUpdateCard() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var params CardParamUpdate

		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			server.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		err = s.UpdateCard(r.Context(), params)
		if err != nil {
			server.ErrorResponse(w, http.StatusUnprocessableEntity, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Service) CreateCard(ctx context.Context, params CardParamCreate) error {
	err := s.execTx(ctx, func(r *repository.Repository) error {
		if params.AuthorID <= 0 {
			return fmt.Errorf("author id %s %w", strconv.Itoa(params.AuthorID), ErrNotFound)
		}
		var markedTime *time.Time
		if params.Marked != "" {
			parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", params.Marked)
			if parseErr != nil {
				return fmt.Errorf("invalid date format for Marked: %w", parseErr)
			}
			markedTime = &parsedTime
		} else {
			markedTime = nil
		}

		return r.CreateCard(ctx, repository.Card{
			AuthorID: params.AuthorID,
			Title:    params.Title,
			Content:  params.Content,
			Marked:   markedTime,
		})
	})

	return err
}

func (s *Service) HandleCreateCard() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var input CardParamCreate
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			server.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		err = s.CreateCard(r.Context(), input)
		if err != nil {
			server.ErrorResponse(w, http.StatusUnprocessableEntity, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Service) GetAllCards(ctx context.Context, param CardsParam) ([]Card, int) {
	repo := repository.New(s.db)
	repoParam := repository.CardsParam{
		AuthorID: param.AuthorID,
	}
	repoParam.Page = param.Page
	repoParam.Size = param.Size
	cs, total := repo.GetCards(ctx, repoParam)
	res := make([]Card, 0, len(cs))
	for _, c := range cs {
		res = append(res, mapCardRepoToService(c))
	}

	return res, total
}

func (s *Service) HandleGetAllCards() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		urlParams := r.URL.Query()
		var errs []error

		authorIDStr := urlParams.Get("author_id")
		authorID := 0
		if authorIDStr != "" {
			var err error
			authorID, err = strconv.Atoi(authorIDStr)
			if err != nil {
				errs = append(errs, err)
			}
		}

		pageStr := urlParams.Get("page")
		page := 1
		if pageStr != "" {
			var err error
			page, err = strconv.Atoi(pageStr)
			if err != nil {
				errs = append(errs, err)
			}
		}

		sizeStr := urlParams.Get("size")
		size := 10
		if sizeStr != "" {
			var err error
			size, err = strconv.Atoi(sizeStr)
			if err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			server.ErrorResponse(w, http.StatusBadRequest, errors.Join(errs...))
			return
		}

		params := CardsParam{
			AuthorID: authorID,
			PaginationParam: PaginationParam{
				Page: page,
				Size: size,
			},
		}

		cs, total := s.GetAllCards(r.Context(), params)
		output := struct {
			Next  string
			Prev  string
			Total int
			Data  []Card
		}{
			Total: total,
			Data:  cs,
		}
		server.JSONResponse(w, http.StatusOK, output)
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

func mapCardRepoToService(data repository.Card) Card {
	var marked string
	if data.Marked != nil {
		marked = data.Marked.Format("2006-01-02 15:04:05")
	} else {
		marked = ""
	}

	var deletedAt string
	if data.DeletedAt != nil {
		deletedAt = data.DeletedAt.Format("2006-01-02 15:04:05")
	} else {
		deletedAt = "" // or handle as needed
	}

	return Card{
		ActivitiesNo: data.ActivitiesNo,
		Title:        data.Title,
		Content:      data.Content,
		CreatedAt:    data.CreatedAt.Format(time.DateTime),
		UpdatedAt:    data.UpdatedAt.Format(time.DateTime),
		DeletedAt:    deletedAt,
		Marked:       marked,
	}
}
