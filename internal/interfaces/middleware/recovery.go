package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"details": err,
			})
			c.Abort()
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
		}
	})
}
