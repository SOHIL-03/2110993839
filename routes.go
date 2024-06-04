package main

import (
	"github.com/SOHIL-03/2110993839/handlers"
	"github.com/gin-gonic/gin"
)

func (app App) registerRoutes(router *gin.Engine) {
	baseRG := router.Group("/service")

	pinHandler := handlers.NewPingHandler(app.logger)
	baseRG.GET("/ping", pinHandler.Ping)

}
