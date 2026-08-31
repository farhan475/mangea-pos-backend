package report

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"mangea-backend/internal/apperror"
)

type Handler struct {
	repo *Repository
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{repo: NewRepository(db)}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/reports")
	{
		reports.GET("/daily", h.GetDailyReport)
		reports.GET("/weekly", h.GetWeeklyReport)
		reports.GET("/top-products", h.GetTopProducts)
	}
}

// GetDailyReport returns daily report for a given date (default: today)
func (h *Handler) GetDailyReport(c *gin.Context) {
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		apperror.RespondBadRequest(c, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	report, err := h.repo.GetDailyReport(date)
	if err != nil {
		apperror.RespondInternalServerError(c, "Failed to generate daily report")
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetWeeklyReport returns weekly report for the week containing the given date
func (h *Handler) GetWeeklyReport(c *gin.Context) {
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		apperror.RespondBadRequest(c, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	report, err := h.repo.GetWeeklyReport(date)
	if err != nil {
		apperror.RespondInternalServerError(c, "Failed to generate weekly report")
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetTopProducts returns top selling products for a date range
func (h *Handler) GetTopProducts(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	// Validate date formats
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		apperror.RespondBadRequest(c, "Invalid start_date format. Use YYYY-MM-DD")
		return
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		apperror.RespondBadRequest(c, "Invalid end_date format. Use YYYY-MM-DD")
		return
	}

	products, err := h.repo.GetTopProducts(startDate, endDate, 10)
	if err != nil {
		apperror.RespondInternalServerError(c, "Failed to get top products")
		return
	}

	c.JSON(http.StatusOK, products)
}
