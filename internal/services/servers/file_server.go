package servers

import (
	"context"
	"errors"
	"fmt"
	"github.com/Kapeland/task-Astral/internal/models"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	svStruct "github.com/Kapeland/task-Astral/internal/services/structs"
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

type FileServer struct {
	F FileModelManager
	A AuthModelManager
}

func (s *FileServer) UploadDoc(c *gin.Context) {
	lgr := logger.GetLogger()

	form, _ := c.MultipartForm()

	meta := form.Value["meta"]

	jsn, containsJSON := form.Value["json"]

	if !containsJSON {
		lgr.Info("no 'json' in multipart/form", "fileServer", "UploadDoc", "MultipartForm")

		jsn = append(jsn, "")
	}

	file, ok := form.File["file"]

	if !ok {
		lgr.Error("no 'file' in multipart/form", "fileServer", "UploadDoc", "MultipartForm")

		c.JSON(http.StatusBadRequest, svStruct.ErrResponse{Err: svStruct.ErrBody{
			Code: 400,
			Text: "no 'file' in multipart/form",
		}})
		return
	}

	varMeta := svStruct.DocMeta{}

	jsoniter.Unmarshal([]byte(meta[0]), &varMeta) // mw will check error

	doc := svStruct.AddDocForm{
		Meta: varMeta,
		Json: jsoniter.RawMessage(jsn[0]),
	}

	if doc.Meta.File {
		err := c.SaveUploadedFile(file[0], "./file-storage/"+doc.Meta.Name)
		if err != nil {
			lgr.Error(err.Error(), "fileServer", "UploadDoc", "SaveUploadedFile")

			c.JSON(http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
				Code: 500,
				Text: "Can't save uploaded file",
			}})
			return
		}
	} else {
		err := c.SaveUploadedFile(file[0], "./file-storage/json/"+doc.Meta.Name)
		if err != nil {
			lgr.Error(err.Error(), "fileServer", "UploadDoc", "SaveUploadedFile")

			c.JSON(http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
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
	var err error
	if containsJSON {
		newData, err = jsoniter.Marshal(svStruct.AddDocResp{Data: svStruct.DocData{
			JSON: jsoniter.RawMessage(jsn[0]),
			File: doc.Meta.Name,
		}})
	} else {
		newData, err = jsoniter.Marshal(svStruct.AddDocResp{Data: svStruct.DocData{
			JSON: jsoniter.RawMessage("{}"),
			File: doc.Meta.Name,
		}})
	}

	if err != nil {
		lgr.Error(err.Error(), "fileServer", "UploadDoc", "jsoniter.Marshal")

		c.JSON(http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
			Code: 500,
			Text: "Can't marshal response",
		}})
		return
	}
	var tmpMap map[string]interface{}
	err = jsoniter.Unmarshal(newData, &tmpMap)
	if err != nil {
		lgr.Error(err.Error(), "fileServer", "UploadDoc", "jsoniter.Unmarshal")

		c.JSON(http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
			Code: 500,
			Text: "Can't unmarshal response",
		}})
		return
	}

	c.JSON(status, tmpMap)
}

func (s *FileServer) uploadDoc(ctx context.Context, doc svStruct.AddDocForm) (int, svStruct.ErrResponse) {
	lgr := logger.GetLogger()

	err := s.F.AddNewDoc(ctx, structs.FileDTO{
		Meta: structs.DocMetaDTO(doc.Meta),
		Json: doc.Json,
	})
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			lgr.Error("Can't find login mathing provided token", "fileServer", "uploadDoc", "AddNewDoc")
			// На самом деле странно получить такую ошибку. То есть токен есть и он норм, а логина нет.
			return http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
				Code: 500,
				Text: "Can't find login mathing provided token",
			}}

		case errors.Is(err, models.ErrConflict):
			lgr.Info("Document already exists", "fileServer", "uploadDoc", "AddNewDoc")

			return http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
				Code: 400,
				Text: "Duplicated doc",
			}}
		case errors.Is(err, models.ErrInvalidInput):
			lgr.Info("Can't set grants to unexisting user", "fileServer", "uploadDoc", "AddNewDoc")

			return http.StatusBadRequest, svStruct.ErrResponse{Err: svStruct.ErrBody{
				Code: 400,
				Text: "Bad grants",
			}}
		default:
			lgr.Error(err.Error(), "fileServer", "uploadDoc", "AddNewDoc")

			return http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
				Code: 500,
				Text: "Internal server error",
			}}
		}

	}

	return http.StatusOK, svStruct.ErrResponse{}
}

