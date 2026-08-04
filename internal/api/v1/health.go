package v1

import "github.com/gin-gonic/gin"

func Health(c *gin.Context) {

	Success(c, "Service is healthy", gin.H{
		"service": "Certificate Manager",
		"status":  "UP",
	})

}
