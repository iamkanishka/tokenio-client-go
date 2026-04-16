package errors

import (
	"encoding/json"
	"io"
	"net/http"
)

type apiErrorBody struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"requestId"`
	Details   map[string]any `json:"details"`
}

// FromResponse converts a non-2xx *http.Response to a typed *APIError.
// It reads and closes the response body.
func FromResponse(resp *http.Response) error {
	if resp == nil {
		return &APIError{Code: CodeUnknown, Message: "nil response"}
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	resp.Body.Close()

	var eb apiErrorBody
	_ = json.Unmarshal(body, &eb)

	if eb.Message == "" {
		eb.Message = http.StatusText(resp.StatusCode)
		if len(body) > 0 && len(body) < 512 {
			eb.Message = string(body)
		}
	}

	return &APIError{
		Code:      resolveCode(resp.StatusCode, Code(eb.Code)),
		Message:   eb.Message,
		Status:    resp.StatusCode,
		RequestID: resp.Header.Get("X-Request-ID"),
		Details:   eb.Details,
	}
}

func resolveCode(status int, apiCode Code) Code {
	if apiCode != "" && apiCode != CodeUnknown {
		return apiCode
	}
	switch status {
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusConflict:
		return CodeConflict
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return CodeValidation
	case http.StatusTooManyRequests:
		return CodeRateLimit
	case http.StatusInternalServerError:
		return CodeInternalServer
	case http.StatusBadGateway:
		return CodeBadGateway
	case http.StatusServiceUnavailable:
		return CodeServiceUnavailable
	case http.StatusGatewayTimeout:
		return CodeTimeout
	default:
		return CodeUnknown
	}
}
