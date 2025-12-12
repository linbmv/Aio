package cooldown

import "net/http"

// Category 表示错误归类结果
type Category int

const (
	CategoryNone Category = iota
	CategoryKey
	CategoryProvider
	CategoryClient
)

// ClassifyStatus 根据状态码区分错误类型
func ClassifyStatus(code int) Category {
	switch {
	case code == http.StatusTooManyRequests: // 429
		return CategoryKey
	case code == http.StatusUnauthorized || code == http.StatusForbidden: // 401, 403
		return CategoryKey // Key 失效
	case code >= 500:
		return CategoryProvider
	case code >= 400:
		return CategoryClient
	default:
		return CategoryNone
	}
}
