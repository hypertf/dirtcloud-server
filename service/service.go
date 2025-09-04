package service

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"github.com/nicolas/dirtcloud/domain"
)

// Service provides business logic for DirtCloud operations
type Service struct {
	projectRepo  ProjectRepository
	instanceRepo InstanceRepository
	metadataRepo MetadataRepository
}

// ProjectRepository defines the interface for project data operations
type ProjectRepository interface {
	Create(project *domain.Project) error
	GetByID(id string) (*domain.Project, error)
	GetByName(name string) (*domain.Project, error)
	List(opts domain.ProjectListOptions) ([]*domain.Project, error)
	Update(id string, req domain.UpdateProjectRequest) (*domain.Project, error)
	Delete(id string) error
}

// InstanceRepository defines the interface for instance data operations
type InstanceRepository interface {
	Create(instance *domain.Instance) error
	GetByID(id string) (*domain.Instance, error)
	List(opts domain.InstanceListOptions) ([]*domain.Instance, error)
	Update(id string, req domain.UpdateInstanceRequest) (*domain.Instance, error)
	Delete(id string) error
}

// MetadataRepository defines the interface for metadata data operations
type MetadataRepository interface {
	Set(path, value string) (*domain.Metadata, error)
	Get(path string) (*domain.Metadata, error)
	List(opts domain.MetadataListOptions) ([]string, error)
	Delete(path string) error
}

// NewService creates a new service instance
func NewService(projectRepo ProjectRepository, instanceRepo InstanceRepository, metadataRepo MetadataRepository) *Service {
	return &Service{
		projectRepo:  projectRepo,
		instanceRepo: instanceRepo,
		metadataRepo: metadataRepo,
	}
}

// generateID generates a random hex ID
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// validateProjectName validates a project name
func validateProjectName(name string) error {
	if name == "" {
		return domain.InvalidInputError("project name cannot be empty", nil)
	}
	if len(name) > 255 {
		return domain.InvalidInputError("project name too long", map[string]interface{}{
			"max_length": 255,
			"actual":     len(name),
		})
	}
	// Simple alphanumeric + dash/underscore validation
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
		return domain.InvalidInputError("project name can only contain alphanumeric characters, dashes, and underscores", nil)
	}
	return nil
}

// validateInstanceName validates an instance name
func validateInstanceName(name string) error {
	if name == "" {
		return domain.InvalidInputError("instance name cannot be empty", nil)
	}
	if len(name) > 255 {
		return domain.InvalidInputError("instance name too long", map[string]interface{}{
			"max_length": 255,
			"actual":     len(name),
		})
	}
	// Simple alphanumeric + dash/underscore validation
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
		return domain.InvalidInputError("instance name can only contain alphanumeric characters, dashes, and underscores", nil)
	}
	return nil
}

// validateInstanceSpecs validates instance specifications
func validateInstanceSpecs(cpu int, memoryMB int, image string) error {
	if cpu <= 0 {
		return domain.InvalidInputError("CPU must be positive", map[string]interface{}{"cpu": cpu})
	}
	if cpu > 64 {
		return domain.InvalidInputError("CPU too high", map[string]interface{}{
			"max_cpu": 64,
			"actual":  cpu,
		})
	}
	if memoryMB <= 0 {
		return domain.InvalidInputError("memory must be positive", map[string]interface{}{"memory_mb": memoryMB})
	}
	if memoryMB > 512*1024 { // 512GB
		return domain.InvalidInputError("memory too high", map[string]interface{}{
			"max_memory_mb": 512 * 1024,
			"actual":        memoryMB,
		})
	}
	if image == "" {
		return domain.InvalidInputError("image cannot be empty", nil)
	}
	if len(image) > 255 {
		return domain.InvalidInputError("image name too long", map[string]interface{}{
			"max_length": 255,
			"actual":     len(image),
		})
	}
	return nil
}

// validateInstanceStatus validates instance status
func validateInstanceStatus(status string) error {
	if status != domain.StatusRunning && status != domain.StatusStopped {
		return domain.InvalidInputError("invalid status", map[string]interface{}{
			"valid_statuses": []string{domain.StatusRunning, domain.StatusStopped},
			"actual":         status,
		})
	}
	return nil
}

// Project operations

