package product

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
	products := rg.Group("/products")
	{
		products.GET("", h.List)
		products.POST("", h.Create)
		products.GET("/:id", h.GetByID)
		products.PUT("/:id", h.Update)
		products.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) List(c *gin.Context) {
	var categoryID *string
	var isAvailable *bool

	if cid := c.Query("category_id"); cid != "" {
		categoryID = &cid
	}

	if ia := c.Query("is_available"); ia != "" {
		val := ia == "true"
		isAvailable = &val
	}

	products, err := h.repo.List(categoryID, isAvailable)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if products == nil {
		products = []Product{}
	}

	c.JSON(http.StatusOK, products)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	product, err := h.repo.GetByID(id)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if product == nil {
		apperror.RespondNotFound(c, "product not found")
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	product, err := h.repo.Create(req)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	product, err := h.repo.Update(id, req)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if product == nil {
		apperror.RespondNotFound(c, "product not found")
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.repo.Delete(id)
	if err != nil {
		if err.Error() == "product not found" {
			apperror.RespondNotFound(c, err.Error())
			return
		}
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
