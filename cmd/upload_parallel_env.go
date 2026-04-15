package main

import (
	"os"
	"strconv"
	"strings"
)

const uploadParallelEnvMax = 16

// resolveUploadParallelForProcess 返回应写入 KUAKE_UPLOAD_PARALLEL 的十进制字符串，或 "" 表示不设置。
// 优先级（design）：命令行 flag 已解析出的值 > 环境变量 KUAKE_UPLOAD_PARALLEL（1–16）；非法 env 视为未设置。
func resolveUploadParallelForProcess(flagValue string) string {
	if strings.TrimSpace(flagValue) != "" {
		return strings.TrimSpace(flagValue)
	}
	s := strings.TrimSpace(os.Getenv("KUAKE_UPLOAD_PARALLEL"))
	if s == "" {
		return ""
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > uploadParallelEnvMax {
		return ""
	}
	return strconv.Itoa(n)
}
