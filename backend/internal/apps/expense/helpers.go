package expense

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
)

var errUnauthorized = errors.New("unauthorized")

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, errUnauthorized
	}
	return uuid.Parse(userIDStr.(string))
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid id format",
		})
		return uuid.Nil, err
	}
	return id, nil
}

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
