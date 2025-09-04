package sqlite

import (
	"database/sql"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/nicolas/dirtcloud/domain"
)

// MetadataRepository handles metadata data operations
type MetadataRepository struct {
	db *DB
}

// NewMetadataRepository creates a new metadata repository
func NewMetadataRepository(db *DB) *MetadataRepository {
	return &MetadataRepository{db: db}
}

// normalizePath normalizes a metadata path according to the rules:
// - no empty segments
// - no .. segments
// - no trailing slash duplicates
// - always starts with /
func normalizePath(p string) string {
	// Ensure path starts with /
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	// Clean the path using path.Clean which handles .. and empty segments
	cleaned := path.Clean(p)
	
	// path.Clean might return "." for empty paths, convert back to "/"
	if cleaned == "." {
		cleaned = "/"
	}

	return cleaned
}

// Set creates or updates metadata at the given path
func (r *MetadataRepository) Set(metadataPath, value string) (*domain.Metadata, error) {
	normalizedPath := normalizePath(metadataPath)
	now := time.Now()

	metadata := &domain.Metadata{
		Path:      normalizedPath,
		Value:     value,
		UpdatedAt: now,
	}

	query := `INSERT OR REPLACE INTO metadata (path, value, updated_at) VALUES (?, ?, ?)`
	
	_, err := r.db.Exec(query, metadata.Path, metadata.Value, metadata.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to set metadata: %w", err)
	}

	return metadata, nil
}

// Get retrieves metadata by path
func (r *MetadataRepository) Get(metadataPath string) (*domain.Metadata, error) {
	normalizedPath := normalizePath(metadataPath)
	
	metadata := &domain.Metadata{}
	query := `SELECT path, value, updated_at FROM metadata WHERE path = ?`
	
	err := r.db.QueryRow(query, normalizedPath).Scan(
		&metadata.Path,
		&metadata.Value,
		&metadata.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NotFoundError("metadata", normalizedPath)
		}
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return metadata, nil
}

// List retrieves metadata entries with optional prefix filtering
func (r *MetadataRepository) List(opts domain.MetadataListOptions) ([]string, error) {
	var paths []string
	var args []interface{}
	
	query := `SELECT path FROM metadata`
	var conditions []string

	if opts.Prefix != "" {
		normalizedPrefix := normalizePath(opts.Prefix)
		// For prefix matching, we want paths that start with the prefix
		conditions = append(conditions, "path LIKE ?")
		args = append(args, normalizedPrefix+"%")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY path"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		err := rows.Scan(&path)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metadata path: %w", err)
		}
		paths = append(paths, path)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metadata: %w", err)
	}

	return paths, nil
}

// Delete deletes metadata by path
func (r *MetadataRepository) Delete(metadataPath string) error {
	normalizedPath := normalizePath(metadataPath)
	
	// First check if metadata exists
	_, err := r.Get(metadataPath)
	if err != nil {
		return err
	}

	query := `DELETE FROM metadata WHERE path = ?`
	
	_, err = r.db.Exec(query, normalizedPath)
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}