// CreateProject creates a new project
func (s *Service) CreateProject(req domain.CreateProjectRequest) (*domain.Project, error) {
	if err := validateProjectName(req.Name); err != nil {
		return nil, err
	}

	id, err := generateID()
	if err != nil {
		return nil, domain.InternalError("failed to generate ID")
	}

	project := &domain.Project{
		ID:   id,
		Name: req.Name,
	}

	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject retrieves a project by ID
func (s *Service) GetProject(id string) (*domain.Project, error) {
	return s.projectRepo.GetByID(id)
}

// ListProjects lists projects with optional filtering
func (s *Service) ListProjects(opts domain.ProjectListOptions) ([]*domain.Project, error) {
	return s.projectRepo.List(opts)
}

// UpdateProject updates an existing project
func (s *Service) UpdateProject(id string, req domain.UpdateProjectRequest) (*domain.Project, error) {
	if err := validateProjectName(req.Name); err != nil {
		return nil, err
	}

	return s.projectRepo.Update(id, req)
}

// DeleteProject deletes a project
func (s *Service) DeleteProject(id string) error {
	return s.projectRepo.Delete(id)
}

// Instance operations

// CreateInstance creates a new instance
func (s *Service) CreateInstance(req domain.CreateInstanceRequest) (*domain.Instance, error) {
	if err := validateInstanceName(req.Name); err != nil {
		return nil, err
	}

	if err := validateInstanceSpecs(req.CPU, req.MemoryMB, req.Image); err != nil {
		return nil, err
	}

	status := req.Status
	if status == "" {
		status = domain.StatusRunning
	}
	if err := validateInstanceStatus(status); err != nil {
		return nil, err
	}

	// Verify project exists
	_, err := s.projectRepo.GetByID(req.ProjectID)
	if err != nil {
		if domain.IsNotFound(err) {
			return nil, domain.ForeignKeyViolationError("project", "id", req.ProjectID)
		}
		return nil, err
	}

	id, err := generateID()
	if err != nil {
		return nil, domain.InternalError("failed to generate ID")
	}

	instance := &domain.Instance{
		ID:        id,
		ProjectID: req.ProjectID,
		Name:      req.Name,
		CPU:       req.CPU,
		MemoryMB:  req.MemoryMB,
		Image:     req.Image,
		Status:    status,
	}

	if err := s.instanceRepo.Create(instance); err != nil {
		return nil, err
	}

	return instance, nil
}

// GetInstance retrieves an instance by ID
func (s *Service) GetInstance(id string) (*domain.Instance, error) {
	return s.instanceRepo.GetByID(id)
}

// ListInstances lists instances with optional filtering
func (s *Service) ListInstances(opts domain.InstanceListOptions) ([]*domain.Instance, error) {
	return s.instanceRepo.List(opts)
}

// UpdateInstance updates an existing instance
func (s *Service) UpdateInstance(id string, req domain.UpdateInstanceRequest) (*domain.Instance, error) {
	if req.Name != nil {
		if err := validateInstanceName(*req.Name); err != nil {
			return nil, err
		}
	}

	if req.CPU != nil || req.MemoryMB != nil {
		// Get current instance to validate complete specs
		current, err := s.instanceRepo.GetByID(id)
		if err != nil {
			return nil, err
		}

		cpu := current.CPU
		memory := current.MemoryMB
		image := current.Image

		if req.CPU != nil {
			cpu = *req.CPU
		}
		if req.MemoryMB != nil {
			memory = *req.MemoryMB
		}
		if req.Image != nil {
			image = *req.Image
		}

		if err := validateInstanceSpecs(cpu, memory, image); err != nil {
			return nil, err
		}
	}

	if req.Status != nil {
		if err := validateInstanceStatus(*req.Status); err != nil {
			return nil, err
		}
	}

	return s.instanceRepo.Update(id, req)
}

// DeleteInstance deletes an instance
func (s *Service) DeleteInstance(id string) error {
	return s.instanceRepo.Delete(id)
}

// Metadata operations

// SetMetadata creates or updates metadata
func (s *Service) SetMetadata(path, value string) (*domain.Metadata, error) {
	if path == "" {
		return nil, domain.InvalidInputError("metadata path cannot be empty", nil)
	}

	return s.metadataRepo.Set(path, value)
}

// GetMetadata retrieves metadata by path
func (s *Service) GetMetadata(path string) (*domain.Metadata, error) {
	if path == "" {
		return nil, domain.InvalidInputError("metadata path cannot be empty", nil)
	}

	return s.metadataRepo.Get(path)
}

// ListMetadata lists metadata paths with optional prefix filtering
func (s *Service) ListMetadata(opts domain.MetadataListOptions) ([]string, error) {
	return s.metadataRepo.List(opts)
}

// DeleteMetadata deletes metadata by path
func (s *Service) DeleteMetadata(path string) error {
	if path == "" {
		return domain.InvalidInputError("metadata path cannot be empty", nil)
	}

	return s.metadataRepo.Delete(path)
}

// GetMetadataValue retrieves just the value from metadata
func (s *Service) GetMetadataValue(path string) (string, error) {
	metadata, err := s.GetMetadata(path)
	if err != nil {
		return "", err
	}
	return metadata.Value, nil
}