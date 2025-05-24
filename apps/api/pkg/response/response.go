package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Meta struct {
	Page     int   `json:"page"`
	Limit    int   `json:"limit"`
	Total    int64 `json:"total"`
	HasMore  bool  `json:"hasMore"`
}

type Error struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func BadRequest(c *gin.Context, message string, details ...map[string]interface{}) {
	resp := Response{
		Success: false,
		Error: &Error{
			Code:    "BAD_REQUEST",
			Message: message,
		},
	}
	
	if len(details) > 0 {
		resp.Error.Details = details[0]
	}
	
	c.JSON(http.StatusBadRequest, resp)
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error: &Error{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error: &Error{
			Code:    "FORBIDDEN",
			Message: message,
		},
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &Error{
			Code:    "NOT_FOUND",
			Message: message,
		},
	})
}

func Conflict(c *gin.Context, message string, details ...map[string]interface{}) {
	resp := Response{
		Success: false,
		Error: &Error{
			Code:    "CONFLICT",
			Message: message,
		},
	}
	
	if len(details) > 0 {
		resp.Error.Details = details[0]
	}
	
	c.JSON(http.StatusConflict, resp)
}

func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &Error{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: message,
		},
	})
}

func ValidationError(c *gin.Context, errors map[string]interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &Error{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: errors,
		},
	})
}

func NewMeta(page, limit int, total int64) *Meta {
	hasMore := int64(page*limit) < total
	return &Meta{
		Page:    page,
		Limit:   limit,
		Total:   total,
		HasMore: hasMore,
	}
}