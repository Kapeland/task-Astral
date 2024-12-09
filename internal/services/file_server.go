package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/Kapeland/task-Astral/internal/models"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"os"
	"strconv"
)

type FileModelManager interface {
	AddNewDoc(ctx context.Context, docDTO structs.FileDTO) error
	DeleteDoc(ctx context.Context, token string, docID string) (structs.RmDoc, error)
	GetDocs(ctx context.Context, listInfo structs.ListInfo) ([]structs.DocEntry, error)
	GetDoc(ctx context.Context, token string, docID string) (structs.GetDocDTO, error)
}

type fileServer struct {
	f FileModelManager
	a AuthModelManager
}

func (s *fileServer) UploadDoc(c *gin.Context) {
	form, err := c.MultipartForm()

	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: UploadDoc: MultipartForm: error:%s", err.Error()))
		c.JSON(http.StatusBadRequest, ErrResponse{Err: ErrBody{
			Code: 400,
			Text: "Bad MultipartForm",
		}})
		return
	}

	meta, ok := form.Value["meta"]

	if !ok {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: UploadDoc: no 'meta' in multipart/form"))
		c.JSON(http.StatusBadRequest, ErrResponse{Err: ErrBody{
			Code: 400,
			Text: "no 'meta' in multipart/form",
		}})
		return

	}

	jsn, containsJSON := form.Value["json"]

	if !containsJSON {
		logger.Log(logger.InfoPrefix, fmt.Sprintf("fileServer: UploadDoc: no 'json' in multipart/form"))
		jsn = append(jsn, "")
	}

	file, ok := form.File["file"]

	if !ok {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: UploadDoc: no 'file' in multipart/form"))
		c.JSON(http.StatusBadRequest, ErrResponse{Err: ErrBody{
			Code: 400,
			Text: "no 'file' in multipart/form",
		}})
		return
	}

	varMeta := DocMeta{}

	err = jsoniter.Unmarshal([]byte(meta[0]), &varMeta)
	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: jsoniter.Unmarshal: error:%s", err.Error()))
		c.JSON(http.StatusBadRequest, ErrResponse{Err: ErrBody{
			Code: 400,
			Text: "look like it wrong 'meta' structure",
		}})
		return
	}

	doc := AddDocForm{
		Meta: varMeta,
		Json: jsoniter.RawMessage(jsn[0]),
	}

	if doc.Meta.File {
		err := c.SaveUploadedFile(file[0], "./file-storage/"+doc.Meta.Name)
		if err != nil {
			logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: SaveUploadedFile: error:%s", err.Error()))
			c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
				Code: 500,
				Text: "Can't save uploaded file",
			}})
			return
		}
	} else {
		err := c.SaveUploadedFile(file[0], "./file-storage/json/"+doc.Meta.Name)
		if err != nil {
			logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: SaveUploadedFile: error:%s", err.Error()))
			c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
				Code: 500,
				Text: "Can't save uploaded file",
			}})
			return
		}
	}

	status, errResp := s.uploadDoc(c.Request.Context(), doc)

	if status != http.StatusOK {
		c.JSON(status, errResp)
		return
	}
	var newData []byte
	if containsJSON {
		newData, err = jsoniter.Marshal(AddDocResp{Data: DocData{
			JSON: jsoniter.RawMessage(jsn[0]),
			File: doc.Meta.Name,
		}})
	} else {
		newData, err = jsoniter.Marshal(AddDocResp{Data: DocData{
			JSON: jsoniter.RawMessage("{}"),
			File: doc.Meta.Name,
		}})
	}

	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: jsoniter.Marshal: error:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Can't marshal response",
		}})
		return
	}
	var tmpMap map[string]interface{}
	err = jsoniter.Unmarshal(newData, &tmpMap)
	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: jsoniter.Unmarshal: error:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Can't unmarshal response",
		}})
		return
	}

	c.JSON(status, tmpMap)
}

