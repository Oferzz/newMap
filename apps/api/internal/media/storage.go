package media

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/config"
)

// Storage defines the interface for media storage
type Storage interface {
	Upload(file *multipart.FileHeader, userID string) (*MediaFile, error)
	Delete(filePath string) error
	GetURL(filePath string) string
	GetFullPath(filePath string) string
	EnsureDirectories() error
}

// MediaFile represents a stored media file
type MediaFile struct {
	ID              string    `json:"id"`
	Filename        string    `json:"filename"`
	OriginalName    string    `json:"original_name"`
	MimeType        string    `json:"mime_type"`
	Size            int64     `json:"size"`
	StoragePath     string    `json:"storage_path"`
	URL             string    `json:"url"`
	ThumbnailSmall  string    `json:"thumbnail_small,omitempty"`
	ThumbnailMedium string    `json:"thumbnail_medium,omitempty"`
	ThumbnailLarge  string    `json:"thumbnail_large,omitempty"`
	Width           int       `json:"width,omitempty"`
	Height          int       `json:"height,omitempty"`
	UploadedBy      string    `json:"uploaded_by"`
	UploadedAt      time.Time `json:"uploaded_at"`
}

// DiskStorage implements Storage interface using filesystem
type DiskStorage struct {
	basePath string
	cdnURL   string
	config   *config.MediaConfig
}

// NewDiskStorage creates a new disk storage instance
func NewDiskStorage(cfg *config.MediaConfig) (*DiskStorage, error) {
	storage := &DiskStorage{
		basePath: cfg.StoragePath,
		cdnURL:   cfg.CDNURL,
		config:   cfg,
	}

	// Ensure base directories exist
	if err := storage.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create storage directories: %w", err)
	}

	return storage, nil
}

// EnsureDirectories creates necessary directories
func (s *DiskStorage) EnsureDirectories() error {
	dirs := []string{
		s.basePath,
		filepath.Join(s.basePath, "images"),
		filepath.Join(s.basePath, "images", "original"),
		filepath.Join(s.basePath, "images", "thumbnails"),
		filepath.Join(s.basePath, "images", "thumbnails", "small"),
		filepath.Join(s.basePath, "images", "thumbnails", "medium"),
		filepath.Join(s.basePath, "images", "thumbnails", "large"),
		filepath.Join(s.basePath, "videos"),
		filepath.Join(s.basePath, "temp"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Upload stores a file to disk
func (s *DiskStorage) Upload(fileHeader *multipart.FileHeader, userID string) (*MediaFile, error) {
	// Validate file size
	if fileHeader.Size > s.config.MaxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", fileHeader.Size, s.config.MaxFileSize)
	}

	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Detect mime type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}
	file.Seek(0, 0)

	mimeType := detectMimeType(buffer)
	if !s.isAllowedMimeType(mimeType) {
		return nil, fmt.Errorf("mime type %s is not allowed", mimeType)
	}

	// Generate unique filename
	fileID := generateFileID(fileHeader.Filename, userID)
	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		ext = getExtensionForMimeType(mimeType)
	}
	filename := fmt.Sprintf("%s%s", fileID, ext)

	// Determine storage path based on file type
	var relativePath string
	if strings.HasPrefix(mimeType, "image/") {
		relativePath = filepath.Join("images", "original", getDatePath(), filename)
	} else if strings.HasPrefix(mimeType, "video/") {
		relativePath = filepath.Join("videos", getDatePath(), filename)
	} else {
		return nil, fmt.Errorf("unsupported file type: %s", mimeType)
	}

	fullPath := filepath.Join(s.basePath, relativePath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(fullPath) // Clean up on error
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create media file record
	mediaFile := &MediaFile{
		ID:           fileID,
		Filename:     filename,
		OriginalName: fileHeader.Filename,
		MimeType:     mimeType,
		Size:         written,
		StoragePath:  relativePath,
		URL:          s.GetURL(relativePath),
		UploadedBy:   userID,
		UploadedAt:   time.Now(),
	}

	// Process thumbnails for images asynchronously
	if strings.HasPrefix(mimeType, "image/") {
		// In a real implementation, this would be done asynchronously
		go s.generateThumbnails(fullPath, relativePath, mediaFile)
	}

	return mediaFile, nil
}

// Delete removes a file from disk
func (s *DiskStorage) Delete(filePath string) error {
	fullPath := filepath.Join(s.basePath, filePath)
	
	// Delete original file
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Delete thumbnails if it's an image
	if strings.Contains(filePath, "images/original/") {
		s.deleteThumbnails(filePath)
	}

	return nil
}

// GetURL returns the public URL for a file
func (s *DiskStorage) GetURL(filePath string) string {
	return fmt.Sprintf("%s/%s", s.cdnURL, filePath)
}

// GetFullPath returns the full filesystem path
func (s *DiskStorage) GetFullPath(filePath string) string {
	return filepath.Join(s.basePath, filePath)
}

// isAllowedMimeType checks if the mime type is allowed
func (s *DiskStorage) isAllowedMimeType(mimeType string) bool {
	for _, allowed := range s.config.AllowedMimeTypes {
		if mimeType == allowed {
			return true
		}
	}
	return false
}

// generateThumbnails creates thumbnails for an image
func (s *DiskStorage) generateThumbnails(originalPath, relativePath string, mediaFile *MediaFile) {
	// This is a placeholder - in production, use an image processing library
	// like github.com/disintegration/imaging
	
	// For now, just set the thumbnail URLs to the original
	mediaFile.ThumbnailSmall = mediaFile.URL
	mediaFile.ThumbnailMedium = mediaFile.URL
	mediaFile.ThumbnailLarge = mediaFile.URL
}

// deleteThumbnails removes all thumbnails for an image
func (s *DiskStorage) deleteThumbnails(originalPath string) {
	// This would delete the actual thumbnail files
}

// Helper functions

func generateFileID(filename, userID string) string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%s-%s-%d", filename, userID, timestamp)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func getDatePath() string {
	now := time.Now()
	return filepath.Join(
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
	)
}

func detectMimeType(buffer []byte) string {
	// Simple mime type detection based on magic numbers
	if len(buffer) < 4 {
		return "application/octet-stream"
	}

	// JPEG
	if buffer[0] == 0xFF && buffer[1] == 0xD8 && buffer[2] == 0xFF {
		return "image/jpeg"
	}

	// PNG
	if buffer[0] == 0x89 && buffer[1] == 0x50 && buffer[2] == 0x4E && buffer[3] == 0x47 {
		return "image/png"
	}

	// WebP
	if len(buffer) >= 12 && string(buffer[0:4]) == "RIFF" && string(buffer[8:12]) == "WEBP" {
		return "image/webp"
	}

	// MP4
	if len(buffer) >= 12 && (string(buffer[4:8]) == "ftyp" || string(buffer[4:8]) == "moov") {
		return "video/mp4"
	}

	return "application/octet-stream"
}

func getExtensionForMimeType(mimeType string) string {
	extensions := map[string]string{
		"image/jpeg":  ".jpg",
		"image/png":   ".png",
		"image/webp":  ".webp",
		"video/mp4":   ".mp4",
	}
	
	if ext, ok := extensions[mimeType]; ok {
		return ext
	}
	return ".bin"
}