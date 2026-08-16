package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	// ErrInvalidGeneratedFile 表示客户端提供的文件标识不满足受限下载规则。
	ErrInvalidGeneratedFile = errors.New("invalid generated file")
	// ErrGeneratedFileNotFound 表示受限目录中不存在可下载的普通文件。
	ErrGeneratedFileNotFound = errors.New("generated file not found")
)

// GeneratedFileStore 统一管理服务器生成的报表和导出文件。
// 它只向调用方暴露不含目录信息的随机 basename，避免泄露部署磁盘路径。
type GeneratedFileStore struct {
	exportDir string
	reportDir string
}

// NewGeneratedFileStore 使用配置中的导出和报表目录创建文件存储服务。
func NewGeneratedFileStore(exportDir, reportDir string) *GeneratedFileStore {
	return &GeneratedFileStore{exportDir: exportDir, reportDir: reportDir}
}

// NewReportFile 分配一个仅由服务端生成的 PNG 文件路径。
func (s *GeneratedFileStore) NewReportFile() (basename, path string, err error) {
	return s.newFile(s.reportDir, ".png")
}

// NewExportFile 分配一个仅由服务端生成的导出文件路径。
func (s *GeneratedFileStore) NewExportFile(extension string) (basename, path string, err error) {
	if extension != ".json" && extension != ".xlsx" {
		return "", "", ErrInvalidGeneratedFile
	}
	return s.newFile(s.exportDir, extension)
}

// ReportDownloadURL 返回报表受限下载 API，而不是文件系统路径。
func (s *GeneratedFileStore) ReportDownloadURL(basename string) string {
	return buildDownloadURL("/api/v1/billing/download", basename)
}

// ExportDownloadURL 返回导出文件受限下载 API，而不是文件系统路径。
func (s *GeneratedFileStore) ExportDownloadURL(basename string) string {
	return buildDownloadURL("/api/v1/billing/export/download", basename)
}

// ResolveReportDownload 验证 PNG basename 并解析为 reportDir 内的真实普通文件。
func (s *GeneratedFileStore) ResolveReportDownload(basename string) (string, error) {
	return s.resolveDownload(s.reportDir, basename, ".png")
}

// ResolveExportDownload 验证导出 basename 并解析为 exportDir 内的真实普通文件。
func (s *GeneratedFileStore) ResolveExportDownload(basename string) (string, error) {
	return s.resolveDownload(s.exportDir, basename, ".json", ".xlsx")
}

func (s *GeneratedFileStore) newFile(dir, extension string) (basename, path string, err error) {
	root, err := prepareDirectory(dir)
	if err != nil {
		return "", "", err
	}

	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return "", "", err
	}
	basename = fmt.Sprintf("generated_%s_%x%s", time.Now().UTC().Format("20060102150405"), token, extension)
	return basename, filepath.Join(root, basename), nil
}

func (s *GeneratedFileStore) resolveDownload(dir, basename string, allowedExtensions ...string) (string, error) {
	if !isSafeBasename(basename, allowedExtensions...) {
		return "", ErrInvalidGeneratedFile
	}

	root, err := prepareDirectory(dir)
	if err != nil {
		return "", err
	}
	candidate := filepath.Clean(filepath.Join(root, basename))
	rel, err := filepath.Rel(root, candidate)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", ErrInvalidGeneratedFile
	}

	// 先使用 Lstat 拒绝直接符号链接，避免 Windows 上 Stat/EvalSymlinks
	// 对不同链接类型的解析差异将目录内文件重定向到允许目录之外。
	info, err := os.Lstat(candidate)
	if err != nil {
		return "", ErrGeneratedFileNotFound
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", ErrInvalidGeneratedFile
	}
	if !info.Mode().IsRegular() {
		return "", ErrGeneratedFileNotFound
	}

	// 再解析父目录中的链接，确保真实目标仍在受限目录内。
	realCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", ErrGeneratedFileNotFound
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	rel, err = filepath.Rel(realRoot, realCandidate)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", ErrInvalidGeneratedFile
	}
	return realCandidate, nil
}

func prepareDirectory(dir string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", errors.New("generated file directory is empty")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Abs(dir)
}

func isSafeBasename(name string, allowedExtensions ...string) bool {
	if name == "" || filepath.Base(name) != name || filepath.Clean(name) != name || filepath.IsAbs(name) {
		return false
	}

	extension := filepath.Ext(name)
	allowed := false
	for _, candidate := range allowedExtensions {
		if extension == candidate {
			allowed = true
			break
		}
	}
	if !allowed {
		return false
	}

	stem := strings.TrimSuffix(name, extension)
	if !strings.HasPrefix(stem, "generated_") {
		return false
	}
	for _, r := range stem {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func buildDownloadURL(path, basename string) string {
	return path + "?" + url.Values{"file": {basename}}.Encode()
}
