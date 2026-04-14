package validation

// SafeString 安全字符串转换
func SafeString(data map[string]interface{}, key string) (string, bool) {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

// SafeFloat64 安全浮点数转换
func SafeFloat64(data map[string]interface{}, key string) (float64, bool) {
	if v, ok := data[key]; ok {
		if f, ok := v.(float64); ok {
			return f, true
		}
		// JSON unmarshal 数字可能为 float64，但 int 也可能直接存储为 int
		// 需要支持 int 类型到 float64 的转换
		if i, ok := v.(int); ok {
			return float64(i), true
		}
		// int64 类型（Go 1.21+ 可能使用）
		if i, ok := v.(int64); ok {
			return float64(i), true
		}
	}
	return 0, false
}

// SafeInt 安全整数转换
func SafeInt(data map[string]interface{}, key string) (int, bool) {
	if v, ok := data[key]; ok {
		if i, ok := v.(int); ok {
			return i, true
		}
		// JSON unmarshal 数字可能为 float64
		if f, ok := v.(float64); ok {
			return int(f), true
		}
		// 也支持 int64 类型
		if i, ok := v.(int64); ok {
			return int(i), true
		}
	}
	return 0, false
}

// SafeMap 安全地提取嵌套 map[string]interface{}。
func SafeMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	if v, ok := data[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m, true
		}
	}
	return nil, false
}
