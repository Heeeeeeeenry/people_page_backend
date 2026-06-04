package controller

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// FileInfo represents uploaded file metadata returned to client
type FileInfo struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

// getMediaRoot returns the media root directory from MEDIA_ROOT env or default "./media"
func getMediaRoot() string {
	root := os.Getenv("MEDIA_ROOT")
	if root == "" {
		root = "./media"
	}
	return root
}

// getFileType determines file type from extension
func getFileType(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "image"
	case ".mp3", ".wav", ".m4a":
		return "audio"
	case ".mp4":
		return "video"
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx":
		return "document"
	default:
		return "other"
	}
}

// ValidateFileExt checks if the extension is in the allowed list
func ValidateFileExt(ext string) bool {
	ext = strings.ToLower(ext)
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
		".mp3": true, ".wav": true, ".m4a": true,
		".mp4": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	}
	return allowed[ext]
}

// UploadFile handles file upload (multipart/form-data, field name "file")
// POST /api/upload
func UploadFile(c *gin.Context) {
	// Limit file size to 50MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 50<<20)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取文件失败: " + err.Error()})
		return
	}
	defer file.Close()

	// Validate extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !ValidateFileExt(ext) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型: " + ext})
		return
	}

	// Build save path: {MEDIA_ROOT}/letters/
	mediaRoot := getMediaRoot()
	dir := filepath.Join(mediaRoot, "letters")
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败"})
		return
	}

	// Generate unique filename: {timestamp}_{random}_{original}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	timestamp := time.Now().UnixMilli()
	random := rng.Int63n(10000)
	safeName := fmt.Sprintf("%d_%04d_%s", timestamp, random, header.Filename)
	savePath := filepath.Join(dir, safeName)

	// Save file to disk
	dst, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入文件失败"})
		return
	}

	// Build public URL path
	url := "/media/letters/" + safeName
	fileType := getFileType(ext)

	c.JSON(http.StatusOK, gin.H{
		"url":  url,
		"name": header.Filename,
		"size": written,
		"type": fileType,
	})
}
