package table

import (
	"net/http"

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
	tables := rg.Group("/tables")
	{
		tables.GET("", h.List)
		tables.POST("", h.Create)
		tables.GET("/:id", h.GetByID)
		tables.PUT("/:id", h.Update)
		tables.PATCH("/:id/status", h.UpdateStatus)
		tables.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) List(c *gin.Context) {
	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	tables, err := h.repo.List(status)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if tables == nil {
		tables = []Table{}
	}

	c.JSON(http.StatusOK, tables)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	table, err := h.repo.GetByID(id)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if table == nil {
		apperror.RespondNotFound(c, "table not found")
		return
	}

	c.JSON(http.StatusOK, table)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateTableRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	table, err := h.repo.Create(req)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, table)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	table, err := h.repo.Update(id, req)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if table == nil {
		apperror.RespondNotFound(c, "table not found")
		return
	}

	c.JSON(http.StatusOK, table)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateTableStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	table, err := h.repo.UpdateStatus(id, req.Status)
	if err != nil {
		apperror.RespondUnprocessableEntity(c, err.Error())
		return
	}

	if table == nil {
		apperror.RespondNotFound(c, "table not found")
		return
	}

	c.JSON(http.StatusOK, table)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.repo.Delete(id)
	if err != nil {
		if err.Error() == "table not found" {
			apperror.RespondNotFound(c, err.Error())
			return
		}
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
