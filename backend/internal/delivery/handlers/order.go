package handlers

import (
	resp "WB/internal/lib/api/response"
	"WB/internal/models"
	usecase "WB/internal/usecase"
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func NewOrder(log *slog.Logger, orderUseCase *usecase.OrderUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order.NewOrder"

		ctx := r.Context()

		var order models.Order

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if err := render.DecodeJSON(r.Body, &order); err != nil {
			log.Error("failed to unmarshal order", "op", op, "error", err)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		if err := orderUseCase.CreateOrder(ctx, order); err != nil {
			log.Error("failed to unmarshal order", "op", op, "error", err)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		log.Info("order creating success")
		render.JSON(w, r, resp.OK())
	}
}

func GetOrder(log *slog.Logger, orderUseCase *usecase.OrderUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order.GetOrder"

		var order models.Order

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		//error track
		OrderID := chi.URLParam(r, "id")
		if OrderID == "" {
			http.Error(w, "id parameter missing", http.StatusBadRequest)
			return
		}

		order, err := orderUseCase.GetOrder(context.Background(), OrderID)
		if err != nil {
			log.Error("failed to unmarshal order", "op", op, "error", err)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		log.Info("order getting success")
		render.JSON(w, r, order)
	}
}
