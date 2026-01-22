package bot

import (
	"strings"
)

// escapeMarkdown экранирует символы, которые ломают Markdown форматирование
func EscapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"`", "\\`",
		"~", "\\~",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

// copyStateData создаёт поверхностную копию карты состояния
// Важно: защищает от случайной мутации общего объекта
func CopyStateData(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// GetStateInt64 безопасно извлекает int64 из данных состояния
func GetStateInt64(data map[string]interface{}, key string) (int64, bool) {
	if data == nil {
		return 0, false
	}
	v, ok := data[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case int64:
		return val, true
	case int:
		return int64(val), true
	case float64:
		return int64(val), true
	default:
		return 0, false
	}
}

// GetStateString безопасно извлекает string из данных состояния
func GetStateString(data map[string]interface{}, key string) (string, bool) {
	if data == nil {
		return "", false
	}
	v, ok := data[key]
	if !ok {
		return "", false
	}
	str, ok := v.(string)
	return str, ok
}
