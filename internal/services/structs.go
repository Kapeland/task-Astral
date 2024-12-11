package services

import (
	jsoniter "github.com/json-iterator/go"
	"time"
)

type RegisterResp struct {
	Resp RegisterRespBody `json:"response"`
}

type RegisterRespBody struct {
	Login string `json:"login"`
}

type AuthResp struct {
	Resp AuthRespBody `json:"response"`
}

type AuthRespBody struct {
	Token string `json:"token"`
}

type LogoutResp struct {
	Resp jsoniter.RawMessage `json:"response"`
}

type AddDocForm struct {
	Meta DocMeta
	Json jsoniter.RawMessage
}

type DocMeta struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
}

type DocData struct {
	JSON jsoniter.RawMessage `json:"json,omitempty"`
	File string              `json:"file"`
}

type AddDocResp struct {
	Data DocData `json:"data"`
}

type GetDocListReq struct {
	Token string `json:"token"`
	Login string `json:"login"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Limit int    `json:"limit"`
}

type DataResp struct {
	Data DataBody `json:"data,omitempty"`
}
type DataBody struct {
	Docs []Doc `json:"docs,omitempty"`
}
type Doc struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Mime    string    `json:"mime"`
	IsFile  bool      `json:"file"`
	Public  bool      `json:"public"`
	Created time.Time `json:"created"`
	Granted []string  `json:"grant"`
}

type Response struct {
	DataResp
}

type ErrResponse struct {
	Err ErrBody `json:"error"`
}

type ErrBody struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}
