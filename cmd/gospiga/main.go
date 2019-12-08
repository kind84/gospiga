package main

import (
	"github.com/gin-gonic/gin"

	"github.com/kind84/gospiga/api"
	"github.com/kind84/gospiga/usecase"
)

func main() {
	app := usecase.NewApp()
	service := api.NewService(app)
	r := gin.Default()
	r.GET("/ping", service.Ping)
	r.POST("/new-recipe", service.NewRecipe)
	r.Run()
}
