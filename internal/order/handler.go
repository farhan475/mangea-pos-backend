package order

import (
	"net/http"
	"strings"

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
	orders := rg.Group("/orders")
	{
		orders.GET("", h.List)
		orders.POST("", h.Create)
		orders.GET("/:id", h.GetByID)
		orders.PUT("/:id", h.Update)
		orders.PATCH("/:id/status", h.UpdateStatus)
		orders.POST("/:id/pay", h.Pay)
		orders.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) List(c *gin.Context) {
	var status *string
	var tableNumber *string
	var customerName *string
	var dateFilter *string

	if s := c.Query("status"); s != "" {
		status = &s
	}
	if tn := c.Query("table_number"); tn != "" {
		tableNumber = &tn
	}
	if cn := c.Query("customer_name"); cn != "" {
		customerName = &cn
	}
	if d := c.Query("date"); d != "" {
		dateFilter = &d
	}

	orders, err := h.repo.List(status, tableNumber, customerName, dateFilter)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if orders == nil {
		orders = []Order{}
	}

	c.JSON(http.StatusOK, orders)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	order, err := h.repo.GetByID(id)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if order == nil {
		apperror.RespondNotFound(c, "order not found")
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	order, err := h.repo.Create(req)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	order, err := h.repo.Update(id, req)
	if err != nil {
		apperror.RespondUnprocessableEntity(c, err.Error())
		return
	}

	if order == nil {
		apperror.RespondNotFound(c, "order not found")
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	order, err := h.repo.UpdateStatus(id, req.Status)
	if err != nil {
		if strings.Contains(err.Error(), "cannot transition") || strings.Contains(err.Error(), "concurrently") {
			apperror.RespondUnprocessableEntity(c, err.Error())
			return
		}
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if order == nil {
		apperror.RespondNotFound(c, "order not found")
		return
	}

	c.JSON(http.StatusOK, order)
}

// Pay records a payment for an order and marks it paid.
func (h *Handler) Pay(c *gin.Context) {
	id := c.Param("id")

	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	order, err := h.repo.Pay(id, req)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "already paid") || strings.Contains(msg, "cannot pay") ||
			strings.Contains(msg, "less than total") || strings.Contains(msg, "concurrently") {
			apperror.RespondUnprocessableEntity(c, msg)
			return
		}
		apperror.RespondInternalServerError(c, msg)
		return
	}

	if order == nil {
		apperror.RespondNotFound(c, "order not found")
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.repo.Delete(id)
	if err != nil {
		if err.Error() == "order not found" {
			apperror.RespondNotFound(c, err.Error())
			return
		}
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