func (s *FileServer) DeleteDoc(c *gin.Context) {
	lgr := logger.GetLogger()

	docID := c.Param("id")
	token := c.Query("token")

	doc, status, errResp := s.deleteDoc(c.Request.Context(), token, docID)

	if status != http.StatusOK {
		c.JSON(status, errResp)
		return
	}

	newData, err := jsoniter.Marshal(svStruct.LogoutResp{jsoniter.RawMessage(fmt.Sprintf("{\"%s\":true}", doc.ID))})
	if err != nil {
		lgr.Error(err.Error(), "fileServer", "DeleteDoc", "jsoniter.Marshal")

		c.JSON(http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
			Code: 500,
			Text: "Can't marshal response",
		}})
		return
	}
	var tmp map[string]interface{}
	err = jsoniter.Unmarshal(newData, &tmp)

	if err != nil {
		lgr.Error(err.Error(), "fileServer", "DeleteDoc", "jsoniter.Unmarshal")

		c.JSON(http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
			Code: 500,
			Text: "Can't unmarshal response",
		}})
		return
	}

	err = os.Remove("./file-storage/" + doc.Name)
	if err != nil {
		lgr.Error(err.Error(), "fileServer", "DeleteDoc", "Remove")

		c.JSON(http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
			Code: 500,
			Text: "Can't remove file from FS",
		}})
		return
	}

	c.JSON(status, tmp)

}

func (s *FileServer) deleteDoc(ctx context.Context, token string, docID string) (structs.RmDoc, int, svStruct.ErrResponse) {
	lgr := logger.GetLogger()

	doc, err := s.F.DeleteDoc(ctx, token, docID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return structs.RmDoc{}, http.StatusBadRequest, svStruct.ErrResponse{Err: svStruct.ErrBody{
				Code: 400,
				Text: "Look like there is no such document",
			}}
			// На самом деле потенциально может быть такое, что логина нет, а токен есть. Но не ясно как такое может получиться.
		}
		lgr.Error(err.Error(), "fileServer", "deleteDoc", "DeleteDoc")

		return structs.RmDoc{}, http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
			Code: 500,
			Text: "Internal error during document deletion",
		}}
	}

	return doc, http.StatusOK, svStruct.ErrResponse{}
}

func (s *FileServer) GetDocsList(c *gin.Context) {
	lgr := logger.GetLogger()

	token := c.Query("token")
	login := c.Query("login")
	key := c.Query("key")
	val := c.Query("value")
	limit, err := strconv.Atoi(c.Query("limit"))

	if err != nil {
		lgr.Error(err.Error(), "fileServer", "GetDocsList", "strconv.Atoi")

		c.JSON(http.StatusBadRequest, svStruct.ErrResponse{Err: svStruct.ErrBody{
			Code: 400,
			Text: "bad limit val",
		}})
		return
	}

	listReq := svStruct.GetDocListReq{
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
		dataResp := svStruct.DataResp{}
		dataResp.Data.Docs = make([]svStruct.Doc, len(docs))
		for i, doc := range docs {
			dataResp.Data.Docs[i] = svStruct.Doc(doc)
		}
		c.JSON(status, svStruct.Response{dataResp})
	}

}

func (s *FileServer) getDocsList(ctx context.Context, listInfo svStruct.GetDocListReq) ([]structs.DocEntry, int, svStruct.ErrResponse) {
	lgr := logger.GetLogger()

	docs, err := s.F.GetDocs(ctx, structs.ListInfo(listInfo))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		lgr.Error(err.Error(), "fileServer", "getDocsList", "GetDocs")

		return []structs.DocEntry{}, http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
			Code: 500,
			Text: "Intermal server error",
		}}
	}

	return docs, http.StatusOK, svStruct.ErrResponse{}
}

func (s *FileServer) GetDoc(c *gin.Context) {
	lgr := logger.GetLogger()

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
			lgr.Error(err.Error(), "fileServer", "GetDoc", "jsoniter.Marshal")

			c.JSON(http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
				Code: 500,
				Text: "Can't marshal response",
			}})
			return
		}
		var tmp map[string]interface{}
		err = jsoniter.Unmarshal(newData, &tmp)

		if err != nil {
			lgr.Error(err.Error(), "fileServer", "GetDoc", "jsoniter.Unmarshal")

			c.JSON(http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
				Code: 500,
				Text: "Can't unmarshal response",
			}})
			return
		}
		c.JSON(status, tmp)
	}

}

func (s *FileServer) getDoc(ctx context.Context, token string, docID string) (structs.GetDocDTO, int, svStruct.ErrResponse) {
	lgr := logger.GetLogger()

	doc, err := s.F.GetDoc(ctx, token, docID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return structs.GetDocDTO{}, http.StatusBadRequest, svStruct.ErrResponse{Err: svStruct.ErrBody{
				Code: 400,
				Text: "Looks like there is no such document",
			}}
		}
		if errors.Is(err, models.ErrForbidden) {
			return structs.GetDocDTO{}, http.StatusForbidden, svStruct.ErrResponse{Err: svStruct.ErrBody{
				Code: 403,
				Text: "Forbidden",
			}}
		}
		lgr.Error(err.Error(), "fileServer", "getDoc", "GetDoc")

		return structs.GetDocDTO{}, http.StatusInternalServerError, svStruct.ErrResponse{Err: svStruct.ErrBody{
			Code: 500,
			Text: "Internal server error during getting document",
		}}
	}

	return doc, http.StatusOK, svStruct.ErrResponse{}
}
