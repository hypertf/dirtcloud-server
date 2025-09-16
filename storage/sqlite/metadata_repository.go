package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hypertf/dirtcloud-server/domain"
)

// MetadataRepository handles metadata data operations
type MetadataRepository struct {
	db *DB
}

// NewMetadataRepository creates a new metadata repository
func NewMetadataRepository(db *DB) *MetadataRepository {
	return &MetadataRepository{db: db}
}

// Create creates new metadata
func (r *MetadataRepository) Create(req domain.CreateMetadataRequest) (*domain.Metadata, error) {
	// Check if path already exists
	exists, err := r.pathExists(req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to check path existence: %w", err)
	}
	if exists {
		return nil, domain.AlreadyExistsError("metadata with path", req.Path)
	}

	id := uuid.New().String()
	now := time.Now()

	metadata := &domain.Metadata{
		ID:        id,
		Path:      req.Path,
		Value:     req.Value,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO metadata (id, path, value, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	
	_, err = r.db.Exec(query, metadata.ID, metadata.Path, metadata.Value, metadata.CreatedAt, metadata.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata: %w", err)
	}

	return metadata, nil
}

// GetByID retrieves metadata by ID
func (r *MetadataRepository) GetByID(id string) (*domain.Metadata, error) {
	metadata := &domain.Metadata{}
	query := `SELECT id, path, value, created_at, updated_at FROM metadata WHERE id = ?`
	
	err := r.db.QueryRow(query, id).Scan(
		&metadata.ID,
		&metadata.Path,
		&metadata.Value,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NotFoundError("metadata", id)
		}
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return metadata, nil
}

// Update updates existing metadata
func (r *MetadataRepository) Update(id string, req domain.UpdateMetadataRequest) (*domain.Metadata, error) {
	// First get the existing metadata
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check if path is being changed and if new path already exists
	if req.Path != nil && *req.Path != existing.Path {
		exists, err := r.pathExists(*req.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to check path existence: %w", err)
		}
		if exists {
			return nil, domain.AlreadyExistsError("metadata with path", *req.Path)
		}
	}

	// Update fields
	if req.Path != nil {
		existing.Path = *req.Path
	}
	if req.Value != nil {
		existing.Value = *req.Value
	}
	existing.UpdatedAt = time.Now()

	query := `UPDATE metadata SET path = ?, value = ?, updated_at = ? WHERE id = ?`
	
	_, err = r.db.Exec(query, existing.Path, existing.Value, existing.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update metadata: %w", err)
	}

	return existing, nil
}

// List retrieves metadata entries with optional prefix filtering
func (r *MetadataRepository) List(opts domain.MetadataListOptions) ([]*domain.Metadata, error) {
	var metadata []*domain.Metadata
	var args []interface{}
	
	query := `SELECT id, path, value, created_at, updated_at FROM metadata`
	var conditions []string

	if opts.Prefix != "" {
		// For prefix matching, we want paths that start with the prefix
		conditions = append(conditions, "path LIKE ?")
		args = append(args, opts.Prefix+"%")
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
		m := &domain.Metadata{}
		err := rows.Scan(&m.ID, &m.Path, &m.Value, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metadata: %w", err)
		}
		metadata = append(metadata, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metadata: %w", err)
	}

	return metadata, nil
}

// Delete deletes metadata by ID
func (r *MetadataRepository) Delete(id string) error {
	// First check if metadata exists
	_, err := r.GetByID(id)
	if err != nil {
		return err
	}

	query := `DELETE FROM metadata WHERE id = ?`
	
	_, err = r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

// pathExists checks if a path already exists in the database
func (r *MetadataRepository) pathExists(path string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM metadata WHERE path = ?`
	
	err := r.db.QueryRow(query, path).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}