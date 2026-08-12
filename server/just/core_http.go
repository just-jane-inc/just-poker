package just

import (
	"encoding/json"
	"net/http"
	"strings"
)

type MessageType string

type ResponseMessage[T any] struct {
	Type string `json:"type" example:"user.created"`
	Data T      `json:"data"`
}

type ErrorDTO struct {
	ErrorCode ErrorCode `json:"error_code"`
	Error     string    `json:"error"`
}

type HTTPResponse struct {
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

func NotFound(message string, code ErrorCode) HTTPResponse {
	return HTTPResponse{
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

func Unauthorized() HTTPResponse {
	return HTTPResponse{
		Code: http.StatusUnauthorized,
		Object: ResponseMessage[ErrorDTO]{
			Type: "error",
			Data: ErrorDTO{
				ErrorCode: 67,
				Error:     "=^-^=",
			},
		},
	}
}

func MissingToken() HTTPResponse {
	return HTTPResponse{
		Code: http.StatusUnauthorized,
		Object: ResponseMessage[ErrorDTO]{
			Type: "error",
			Data: ErrorDTO{
				ErrorCode: TokenMissing,
				Error:     "access token not provided with request",
			},
		},
	}
}

func InternalError(message string) HTTPResponse {
	return HTTPResponse{
		Code: http.StatusInternalServerError,
		Object: ResponseMessage[ErrorDTO]{
			Type: "error",
			Data: ErrorDTO{
				ErrorCode: Unknown,
				Error:     message,
			},
		},
	}
}

func BadRequest(message string, code ErrorCode) HTTPResponse {
	return HTTPResponse{
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

func InvalidPlayerActionGameIsPaused() HTTPResponse {
	return HTTPResponse{
		Code: http.StatusBadRequest,
		Object: ResponseMessage[ErrorDTO]{
			Type: "error",
			Data: ErrorDTO{
				ErrorCode: GameIsPaused,
				Error:     "the game is currently paused",
			},
		},
	}
}

func OK(responseType string, obj any) HTTPResponse {
	return HTTPResponse{
		Code: http.StatusOK,
		Object: ResponseMessage[any]{
			Type: responseType,
			Data: obj,
		},
	}
}

func (r HTTPResponse) WriteJSONResponse(w http.ResponseWriter) {
	WriteJSONResponse(w, r.Code, r.Object)
}

func WriteJSONResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		Logger.Errorf("encountered error when writing json response: %s", err.Error())
	}
}
