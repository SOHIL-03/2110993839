package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type PingHandler struct {
	logger *zap.Logger
}

func NewPingHandler(logger *zap.Logger) *PingHandler {
	return &PingHandler{
		logger: logger,
	}
}

func (h PingHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": "PONG",
	})

}
