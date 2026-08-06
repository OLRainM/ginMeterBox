package repository

import (
	"os"
	"path/filepath"
)

func writeJSONAtomically(filename string, data []byte) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".json-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := tmp.Chmod(0644); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// Windows 不允许 Rename 覆盖已有文件，先移除旧目标；临时文件已完整落盘，失败时原始数据仍保留。
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(tmpName, filename)
}
