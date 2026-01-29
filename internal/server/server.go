package server

import "github.com/gin-gonic/gin"

type Server struct {
	router *gin.Engine
}

func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware())
	registerHandlers(r)
	return &Server{router: r}
}

func (s Server) Start() error {
	return s.router.Run()
}

func registerHandlers(router *gin.Engine) {
	router.GET("/hello", func(context *gin.Context) {
		context.String(200, "hello")
	})
}
