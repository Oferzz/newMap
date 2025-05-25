package media

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"

	"github.com/jmoiron/sqlx"
)

// Service handles media operations
type Service struct {
	db      *sqlx.DB
	storage Storage
}

// NewService creates a new media service
func NewService(db *sqlx.DB, storage Storage) *Service {
	return &Service{
		db:      db,
		storage: storage,
	}
}

// UploadMedia handles file upload and database record creation
func (s *Service) UploadMedia(ctx context.Context, file *multipart.FileHeader, userID string) (*MediaFile, error) {
	// Upload file to storage
	mediaFile, err := s.storage.Upload(file, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Save to database
	query := `
		INSERT INTO media (
			id, filename, original_name, mime_type, size_bytes,
			storage_path, cdn_url, thumbnail_small, thumbnail_medium,
			thumbnail_large, width, height, uploaded_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	_, err = s.db.ExecContext(ctx, query,
		mediaFile.ID,
		mediaFile.Filename,
		mediaFile.OriginalName,
		mediaFile.MimeType,
		mediaFile.Size,
		mediaFile.StoragePath,
		mediaFile.URL,
		mediaFile.ThumbnailSmall,
		mediaFile.ThumbnailMedium,
		mediaFile.ThumbnailLarge,
		mediaFile.Width,
		mediaFile.Height,
		mediaFile.UploadedBy,
	)

	if err != nil {
		// Clean up uploaded file on database error
		_ = s.storage.Delete(mediaFile.StoragePath)
		return nil, fmt.Errorf("failed to save media record: %w", err)
	}

	return mediaFile, nil
}

// GetMedia retrieves media information by ID
func (s *Service) GetMedia(ctx context.Context, mediaID string) (*MediaFile, error) {
	var media MediaFile
	
	query := `
		SELECT 
			id, filename, original_name, mime_type, size_bytes,
			storage_path, cdn_url, thumbnail_small, thumbnail_medium,
			thumbnail_large, width, height, uploaded_by, created_at
		FROM media
		WHERE id = $1`

	err := s.db.GetContext(ctx, &media, query, mediaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("media not found")
		}
		return nil, fmt.Errorf("failed to get media: %w", err)
	}

	return &media, nil
}

// DeleteMedia removes media from storage and database
func (s *Service) DeleteMedia(ctx context.Context, mediaID string, userID string) error {
	// Start transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get media info and verify ownership
	var media struct {
		StoragePath string `db:"storage_path"`
		UploadedBy  string `db:"uploaded_by"`
	}

	query := `SELECT storage_path, uploaded_by FROM media WHERE id = $1`
	err = tx.GetContext(ctx, &media, query, mediaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("media not found")
		}
		return fmt.Errorf("failed to get media: %w", err)
	}

	// Check ownership
	if media.UploadedBy != userID {
		return fmt.Errorf("unauthorized to delete this media")
	}

	// Delete from database
	_, err = tx.ExecContext(ctx, "DELETE FROM media WHERE id = $1", mediaID)
	if err != nil {
		return fmt.Errorf("failed to delete media record: %w", err)
	}

	// Delete related records
	_, err = tx.ExecContext(ctx, "DELETE FROM media_usage WHERE media_id = $1", mediaID)
	if err != nil {
		return fmt.Errorf("failed to delete media usage records: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Delete from storage (after successful DB deletion)
	if err = s.storage.Delete(media.StoragePath); err != nil {
		// Log error but don't fail - DB is already updated
		fmt.Printf("Warning: failed to delete file from storage: %v\n", err)
	}

	return nil
}

// AttachMediaToEntity creates a relationship between media and an entity
func (s *Service) AttachMediaToEntity(ctx context.Context, mediaID, entityType, entityID string) error {
	query := `
		INSERT INTO media_usage (media_id, entity_type, entity_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (media_id, entity_type, entity_id) DO NOTHING`

	_, err := s.db.ExecContext(ctx, query, mediaID, entityType, entityID)
	if err != nil {
		return fmt.Errorf("failed to attach media: %w", err)
	}

	return nil
}

// GetEntityMedia retrieves all media for a specific entity
func (s *Service) GetEntityMedia(ctx context.Context, entityType, entityID string) ([]*MediaFile, error) {
	query := `
		SELECT 
			m.id, m.filename, m.original_name, m.mime_type, m.size_bytes,
			m.storage_path, m.cdn_url, m.thumbnail_small, m.thumbnail_medium,
			m.thumbnail_large, m.width, m.height, m.uploaded_by, m.created_at
		FROM media m
		JOIN media_usage mu ON m.id = mu.media_id
		WHERE mu.entity_type = $1 AND mu.entity_id = $2
		ORDER BY m.created_at DESC`

	var media []*MediaFile
	err := s.db.SelectContext(ctx, &media, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity media: %w", err)
	}

	return media, nil
}

// GetUserMedia retrieves all media uploaded by a user
func (s *Service) GetUserMedia(ctx context.Context, userID string, limit, offset int) ([]*MediaFile, error) {
	query := `
		SELECT 
			id, filename, original_name, mime_type, size_bytes,
			storage_path, cdn_url, thumbnail_small, thumbnail_medium,
			thumbnail_large, width, height, uploaded_by, created_at
		FROM media
		WHERE uploaded_by = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var media []*MediaFile
	err := s.db.SelectContext(ctx, &media, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user media: %w", err)
	}

	return media, nil
}

// CleanupUnusedMedia removes media that isn't attached to any entity
func (s *Service) CleanupUnusedMedia(ctx context.Context, olderThanDays int) error {
	query := `
		DELETE FROM media
		WHERE id IN (
			SELECT m.id
			FROM media m
			LEFT JOIN media_usage mu ON m.id = mu.media_id
			WHERE mu.media_id IS NULL
			AND m.created_at < NOW() - INTERVAL '%d days'
		)`

	_, err := s.db.ExecContext(ctx, fmt.Sprintf(query, olderThanDays))
	if err != nil {
		return fmt.Errorf("failed to cleanup unused media: %w", err)
	}

	return nil
}