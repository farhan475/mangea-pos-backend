package category

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
	rg.GET("/categories", h.List)
	rg.POST("/categories", h.Create)
}

func (h *Handler) List(c *gin.Context) {
	categories, err := h.repo.List()
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if categories == nil {
		categories = []Category{}
	}

	c.JSON(http.StatusOK, categories)
}

// Create adds a new category
func (h *Handler) Create(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	category, err := h.repo.Create(req)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, category)
}
