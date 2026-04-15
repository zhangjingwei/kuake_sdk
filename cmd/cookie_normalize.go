package main

import "github.com/zhangjingwei/kuake_cli/sdk"

// normalizeQuarkCookieInput 委托 sdk，与 KUAKE_COOKIE / -cookies 规范化一致。
func normalizeQuarkCookieInput(raw string) string {
	return sdk.NormalizeQuarkCookieInput(raw)
}
