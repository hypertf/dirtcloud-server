package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// setupTestDB creates a new in-memory SQLite database for testing
func setupTestDB(t *testing.T) *DB {
	t.Helper()

	// Use in-memory database for tests
	db, err := NewDB(":memory:")
	require.NoError(t, err, "Failed to create test database")

	return db
}

// setupTestDBWithData creates a test database and populates it with sample data
func setupTestDBWithData(t *testing.T) (*DB, map[string]interface{}) {
	t.Helper()

	db := setupTestDB(t)
	
	// Sample data that can be used by tests
	data := map[string]interface{}{
		"project_id_1": "proj-123",
		"project_id_2": "proj-456",
		"instance_id_1": "inst-123",
		"instance_id_2": "inst-456",
	}

	return db, data
}

// cleanupTestDB closes the test database
func cleanupTestDB(t *testing.T, db *DB) {
	t.Helper()
	
	if db != nil {
		err := db.Close()
		require.NoError(t, err, "Failed to close test database")
	}
}