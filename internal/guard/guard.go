package guard

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// Guard enforces operation and path blacklists loaded from environment variables.
type Guard struct {
	denyOps     map[string]bool
	denyPaths   []string
	denyExts    map[string]bool
	maxUploadMB int64
	downloadDir string
}

// NewGuard reads KUAKE_DENY_OPS, KUAKE_DENY_PATHS, KUAKE_DENY_EXTS,
// KUAKE_MAX_UPLOAD_MB, and KUAKE_DOWNLOAD_DIR from the environment.
func NewGuard() *Guard {
	g := &Guard{
		denyOps:  make(map[string]bool),
		denyExts: make(map[string]bool),
	}
	for _, op := range splitColon(os.Getenv("KUAKE_DENY_OPS")) {
		g.denyOps[op] = true
	}
	g.denyPaths = splitColon(os.Getenv("KUAKE_DENY_PATHS"))
	for _, ext := range splitColon(os.Getenv("KUAKE_DENY_EXTS")) {
		g.denyExts[strings.ToLower(ext)] = true
	}
	if v := strings.TrimSpace(os.Getenv("KUAKE_MAX_UPLOAD_MB")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			g.maxUploadMB = n
		}
	}
	if v := os.Getenv("KUAKE_DOWNLOAD_DIR"); v != "" {
		g.downloadDir = v
	} else {
		wd, _ := os.Getwd()
		g.downloadDir = wd
	}
	return g
}

// splitColon splits a colon-separated env value, trimming spaces, dropping empty segments.
func splitColon(v string) []string {
	var out []string
	for _, s := range strings.Split(v, ":") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// CheckOp returns an error if the named operation is in KUAKE_DENY_OPS.
func (g *Guard) CheckOp(op string) error {
	if g.denyOps[op] {
		return fmt.Errorf("operation %q is denied (KUAKE_DENY_OPS)", op)
	}
	return nil
}

// CheckPath returns an error if remotePath falls under any KUAKE_DENY_PATHS prefix.
func (g *Guard) CheckPath(remotePath string) error {
	for _, denied := range g.denyPaths {
		if isPathDenied(remotePath, denied) {
			return fmt.Errorf("path %q is protected (KUAKE_DENY_PATHS)", remotePath)
		}
	}
	return nil
}

// isPathDenied reports whether target equals deny or is a descendant of it.
// Uses path.Clean to normalise both sides, then requires a "/" boundary to
// avoid false positives on sibling paths (e.g. "/backup2" vs "/backup").
func isPathDenied(target, deny string) bool {
	t := path.Clean(target)
	d := path.Clean(deny)
	return t == d || strings.HasPrefix(t, d+"/")
}

// CheckUpload returns an error if filename or sizeBytes violates upload rules.
// filename is the bare filename (no directory component).
func (g *Guard) CheckUpload(filename string, sizeBytes int64) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" && g.denyExts[ext] {
		return fmt.Errorf("file extension %q is denied (KUAKE_DENY_EXTS)", ext)
	}
	if g.maxUploadMB > 0 && sizeBytes > g.maxUploadMB*1024*1024 {
		return fmt.Errorf("file size %d bytes exceeds limit of %d MB (KUAKE_MAX_UPLOAD_MB)",
			sizeBytes, g.maxUploadMB)
	}
	return nil
}

// CheckDownloadDir returns an error if localSubPath contains ".." (path traversal).
func (g *Guard) CheckDownloadDir(localSubPath string) error {
	if strings.Contains(localSubPath, "..") {
		return fmt.Errorf("local_sub_dir must not contain '..' (path traversal)")
	}
	return nil
}

// DownloadDir returns the base download sandbox directory.
func (g *Guard) DownloadDir() string {
	return g.downloadDir
}
