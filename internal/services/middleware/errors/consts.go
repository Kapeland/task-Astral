package errors

const (
	NoMeta           = "no 'meta' in multipart/form"
	BadMeta          = "looks like wrong 'meta' structure"
	BadMultipartForm = "Bad MultipartForm"
	NotAuthToken     = "There is no authorized person with this token"
	TokenExpired     = "Token expired"
	ServErr          = "Internal server error"
)
