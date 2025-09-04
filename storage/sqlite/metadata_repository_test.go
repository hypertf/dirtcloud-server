package sqlite

import (
	"fmt"
	"testing"

	"github.com/nicolas/dirtcloud/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataRepository_normalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "root path",
			input:    "/",
			expected: "/",
		},
		{
			name:     "simple path",
			input:    "/config",
			expected: "/config",
		},
		{
			name:     "nested path",
			input:    "/config/app/settings",
			expected: "/config/app/settings",
		},
		{
			name:     "path without leading slash",
			input:    "config/app",
			expected: "/config/app",
		},
		{
			name:     "path with trailing slash",
			input:    "/config/app/",
			expected: "/config/app",
		},
		{
			name:     "path with double slashes",
			input:    "/config//app",
			expected: "/config/app",
		},
		{
			name:     "path with dot segments",
			input:    "/config/./app",
			expected: "/config/app",
		},
		{
			name:     "path with dotdot segments",
			input:    "/config/app/../settings",
			expected: "/config/settings",
		},
		{
			name:     "complex path",
			input:    "/config/../app/./settings//database/",
			expected: "/app/settings/database",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "/",
		},
		{
			name:     "just dots",
			input:    ".",
			expected: "/",
		},
		{
			name:     "relative path",
			input:    "app/config",
			expected: "/app/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetadataRepository_Set(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	tests := []struct {
		name         string
		path         string
		value        string
		expectedPath string
	}{
		{
			name:         "simple set",
			path:         "/config/app.yaml",
			value:        "database: localhost",
			expectedPath: "/config/app.yaml",
		},
		{
			name:         "path normalization",
			path:         "config//app/../settings",
			value:        "setting: value",
			expectedPath: "/config/settings",
		},
		{
			name:         "empty value",
			path:         "/empty",
			value:        "",
			expectedPath: "/empty",
		},
		{
			name:         "large value",
			path:         "/large",
			value:        string(make([]byte, 10000)),
			expectedPath: "/large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := repo.Set(tt.path, tt.value)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedPath, metadata.Path)
			assert.Equal(t, tt.value, metadata.Value)
			assert.False(t, metadata.UpdatedAt.IsZero())
		})
	}
}

func TestMetadataRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	// Set up test data
	_, err := repo.Set("/config/app.yaml", "database: localhost")
	require.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		expectError bool
		errorCode   string
		expectValue string
	}{
		{
			name:        "existing path",
			path:        "/config/app.yaml",
			expectError: false,
			expectValue: "database: localhost",
		},
		{
			name:        "path normalization",
			path:        "config//app.yaml",
			expectError: false,
			expectValue: "database: localhost",
		},
		{
			name:        "non-existing path",
			path:        "/nonexistent",
			expectError: true,
			errorCode:   domain.ErrorCodeNotFound,
		},
		{
			name:        "empty path after normalization",
			path:        "/",
			expectError: true,
			errorCode:   domain.ErrorCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := repo.Get(tt.path)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCode != "" {
					dirtErr, ok := err.(*domain.DirtError)
					require.True(t, ok)
					assert.Equal(t, tt.errorCode, dirtErr.Code)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectValue, metadata.Value)
				assert.False(t, metadata.UpdatedAt.IsZero())
			}
		})
	}
}

func TestMetadataRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	// Set up test data
	testPaths := []string{
		"/config/app.yaml",
		"/config/database.yaml",
		"/config/auth/ldap.yaml",
		"/config/auth/oauth.yaml",
		"/data/users.json",
		"/data/logs/app.log",
	}

	for _, path := range testPaths {
		_, err := repo.Set(path, "test-value")
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		prefix         string
		expectedPaths  []string
		expectedLength int
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
		},
		{
			name:   "list with auth prefix",
			prefix: "/config/auth",
			expectedPaths: []string{
				"/config/auth/ldap.yaml",
				"/config/auth/oauth.yaml",
			},
		},
		{
			name:   "list with data prefix",
			prefix: "/data",
			expectedPaths: []string{
				"/data/logs/app.log",
				"/data/users.json",
			},
		},
		{
			name:          "list with non-matching prefix",
			prefix:        "/nonexistent",
			expectedPaths: nil,
		},
		{
			name:   "list with specific file prefix",
			prefix: "/config/app",
			expectedPaths: []string{
				"/config/app.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := domain.MetadataListOptions{
				Prefix: tt.prefix,
			}

			paths, err := repo.List(opts)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedPaths, paths)
		})
	}
}

func TestMetadataRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	// Set up test data
	_, err := repo.Set("/config/app.yaml", "database: localhost")
	require.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		expectError bool
		errorCode   string
	}{
		{
			name:        "delete existing path",
			path:        "/config/app.yaml",
			expectError: false,
		},
		{
			name:        "delete non-existing path",
			path:        "/nonexistent",
			expectError: true,
			errorCode:   domain.ErrorCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.path)

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
				_, err = repo.Get(tt.path)
				require.Error(t, err)
				assert.True(t, domain.IsNotFound(err))
			}
		})
	}
}

func TestMetadataRepository_SetUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	path := "/config/app.yaml"
	originalValue := "database: localhost"
	updatedValue := "database: production"

	// Create initial metadata
	metadata1, err := repo.Set(path, originalValue)
	require.NoError(t, err)
	assert.Equal(t, originalValue, metadata1.Value)

	// Update the same path
	metadata2, err := repo.Set(path, updatedValue)
	require.NoError(t, err)
	assert.Equal(t, updatedValue, metadata2.Value)
	assert.Equal(t, path, metadata2.Path)

	// Verify the update
	metadata3, err := repo.Get(path)
	require.NoError(t, err)
	assert.Equal(t, updatedValue, metadata3.Value)
	assert.True(t, metadata3.UpdatedAt.After(metadata1.UpdatedAt) || metadata3.UpdatedAt.Equal(metadata2.UpdatedAt))

	// Verify there's only one record
	opts := domain.MetadataListOptions{}
	paths, err := repo.List(opts)
	require.NoError(t, err)
	assert.Len(t, paths, 1)
	assert.Equal(t, path, paths[0])
}

func TestMetadataRepository_PathNormalizationConsistency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMetadataRepository(db)

	// Test that different representations of the same path work consistently
	pathVariants := []string{
		"/config/app.yaml",
		"config/app.yaml",
		"/config//app.yaml",
		"/config/./app.yaml",
		"/config/other/../app.yaml",
	}

	value := "test-value"

	// Set using the first variant
	_, err := repo.Set(pathVariants[0], value)
	require.NoError(t, err)

	// Try to get using all variants
	for i, variant := range pathVariants {
		t.Run(fmt.Sprintf("variant_%d", i), func(t *testing.T) {
			metadata, err := repo.Get(variant)
			require.NoError(t, err)
			assert.Equal(t, value, metadata.Value)
			assert.Equal(t, "/config/app.yaml", metadata.Path)
		})
	}

	// Try to set using another variant (should update, not create new)
	newValue := "updated-value"
	_, err = repo.Set(pathVariants[2], newValue)
	require.NoError(t, err)

	// Verify there's still only one record
	opts := domain.MetadataListOptions{}
	paths, err := repo.List(opts)
	require.NoError(t, err)
	assert.Len(t, paths, 1)
	assert.Equal(t, "/config/app.yaml", paths[0])

	// Verify the value was updated
	metadata, err := repo.Get(pathVariants[0])
	require.NoError(t, err)
	assert.Equal(t, newValue, metadata.Value)
}