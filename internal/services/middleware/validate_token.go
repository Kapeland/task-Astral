package middleware

import (
	"github.com/Kapeland/task-Astral/internal/models"
	myErrs "github.com/Kapeland/task-Astral/internal/services/middleware/errors"
	"github.com/Kapeland/task-Astral/internal/services/middleware/structs"
	"github.com/Kapeland/task-Astral/internal/services/servers"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"net/http"
)

func ValidateTokenInMultipartFrom(a servers.AuthModelManager, lgr *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		form, err := c.MultipartForm()

		if err != nil {
			lgr.Error(err.Error(), "validate_token", "ValidateTokenInMultipartFrom", "MultipartForm")
			c.AbortWithStatusJSON(http.StatusBadRequest, structs.ErrResponse{Err: structs.ErrBody{
				Code: 400,
				Text: myErrs.BadMultipartForm,
			}})
		}

		meta, ok := form.Value["meta"]

		if !ok {
			lgr.Error(myErrs.NoMeta, "validate_token", "ValidateTokenInMultipartFrom", "MultipartForm")

			c.AbortWithStatusJSON(http.StatusBadRequest, structs.ErrResponse{Err: structs.ErrBody{
				Code: 400,
				Text: myErrs.NoMeta,
			}})
		}

		varMeta := structs.DocMeta{}

		err = jsoniter.Unmarshal([]byte(meta[0]), &varMeta)
		if err != nil {
			lgr.Error(err.Error(), "validate_token", "ValidateTokenInMultipartFrom", "jsoniter.Unmarshal")

			c.AbortWithStatusJSON(http.StatusBadRequest, structs.ErrResponse{Err: structs.ErrBody{
				Code: 400,
				Text: myErrs.BadMeta,
			}})
		}

		valid, err := a.ValidateToken(c.Request.Context(), varMeta.Token)

		if !valid {
			switch {
			case errors.Is(err, models.ErrInvalidToken):
				c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrResponse{Err: structs.ErrBody{
					Code: 401,
					Text: myErrs.NotAuthToken,
				}})
			case errors.Is(err, models.ErrTokenExpired):
				c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrResponse{Err: structs.ErrBody{
					Code: 401,
					Text: myErrs.TokenExpired,
				}})
			default:
				lgr.Error(err.Error(), "validate_token", "ValidateTokenInMultipartFrom", "ValidateToken")

				c.AbortWithStatusJSON(http.StatusInternalServerError, structs.ErrResponse{Err: structs.ErrBody{
					Code: 500,
					Text: myErrs.ServErr,
				}})
			}
		}
		c.Next()
	}
}

func ValidateTokenInQuery(a servers.AuthModelManager, lgr *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")

		valid, err := a.ValidateToken(c.Request.Context(), token)

		if !valid {
			switch {
			case errors.Is(err, models.ErrInvalidToken):
				c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrResponse{Err: structs.ErrBody{
					Code: 401,
					Text: myErrs.NotAuthToken,
				}})
				return
			case errors.Is(err, models.ErrTokenExpired):
				c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrResponse{Err: structs.ErrBody{
					Code: 401,
					Text: myErrs.TokenExpired,
				}})
				return
			default:
				lgr.Error(err.Error(), "validate_token", "ValidateTokenInQuery", "ValidateToken")

				c.AbortWithStatusJSON(http.StatusInternalServerError, structs.ErrResponse{Err: structs.ErrBody{
					Code: 500,
					Text: myErrs.ServErr,
				}})
				return
			}
		}
		c.Next()
	}
}
