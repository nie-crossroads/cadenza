package music

import "net/http"

const (
	CodeDownloadFailed = "ADD_DOWNLOAD_FAILED"
	CodeCOSFailed      = "ADD_COS_UPLOAD_FAILED"
	CodeTooLarge       = "ADD_FILE_TOO_LARGE"
	CodeTimeout        = "ADD_TIMEOUT"
	CodeBlogToken      = "ADD_BLOG_TOKEN_INVALID"
	CodeBlogValidation = "BLOG_VALIDATION_FAILED"
	CodeDuplicate      = "ADD_DUPLICATE"
	CodeUnknown        = "ADD_UNKNOWN"
	CodeSearchFailed   = "SEARCH_FAILED"
)

// Fail 是返回给调用方的白名单错误。Error 只含错误码，不含上游原文。
type Fail struct {
	Code   string
	Status int
}

func (f *Fail) Error() string {
	if f == nil {
		return ""
	}
	return f.Code
}

func fail(code string, status int) *Fail {
	return &Fail{Code: code, Status: status}
}

func statusOf(code string) int {
	switch code {
	case CodeDuplicate:
		return http.StatusConflict
	case CodeTooLarge:
		return http.StatusRequestEntityTooLarge
	case CodeBlogValidation:
		return http.StatusBadRequest
	case CodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusBadGateway
	}
}
