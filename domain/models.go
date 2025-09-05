package domain

import (
	"time"
)

// Project represents a project in the DirtCloud system
type Project struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Instance represents a compute instance within a project
type Instance struct {
	ID        string    `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	Name      string    `json:"name" db:"name"`
	CPU       int       `json:"cpu" db:"cpu"`
	MemoryMB  int       `json:"memory_mb" db:"memory_mb"`
	Image     string    `json:"image" db:"image"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// InstanceStatus constants
const (
	StatusRunning = "running"
	StatusStopped = "stopped"
)

// Metadata represents key-value metadata storage
type Metadata struct {
	ID        string    `json:"id" db:"id"`
	Path      string    `json:"path" db:"path"`
	Value     string    `json:"value" db:"value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateProjectRequest represents the request to create a project
type CreateProjectRequest struct {
	Name string `json:"name"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	Name string `json:"name"`
}

// CreateInstanceRequest represents the request to create an instance
type CreateInstanceRequest struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	CPU       int    `json:"cpu"`
	MemoryMB  int    `json:"memory_mb"`
	Image     string `json:"image"`
	Status    string `json:"status,omitempty"`
}

// UpdateInstanceRequest represents the request to update an instance
type UpdateInstanceRequest struct {
	Name     *string `json:"name,omitempty"`
	CPU      *int    `json:"cpu,omitempty"`
	MemoryMB *int    `json:"memory_mb,omitempty"`
	Image    *string `json:"image,omitempty"`
	Status   *string `json:"status,omitempty"`
}

// ProjectListOptions represents query options for listing projects
type ProjectListOptions struct {
	Name string
}

// InstanceListOptions represents query options for listing instances
type InstanceListOptions struct {
	ProjectID string
	Name      string
	Status    string
}

// CreateMetadataRequest represents the request to create metadata
type CreateMetadataRequest struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

// UpdateMetadataRequest represents the request to update metadata
type UpdateMetadataRequest struct {
	Path  *string `json:"path,omitempty"`
	Value *string `json:"value,omitempty"`
}

// MetadataListOptions represents query options for listing metadata
type MetadataListOptions struct {
	Prefix string
}