package dashboard

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"mangea-backend/internal/apperror"
)

type Handler struct {
	repo *Repository
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{
		repo: NewRepository(db),
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	dashboard := rg.Group("/dashboard")
	{
		dashboard.GET("/metrics", h.GetMetrics)
		dashboard.GET("/popular-dishes", h.GetPopularDishes)
		dashboard.GET("/out-of-stock", h.GetOutOfStock)
	}
}

func (h *Handler) GetMetrics(c *gin.Context) {
	metrics, err := h.repo.GetMetrics()
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) GetPopularDishes(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	dishes, err := h.repo.GetPopularDishes(limit)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if dishes == nil {
		dishes = []PopularDish{}
	}

	c.JSON(http.StatusOK, dishes)
}

func (h *Handler) GetOutOfStock(c *gin.Context) {
	alerts, err := h.repo.GetOutOfStock()
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if alerts == nil {
		alerts = []StockAlert{}
	}

	c.JSON(http.StatusOK, alerts)
}
