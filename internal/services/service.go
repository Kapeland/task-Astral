package services

import (
	mw "github.com/Kapeland/task-Astral/internal/services/middleware"
	"github.com/Kapeland/task-Astral/internal/services/servers"
	"github.com/Kapeland/task-Astral/internal/utils/config"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

type Service struct {
	fm servers.FileModelManager
	am servers.AuthModelManager
	um UsersModelManager
}

func NewService(fm servers.FileModelManager, am servers.AuthModelManager, um UsersModelManager) Service {
	return Service{fm: fm, am: am, um: um}
}

func (s Service) Launch(cfg *config.Config, lgr *logger.Logger) {
	implAuth := servers.AuthServer{A: s.am}
	implFile := servers.FileServer{F: s.fm, A: s.am}

	memoryStore := persist.NewMemoryStore(2 * time.Minute)
	if !cfg.Project.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.HandleMethodNotAllowed = true // Обрабатывает 405 код
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	usrGr := router.Group("/api")
	{
		usrGr.POST("/register", mw.CachePurge(memoryStore, lgr), implAuth.Register)
	}
	authGR := router.Group("/api")
	{
		authGR.POST("/auth", mw.CachePurge(memoryStore, lgr), implAuth.Auth)
		authGR.DELETE("/auth/:token", mw.CachePurge(memoryStore, lgr), implAuth.Logout)
	}

	docsGr := router.Group("/api")
	{
		docsGr.POST("/docs", mw.ValidateTokenInMultipartFrom(s.am, lgr), mw.CachePurge(memoryStore, lgr), implFile.UploadDoc)
		docsGr.GET("/docs", cache.CacheByRequestURI(memoryStore, 2*time.Minute), mw.ValidateTokenInQuery(s.am, lgr), implFile.GetDocsList)
		docsGr.HEAD("/docs", cache.CacheByRequestURI(memoryStore, 2*time.Minute), mw.ValidateTokenInQuery(s.am, lgr), implFile.GetDocsList)
		docsGr.GET("/docs/:id", cache.CacheByRequestURI(memoryStore, 2*time.Minute), mw.ValidateTokenInQuery(s.am, lgr), implFile.GetDoc)
		docsGr.HEAD("/docs/:id", cache.CacheByRequestURI(memoryStore, 2*time.Minute), mw.ValidateTokenInQuery(s.am, lgr), implFile.GetDoc)
		docsGr.DELETE("/docs/:id", mw.ValidateTokenInQuery(s.am, lgr), mw.CachePurge(memoryStore, lgr), implFile.DeleteDoc)
	}

	if err := router.Run(":" + strconv.Itoa(cfg.Rest.Port)); err != nil {
		lgr.Error(err.Error(), "Service", "Launch", "router.Run")
	}
}
