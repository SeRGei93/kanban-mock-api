package card

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	resp "kanban-app/internal/lib/api/response"
	"kanban-app/internal/lib/logger/sl"
	"kanban-app/internal/storage"
	"log/slog"
	"net/http"
)

type AddCardRequest struct {
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content,omitempty"`
	Sort     int64  `json:"sort" validate:"required"`
	ColumnId int64  `json:"columnId" validate:"required"`
}

type ResponseCardAdd struct {
	resp.Response
	Id int64 `json:"id" validate:"required,omitempty"`
}

type SaverCard interface {
	SaveCard(name string, content string, sort int64, columnId int64) (int64, error)
}

func Add(log *slog.Logger, saver SaverCard) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handler.card.Add"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		var req AddCardRequest
		err := render.DecodeJSON(request.Body, &req)
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

		id, err := saver.SaveCard(req.Name, req.Content, req.Sort, req.ColumnId)
		if errors.Is(err, storage.ErrCardExists) {
			log.Info("card already exists", slog.String("name", req.Name))

			render.JSON(writer, request, resp.Error("card already exists"))
			return
		}

		if err != nil {
			log.Error("failed to save card", sl.Err(err))
			render.JSON(writer, request, resp.Error("Failed to add card"))
			return
		}

		log.Info("card saved", slog.Int64("id", id))

		render.JSON(writer, request, ResponseCardAdd{
			Response: resp.Ok(),
			Id:       id,
		})
	}
}
