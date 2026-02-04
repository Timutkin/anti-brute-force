package server

import (
	"github.com/gin-gonic/gin"
	"github.com/timutkin/anti-brute-force/internal/config"
	"github.com/timutkin/anti-brute-force/internal/handler"
)

type Server struct {
	router            *gin.Engine
	blackListHandler  handler.BlackListHandler
	whiteListHandler  handler.WhiteListHandler
	bruteForceHandler handler.BruteForceHandler
	cfg               config.Server
}

func (s *Server) registerHandlers(router *gin.Engine) {
	router.POST("/blacklists", s.blackListHandler.AddCIDR())
	router.DELETE("/blacklists", s.blackListHandler.DeleteCIDR())
	router.GET("/blacklists", s.blackListHandler.GetCIDRs())

	router.POST("/whitelists", s.whiteListHandler.AddCIDR())
	router.DELETE("/whitelists", s.blackListHandler.DeleteCIDR())
	router.GET("/whitelists", s.whiteListHandler.GetCIDRs())

	router.POST("/allows", s.bruteForceHandler.AllowAuthorization())
	router.GET("/buckets", s.bruteForceHandler.GetBuckets())
	router.DELETE("/buckets", s.bruteForceHandler.DeleteBuckets())
}

func NewServer(
	blackListHandler handler.BlackListHandler,
	whiteListHandler handler.WhiteListHandler,
	bruteForceHandler handler.BruteForceHandler,
	cfg config.Server,
) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware())
	s := &Server{
		router:            r,
		blackListHandler:  blackListHandler,
		whiteListHandler:  whiteListHandler,
		bruteForceHandler: bruteForceHandler,
		cfg:               cfg,
	}
	s.registerHandlers(r)
	return s
}

func (s *Server) Start() error {
	return s.router.Run(s.cfg.ListenAddress)
}
