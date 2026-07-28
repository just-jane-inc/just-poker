package just

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type MessageType string

const (
	user_created_message    MessageType = "user.created"
	not_found_error_message MessageType = "error.not-found"
)

type ResponseMessage[T any] struct {
	Type string `json:"type" example:"user.created"`
	Data T      `json:"data"`
}

type ErrorDTO struct {
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

type HttpResponse struct {
	Code   int
	Object any
}

func IgnoreTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimRight(r.URL.Path, "/")
			r.URL.RawPath = strings.TrimRight(r.URL.RawPath, "/")
		}

		next.ServeHTTP(w, r)
	})
}

func NotFound(message string, code int) HttpResponse {
	return HttpResponse{
		Code: http.StatusNotFound,
		Object: ResponseMessage[ErrorDTO]{
			Type: "error",
			Data: ErrorDTO{
				ErrorCode: code,
				Error:     message,
			},
		},
	}
}

func BadRequest(message string, code int) HttpResponse {
	return HttpResponse{
		Code: http.StatusBadRequest,
		Object: ResponseMessage[ErrorDTO]{
			Type: "error",
			Data: ErrorDTO{
				ErrorCode: code,
				Error:     message,
			},
		},
	}
}

func OK(responseType string, obj any) HttpResponse {
	return HttpResponse{
		Code: http.StatusOK,
		Object: ResponseMessage[any]{
			Type: responseType,
			Data: obj,
		},
	}
}

func (r HttpResponse) WriteJSONResponse(w http.ResponseWriter) {
	WriteJSONResponse(w, r.Code, r.Object)
}

func WriteJSONResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		Logger.Errorf("encountered error when writing json response: %s", err.Error())
	}
}
