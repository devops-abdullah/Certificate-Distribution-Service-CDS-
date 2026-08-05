// Package dashboard serves a small same-origin HTML page that exercises
// every route this service exposes and reports pass/fail against each
// endpoint's expected status code. It's a development convenience for
// manually checking the API without hand-writing curl commands, not part
// of the public API contract.
package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed dashboard.html
var page []byte

// Handler serves the dashboard page.
func Handler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", page)
}
