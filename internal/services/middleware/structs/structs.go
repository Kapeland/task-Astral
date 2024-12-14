package structs

type ErrResponse struct {
	Err ErrBody `json:"error"`
}

type ErrBody struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type DocMeta struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
}
