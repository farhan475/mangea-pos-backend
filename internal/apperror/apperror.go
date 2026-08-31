package apperror

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{Error: message})
}

func RespondBadRequest(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, message)
}

func RespondNotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, message)
}

func RespondUnprocessableEntity(c *gin.Context, message string) {
	RespondError(c, http.StatusUnprocessableEntity, message)
}

func RespondInternalServerError(c *gin.Context, message string) {
	RespondError(c, http.StatusInternalServerError, message)
}

func RespondUnauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, message)
}

func RespondConflict(c *gin.Context, message string) {
	RespondError(c, http.StatusConflict, message)
}
