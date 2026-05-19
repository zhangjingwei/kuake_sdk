package guard

import (
	"os"
	"testing"
)

func TestCheckOp_DeniedOp(t *testing.T) {
	t.Setenv("KUAKE_DENY_OPS", "delete:move")
	g := NewGuard()
	if err := g.CheckOp("delete"); err == nil {
		t.Error("expected error for denied op 'delete'")
	}
	if err := g.CheckOp("move"); err == nil {
		t.Error("expected error for denied op 'move'")
	}
	if err := g.CheckOp("list"); err != nil {
		t.Errorf("unexpected error for allowed op 'list': %v", err)
	}
}

func TestCheckPath_SiblingNotBlocked(t *testing.T) {
	t.Setenv("KUAKE_DENY_PATHS", "/backup")
	g := NewGuard()
	if err := g.CheckPath("/backup2/file.txt"); err != nil {
		t.Errorf("/backup2 should not be blocked by /backup rule: %v", err)
	}
	if err := g.CheckPath("/backup"); err == nil {
		t.Error("expected exact match /backup to be blocked")
	}
	if err := g.CheckPath("/backup/2024/file.txt"); err == nil {
		t.Error("expected child /backup/2024/file.txt to be blocked")
	}
}

func TestCheckPath_MultiplePaths(t *testing.T) {
	t.Setenv("KUAKE_DENY_PATHS", "/重要资料:/备份")
	g := NewGuard()
	if err := g.CheckPath("/重要资料/doc.pdf"); err == nil {
		t.Error("expected /重要资料/doc.pdf to be blocked")
	}
	if err := g.CheckPath("/备份/old.zip"); err == nil {
		t.Error("expected /备份/old.zip to be blocked")
	}
	if err := g.CheckPath("/documents/file.txt"); err != nil {
		t.Errorf("unexpected block of /documents/file.txt: %v", err)
	}
}

func TestCheckUpload_ExtBlocked(t *testing.T) {
	t.Setenv("KUAKE_DENY_EXTS", ".env:.key:.pem")
	g := NewGuard()
	if err := g.CheckUpload(".env", 100); err == nil {
		t.Error("expected .env to be blocked")
	}
	if err := g.CheckUpload("server.key", 100); err == nil {
		t.Error("expected .key to be blocked")
	}
	if err := g.CheckUpload("cert.pem", 100); err == nil {
		t.Error("expected .pem to be blocked")
	}
	if err := g.CheckUpload("main.go", 100); err != nil {
		t.Errorf("unexpected block of main.go: %v", err)
	}
}

func TestCheckUpload_SizeLimit(t *testing.T) {
	t.Setenv("KUAKE_MAX_UPLOAD_MB", "10")
	g := NewGuard()
	if err := g.CheckUpload("big.zip", 11*1024*1024); err == nil {
		t.Error("expected 11MB file to be blocked by 10MB limit")
	}
	if err := g.CheckUpload("small.zip", 5*1024*1024); err != nil {
		t.Errorf("unexpected block of 5MB file: %v", err)
	}
	if err := g.CheckUpload("exact.zip", 10*1024*1024); err != nil {
		t.Errorf("unexpected block of exactly-at-limit file: %v", err)
	}
}

func TestCheckDownloadDir_PathTraversal(t *testing.T) {
	g := NewGuard()
	cases := []struct {
		input   string
		wantErr bool
	}{
		{"../etc", true},
		{"sub/../other", true},
		{"subdir", false},
		{"a/b/c", false},
		{"", false},
	}
	for _, c := range cases {
		err := g.CheckDownloadDir(c.input)
		if c.wantErr && err == nil {
			t.Errorf("CheckDownloadDir(%q): expected error", c.input)
		}
		if !c.wantErr && err != nil {
			t.Errorf("CheckDownloadDir(%q): unexpected error: %v", c.input, err)
		}
	}
}

func TestDownloadDir_Default(t *testing.T) {
	os.Unsetenv("KUAKE_DOWNLOAD_DIR")
	g := NewGuard()
	wd, _ := os.Getwd()
	if g.DownloadDir() != wd {
		t.Errorf("expected DownloadDir=%q (cwd), got %q", wd, g.DownloadDir())
	}
}

func TestDownloadDir_Custom(t *testing.T) {
	t.Setenv("KUAKE_DOWNLOAD_DIR", "/tmp/quark")
	g := NewGuard()
	if g.DownloadDir() != "/tmp/quark" {
		t.Errorf("expected /tmp/quark, got %q", g.DownloadDir())
	}
}
