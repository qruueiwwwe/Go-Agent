package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"agent/global"
	"agent/library/log"
)

// ========== FileUploadController ==========

// FileUploadController 文件上传控制器
type FileUploadController struct{}

// NewFileUploadController 创建文件上传控制器
func NewFileUploadController() *FileUploadController {
	return &FileUploadController{}
}

// UploadRequest 上传请求
type UploadRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// UploadResponse 上传响应
type UploadResponse struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Message  string `json:"message"`
}

// Upload 处理文件上传请求
// @Summary 文件上传接口
// @Tags file
// @Accept multipart/form-data
// @Param file formData file true "上传文件"
// @Success 200 {object} ReplyResponse
// @Router /api/upload [post]
func (f *FileUploadController) Upload(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if ctx == nil || r == nil {
		log.Error(context.Background(), "params is nil")
		ReplyError(w, global.ErrCodeInvalidParams, global.GetErrorMsg(global.ErrCodeInvalidParams))
		return
	}

	ctx = log.WithContext(ctx)
	start := time.Now()

	// 只支持 POST 和 multipart/form-data
	if r.Method != http.MethodPost {
		log.Warn(ctx, "收到非POST请求: method=%s", r.Method)
		ReplyError(w, global.ErrCodeInvalidParams, "只支持 POST 方法")
		return
	}

	// 解析 multipart form
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil { // 10MB 限制
		log.Error(ctx, "解析文件上传失败: %v", err)
		ReplyError(w, global.ErrCodeInvalidParams, "文件过大或格式错误")
		return
	}

	// 获取上传的文件
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Error(ctx, "获取上传文件失败: %v", err)
		ReplyError(w, global.ErrCodeInvalidParams, "获取上传文件失败")
		return
	}
	defer file.Close()

	// 校验文件名
	filename := fileHeader.Filename
	if filename == "" {
		ReplyError(w, global.ErrCodeInvalidParams, "文件名不能为空")
		return
	}

	// 只允许特定的文件类型
	allowedExts := map[string]bool{
		".txt":  true,
		".md":   true,
		".json": true,
		".go":   true,
		".py":   true,
		".js":   true,
		".pdf":  true,
	}

	ext := filepath.Ext(filename)
	if !allowedExts[ext] {
		log.Warn(ctx, "不支持的文件类型: %s", ext)
		ReplyError(w, global.ErrCodeInvalidParams, fmt.Sprintf("不支持的文件类型: %s", ext))
		return
	}

	// 确保 data 目录存在
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Error(ctx, "创建 data 目录失败: %v", err)
		ReplyError(w, global.ErrCodeServiceInit, "创建目录失败")
		return
	}

	// 保存文件
	filePath := filepath.Join(dataDir, filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		log.Error(ctx, "创建文件失败: %v", err)
		ReplyError(w, global.ErrCodeServiceInit, "创建文件失败")
		return
	}
	defer outFile.Close()

	// 复制文件内容
	written, err := io.Copy(outFile, file)
	if err != nil {
		log.Error(ctx, "写入文件失败: %v", err)
		ReplyError(w, global.ErrCodeServiceInit, "写入文件失败")
		return
	}

	log.Info(ctx, "文件上传成功: filename=%s, size=%d, duration=%v", filename, written, time.Since(start))

	// 返回成功结果
	ReplySuccess(w, UploadResponse{
		Filename: filename,
		Path:     filePath,
		Size:     written,
		Message:  fmt.Sprintf("文件上传成功，大小 %d 字节", written),
	})
}

// ListFiles 获取已上传的文件列表
// @Summary 获取文件列表
// @Tags file
// @Success 200 {object} ReplyResponse
// @Router /api/files [get]
func (f *FileUploadController) ListFiles(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if ctx == nil || r == nil {
		log.Error(context.Background(), "params is nil")
		ReplyError(w, global.ErrCodeInvalidParams, global.GetErrorMsg(global.ErrCodeInvalidParams))
		return
	}

	ctx = log.WithContext(ctx)

	dataDir := "data"
	files, err := os.ReadDir(dataDir)
	if err != nil {
		log.Error(ctx, "读取 data 目录失败: %v", err)
		ReplyError(w, global.ErrCodeServiceInit, "读取文件列表失败")
		return
	}

	type FileInfo struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}

	var fileList []FileInfo
	for _, file := range files {
		if !file.IsDir() {
			info, _ := file.Info()
			fileList = append(fileList, FileInfo{
				Name: file.Name(),
				Size: info.Size(),
			})
		}
	}

	log.Info(ctx, "获取文件列表成功: count=%d", len(fileList))
	ReplySuccess(w, map[string]interface{}{
		"files": fileList,
		"count": len(fileList),
	})
}

// DeleteFile 删除已上传的文件
// @Summary 删除文件
// @Tags file
// @Param filename query string true "文件名"
// @Success 200 {object} ReplyResponse
// @Router /api/file/{filename} [delete]
func (f *FileUploadController) DeleteFile(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if ctx == nil || r == nil {
		log.Error(context.Background(), "params is nil")
		ReplyError(w, global.ErrCodeInvalidParams, global.GetErrorMsg(global.ErrCodeInvalidParams))
		return
	}

	ctx = log.WithContext(ctx)

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		ReplyError(w, global.ErrCodeInvalidParams, "文件名不能为空")
		return
	}

	filePath := filepath.Join("data", filename)

	// 校验路径安全性
	if !filepath.IsAbs(filepath.Join("data", filename)) {
		// 防止路径穿透
		absPath, _ := filepath.Abs(filePath)
		absDataDir, _ := filepath.Abs("data")
		if !filepath.HasPrefix(absPath, absDataDir) {
			ReplyError(w, global.ErrCodeInvalidParams, "不允许访问该文件")
			return
		}
	}

	if err := os.Remove(filePath); err != nil {
		log.Error(ctx, "删除文件失败: %v", err)
		ReplyError(w, global.ErrCodeServiceInit, "删除文件失败")
		return
	}

	log.Info(ctx, "文件删除成功: filename=%s", filename)
	ReplySuccess(w, map[string]string{
		"message": "文件删除成功",
	})
}
