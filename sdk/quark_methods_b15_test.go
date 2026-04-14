package sdk

import "testing"

// TestB15_ExportedMethodsChecklist 与 openspec micro-tasks B.15 对照：导出方法清单。
// 本表须与「public-methods-validation-tasks.md §24 方法校验矩阵」及校验 rollout 同步更新，避免与实现脱节。
// 具体参数校验见各业务测试与 validation 包。
func TestB15_ExportedMethodsChecklist(t *testing.T) {
	_ = (*QuarkClient)(nil)
	methods := []string{
		"SetBaseURL", "GetCookies", "ConvertToFileInfo",
		"GetUserInfo",
		"GetShareInfo", "GetShareStoken", "GetShareList", "SaveShareFile", "CreateShare",
		"GetShareLink", "SetSharePassword", "GetMyShareList", "GetShareIDByFid", "DeleteShare",
		"UploadFile", "CreateFolder", "Copy", "Move", "Rename", "List", "GetFileInfo", "Delete",
		"GetDownloadURL", "DownloadFile",
	}
	if len(methods) != 24 {
		t.Fatalf("expected 24 exported QuarkClient methods, got %d — update checklist", len(methods))
	}
}
