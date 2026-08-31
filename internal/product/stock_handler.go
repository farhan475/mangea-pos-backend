package product

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"mangea-backend/internal/apperror"
)

// Stock management endpoints

// GetLowStockProducts returns products with low stock
func (h *Handler) GetLowStockProducts(c *gin.Context) {
	products, err := h.repo.GetLowStockProducts()
	if err != nil {
		apperror.RespondInternalServerError(c, "Failed to fetch low stock products")
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetOutOfStockProducts returns products that are out of stock
func (h *Handler) GetOutOfStockProducts(c *gin.Context) {
	products, err := h.repo.GetOutOfStockProducts()
	if err != nil {
		apperror.RespondInternalServerError(c, "Failed to fetch out of stock products")
		return
	}

	c.JSON(http.StatusOK, products)
}

// UpdateStock updates the stock quantity for a product
func (h *Handler) UpdateStock(c *gin.Context) {
	id := c.Param("id")

	// Pointer allows explicitly setting stock to 0 (binding:"required" rejects zero values)
	var req struct {
		Stock *int `json:"stock" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	// Check if product exists
	product, err := h.repo.GetByID(id)
	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}
	if product == nil {
		apperror.RespondNotFound(c, "Product not found")
		return
	}

	// Update stock
	if err := h.repo.UpdateStock(id, *req.Stock); err != nil {
		apperror.RespondInternalServerError(c, "Failed to update stock")
		return
	}

	// Fetch updated product
	updatedProduct, _ := h.repo.GetByID(id)
	c.JSON(http.StatusOK, updatedProduct)
}

// AddStock adds stock quantity to a product
func (h *Handler) AddStock(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	// Check if product exists
	product, err := h.repo.GetByID(id)
	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}
	if product == nil {
		apperror.RespondNotFound(c, "Product not found")
		return
	}

	// Add stock atomically (stock = stock + quantity)
	if _, err := h.repo.AddStockAtomically(id, req.Quantity); err != nil {
		if strings.Contains(err.Error(), "product not found") {
			apperror.RespondNotFound(c, "Product not found")
			return
		}
		apperror.RespondInternalServerError(c, "Failed to add stock")
		return
	}

	// Fetch updated product
	updatedProduct, _ := h.repo.GetByID(id)
	c.JSON(http.StatusOK, updatedProduct)
}

// ReduceStock reduces stock quantity from a product
func (h *Handler) ReduceStock(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	// Check if product exists
	product, err := h.repo.GetByID(id)
	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}
	if product == nil {
		apperror.RespondNotFound(c, "Product not found")
		return
	}

	// Reduce stock atomically (fails if insufficient)
	if _, err := h.repo.ReduceStockAtomically(id, req.Quantity); err != nil {
		if strings.Contains(err.Error(), "insufficient stock") {
			apperror.RespondBadRequest(c, "Insufficient stock")
			return
		}
		apperror.RespondInternalServerError(c, "Failed to reduce stock")
		return
	}

	// Fetch updated product
	updatedProduct, _ := h.repo.GetByID(id)
	c.JSON(http.StatusOK, updatedProduct)
}

// GetStockStatistics returns stock statistics
func (h *Handler) GetStockStatistics(c *gin.Context) {
	stats, err := h.repo.GetStockStatistics()
	if err != nil {
		apperror.RespondInternalServerError(c, "Failed to fetch stock statistics")
		return
	}

	c.JSON(http.StatusOK, stats)
}
