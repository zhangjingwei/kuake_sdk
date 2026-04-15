package sdk

import (
	"os"
	"strings"
)

// NormalizeQuarkCookieInput 对「非空原始凭证串」应用与 CLI 一致的规则：trim 后若仍为空则返回 ""；
// 若既不包含 "__pus=" 也不包含 "__puus="，则视为裸 token，加前缀 "__pus="；若末尾无 ";" 则追加 ";"。
func NormalizeQuarkCookieInput(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "__pus=") && !strings.Contains(s, "__puus=") {
		s = "__pus=" + s
	}
	if !strings.HasSuffix(s, ";") {
		s = s + ";"
	}
	return s
}

// cookieFromSplitEnv 将 KUAKE_PUS / KUAKE_PUUS 拼成一段 Cookie 子串（值为裸串，不含 __pus= 前缀）。
func cookieFromSplitEnv() string {
	pus := strings.TrimSpace(os.Getenv("KUAKE_PUS"))
	puus := strings.TrimSpace(os.Getenv("KUAKE_PUUS"))
	if pus == "" && puus == "" {
		return ""
	}
	var b strings.Builder
	if pus != "" {
		b.WriteString("__pus=")
		b.WriteString(pus)
		b.WriteString("; ")
	}
	if puus != "" {
		b.WriteString("__puus=")
		b.WriteString(puus)
		b.WriteString("; ")
	}
	return strings.TrimSpace(b.String())
}

// ResolveEnvCookieString 从环境变量解析会话 Cookie：整段 KUAKE_COOKIE 优先；否则 KUAKE_PUS + KUAKE_PUUS 拼接后再规范化。
// 供 cmd 与 E2E 等共用。
func ResolveEnvCookieString() string {
	if s := NormalizeQuarkCookieInput(os.Getenv("KUAKE_COOKIE")); s != "" {
		return s
	}
	return NormalizeQuarkCookieInput(cookieFromSplitEnv())
}
