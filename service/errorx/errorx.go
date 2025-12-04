package errorx

import (
	"net/http"
	"strings"
)

type ErrorLevel int

const (
	ErrorNone ErrorLevel = iota
	ErrorClient
	ErrorKey
	ErrorChannel
)

type ClassifiedError struct {
	Level      ErrorLevel
	HTTPStatus int
	Code       string
	Message    string
	Retryable  bool
}

// ClassifyHTTPError 分类 HTTP 错误
func ClassifyHTTPError(status int, body []byte, header http.Header) ClassifiedError {
	if status >= 200 && status < 300 {
		return ClassifiedError{Level: ErrorNone, HTTPStatus: status}
	}

	bodyStr := strings.ToLower(string(body))

	switch status {
	case 400, 422:
		return ClassifiedError{
			Level:      ErrorClient,
			HTTPStatus: status,
			Code:       "bad_request",
			Retryable:  false,
		}
	case 404:
		return ClassifiedError{
			Level:      ErrorClient,
			HTTPStatus: status,
			Code:       "not_found",
			Retryable:  false,
		}
	case 401, 403:
		return ClassifiedError{
			Level:      ErrorKey,
			HTTPStatus: status,
			Code:       "auth_invalid",
			Retryable:  true,
		}
	case 429:
		if isGlobalRateLimit(bodyStr, header) {
			return ClassifiedError{
				Level:      ErrorChannel,
				HTTPStatus: status,
				Code:       "rate_limit_global",
				Retryable:  true,
			}
		}
		return ClassifiedError{
			Level:      ErrorKey,
			HTTPStatus: status,
			Code:       "rate_limit_key",
			Retryable:  true,
		}
	case 413:
		return ClassifiedError{
			Level:      ErrorClient,
			HTTPStatus: status,
			Code:       "payload_too_large",
			Retryable:  false,
		}
	case 500, 502, 503, 504, 520, 521, 524:
		return ClassifiedError{
			Level:      ErrorChannel,
			HTTPStatus: status,
			Code:       "upstream_unavailable",
			Retryable:  true,
		}
	default:
		if strings.Contains(bodyStr, "invalid") || strings.Contains(bodyStr, "expired") {
			return ClassifiedError{
				Level:      ErrorKey,
				HTTPStatus: status,
				Code:       "key_invalid",
				Retryable:  true,
			}
		}
		return ClassifiedError{
			Level:      ErrorChannel,
			HTTPStatus: status,
			Code:       "upstream_error",
			Retryable:  true,
		}
	}
}

func isGlobalRateLimit(body string, header http.Header) bool {
	scope := header.Get("X-RateLimit-Scope")
	if scope == "account" || scope == "organization" {
		return true
	}
	if strings.Contains(body, "account") && strings.Contains(body, "rate limit") {
		return true
	}
	return false
}
