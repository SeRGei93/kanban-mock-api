package card

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	resp "kanban-app/internal/lib/api/response"
	"kanban-app/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"strconv"
)

type RemoveCardResponse struct {
	resp.Response
}

type RemoveCard interface {
	RemoveCard(id int64) error
}

func Remove(log *slog.Logger, remover RemoveCard) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handler.card.Remove"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		idParam := chi.URLParam(request, "id")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			http.Error(writer, "Invalid ID format, must be int64", http.StatusBadRequest)
			return
		}

		err = remover.RemoveCard(id)
		if err != nil {
			log.Error("failed to remove card", sl.Err(err))
			render.JSON(writer, request, resp.Error("Failed to remove card"))
			return
		}

		log.Info("card remove", slog.Int64("id", id))

		render.JSON(writer, request, RemoveCardResponse{
			Response: resp.Ok(),
		})
	}
}
