package main

import (
	"fmt"
	"github.com/SOHIL-03/2110993839/config"
	"github.com/SOHIL-03/2110993839/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

var router *gin.Engine

type App struct {
	logger *zap.Logger
}

func main() {
	config.Init()
	appLogger := logger.Init()
	app := App{logger: appLogger}

	app.startServer()
}

func (app App) startServer() {
	environment := config.GetString("environment")
	port := config.GetString("server.port")
	readTimeout := config.GetInt("server.readTimeout")
	writeTimeout := config.GetInt("server.writeTimeout")

	if environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router = gin.New()
	router.UseH2C = true
	router.RemoveExtraSlash = true
	err := router.SetTrustedProxies([]string{})
	if err != nil {
		app.logger.Fatal("Error in router.SetTrustedProxies", zap.Error(err))
		return
	}

	router.Use(ginLogger(app.logger), ginRecovery(app.logger))
	app.registerRoutes(router)
	router.HandleMethodNotAllowed = true

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": "method not allowed"})
	})
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not found"})
	})

	h2s := &http2.Server{}
	s := &http.Server{
		Addr:           port,
		Handler:        h2c.NewHandler(router, h2s),
		ReadTimeout:    time.Duration(readTimeout) * time.Second,
		WriteTimeout:   time.Duration(writeTimeout) * time.Second,
		MaxHeaderBytes: 1 << 10,
	}

	app.logger.Info(fmt.Sprintf("Starting server on port %s", port))
	err = s.ListenAndServe()
	if err != nil {
		app.logger.Fatal("failed to start server", zap.Error(err))
	}
}

func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		logger.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

func ginRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, true)
				if brokenPipe {
					logger.Sugar().Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error))
					c.Abort()
					return
				}
				logger.Sugar().Error(err)
				logger.Sugar().Error(string(debug.Stack()))
				logger.Sugar().Error("[raw http request] ", string(httpRequest))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			}
		}()
		c.Next()
	}
}
