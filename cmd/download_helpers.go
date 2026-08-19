package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zhangjingwei/kuake_cli/sdk"
)

const defaultDownloadWorkers = 4

type directoryDownloadItem struct {
	RemotePath string
	LocalPath  string
	FID        string
	Size       int64
}

type directoryDownloadResult struct {
	Total      int
	Downloaded int
	Skipped    int
	Failed     int
	TotalBytes int64
	LocalPath  string
}

// resolveFileDownloadPath treats an existing directory, a trailing separator,
// or a destination without a file extension as a directory destination.
func resolveFileDownloadPath(destPath, fileName string) (string, error) {
	if destPath == "" || destPath == "." {
		return filepath.Join(destPath, fileName), nil
	}
	info, err := os.Stat(destPath)
	if err == nil && info.IsDir() {
		return filepath.Join(destPath, fileName), nil
	}
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	isDirectory := strings.HasSuffix(destPath, "/") ||
		strings.HasSuffix(destPath, string(filepath.Separator)) ||
		filepath.Ext(filepath.Base(filepath.Clean(destPath))) == ""
	if isDirectory {
		if err := os.MkdirAll(destPath, 0o755); err != nil {
			return "", fmt.Errorf("create destination directory: %w", err)
		}
		return filepath.Join(destPath, fileName), nil
	}
	if parent := filepath.Dir(destPath); parent != "." {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return "", fmt.Errorf("create destination parent: %w", err)
		}
	}
	return destPath, nil
}

// resolveDirectoryDownloadPath always treats destPath as a directory, even if
// its final component has a file extension.
func resolveDirectoryDownloadPath(destPath, remoteName string) (string, error) {
	if destPath == "" || destPath == "." {
		destPath = filepath.Join(destPath, remoteName)
	}
	if err := os.MkdirAll(destPath, 0o755); err != nil {
		return "", fmt.Errorf("create destination directory: %w", err)
	}
	return filepath.Clean(destPath), nil
}

func parseDownloadArgs(args []string) ([]string, int, error) {
	workers := defaultDownloadWorkers
	positional := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] != "--workers" {
			positional = append(positional, args[i])
			continue
		}
		if i+1 >= len(args) {
			return nil, 0, fmt.Errorf("--workers requires a value")
		}
		i++
		var parsed int
		if _, err := fmt.Sscanf(args[i], "%d", &parsed); err != nil || parsed < 1 || parsed > 16 {
			return nil, 0, fmt.Errorf("--workers must be an integer between 1 and 16")
		}
		workers = parsed
	}
	return positional, workers, nil
}