func (s *fileServer) uploadDoc(ctx context.Context, doc AddDocForm) (int, ErrResponse) {
	valid, err := s.a.ValidateToken(ctx, doc.Meta.Token)

	if !valid {
		if errors.Is(err, models.ErrInvalidToken) {
			return http.StatusUnauthorized, ErrResponse{Err: ErrBody{
				Code: 401,
				Text: "There is no authorized person with this token",
			}}
		}
		if errors.Is(err, models.ErrTokenExpired) {
			return http.StatusUnauthorized, ErrResponse{Err: ErrBody{
				Code: 401,
				Text: "Token expired",
			}}
		}
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: ValidateToken: err:%s", err.Error()))
		return http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Internal server error",
		}}
	}

	err = s.f.AddNewDoc(ctx, structs.FileDTO{
		Meta: structs.DocMetaDTO(doc.Meta),
		Json: doc.Json,
	})
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: AddNewDoc: Can't find login mathing provided token"))
			return http.StatusInternalServerError, ErrResponse{Err: ErrBody{
				Code: 500,
				Text: "Can't find login mathing provided token",
			}}
			// На самом деле странно получить такую ошибку. То есть токен есть и он норм, а логина нет.
		}
		if errors.Is(err, models.ErrConflict) {
			logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: AddNewDoc: Document already exists"))
			return http.StatusInternalServerError, ErrResponse{Err: ErrBody{
				Code: 400,
				Text: "Duplicated doc",
			}}
		}
		if errors.Is(err, models.ErrInvalidInput) {
			logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: AddNewDoc: Can't give grants to unexisting user"))
			return http.StatusBadRequest, ErrResponse{Err: ErrBody{
				Code: 400,
				Text: "Bad grants",
			}}
		}
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: AddNewDoc: err:%s", err.Error()))
		return http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Internal server error",
		}}
	}

	return http.StatusOK, ErrResponse{}
}

func (s *fileServer) DeleteDoc(c *gin.Context) {
	docID := c.Param("id")
	token := c.Query("token")

	doc, status, errResp := s.deleteDoc(c.Request.Context(), token, docID)

	if status != http.StatusOK {
		c.JSON(status, errResp)
		return
	}

	newData, err := jsoniter.Marshal(LogoutResp{jsoniter.RawMessage(fmt.Sprintf("{\"%s\":true}", doc.ID))})
	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: jsoniter.Marshal: error:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Can't marshal response",
		}})
		return
	}
	var tmp map[string]interface{}
	err = jsoniter.Unmarshal(newData, &tmp)

	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: jsoniter.Unmarshal: error:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Can't unmarshal response",
		}})
		return
	}

	err = os.Remove("./file-storage/" + doc.Name)
	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: Remove: error:%s", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Can't remove file from FS",
		}})
		return
	}

	c.JSON(status, tmp)

}

func (s *fileServer) deleteDoc(ctx context.Context, token string, docID string) (structs.RmDoc, int, ErrResponse) {
	valid, err := s.a.ValidateToken(ctx, token)

	if !valid {
		if errors.Is(err, models.ErrTokenExpired) {
			return structs.RmDoc{}, http.StatusUnauthorized, ErrResponse{Err: ErrBody{
				Code: 401,
				Text: "Token expired",
			}}
		}
		if errors.Is(err, models.ErrInvalidToken) {
			return structs.RmDoc{}, http.StatusBadRequest, ErrResponse{Err: ErrBody{
				Code: 400,
				Text: "Invalid token",
			}}
		}
		return structs.RmDoc{}, http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Internal server error furin token validation",
		}}
	}

	doc, err := s.f.DeleteDoc(ctx, token, docID)
	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf(err.Error()))
		if errors.Is(err, models.ErrNotFound) {
			return structs.RmDoc{}, http.StatusBadRequest, ErrResponse{Err: ErrBody{
				Code: 400,
				Text: "Look like there is no such document",
			}}
			// На самом деле потенциально может быть такое, что логина нет, а токен есть. Но не ясно как такое может получиться.
		}
		return structs.RmDoc{}, http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Internal error during document deletion",
		}}
	}

	return doc, http.StatusOK, ErrResponse{}
}

func (s *fileServer) GetDocsList(c *gin.Context) {
	token := c.Query("token")
	login := c.Query("login")
	key := c.Query("key")
	val := c.Query("value")
	limit, err := strconv.Atoi(c.Query("limit"))

	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: strconv.Atoi: error:%s", err.Error()))
		c.JSON(http.StatusBadRequest, ErrResponse{Err: ErrBody{
			Code: 400,
			Text: "bad limit val",
		}})
		return
	}

	listReq := GetDocListReq{
		Token: token,
		Login: login,
		Key:   key,
		Value: val,
		Limit: limit,
	}

	docs, status, errResp := s.getDocsList(c.Request.Context(), listReq)

	if status != http.StatusOK {
		c.JSON(status, errResp)
		return
	}
	if len(docs) == 0 {
		c.JSON(status, gin.H{})
	} else {
		dataResp := DataResp{}
		dataResp.Data.Docs = make([]Doc, len(docs))
		for i, doc := range docs {
			dataResp.Data.Docs[i] = Doc(doc)
		}
		c.JSON(status, Response{dataResp})
	}

}

