package json_format

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Style struct {
	Null    string `json:"null"`
	Boolean string `json:"boolean"`
	Numeric string `json:"numeric"`
	String  string `json:"string"`
	Key     string `json:"key"`
	Generic string `json:"generic"`
}

const resetStyle = "\033[0m"

var DarkStyle = Style{
	Null:    "\033[38;5;94m",
	Boolean: "\033[38;5;45m",
	Numeric: "\033[38;5;141m",
	String:  "\033[38;5;228m",
	Key:     "\033[1m\033[38;5;197m",
	Generic: "\033[0m",
}

var LightStyle = Style{
	Null:    "\033[38;5;118m",
	Boolean: "\033[38;5;18m",
	Numeric: "\033[38;5;55m",
	String:  "\033[38;5;22m",
	Key:     "\033[1m\033[38;5;52m",
	Generic: "\033[0m",
}

func PrettyPrintJSON(obj map[string]interface{}, indent int, style *Style) string {
	return formatJSONValue(obj, indent, 0, style)
}

func formatJSONValue(value interface{}, indent int, currentIndent int, style *Style) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%s\"%s\"%s", style.String, v, resetStyle)
	case bool:
		return fmt.Sprintf("%s%t%s", style.Boolean, v, resetStyle)
	case nil:
		return fmt.Sprintf("%snull%s", style.Null, resetStyle)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%s%v%s", style.Numeric, v, resetStyle)
	case time.Time:
		return fmt.Sprintf("%s\"%s\"%s", style.String, v.Format(time.RFC3339), resetStyle)
	case map[string]interface{}:
		return formatJSONObject(v, indent, currentIndent, style)
	case []interface{}:
		return formatJSONArray(v, indent, currentIndent, style)
	default:
		return fmt.Sprintf("\"%v\"", v)
	}
}

func formatJSONObject(obj map[string]interface{}, indent int, currentIndent int, style *Style) string {
	if len(obj) == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{\n")
	nextLevel := currentIndent + indent

	// Sorting the keys for consistent output
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for i, key := range keys {
		val := obj[key]
		linePrefix := strings.Repeat(" ", nextLevel) + style.Key + "\"" + key + "\"" + resetStyle + ": "
		valueStr := formatJSONValue(val, indent, nextLevel, style)

		sb.WriteString(linePrefix + valueStr)
		if i < len(keys)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat(" ", currentIndent) + "}")
	return sb.String()
}

func formatJSONArray(arr []interface{}, indent int, currentIndent int, style *Style) string {
	if len(arr) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[\n")
	nextLevel := currentIndent + indent

	itemCount := len(arr)
	count := 0

	for _, val := range arr {
		count++
		linePrefix := strings.Repeat(" ", nextLevel)
		valueStr := formatJSONValue(val, indent, nextLevel, style)

		sb.WriteString(linePrefix + valueStr)
		if count < itemCount {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat(" ", currentIndent) + "]")
	return sb.String()
}
