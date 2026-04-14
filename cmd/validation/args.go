package validation

import (
	"fmt"
	"strconv"
	"strings"
)

// singleQuoted 将 s 格式化为单引号包裹的展示形式（内部单引号转义为 \'）。
func singleQuoted(s string) string {
	var b strings.Builder
	b.WriteByte('\'')
	for _, r := range s {
		if r == '\'' {
			b.WriteString(`\'`)
		} else {
			b.WriteRune(r)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

func errIntArg(name, value string) error {
	return fmt.Errorf("%s must be int (例如：'1'); got %s", name, singleQuoted(value))
}

func errBoolArg(name, value string) error {
	return fmt.Errorf("%s must be bool (例如：'true', 'false'); got %s", name, singleQuoted(value))
}

// ParseIntArg 解析整数字符串参数，返回更明确的错误信息。
func ParseIntArg(value, name string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, errIntArg(name, value)
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, errIntArg(name, value)
	}
	return n, nil
}

// ParseOptionalIntArg 解析可选整数参数，如果值为空则返回默认值。
func ParseOptionalIntArg(value, name string, defaultValue int) (int, error) {
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, errIntArg(name, value)
	}
	return n, nil
}

// ParseBoolArg 解析布尔字符串参数，支持 true/false。
func ParseBoolArg(value, name string) (bool, error) {
	switch value {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errBoolArg(name, value)
	}
}
