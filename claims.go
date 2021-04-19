package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)


func customClaims(c *gin.Context) {

	if c.MustGet("isAdmin").(bool) {
		c.JSON(http.StatusOK, nil)
	} else {
		c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
			Code:   http.StatusForbidden,
			Errors: []string{"you are not an admin"},
		})
	}
}
