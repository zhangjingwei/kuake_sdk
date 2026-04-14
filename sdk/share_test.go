package sdk

import (
	"errors"
	"testing"

	"kuake_sdk/sdk/validation"
)

func TestGetMyShareList_InvalidPage_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.GetMyShareList(1001, 50, "created_at", "desc")
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *validation.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want *validation.ValidationError, got %T: %v", err, err)
	}
	if ve.Code != "INVALID_ARG" {
		t.Errorf("Code = %q, want INVALID_ARG", ve.Code)
	}
}

func TestGetShareList_InvalidPage_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.GetShareList("p", "s", "0", 1001, 50, "file_name", "asc")
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *validation.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want *validation.ValidationError, got %T: %v", err, err)
	}
	if ve.Code != "INVALID_ARG" {
		t.Errorf("Code = %q, want INVALID_ARG", ve.Code)
	}
}

func TestGetShareList_InvalidSortBy_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.GetShareList("p", "s", "0", 1, 50, "invalid_sort_field", "asc")
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *validation.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want *validation.ValidationError, got %T: %v", err, err)
	}
	if ve.Code != "INVALID_ARG" {
		t.Errorf("Code = %q, want INVALID_ARG", ve.Code)
	}
}

func TestGetShareList_InvalidSortOrder_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.GetShareList("p", "s", "0", 1, 50, "file_name", "invalid_order")
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *validation.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want *validation.ValidationError, got %T: %v", err, err)
	}
	if ve.Code != "INVALID_ARG" {
		t.Errorf("Code = %q, want INVALID_ARG", ve.Code)
	}
}

func TestGetShareInfo(t *testing.T) {
	t.Skip("Skipping test that requires network access. Use integration tests instead.")

	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{
			name:    "get share info from valid link",
			text:    "https://pan.quark.cn/s/test123",
			wantErr: false,
		},
		{
			name:    "get share info from invalid link",
			text:    "invalid_link",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shareInfo, err := client.GetShareInfo(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetShareInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && shareInfo == nil {
				t.Error("GetShareInfo() returned nil shareInfo")
			}
		})
	}
}

func TestCreateShare(t *testing.T) {
	t.Skip("Skipping test that requires network access. Use integration tests instead.")

	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	tests := []struct {
		name        string
		filePath    string
		expireDays  int
		needPasscode bool
		wantErr     bool
	}{
		{
			name:        "create share without passcode",
			filePath:    "/test_file.txt",
			expireDays:  7,
			needPasscode: false,
			wantErr:     false,
		},
		{
			name:        "create share with passcode",
			filePath:    "/test_file.txt",
			expireDays:  30,
			needPasscode: true,
			wantErr:     false,
		},
		{
			name:        "create permanent share",
			filePath:    "/test_file.txt",
			expireDays:  0,
			needPasscode: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shareLink, err := client.CreateShare(tt.filePath, tt.expireDays, tt.needPasscode)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateShare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && shareLink == nil {
				t.Error("CreateShare() returned nil shareLink")
			}
		})
	}
}

func TestGetShareLink(t *testing.T) {
	t.Skip("Skipping test that requires network access. Use integration tests instead.")

	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	tests := []struct {
		name    string
		shareID string
		wantErr bool
	}{
		{
			name:    "get share link",
			shareID: "test_share_id",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shareLink, err := client.GetShareLink(tt.shareID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetShareLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && shareLink == nil {
				t.Error("GetShareLink() returned nil shareLink")
			}
		})
	}
}

func TestGetShareStoken(t *testing.T) {
	t.Skip("Skipping test that requires network access. Use integration tests instead.")

	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	tests := []struct {
		name     string
		pwdID    string
		passcode string
		wantErr  bool
	}{
		{
			name:     "get share stoken",
			pwdID:    "test_pwd_id",
			passcode: "1234",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.GetShareStoken(tt.pwdID, tt.passcode)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetShareStoken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("GetShareStoken() returned nil result")
			}
		})
	}
}

func TestGetShareList(t *testing.T) {
	t.Skip("Skipping test that requires network access. Use integration tests instead.")

	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	tests := []struct {
		name    string
		pwdID   string
		stoken  string
		pdirFid string
		page    int
		size    int
		wantErr bool
	}{
		{
			name:    "get share list",
			pwdID:   "test_pwd_id",
			stoken:  "test_stoken",
			pdirFid: "0",
			page:    1,
			size:    20,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.GetShareList(tt.pwdID, tt.stoken, tt.pdirFid, tt.page, tt.size, "file_name", "asc")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetShareList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("GetShareList() returned nil result")
			}
		})
	}
}

func TestSaveShareFile(t *testing.T) {
	t.Skip("Skipping test that requires network access. Use integration tests instead.")

	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	tests := []struct {
		name           string
		pwdID          string
		stoken         string
		fidList        []string
		shareTokenList []string
		toPdirFid      string
		pdirSaveAll    bool
		wantErr        bool
	}{
		{
			name:           "save share file",
			pwdID:          "test_pwd_id",
			stoken:         "test_stoken",
			fidList:        []string{"fid1", "fid2"},
			shareTokenList: []string{"token1", "token2"},
			toPdirFid:      "0",
			pdirSaveAll:    false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.SaveShareFile(tt.pwdID, tt.stoken, tt.fidList, tt.shareTokenList, tt.toPdirFid, tt.pdirSaveAll)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveShareFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("SaveShareFile() returned nil result")
			}
		})
	}
}

func TestSetSharePassword(t *testing.T) {
	t.Skip("Skipping test that requires network access. Use integration tests instead.")

	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	tests := []struct {
		name     string
		pwdID    string
		passcode string
		wantErr  bool
	}{
		{
			name:     "set share password",
			pwdID:    "test_pwd_id",
			passcode: "1234",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.SetSharePassword(tt.pwdID, tt.passcode)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetSharePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetShareInfo_EmptyText_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.GetShareInfo("   ")
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *validation.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want *validation.ValidationError, got %T", err)
	}
}

func TestGetShareStoken_InvalidPwdID_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.GetShareStoken("not@valid", "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGetShareList_EmptyStoken_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.GetShareList("validpwd012", "   ", "0", 1, 50, "file_name", "asc")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateShare_InvalidExpireDays_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.CreateShare("/foo", 99, false)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *validation.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want *validation.ValidationError, got %T: %v", err, err)
	}
}

func TestGetShareLink_InvalidShareID_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.GetShareLink("bad/id")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSetSharePassword_EmptyPasscode_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	err := qc.SetSharePassword("pwd012345678", "  ")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGetMyShareList_InvalidOrderField_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.GetMyShareList(1, 50, "not_a_field", "desc")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGetShareIDByFid_InvalidFid_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	_, err := qc.GetShareIDByFid("bad/fid")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestDeleteShare_InvalidShareID_ReturnsValidationError(t *testing.T) {
	qc := &QuarkClient{}
	err := qc.DeleteShare([]string{"validfid00", "bad/id"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

