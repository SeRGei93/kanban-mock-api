package card

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	resp "kanban-app/internal/lib/api/response"
	"kanban-app/internal/lib/logger/sl"
	"kanban-app/internal/storage/sqlite"
	"log/slog"
	"net/http"
	"strconv"
)

type UpdateCardRequest struct {
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content,omitempty"`
	Sort     int64  `json:"sort" validate:"required"`
	ColumnId int64  `json:"column_id" validate:"required"`
}

type UpdateCardResponse struct {
	resp.Response
}

type UpdaterCard interface {
	UpdateCard(card *sqlite.Card) error
	FindCard(id int64) (*sqlite.Card, error)
}

func Update(log *slog.Logger, updater UpdaterCard) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handler.card.Update"

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

		var req UpdateCardRequest
		err = render.DecodeJSON(request.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(writer, request, resp.Error("failed to decode request"))
			return
		}

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("validation error", sl.Err(err))
			render.JSON(writer, request, resp.ValidationError(validateErr))

			return
		}

		existCard, err := updater.FindCard(id)
		if err != nil {
			log.Error("failed to find card", sl.Err(err))
			render.JSON(writer, request, resp.Error("Failed to update card"))
		}

		existCard.Name = req.Name
		existCard.Content = req.Content
		existCard.Sort = req.Sort
		existCard.ColumnId = req.ColumnId

		err = updater.UpdateCard(existCard)

		if err != nil {
			log.Error("failed to update card", sl.Err(err))
			render.JSON(writer, request, resp.Error("Failed to update card"))
			return
		}

		log.Info("card update", slog.Int64("id", id))

		render.JSON(writer, request, UpdateCardResponse{
			Response: resp.Ok(),
		})
	}
}
