package card

import (
	"database/sql"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	resp "kanban-app/internal/lib/api/response"
	"kanban-app/internal/lib/logger/sl"
	"kanban-app/internal/storage/sqlite"
	"log/slog"
	"net/http"
)

type ResponseCardAll struct {
	resp.Response
	Cards []sqlite.Card `json:"cards"`
}

type GetterCard interface {
	GetCards() ([]sqlite.Card, error)
}

func GetAll(log *slog.Logger, getter GetterCard) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handler.card.GetAll"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		cards, err := getter.GetCards()
		if errors.Is(err, sql.ErrNoRows) {
			log.Error("cards not found", sl.Err(err))
			render.JSON(writer, request, resp.Error("cards not found"))

			return
		}

		render.JSON(writer, request, ResponseCardAll{
			Response: resp.Ok(),
			Cards:    cards,
		})
	}
}
