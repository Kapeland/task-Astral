package middleware

import (
	"context"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
)

func CachePurge(ctx context.Context, store *persist.RedisStore, lgr *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := store.RedisClient.FlushAll(ctx).Err()
		if err != nil {
			lgr.Error(err.Error(), "cache_purge", "CachePurge", "FlushAll")
		}
	}
}
