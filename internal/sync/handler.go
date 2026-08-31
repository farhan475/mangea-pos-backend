package sync

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"mangea-backend/internal/apperror"
	"mangea-backend/internal/order"
)

type Handler struct {
	orderRepo *order.Repository
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{
		orderRepo: order.NewRepository(db),
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/sync/orders", h.SyncOrders)
}

func (h *Handler) SyncOrders(c *gin.Context) {
	var orders []order.Order

	if err := c.ShouldBindJSON(&orders); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	if len(orders) == 0 {
		apperror.RespondBadRequest(c, "orders array cannot be empty")
		return
	}

	syncedOrders, err := h.orderRepo.UpsertOrders(orders)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if syncedOrders == nil {
		syncedOrders = []order.Order{}
	}

	c.JSON(http.StatusOK, syncedOrders)
}
