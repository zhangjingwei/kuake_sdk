package sdk

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

// 端到端回归：真实调用夸克网盘 API，验证核心链路。
//
// 启用方式（任选其一）：
//   - E2E_REGRESSION=1
//   - INTEGRATION_TEST=1（与现有集成测试共用开关）
//
// 凭证：仅使用与 kuake CLI 相同的环境变量（见 ResolveEnvCookieString）——KUAKE_COOKIE 或 KUAKE_PUS+KUAKE_PUUS。
// 本测试会尝试加载「当前工作目录」与「上级目录」下的 `.env`（不覆盖已在环境中的变量），便于与仓库根目录 `.env` 对齐。
//
// Windows PowerShell 示例（仅用环境变量，无需 json）：
//   $env:E2E_REGRESSION=1; $env:KUAKE_COOKIE="..."; go test ./sdk -run TestE2E -count=1 -v

func e2eRegressionEnabled() bool {
	return os.Getenv("E2E_REGRESSION") == "1" || os.Getenv("INTEGRATION_TEST") == "1"
}

// e2eTryLoadDotEnv 尝试加载 cwd 与 ../.env，与本地开发习惯一致；不覆盖已 export 的变量。
func e2eTryLoadDotEnv() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	for _, rel := range []string{".env", filepath.Join("..", ".env")} {
		p := filepath.Join(wd, rel)
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			continue
		}
		_ = godotenv.Load(p)
	}
}

func mustFid(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	switch v := data["fid"].(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return ""
	}
}

// TestE2E_Regression_CoreFlow 核心链路（顺序固定）：
// CreateFolder（先建目录）-> 校验目录可 GetFileInfo -> Upload -> GetFileInfo(文件) -> DownloadFile -> Delete(文件) -> Delete(目录)。
func TestE2E_Regression_CoreFlow(t *testing.T) {
	if !e2eRegressionEnabled() {
		t.Skip("Skipping e2e regression. Set E2E_REGRESSION=1 or INTEGRATION_TEST=1.")
	}
	e2eTryLoadDotEnv()

	cookie := ResolveEnvCookieString()
	if cookie == "" {
		t.Skip("No credentials: set KUAKE_COOKIE or KUAKE_PUS/KUAKE_PUUS, or place .env in cwd or parent (loaded before this check).")
	}
	client := NewQuarkClient(cookie)
	if client == nil {
		t.Fatal("NewQuarkClient returned nil")
	}

	var randSuffix [8]byte
	if _, err := rand.Read(randSuffix[:]); err != nil {
		t.Fatalf("rand: %v", err)
	}
	folderName := "kuake_e2e_" + hex.EncodeToString(randSuffix[:])
	createFolder, err := client.CreateFolder(folderName, "/")
	if err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}
	if !createFolder.Success {
		t.Fatalf("CreateFolder: %s %s", createFolder.Code, createFolder.Message)
	}
	dirFid := mustFid(createFolder.Data)
	if dirFid == "" {
		t.Fatalf("CreateFolder: missing fid in response data")
	}

	remoteDirPath := "/" + folderName
	var dirInfo *StandardResponse
	for attempt := 0; attempt < 40; attempt++ {
		var err error
		dirInfo, err = client.GetFileInfo(remoteDirPath)
		if err != nil {
			t.Fatalf("GetFileInfo(dir): %v", err)
		}
		if dirInfo.Success {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if dirInfo == nil || !dirInfo.Success {
		t.Fatalf("GetFileInfo(dir): %s %s", dirInfo.Code, dirInfo.Message)
	}
	if got := mustFid(dirInfo.Data); got != dirFid {
		t.Fatalf("GetFileInfo(dir) fid: got %q want %q", got, dirFid)
	}

	remoteFilePath := remoteDirPath + "/e2e_regression.txt"

	payload := []byte("kuake e2e regression " + time.Now().UTC().Format(time.RFC3339Nano) + "\n")
	tmpFile := filepath.Join(t.TempDir(), "upload.txt")
	if err := os.WriteFile(tmpFile, payload, 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}

	uploadResp, err := client.UploadFile(tmpFile, remoteFilePath, nil, nil)
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if !uploadResp.Success {
		t.Fatalf("UploadFile: %s %s", uploadResp.Code, uploadResp.Message)
	}
	fileFid := mustFid(uploadResp.Data)
	if fileFid == "" {
		t.Fatalf("UploadFile: missing fid in response data")
	}

	var info *StandardResponse
	for attempt := 0; attempt < 40; attempt++ {
		var err error
		info, err = client.GetFileInfo(remoteFilePath)
		if err != nil {
			t.Fatalf("GetFileInfo(file): %v", err)
		}
		if info.Success {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if info == nil || !info.Success {
		t.Fatalf("GetFileInfo(file): %s %s", info.Code, info.Message)
	}
	if got := mustFid(info.Data); got != fileFid {
		t.Fatalf("GetFileInfo(file) fid: got %q want %q", got, fileFid)
	}

	destDir := t.TempDir()
	if err := client.DownloadFile(fileFid, destDir, "e2e_regression.txt", nil); err != nil {
		t.Fatalf("DownloadFile: %v (若出现 403/OSS Callback 鉴权失败，见仓库 buglist.txt ISSUE-006)", err)
	}
	downloaded, err := os.ReadFile(filepath.Join(destDir, "e2e_regression.txt"))
	if err != nil {
		t.Fatalf("read downloaded: %v", err)
	}
	if string(downloaded) != string(payload) {
		t.Fatalf("downloaded content mismatch")
	}

	delFile, err := client.Delete(remoteFilePath)
	if err != nil {
		t.Fatalf("Delete file: %v", err)
	}
	if !delFile.Success {
		t.Fatalf("Delete file: %s %s", delFile.Code, delFile.Message)
	}

	delDir, err := client.Delete(remoteDirPath)
	if err != nil {
		t.Fatalf("Delete dir: %v", err)
	}
	if !delDir.Success {
		t.Fatalf("Delete dir: %s %s", delDir.Code, delDir.Message)
	}
}
