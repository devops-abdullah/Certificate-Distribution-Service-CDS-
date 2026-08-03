package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Certificate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"message": "Certificate endpoint not implemented yet",
	})
}