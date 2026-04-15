package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func dotEnvPathExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// loadDotEnvFiles 在解析完 CLI 后、创建客户端前调用。
// 依次尝试加载：当前工作目录下的 .env；与 -c/--config 同目录下的 .env（若路径可解析）。
// 已存在于进程环境中的变量不会被覆盖（与 godotenv.Load 行为一致）。
// 设置 KUAKE_LOAD_DOTENV=0 可跳过（仅识别字面量 0，前后空白会被 trim）。
func loadDotEnvFiles(configPath string) {
	if strings.TrimSpace(os.Getenv("KUAKE_LOAD_DOTENV")) == "0" {
		return
	}
	if dotEnvPathExists(".env") {
		_ = godotenv.Load(".env")
	}
	if configPath == "" {
		return
	}
	dir, err := filepath.Abs(filepath.Dir(configPath))
	if err != nil || dir == "" {
		return
	}
	cfgDot := filepath.Join(dir, ".env")
	if dotEnvPathExists(cfgDot) {
		_ = godotenv.Load(cfgDot)
	}
}
