package services

import (
	"context"
	mw "github.com/Kapeland/task-Astral/internal/services/middleware"
	"github.com/Kapeland/task-Astral/internal/services/servers"
	"github.com/Kapeland/task-Astral/internal/utils/config"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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
	ctx := context.Background()

	implAuth := servers.AuthServer{A: s.am}
	implFile := servers.FileServer{F: s.fm, A: s.am}

	redisStore := persist.NewRedisStore(redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		DB:      cfg.Redis.DB,
	}))

	if res, err := redisStore.RedisClient.Ping(ctx).Result(); err != nil {
		lgr.Error(res, "Service", "Launch", "Ping")
		return
	}

	if !cfg.Project.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.HandleMethodNotAllowed = true // Обрабатывает 405 код
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	usrGr := router.Group("/api")
	{
		usrGr.POST("/register", mw.CachePurge(ctx, redisStore, lgr), implAuth.Register)
	}
	authGR := router.Group("/api")
	{
		authGR.POST("/auth", mw.CachePurge(ctx, redisStore, lgr), implAuth.Auth)
		authGR.DELETE("/auth/:token", mw.CachePurge(ctx, redisStore, lgr), implAuth.Logout)
	}

	docsGr := router.Group("/api")
	{
		docsGr.POST("/docs", mw.ValidateTokenInMultipartFrom(s.am, lgr), mw.CachePurge(ctx, redisStore, lgr), implFile.UploadDoc)
		docsGr.GET("/docs", cache.CacheByRequestURI(redisStore, 2*time.Minute), mw.ValidateTokenInQuery(s.am, lgr), implFile.GetDocsList)
		docsGr.HEAD("/docs", cache.CacheByRequestURI(redisStore, 2*time.Minute), mw.ValidateTokenInQuery(s.am, lgr), implFile.GetDocsList)
		docsGr.GET("/docs/:id", cache.CacheByRequestURI(redisStore, 2*time.Minute), mw.ValidateTokenInQuery(s.am, lgr), implFile.GetDoc)
		docsGr.HEAD("/docs/:id", cache.CacheByRequestURI(redisStore, 2*time.Minute), mw.ValidateTokenInQuery(s.am, lgr), implFile.GetDoc)
		docsGr.DELETE("/docs/:id", mw.ValidateTokenInQuery(s.am, lgr), mw.CachePurge(ctx, redisStore, lgr), implFile.DeleteDoc)
	}

	if err := router.Run(":" + strconv.Itoa(cfg.Rest.Port)); err != nil {
		lgr.Error(err.Error(), "Service", "Launch", "router.Run")
	}
}