func listDirectoryItems(client *sdk.QuarkClient, remoteRoot, localRoot string) ([]directoryDownloadItem, int, int64, error) {
	queue := []string{strings.TrimSuffix(remoteRoot, "/")}
	items := make([]directoryDownloadItem, 0)
	skipped := 0
	var totalBytes int64
	for len(queue) > 0 {
		remoteDir := queue[0]
		queue = queue[1:]
		response, err := client.List(remoteDir)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("list %s: %w", remoteDir, err)
		}
		if !response.Success {
			return nil, 0, 0, fmt.Errorf("list %s: %s", remoteDir, response.Message)
		}
		children, ok := response.Data["list"].([]sdk.QuarkFileInfo)
		if !ok {
			return nil, 0, 0, fmt.Errorf("list %s returned an unexpected response", remoteDir)
		}
		for _, child := range children {
			cleanRoot := strings.TrimSuffix(path.Clean(remoteRoot), "/")
			cleanChild := path.Clean(child.Path)
			prefix := cleanRoot + "/"
			if cleanRoot == "" {
				prefix = "/"
			}
			if !strings.HasPrefix(cleanChild, prefix) {
				return nil, 0, 0, fmt.Errorf("unsafe remote path returned: %s", child.Path)
			}
			relative := strings.TrimPrefix(cleanChild, prefix)
			if relative == "" || relative == ".." || strings.HasPrefix(relative, "../") {
				return nil, 0, 0, fmt.Errorf("unsafe remote path returned: %s", child.Path)
			}
			localPath := filepath.Join(localRoot, filepath.FromSlash(relative))
			if child.IsDirectory {
				if err := os.MkdirAll(localPath, 0o755); err != nil {
					return nil, 0, 0, err
				}
				queue = append(queue, child.Path)
				continue
			}
			totalBytes += child.Size
			if info, err := os.Stat(localPath); err == nil && info.Mode().IsRegular() {
				if info.Size() == child.Size {
					skipped++
					continue
				}
				// 兼容旧版本直接写入最终路径所留下的不完整文件。
				// 将其迁移为 .part，随后 DownloadFile 会通过 Range 续传。
				if info.Size() > 0 && info.Size() < child.Size {
					partialPath := localPath + ".part"
					partialInfo, partialErr := os.Stat(partialPath)
					if os.IsNotExist(partialErr) {
						if err := os.Rename(localPath, partialPath); err != nil {
							return nil, 0, 0, fmt.Errorf("migrate partial file %s: %w", localPath, err)
						}
					} else if partialErr == nil && partialInfo.Size() < info.Size() {
						if err := os.Remove(partialPath); err != nil {
							return nil, 0, 0, fmt.Errorf("replace shorter partial file %s: %w", partialPath, err)
						}
						if err := os.Rename(localPath, partialPath); err != nil {
							return nil, 0, 0, fmt.Errorf("migrate partial file %s: %w", localPath, err)
						}
					} else if partialErr == nil {
						if err := os.Remove(localPath); err != nil {
							return nil, 0, 0, fmt.Errorf("remove shorter legacy partial file %s: %w", localPath, err)
						}
					}
				}
			}
			items = append(items, directoryDownloadItem{
				RemotePath: child.Path,
				LocalPath:  localPath,
				FID:        child.Fid,
				Size:       child.Size,
			})
		}
	}
	return items, skipped, totalBytes, nil
}

func humanBytes(value int64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	size := float64(value)
	for _, unit := range units {
		if size < 1024 || unit == "TiB" {
			return fmt.Sprintf("%.1f%s", size, unit)
		}
		size /= 1024
	}
	return fmt.Sprintf("%.1fTiB", size)
}

func directoryProgressLines(done, total int, completedBytes, totalBytes int64, active map[string]*sdk.DownloadProgress, failed int) []string {
	activeBytes := int64(0)
	for _, progress := range active {
		activeBytes += progress.Downloaded
	}
	currentBytes := completedBytes + activeBytes
	percentage := float64(0)
	if totalBytes > 0 {
		percentage = float64(currentBytes) / float64(totalBytes) * 100
	} else if total > 0 {
		percentage = float64(done) / float64(total) * 100
	}
	lines := []string{fmt.Sprintf("%.2f%% files %d/%d %s/%s failed %d", percentage, done, total, humanBytes(currentBytes), humanBytes(totalBytes), failed)}
	paths := make([]string, 0, len(active))
	for remotePath := range active {
		paths = append(paths, remotePath)
	}
	sort.Strings(paths)
	for _, remotePath := range paths {
		progress := active[remotePath]
		name := path.Base(remotePath)
		filePercentage := float64(0)
		if progress.Total > 0 {
			filePercentage = float64(progress.Downloaded) / float64(progress.Total) * 100
		}
		lines = append(lines, fmt.Sprintf("  -> %s %s/%s %.2f%%", name, humanBytes(progress.Downloaded), humanBytes(progress.Total), filePercentage))
	}
	return lines
}

type directoryProgressRenderer struct {
	interactive        bool
	lines              int
	lastNonInteractive time.Time
}

func newDirectoryProgressRenderer() *directoryProgressRenderer {
	info, err := os.Stderr.Stat()
	return &directoryProgressRenderer{interactive: err == nil && info.Mode()&os.ModeCharDevice != 0}
}

