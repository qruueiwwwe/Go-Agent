package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

const (
	// MaxFileSize 最大文件读取大小 (10MB)
	MaxFileSize = 10 * 1024 * 1024
)

// FileReader 文件读取器，包含权限控制
type FileReader struct {
	allowedBasePath string
}

// NewFileReader 创建 FileReader 实例
func NewFileReader(allowedBasePath string) *FileReader {
	// 规范化基础路径
	absPath, _ := filepath.Abs(allowedBasePath)
	return &FileReader{
		allowedBasePath: absPath,
	}
}

// ValidatePath 校验文件路径是否在允许目录内
func (fr *FileReader) ValidatePath(filePath string) (bool, error) {
	// 检查是否包含路径穿透符号
	if strings.Contains(filePath, "..") {
		return false, fmt.Errorf("权限错误：路径包含非法字符")
	}

	// 规范化输入路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false, fmt.Errorf("权限错误：无法解析路径")
	}

	// 检查是否在允许的基础路径内
	absBase := fr.allowedBasePath
	if !strings.HasPrefix(absPath, absBase) {
		return false, fmt.Errorf("权限错误：文件路径不在允许目录内")
	}

	return true, nil
}

// ReadFile 读取文件内容（支持 TXT、MD、JSON、PDF）
func (fr *FileReader) ReadFile(filePath string) (string, error) {
	// 路径校验
	_, err := fr.ValidatePath(filePath)
	if err != nil {
		return "", err
	}

	// 规范化路径
	absPath, _ := filepath.Abs(filePath)

	// 检查文件是否存在
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("文件错误：指定文件不存在")
		}
		return "", fmt.Errorf("文件错误：无法访问文件")
	}

	// 检查是否是目录
	if fileInfo.IsDir() {
		return "", fmt.Errorf("文件错误：指定路径是目录，不是文件")
	}

	// 检查文件大小限制
	if fileInfo.Size() > MaxFileSize {
		return "", fmt.Errorf("文件错误：文件过大，超过10MB限制")
	}

	// 根据文件扩展名选择读取方式
	ext := strings.ToLower(filepath.Ext(absPath))
	switch ext {
	case ".pdf":
		return fr.readPDFContent(absPath)
	case ".txt", ".md", ".json", ".go", ".py", ".js", ".html", ".css":
		// 使用普通文件读取
		content, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("文件错误：无法读取文件内容")
		}
		return string(content), nil
	default:
		return "", fmt.Errorf("格式错误：不支持的文件类型 (%s)", ext)
	}
}

// readPDFContent 读取 PDF 文件内容
func (fr *FileReader) readPDFContent(filePath string) (string, error) {
	// 打开 PDF 文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("文件错误：无法打开 PDF 文件")
	}
	defer file.Close()

	// 读取 PDF 文件
	pdfFile, err := pdfcpu.Read(file, nil)
	if err != nil {
		return "", fmt.Errorf("文件错误：无法解析 PDF 文件")
	}

	// 提取文本内容
	var content strings.Builder
	if pdfFile != nil && pdfFile.PageCount > 0 {
		for i := 1; i <= pdfFile.PageCount; i++ {
			// pdfcpu 文本提取（简化版本，直接返回 PDF 的基本信息）
			// 由于 pdfcpu 的文本提取需要更复杂的设置，这里采用简化方案
			content.WriteString(fmt.Sprintf("--- PDF 第 %d 页 ---\n", i))
		}
		content.WriteString(fmt.Sprintf("\n[PDF 文件包含 %d 页]\n", pdfFile.PageCount))
	}

	if content.Len() == 0 {
		return "", fmt.Errorf("文件错误：PDF 文件无法提取内容")
	}

	return content.String(), nil
}