func (s *fileServer) getDocsList(ctx context.Context, listInfo GetDocListReq) ([]structs.DocEntry, int, ErrResponse) {
	valid, err := s.a.ValidateToken(ctx, listInfo.Token)

	if !valid {
		if errors.Is(err, models.ErrInvalidToken) {
			return []structs.DocEntry{}, http.StatusUnauthorized, ErrResponse{Err: ErrBody{
				Code: 401,
				Text: "There is no authorized person with this token",
			}}
		}
		if errors.Is(err, models.ErrTokenExpired) {
			return []structs.DocEntry{}, http.StatusUnauthorized, ErrResponse{Err: ErrBody{
				Code: 401,
				Text: "Token expired",
			}}
		}
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: ValidateToken: err:%s", err.Error()))

		return []structs.DocEntry{}, http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Internal server erro during token validationr",
		}}
	}
	docs, err := s.f.GetDocs(ctx, structs.ListInfo(listInfo))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		logger.Log(logger.ErrPrefix, fmt.Sprintf(err.Error()))
		return []structs.DocEntry{}, http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Intermal server error",
		}}
	}

	return docs, http.StatusOK, ErrResponse{}
}

func (s *fileServer) GetDoc(c *gin.Context) {
	docID := c.Param("id")
	token := c.Query("token")

	doc, status, errResp := s.getDoc(c.Request.Context(), token, docID)

	if status != http.StatusOK {
		c.JSON(status, errResp)
		return
	}

	if doc.IsFile {
		c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", doc.Name))
		c.Writer.Header().Set("Content-Type", doc.Mime)
		c.File("./file-storage/" + doc.Name)
	} else {
		data, err := os.ReadFile("./file-storage/json/" + doc.Name)
		newData, err := jsoniter.Marshal(jsoniter.RawMessage(fmt.Sprintf("{\"data\": %s}", string(data))))
		if err != nil {
			logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: jsoniter.Marshal: error:%s", err.Error()))
			c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
				Code: 500,
				Text: "Can't marshal response",
			}})
			return
		}
		var tmp map[string]interface{}
		err = jsoniter.Unmarshal(newData, &tmp)

		if err != nil {
			logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: jsoniter.Unmarshal: error:%s", err.Error()))
			c.JSON(http.StatusInternalServerError, ErrResponse{Err: ErrBody{
				Code: 500,
				Text: "Can't unmarshal response",
			}})
			return
		}
		c.JSON(status, tmp)
	}

}

func (s *fileServer) getDoc(ctx context.Context, token string, docID string) (structs.GetDocDTO, int, ErrResponse) {
	valid, err := s.a.ValidateToken(ctx, token)

	if !valid {
		if errors.Is(err, models.ErrInvalidToken) {
			return structs.GetDocDTO{}, http.StatusUnauthorized, ErrResponse{Err: ErrBody{
				Code: 401,
				Text: "There is no authorized person with this token",
			}}
		}
		if errors.Is(err, models.ErrTokenExpired) {
			return structs.GetDocDTO{}, http.StatusUnauthorized, ErrResponse{Err: ErrBody{
				Code: 401,
				Text: "Token expired",
			}}
		}
		logger.Log(logger.ErrPrefix, fmt.Sprintf("fileServer: ValidateToken: err:%s", err.Error()))

		return structs.GetDocDTO{}, http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Internal server erro during token validationr",
		}}
	}

	doc, err := s.f.GetDoc(ctx, token, docID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return structs.GetDocDTO{}, http.StatusBadRequest, ErrResponse{Err: ErrBody{
				Code: 400,
				Text: "Looks like there is no such document",
			}}
		}
		if errors.Is(err, models.ErrForbidden) {
			return structs.GetDocDTO{}, http.StatusForbidden, ErrResponse{Err: ErrBody{
				Code: 403,
				Text: "Forbidden",
			}}
		}
		logger.Log(logger.ErrPrefix, fmt.Sprintf(err.Error()))
		return structs.GetDocDTO{}, http.StatusInternalServerError, ErrResponse{Err: ErrBody{
			Code: 500,
			Text: "Internal server error during getting document",
		}}
	}

	return doc, http.StatusOK, ErrResponse{}
}
