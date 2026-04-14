package validation

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// NonEmpty 非空校验（仅接受 string；nil 与其它类型均报错）
func NonEmpty() *Validator {
	return NewValidator("non_empty", func(value interface{}) error {
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return ErrInvalidArgument("参数不能为空或仅空白")
			}
		case nil:
			return ErrInvalidArgument("参数不能为 nil")
		default:
			return ErrInvalidArgument("非空校验仅支持字符串类型")
		}
		return nil
	})
}

// InRange 整数范围校验：仅当值为 Go 的 int 时做区间判断；其他类型（含 nil、int64、string 等）一律返回 ErrInvalidArgument。浮点数请用 InRangeFloat64。
func InRange(min, max int) *Validator {
	return NewValidator("in_range", func(value interface{}) error {
		if v, ok := value.(int); ok {
			if v < min || v > max {
				return ErrInvalidArgument(fmt.Sprintf("数值须在 [%d, %d] 范围内", min, max))
			}
			return nil
		}
		if _, ok := value.(float64); ok {
			return ErrInvalidArgument("整型范围校验不接受浮点数")
		}
		return ErrInvalidArgument("整型范围校验仅接受 int 类型")
	})
}

// InRangeFloat64 浮点范围校验（接受 float64 及 JSON 常见整型，策略与 SafeFloat64 一致）
func InRangeFloat64(min, max float64) *Validator {
	return NewValidator("in_range_float64", func(value interface{}) error {
		f, ok := floatFromInterface(value)
		if !ok {
			return ErrInvalidArgument("参数必须为数值类型")
		}
		if f < min || f > max {
			return ErrInvalidArgument(fmt.Sprintf("数值须在 [%g, %g] 范围内", min, max))
		}
		return nil
	})
}

func floatFromInterface(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func pathHasNULByte(s string) bool {
	return strings.Contains(s, "\x00")
}

// rawPathHasDotOrDotDotSegment 在 filepath.Clean 之前按路径段检查 "." / ".."，
// 既能在 Windows 上拦截 "/home/../etc"（Clean 后会把 ".." 消掉），又避免把 "file..txt" 当成穿越。
func rawPathHasDotOrDotDotSegment(p string) bool {
	s := filepath.ToSlash(strings.TrimSpace(p))
	if s == "" {
		return false
	}
	for _, seg := range strings.Split(s, "/") {
		if seg == "." || seg == ".." {
			return true
		}
	}
	return false
}

var (
	emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	fidRegexp   = regexp.MustCompile(`^[0-9a-zA-Z_-]{1,128}$`)
)

// ValidEmail 最小邮箱格式校验（与测试中固定策略一致）
func ValidEmail() *Validator {
	return NewValidator("valid_email", func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return ErrInvalidArgument("邮箱参数类型无效")
		}
		s = strings.TrimSpace(s)
		if !emailRegexp.MatchString(s) {
			return ErrInvalidArgument("邮箱格式无效")
		}
		return nil
	})
}

// ValidFID 夸克文件/目录 fid：非空，1–128 位，仅字母数字下划线与连字符
func ValidFID() *Validator {
	return NewValidator("valid_fid", func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return ErrInvalidArgument("fid 参数类型无效")
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return ErrInvalidArgument("fid 不能为空")
		}
		if !fidRegexp.MatchString(s) {
			return ErrInvalidArgument("fid 含非法字符或长度超出限制")
		}
		return nil
	})
}

// ValidPathResult 在通过 ValidPath 安全策略后返回与远程路径一致的规范化路径（斜杠、去尾斜杠等）
func ValidPathResult(path string) (string, error) {
	if err := ValidPath().Validate(path); err != nil {
		return "", err
	}
	return normalizeRemotePath(path), nil
}

func normalizeRemotePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.ReplaceAll(path, "\\", "/")
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func isUNCOrReservedWindowsPath(s string) bool {
	t := strings.TrimSpace(s)
	if t == "" {
		return false
	}
	// UNC（\\host\share）、DOS 设备路径（\\.\）、长路径（\\?\）均以 \\ 开头
	if strings.HasPrefix(t, `\\`) {
		return true
	}
	if runtime.GOOS == "windows" {
		v := filepath.VolumeName(t)
		if len(v) >= 2 && v[0] == '\\' && v[1] == '\\' {
			return true
		}
	}
	// 正斜杠 UNC 形式：//server/share
	sl := filepath.ToSlash(t)
	if strings.HasPrefix(sl, "//") && len(sl) > 2 && sl[2] != '/' {
		return true
	}
	return false
}

// ValidPath 路径安全校验（路径穿越、NUL、UNC / 保留 Windows 路径形式）
func ValidPath() *Validator {
	return NewValidator("valid_path", func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return nil
		}
		if pathHasNULByte(s) {
			return ErrInvalidPath("路径包含 NUL 字符")
		}
		if isUNCOrReservedWindowsPath(s) {
			return ErrInvalidPath(fmt.Sprintf("不支持或保留的路径形式：%s", s))
		}
		if rawPathHasDotOrDotDotSegment(s) {
			return ErrPathTraversal(s)
		}
		return nil
	})
}
