package activity_log

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
	return &Handler{repo: NewRepository(db)}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	logs := rg.Group("/activity-logs")
	{
		logs.GET("", h.List)
		logs.POST("", h.Create)
		logs.POST("/batch", h.BatchCreate)
		logs.DELETE("/cleanup", h.Cleanup)
	}
}

func (h *Handler) List(c *gin.Context) {
	var logType *string
	if t := c.Query("type"); t != "" {
		logType = &t
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	logs, err := h.repo.List(logType, limit)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	if logs == nil {
		logs = []ActivityLog{}
	}

	c.JSON(http.StatusOK, logs)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateActivityLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	log, err := h.repo.Create(req)
	if err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, log)
}

func (h *Handler) BatchCreate(c *gin.Context) {
	var reqs []CreateActivityLogRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	if len(reqs) == 0 {
		apperror.RespondBadRequest(c, "batch cannot be empty")
		return
	}

	var created []ActivityLog
	for _, req := range reqs {
		log, err := h.repo.Create(req)
		if err != nil {
			continue
		}
		created = append(created, *log)
	}

	if created == nil {
		created = []ActivityLog{}
	}

	c.JSON(http.StatusOK, created)
}

func (h *Handler) Cleanup(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	if err := h.repo.DeleteOlderThan(days); err != nil {
		apperror.RespondInternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cleanup completed"})
}
