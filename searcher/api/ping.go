package api

import (
	"github.com/gin-gonic/gin"
)

func (s *GospigaService) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
