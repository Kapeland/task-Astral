package middleware

import (
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
)

func CachePurge(store *persist.MemoryStore, lgr *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := store.Cache.Purge()
		if err != nil {
			lgr.Error(err.Error(), "cache_purge", "CachePurge", "Purge")
		}
	}
}
