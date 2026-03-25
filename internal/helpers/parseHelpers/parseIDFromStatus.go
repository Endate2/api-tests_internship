package parseHelpers

import "strings"

// хелпер для парсинга id из ответа от сервера от ручки создания объявления
func ParseIDFromStatus(status string) string {
	parts := strings.Split(status, " - ")
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
