package card

import (
	"context"
	"database/sql"
	"errors"
	"github.com/febriW/be-to-do/repository"
	"github.com/febriW/be-to-do/server"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrNotAuthorized = errors.New("not authorized")
)

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

func (s *Service) GetAllCards(ctx context.Context, param CardsParam) ([]Card, int) {
	repo := repository.New(s.db)
	repoParam := repository.CardsParam{
		AuthorID: param.AuthorID,
	}
	repoParam.Page = param.Page
	repoParam.Size = param.Size
	cs, total := repo.GetCards(ctx, repoParam)

	res := make([]Card, len(cs))
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

func mapCardRepoToService(data repository.Card) Card {
	return Card{
		ActivitiesNo: data.ActivitiesNo,
		Title:        data.Title,
		Content:      data.Content,
		CreatedAt:    data.CreatedAt,
		UpdatedAt:    data.UpdatedAt,
		DeletedAt:    data.DeletedAt,
		Marked:       data.Marked,
	}
}
