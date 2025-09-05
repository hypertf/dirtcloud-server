package sqlite

import (
	"testing"

	"github.com/jamesnicolas/dirtcloud/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	tests := []struct {
		name        string
		req         domain.CreateMetadataRequest
		expectError bool
		errorCode   string
	}{
		{
			name: "simple create",
			req: domain.CreateMetadataRequest{
				Path:  "/config/app.yaml",
				Value: "database: localhost",
			},
			expectError: false,
		},
		{
			name: "create with nested path",
			req: domain.CreateMetadataRequest{
				Path:  "/config/auth/ldap.yaml",
				Value: "server: ldap.example.com",
			},
			expectError: false,
		},
		{
			name: "create with empty value",
			req: domain.CreateMetadataRequest{
				Path:  "/empty",
				Value: "",
			},
			expectError: false,
		},
		{
			name: "duplicate path should fail",
			req: domain.CreateMetadataRequest{
				Path:  "/config/app.yaml", // Same as first test
				Value: "different value",
			},
			expectError: true,
			errorCode:   domain.ErrorCodeAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := repo.Create(tt.req)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != "" {
					dirtErr, ok := err.(*domain.DirtError)
					require.True(t, ok)
					assert.Equal(t, tt.errorCode, dirtErr.Code)
				}
				assert.Nil(t, metadata)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, metadata.ID)
				assert.Equal(t, tt.req.Path, metadata.Path)
				assert.Equal(t, tt.req.Value, metadata.Value)
				assert.False(t, metadata.CreatedAt.IsZero())
				assert.False(t, metadata.UpdatedAt.IsZero())
				assert.Equal(t, metadata.CreatedAt, metadata.UpdatedAt)
			}
		})
	}
}

func TestMetadataRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	// Create test metadata
	req := domain.CreateMetadataRequest{
		Path:  "/config/app.yaml",
		Value: "database: localhost",
	}
	created, err := repo.Create(req)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		expectError bool
		errorCode   string
	}{
		{
			name:        "existing ID",
			id:          created.ID,
			expectError: false,
		},
		{
			name:        "non-existing ID",
			id:          "non-existent-id",
			expectError: true,
			errorCode:   domain.ErrorCodeNotFound,
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
			errorCode:   domain.ErrorCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := repo.GetByID(tt.id)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != "" {
					dirtErr, ok := err.(*domain.DirtError)
					require.True(t, ok)
					assert.Equal(t, tt.errorCode, dirtErr.Code)
				}
				assert.Nil(t, metadata)
			} else {
				require.NoError(t, err)
				assert.Equal(t, created.ID, metadata.ID)
				assert.Equal(t, created.Path, metadata.Path)
				assert.Equal(t, created.Value, metadata.Value)
				assert.Equal(t, created.CreatedAt.Unix(), metadata.CreatedAt.Unix())
				assert.Equal(t, created.UpdatedAt.Unix(), metadata.UpdatedAt.Unix())
			}
		})
	}
}

func TestMetadataRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	// Create test metadata
	req1 := domain.CreateMetadataRequest{
		Path:  "/config/app.yaml",
		Value: "database: localhost",
	}
	created1, err := repo.Create(req1)
	require.NoError(t, err)

	req2 := domain.CreateMetadataRequest{
		Path:  "/config/other.yaml",
		Value: "other: value",
	}
	created2, err := repo.Create(req2)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		req         domain.UpdateMetadataRequest
		expectError bool
		errorCode   string
		checkResult func(t *testing.T, metadata *domain.Metadata)
	}{
		{
			name: "update value only",
			id:   created1.ID,
			req: domain.UpdateMetadataRequest{
				Value: stringPtr("database: production"),
			},
			expectError: false,
			checkResult: func(t *testing.T, metadata *domain.Metadata) {
				assert.Equal(t, created1.Path, metadata.Path)
				assert.Equal(t, "database: production", metadata.Value)
				assert.Equal(t, created1.CreatedAt.Unix(), metadata.CreatedAt.Unix())
				assert.True(t, metadata.UpdatedAt.After(created1.UpdatedAt))
			},
		},
		{
			name: "update path only",
			id:   created1.ID,
			req: domain.UpdateMetadataRequest{
				Path: stringPtr("/config/app-new.yaml"),
			},
			expectError: false,
			checkResult: func(t *testing.T, metadata *domain.Metadata) {
				assert.Equal(t, "/config/app-new.yaml", metadata.Path)
				assert.Equal(t, "database: production", metadata.Value) // Should keep previous value
			},
		},
		{
			name: "update both path and value",
			id:   created1.ID,
			req: domain.UpdateMetadataRequest{
				Path:  stringPtr("/config/final.yaml"),
				Value: stringPtr("database: final"),
			},
			expectError: false,
			checkResult: func(t *testing.T, metadata *domain.Metadata) {
				assert.Equal(t, "/config/final.yaml", metadata.Path)
				assert.Equal(t, "database: final", metadata.Value)
			},
		},
		{
			name: "update with duplicate path should fail",
			id:   created1.ID,
			req: domain.UpdateMetadataRequest{
				Path: stringPtr(created2.Path), // Try to use path from created2
			},
			expectError: true,
			errorCode:   domain.ErrorCodeAlreadyExists,
		},
		{
			name: "update non-existing ID",
			id:   "non-existent-id",
			req: domain.UpdateMetadataRequest{
				Value: stringPtr("new value"),
			},
			expectError: true,
			errorCode:   domain.ErrorCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := repo.Update(tt.id, tt.req)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != "" {
					dirtErr, ok := err.(*domain.DirtError)
					require.True(t, ok)
					assert.Equal(t, tt.errorCode, dirtErr.Code)
				}
				assert.Nil(t, metadata)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, metadata)
				if tt.checkResult != nil {
					tt.checkResult(t, metadata)
				}
			}
		})
	}
}

func TestMetadataRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	// Set up test data
	testData := []domain.CreateMetadataRequest{
		{Path: "/config/app.yaml", Value: "app config"},
		{Path: "/config/database.yaml", Value: "db config"},
		{Path: "/config/auth/ldap.yaml", Value: "ldap config"},
		{Path: "/config/auth/oauth.yaml", Value: "oauth config"},
		{Path: "/data/users.json", Value: "users data"},
		{Path: "/data/logs/app.log", Value: "log data"},
	}

	var createdMetadata []*domain.Metadata
	for _, req := range testData {
		metadata, err := repo.Create(req)
		require.NoError(t, err)
		createdMetadata = append(createdMetadata, metadata)
	}

	tests := []struct {
		name            string
		prefix          string
		expectedPaths   []string
		expectedLength  int
	}{
		{
			name:   "list all",
			prefix: "",
			expectedPaths: []string{
				"/config/app.yaml",
				"/config/auth/ldap.yaml",
				"/config/auth/oauth.yaml",
				"/config/database.yaml",
				"/data/logs/app.log",
				"/data/users.json",
			},
			expectedLength: 6,
		},
		{
			name:   "list with config prefix",
			prefix: "/config",
			expectedPaths: []string{
				"/config/app.yaml",
				"/config/auth/ldap.yaml",
				"/config/auth/oauth.yaml",
				"/config/database.yaml",
			},
			expectedLength: 4,
		},
		{
			name:   "list with auth prefix",
			prefix: "/config/auth",
			expectedPaths: []string{
				"/config/auth/ldap.yaml",
				"/config/auth/oauth.yaml",
			},
			expectedLength: 2,
		},
		{
			name:   "list with data prefix",
			prefix: "/data",
			expectedPaths: []string{
				"/data/logs/app.log",
				"/data/users.json",
			},
			expectedLength: 2,
		},
		{
			name:           "list with non-matching prefix",
			prefix:         "/nonexistent",
			expectedPaths:  nil,
			expectedLength: 0,
		},
		{
			name:   "list with specific file prefix",
			prefix: "/config/app",
			expectedPaths: []string{
				"/config/app.yaml",
			},
			expectedLength: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := domain.MetadataListOptions{
				Prefix: tt.prefix,
			}

			metadata, err := repo.List(opts)
			require.NoError(t, err)

			assert.Len(t, metadata, tt.expectedLength)

			// Extract paths and verify they match expected
			var actualPaths []string
			for _, m := range metadata {
				actualPaths = append(actualPaths, m.Path)
			}
			assert.Equal(t, tt.expectedPaths, actualPaths)

			// Verify all metadata entries have required fields
			for _, m := range metadata {
				assert.NotEmpty(t, m.ID)
				assert.NotEmpty(t, m.Path)
				assert.NotEmpty(t, m.Value)
				assert.False(t, m.CreatedAt.IsZero())
				assert.False(t, m.UpdatedAt.IsZero())
			}
		})
	}
}

func TestMetadataRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	// Create test metadata
	req := domain.CreateMetadataRequest{
		Path:  "/config/app.yaml",
		Value: "database: localhost",
	}
	created, err := repo.Create(req)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		expectError bool
		errorCode   string
	}{
		{
			name:        "delete existing ID",
			id:          created.ID,
			expectError: false,
		},
		{
			name:        "delete non-existing ID",
			id:          "non-existent-id",
			expectError: true,
			errorCode:   domain.ErrorCodeNotFound,
		},
		{
			name:        "delete already deleted ID",
			id:          created.ID, // Same as first test
			expectError: true,
			errorCode:   domain.ErrorCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.id)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != "" {
					dirtErr, ok := err.(*domain.DirtError)
					require.True(t, ok)
					assert.Equal(t, tt.errorCode, dirtErr.Code)
				}
			} else {
				require.NoError(t, err)

				// Verify it's actually deleted
				_, err = repo.GetByID(tt.id)
				require.Error(t, err)
				assert.True(t, domain.IsNotFound(err))
			}
		})
	}
}

func TestMetadataRepository_PathUniqueness(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	path := "/config/app.yaml"

	// Create first metadata
	req1 := domain.CreateMetadataRequest{
		Path:  path,
		Value: "first value",
	}
	metadata1, err := repo.Create(req1)
	require.NoError(t, err)

	// Try to create another with same path
	req2 := domain.CreateMetadataRequest{
		Path:  path,
		Value: "second value",
	}
	_, err = repo.Create(req2)
	require.Error(t, err)
	assert.True(t, domain.IsAlreadyExists(err))

	// Verify only one exists
	opts := domain.MetadataListOptions{}
	allMetadata, err := repo.List(opts)
	require.NoError(t, err)
	assert.Len(t, allMetadata, 1)
	assert.Equal(t, metadata1.ID, allMetadata[0].ID)
	assert.Equal(t, "first value", allMetadata[0].Value)
}

func TestMetadataRepository_pathExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	path := "/config/app.yaml"

	// Initially should not exist
	exists, err := repo.pathExists(path)
	require.NoError(t, err)
	assert.False(t, exists)

	// Create metadata
	req := domain.CreateMetadataRequest{
		Path:  path,
		Value: "test value",
	}
	_, err = repo.Create(req)
	require.NoError(t, err)

	// Now should exist
	exists, err = repo.pathExists(path)
	require.NoError(t, err)
	assert.True(t, exists)

	// Different path should not exist
	exists, err = repo.pathExists("/different/path")
	require.NoError(t, err)
	assert.False(t, exists)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}