package v1

import "github.com/gin-gonic/gin"

const (
	ServiceName = "Certificate Manager"
	Version     = "0.1.0"
	Build       = "dev"
)

func VersionHandler(c *gin.Context) {

	Success(c, "Version information", gin.H{
		"service": ServiceName,
		"version": Version,
		"build":   Build,
	})

}
