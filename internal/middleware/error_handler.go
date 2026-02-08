package middleware

import (
	"backend-hotlines3/internal/dto"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RecoveryMiddleware catches panics and returns a standard JSON error response
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.JSON(http.StatusInternalServerError, dto.StandardResponse{
					Success: false,
					Error: &dto.ErrorInfo{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "Something went wrong on the server",
						Details: r,
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// HandleValidationError formats validation errors and sends a response
// This helper should be called in handlers when c.ShouldBindJSON returns an error
func HandleValidationError(c *gin.Context, err error) {
	var validationErrors []map[string]string
	var messages []string

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, e := range ve {
			fieldName := e.Field()
			// Simple formatting: lowercase first letter
			if len(fieldName) > 0 {
				fieldName = strings.ToLower(fieldName[:1]) + fieldName[1:]
			}

			var errMsg string
			switch e.Tag() {
			case "required":
				errMsg = fieldName + " is required"
			case "numeric":
				errMsg = fieldName + " must be numeric"
			case "len":
				errMsg = fieldName + " must be " + e.Param() + " characters long"
			case "min":
				errMsg = fieldName + " must be at least " + e.Param() + " characters"
			case "max":
				errMsg = fieldName + " must be at most " + e.Param() + " characters"
			case "oneof":
				errMsg = fieldName + " must be one of: " + e.Param()
			case "email":
				errMsg = fieldName + " must be a valid email address"
			default:
				errMsg = fieldName + " is invalid"
			}

			validationErrors = append(validationErrors, map[string]string{
				"field":   fieldName,
				"message": errMsg,
			})
			messages = append(messages, errMsg)
		}
	} else {
		messages = []string{"Invalid request data format"}
	}

	c.JSON(http.StatusBadRequest, dto.StandardResponse{
		Success: false,
		Error: &dto.ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid input data",
			Details: validationErrors,
		},
	})
	c.Abort()
}
