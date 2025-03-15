package column

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

type ResponseColumnAll struct {
	resp.Response
	Columns []sqlite.Column `json:"columns"`
}

type GetterColumns interface {
	GetColumns() ([]sqlite.Column, error)
}

func GetAll(log *slog.Logger, getter GetterColumns) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handler.column.GetAll"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		cards, err := getter.GetColumns()
		if errors.Is(err, sql.ErrNoRows) {
			log.Error("cards not found", sl.Err(err))
			render.JSON(writer, request, resp.Error("cards not found"))

			return
		}

		render.JSON(writer, request, ResponseColumnAll{
			Response: resp.Ok(),
			Columns:  cards,
		})
	}
}
