package sdk

import (
	"os"
	"strconv"
	"strings"
)

// UploadParallelEnvMax 与 cmd/upload_parallel_env.go 中 uploadParallelEnvMax 保持一致。
const UploadParallelEnvMax = 16

// uploadParallelFromEnv 解析进程环境变量 KUAKE_UPLOAD_PARALLEL（十进制整数）。
// 合法范围为 [1, UploadParallelEnvMax]；未设置或非法时返回 (0, false)。
// kuake 会在解析 upload 子命令后将 --max_upload_parallel 写入该变量，SDK 据此覆盖服务端 part_thread。
func uploadParallelFromEnv() (n int, ok bool) {
	s := strings.TrimSpace(os.Getenv("KUAKE_UPLOAD_PARALLEL"))
	if s == "" {
		return 0, false
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 || v > UploadParallelEnvMax {
		return 0, false
	}
	return v, true
}
