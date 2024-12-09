package services

import (
	"fmt"
	"github.com/Kapeland/task-Astral/internal/utils/configer"
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

func (s Service) Launch() {
	implAuth := authServer{s.am}
	implFile := fileServer{s.fm, s.am}
	cfg, err := configer.GetConfig()
	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("Service: Launch: configer.GetConfig error: %s", err.Error()))
		return
	}
	memoryStore := persist.NewMemoryStore(2 * time.Minute)
	//gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.HandleMethodNotAllowed = true // Обрабатывает 405 код
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	usrGr := router.Group("/api")
	{
		usrGr.POST("/register", CachePurge(memoryStore), implAuth.Register)
	}
	authGR := router.Group("/api")
	{
		authGR.POST("/auth", CachePurge(memoryStore), implAuth.Auth)
		authGR.DELETE("/auth/:token", CachePurge(memoryStore), implAuth.Logout)
	}

	docsGr := router.Group("/api")
	{
		docsGr.POST("/docs", CachePurge(memoryStore), implFile.UploadDoc)
		docsGr.GET("/docs", cache.CacheByRequestURI(memoryStore, 2*time.Minute), implFile.GetDocsList)
		docsGr.HEAD("/docs", cache.CacheByRequestURI(memoryStore, 2*time.Minute), implFile.GetDocsList)
		docsGr.GET("/docs/:id", cache.CacheByRequestURI(memoryStore, 2*time.Minute), implFile.GetDoc)
		docsGr.HEAD("/docs/:id", cache.CacheByRequestURI(memoryStore, 2*time.Minute), implFile.GetDoc)
		docsGr.DELETE("/docs/:id", CachePurge(memoryStore), implFile.DeleteDoc)
	}

	if err = router.Run(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("Service: Launch: router.Run error: %s", err.Error()))
	}
}

func CachePurge(store *persist.MemoryStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := store.Cache.Purge()
		if err != nil {
			logger.Log(logger.ErrPrefix, fmt.Sprintf("CachePurge: error:%s", err.Error()))
		}
	}
}