func (renderer *directoryProgressRenderer) render(lines []string, force bool) {
	if !renderer.interactive {
		if force || time.Since(renderer.lastNonInteractive) >= 5*time.Second {
			fmt.Fprintln(os.Stderr, lines[0])
			renderer.lastNonInteractive = time.Now()
		}
		return
	}
	if renderer.lines > 0 {
		fmt.Fprintf(os.Stderr, "\033[%dA", renderer.lines)
	}
	rows := len(lines)
	if renderer.lines > rows {
		rows = renderer.lines
	}
	for i := 0; i < rows; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		fmt.Fprintf(os.Stderr, "\r\033[K%s\n", line)
	}
	renderer.lines = rows
}

func downloadDirectory(client *sdk.QuarkClient, remoteRoot, remoteName, destPath string, workers int) (*directoryDownloadResult, error) {
	localRoot, err := resolveDirectoryDownloadPath(destPath, remoteName)
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(os.Stderr, "Scanning remote directory...")
	items, skipped, totalBytes, err := listDirectoryItems(client, remoteRoot, localRoot)
	if err != nil {
		return nil, err
	}
	total := skipped + len(items)
	fmt.Fprintf(os.Stderr, "Scan complete: %d files, %s; pending %d, skipped %d; workers %d\n", total, humanBytes(totalBytes), len(items), skipped, workers)

	result := &directoryDownloadResult{Total: total, Skipped: skipped, TotalBytes: totalBytes, LocalPath: localRoot}
	if len(items) == 0 {
		return result, nil
	}

	type outcome struct {
		item directoryDownloadItem
		err  error
	}
	jobs := make(chan directoryDownloadItem)
	outcomes := make(chan outcome)
	var progressMu sync.Mutex
	active := make(map[string]*sdk.DownloadProgress)
	completedBytes := totalBytes
	for _, item := range items {
		completedBytes -= item.Size
	}

	var workersWG sync.WaitGroup
	for i := 0; i < workers; i++ {
		workersWG.Add(1)
		go func() {
			defer workersWG.Done()
			for item := range jobs {
				if err := os.MkdirAll(filepath.Dir(item.LocalPath), 0o755); err != nil {
					outcomes <- outcome{item: item, err: err}
					continue
				}
				progressMu.Lock()
				active[item.RemotePath] = &sdk.DownloadProgress{Total: item.Size}
				progressMu.Unlock()
				err := client.DownloadFile(item.FID, item.LocalPath, path.Base(item.RemotePath), func(p *sdk.DownloadProgress) {
					total := p.Total
					if total <= 0 {
						total = item.Size
					}
					progressMu.Lock()
					active[item.RemotePath] = &sdk.DownloadProgress{Downloaded: p.Downloaded, Total: total}
					progressMu.Unlock()
				})
				outcomes <- outcome{item: item, err: err}
			}
		}()
	}
	go func() {
		for _, item := range items {
			jobs <- item
		}
		close(jobs)
		workersWG.Wait()
		close(outcomes)
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	renderer := newDirectoryProgressRenderer()
	done := skipped
	failures := make([]string, 0)
	for outcomes != nil {
		select {
		case itemOutcome, ok := <-outcomes:
			if !ok {
				outcomes = nil
				continue
			}
			progressMu.Lock()
			delete(active, itemOutcome.item.RemotePath)
			progressMu.Unlock()
			done++
			if itemOutcome.err != nil {
				result.Failed++
				failures = append(failures, fmt.Sprintf("%s: %v", itemOutcome.item.RemotePath, itemOutcome.err))
			} else {
				result.Downloaded++
				completedBytes += itemOutcome.item.Size
			}
		case <-ticker.C:
			progressMu.Lock()
			lines := directoryProgressLines(done, total, completedBytes, totalBytes, active, result.Failed)
			progressMu.Unlock()
			renderer.render(lines, false)
		}
	}
	progressMu.Lock()
	lines := directoryProgressLines(done, total, completedBytes, totalBytes, active, result.Failed)
	progressMu.Unlock()
	renderer.render(lines, true)
	if len(failures) > 0 {
		return result, fmt.Errorf("%d downloads failed; first error: %s", len(failures), failures[0])
	}
	return result, nil
}
