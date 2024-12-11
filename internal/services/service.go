package services

import (
	"github.com/Kapeland/task-Astral/internal/utils/config"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

type Service struct {
	fm FileModelManager
	am AuthModelManager
	um UsersModelManager
}

func NewService(fm FileModelManager, am AuthModelManager, um UsersModelManager) Service {
	return Service{fm: fm, am: am, um: um}
}

func (s Service) Launch(cfg *config.Config, lgr *logger.Logger) {
	implAuth := authServer{s.am}
	implFile := fileServer{s.fm, s.am}

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
		usrGr.POST("/register", CachePurge(memoryStore, lgr), implAuth.Register)
	}
	authGR := router.Group("/api")
	{
		authGR.POST("/auth", CachePurge(memoryStore, lgr), implAuth.Auth)
		authGR.DELETE("/auth/:token", CachePurge(memoryStore, lgr), implAuth.Logout)
	}

	docsGr := router.Group("/api")
	{
		docsGr.POST("/docs", CachePurge(memoryStore, lgr), implFile.UploadDoc)
		docsGr.GET("/docs", cache.CacheByRequestURI(memoryStore, 2*time.Minute), implFile.GetDocsList)
		docsGr.HEAD("/docs", cache.CacheByRequestURI(memoryStore, 2*time.Minute), implFile.GetDocsList)
		docsGr.GET("/docs/:id", cache.CacheByRequestURI(memoryStore, 2*time.Minute), implFile.GetDoc)
		docsGr.HEAD("/docs/:id", cache.CacheByRequestURI(memoryStore, 2*time.Minute), implFile.GetDoc)
		docsGr.DELETE("/docs/:id", CachePurge(memoryStore, lgr), implFile.DeleteDoc)
	}

	if err := router.Run(":" + strconv.Itoa(cfg.Rest.Port)); err != nil {
		lgr.Error(err.Error(), "Service", "Launch", "router.Run")
	}
}

func CachePurge(store *persist.MemoryStore, lgr *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := store.Cache.Purge()
		if err != nil {
			lgr.Error(err.Error(), "Service", "CachePurge", "Purge")
		}
	}
}
