package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnsureDirectoryExists 确保目录存在，如果不存在则创建
func EnsureDirectoryExists(path string) error {
	if path == "" {
		return nil
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径: %v", err)
	}

	// 检查目录是否存在
	info, err := os.Stat(absPath)
	if err == nil {
		// 路径存在，检查是否为目录
		if !info.IsDir() {
			return fmt.Errorf("%s 已存在但不是目录", absPath)
		}
		return nil
	}

	// 如果错误不是「不存在」，则返回错误
	if !os.IsNotExist(err) {
		return fmt.Errorf("检查目录时出错: %v", err)
	}

	// 创建目录
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	return nil
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// GetFileExtension 获取文件扩展名（不包含点）
func GetFileExtension(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}
	return strings.TrimPrefix(ext, ".")
}

// IsYAMLFile 检查文件是否为YAML文件
func IsYAMLFile(path string) bool {
	ext := GetFileExtension(path)
	return ext == "yaml" || ext == "yml"
}

// IsJSONFile 检查文件是否为JSON文件
func IsJSONFile(path string) bool {
	ext := GetFileExtension(path)
	return ext == "json"
}
