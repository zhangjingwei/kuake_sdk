package validation

import (
    "encoding/json"
)

// SafeString 从 map 中安全提取字符串。
func SafeString(data map[string]interface{}, key string) (string, bool) {
    if data == nil {
        return "", false
    }
    raw, ok := data[key]
    if !ok {
        return "", false
    }
    s, ok := raw.(string)
    return s, ok
}

// SafeMap 从 map 中安全提取子对象。
func SafeMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
    if data == nil {
        return nil, false
    }
    raw, ok := data[key]
    if !ok {
        return nil, false
    }
    m, ok := raw.(map[string]interface{})
    return m, ok
}

// ExtractPathFromJSON 从 JSON 字符串中安全提取 path 和 fid。
// 支持完整流式输出格式和简化格式。
func ExtractPathFromJSON(jsonStr string) (string, string, error) {
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
        return "", "", err
    }

    var path, fid string
    if dataObj, ok := SafeMap(data, "data"); ok {
        if p, ok := SafeString(dataObj, "path"); ok {
            path = p
        }
        if f, ok := SafeString(dataObj, "fid"); ok {
            fid = f
        }
    }

    if path == "" {
        if p, ok := SafeString(data, "path"); ok {
            path = p
        }
    }
    if fid == "" {
        if f, ok := SafeString(data, "fid"); ok {
            fid = f
        }
    }

    return path, fid, nil
}
