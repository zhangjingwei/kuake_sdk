package sdk

import (
	"testing"
)

func TestNewQuarkClient(t *testing.T) {
	tests := []struct {
		name      string
		cookie    string
		wantPanic bool
	}{
		{
			name:      "create client with cookie",
			cookie:    "test_token=value1; test_token2=value2;",
			wantPanic: false,
		},
		{
			name:      "panic with no cookie",
			cookie:    "",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("NewQuarkClient() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			var client *QuarkClient
			if tt.cookie == "" {
				client = NewQuarkClient()
			} else {
				client = NewQuarkClient(tt.cookie)
			}
			if !tt.wantPanic {
				if client == nil {
					t.Fatalf("NewQuarkClient() returned nil")
				}
				if client.baseURL == "" {
					t.Errorf("NewQuarkClient() client has empty baseURL")
				}
				if client.accessToken == "" {
					t.Errorf("NewQuarkClient() client has empty accessToken")
				}
				if len(client.accessTokens) == 0 {
					t.Errorf("NewQuarkClient() client has empty accessTokens")
				}
			}
		})
	}
}

func TestSetBaseURL(t *testing.T) {
	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	testURL := "https://test.example.com"
	client.SetBaseURL(testURL)

	if client.baseURL != testURL {
		t.Errorf("SetBaseURL() = %v, want %v", client.baseURL, testURL)
	}
}

func TestSetBaseURL_IgnoresEmptyOrNonHTTP(t *testing.T) {
	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}
	valid := "https://example.com/base"
	client.SetBaseURL(valid)
	client.SetBaseURL("   ")
	if client.baseURL != valid {
		t.Errorf("whitespace-only input must not change baseURL, got %q", client.baseURL)
	}
	client.SetBaseURL("")
	if client.baseURL != valid {
		t.Errorf("empty input must not change baseURL, got %q", client.baseURL)
	}
	client.SetBaseURL("ftp://example.com")
	if client.baseURL != valid {
		t.Errorf("non-http(s) scheme must not change baseURL, got %q", client.baseURL)
	}
}

func TestGetCookies(t *testing.T) {
	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	cookies := client.GetCookies()
	if cookies == nil {
		t.Errorf("GetCookies() returned nil")
	}
}

func TestParseCookie(t *testing.T) {
	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	tests := []struct {
		name      string
		cookieStr string
		wantCount int
	}{
		{
			name:      "parse simple cookie",
			cookieStr: "key1=value1; key2=value2",
			wantCount: 2,
		},
		{
			name:      "parse cookie with spaces",
			cookieStr: "key1 = value1 ; key2 = value2 ",
			wantCount: 2,
		},
		{
			name:      "parse empty cookie",
			cookieStr: "",
			wantCount: 0,
		},
		{
			name:      "parse cookie with empty parts",
			cookieStr: "key1=value1;;key2=value2;",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookies := client.parseCookie(tt.cookieStr)
			if len(cookies) != tt.wantCount {
				t.Errorf("parseCookie() returned %d cookies, want %d", len(cookies), tt.wantCount)
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	tests := []struct {
		name    string
		respMap map[string]interface{}
		target  interface{}
		wantErr bool
	}{
		{
			name: "parse valid response",
			respMap: map[string]interface{}{
				"code":   0,
				"status": 200,
				"data": map[string]interface{}{
					"fid": "test_fid",
				},
			},
			target:  &CreateFolderResponse{},
			wantErr: false,
		},
		{
			name: "parse invalid response",
			respMap: map[string]interface{}{
				"code": "invalid",
			},
			target:  &CreateFolderResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.parseResponse(tt.respMap, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConvertToFileInfo(t *testing.T) {
	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}

	tests := []struct {
		name string
		qf   QuarkFileInfo
		want *FileInfo
	}{
		{
			name: "convert file info",
			qf: QuarkFileInfo{
				Fid:         "fid_file_1",
				Name:        "test.txt",
				Path:        "/test.txt",
				Size:        1024,
				IsDirectory: false,
			},
			want: &FileInfo{
				Name:        "test.txt",
				Path:        "/test.txt",
				Size:        1024,
				IsDirectory: false,
			},
		},
		{
			name: "convert directory info",
			qf: QuarkFileInfo{
				Fid:         "fid_dir_1",
				Name:        "test_dir",
				Path:        "/test_dir",
				Size:        0,
				IsDirectory: true,
			},
			want: &FileInfo{
				Name:        "test_dir",
				Path:        "/test_dir",
				Size:        0,
				IsDirectory: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.ConvertToFileInfo(tt.qf)
			if got == nil {
				t.Fatal("ConvertToFileInfo() returned nil")
			}
			if got.Name != tt.want.Name {
				t.Errorf("ConvertToFileInfo() Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Path != tt.want.Path {
				t.Errorf("ConvertToFileInfo() Path = %v, want %v", got.Path, tt.want.Path)
			}
			if got.Size != tt.want.Size {
				t.Errorf("ConvertToFileInfo() Size = %v, want %v", got.Size, tt.want.Size)
			}
			if got.IsDirectory != tt.want.IsDirectory {
				t.Errorf("ConvertToFileInfo() IsDirectory = %v, want %v", got.IsDirectory, tt.want.IsDirectory)
			}
		})
	}
}

func TestConvertToFileInfo_InvalidReturnsNil(t *testing.T) {
	client := createTestClient(t)
	if client == nil {
		t.Fatal("Failed to create test client")
	}
	if got := client.ConvertToFileInfo(QuarkFileInfo{Name: "x"}); got != nil {
		t.Errorf("missing fid: want nil, got %#v", got)
	}
	if got := client.ConvertToFileInfo(QuarkFileInfo{Fid: "f", Name: ""}); got != nil {
		t.Errorf("missing name: want nil, got %#v", got)
	}
}

// createTestClient 创建测试用的客户端
func createTestClient(t *testing.T) *QuarkClient {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Failed to create client: %v", r)
		}
	}()
	return NewQuarkClient("test_token=value1; test_token2=value2;")
}

