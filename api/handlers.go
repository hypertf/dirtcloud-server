package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nicolas/dirtcloud/domain"
	"github.com/nicolas/dirtcloud/service"
	"github.com/nicolas/dirtcloud/service/chaos"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	service      *service.Service
	chaosService *chaos.ChaosService
	token        string
}

// NewHandler creates a new HTTP handler
func NewHandler(svc *service.Service, chaosService *chaos.ChaosService, token string) *Handler {
	return &Handler{
		service:      svc,
		chaosService: chaosService,
		token:        token,
	}
}

// authenticate checks bearer token authentication
func (h *Handler) authenticate(r *http.Request) error {
	if h.token == "" {
		return nil // No authentication required
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return domain.UnauthorizedError("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return domain.UnauthorizedError("invalid authorization header format")
	}

	if parts[1] != h.token {
		return domain.UnauthorizedError("invalid token")
	}

	return nil
}

// writeError writes a domain error as JSON response
func (h *Handler) writeError(w http.ResponseWriter, err error) {
	var statusCode int
	var dirtErr *domain.DirtError

	if de, ok := err.(*domain.DirtError); ok {
		dirtErr = de
		switch de.Code {
		case domain.ErrorCodeNotFound:
			statusCode = http.StatusNotFound
		case domain.ErrorCodeAlreadyExists:
			statusCode = http.StatusConflict
		case domain.ErrorCodeInvalidInput:
			statusCode = http.StatusBadRequest
		case domain.ErrorCodeForeignKeyViolation:
			statusCode = http.StatusBadRequest
		case domain.ErrorCodeUnauthorized:
			statusCode = http.StatusUnauthorized
		case domain.ErrorCodeTooManyRequests:
			statusCode = http.StatusTooManyRequests
		case domain.ErrorCodeServiceUnavailable:
			statusCode = http.StatusServiceUnavailable
		default:
			statusCode = http.StatusInternalServerError
		}
	} else {
		statusCode = http.StatusInternalServerError
		dirtErr = domain.InternalError(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dirtErr)
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeText writes a plain text response
func (h *Handler) writeText(w http.ResponseWriter, statusCode int, text string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(text))
}

// Project handlers

// CreateProject handles POST /v1/projects
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyProjectsChaos(r.Context(), r, "POST"); err != nil {
		h.writeError(w, err)
		return
	}

	var req domain.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}

	project, err := h.service.CreateProject(req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, project)
}

// GetProject handles GET /v1/projects/{id}
func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyProjectsChaos(r.Context(), r, "GET"); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	project, err := h.service.GetProject(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, project)
}

// ListProjects handles GET /v1/projects
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyProjectsChaos(r.Context(), r, "GET"); err != nil {
		h.writeError(w, err)
		return
	}

	opts := domain.ProjectListOptions{
		Name: r.URL.Query().Get("name"),
	}

	projects, err := h.service.ListProjects(opts)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, projects)
}

// UpdateProject handles PATCH /v1/projects/{id}
func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyProjectsChaos(r.Context(), r, "PATCH"); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var req domain.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}

	project, err := h.service.UpdateProject(id, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, project)
}

// DeleteProject handles DELETE /v1/projects/{id}
func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyProjectsChaos(r.Context(), r, "DELETE"); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	err := h.service.DeleteProject(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Instance handlers

// CreateInstance handles POST /v1/instances
func (h *Handler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyInstancesChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	var req domain.CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}

	instance, err := h.service.CreateInstance(req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, instance)
}

// GetInstance handles GET /v1/instances/{id}
func (h *Handler) GetInstance(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyInstancesChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	instance, err := h.service.GetInstance(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, instance)
}

// ListInstances handles GET /v1/instances
func (h *Handler) ListInstances(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyInstancesChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	opts := domain.InstanceListOptions{
		ProjectID: r.URL.Query().Get("project_id"),
		Name:      r.URL.Query().Get("name"),
		Status:    r.URL.Query().Get("status"),
	}

	instances, err := h.service.ListInstances(opts)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, instances)
}

// UpdateInstance handles PATCH /v1/instances/{id}
func (h *Handler) UpdateInstance(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyInstancesChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var req domain.UpdateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}

	instance, err := h.service.UpdateInstance(id, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, instance)
}

// DeleteInstance handles DELETE /v1/instances/{id}
func (h *Handler) DeleteInstance(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyInstancesChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	err := h.service.DeleteInstance(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Metadata handlers

// SetMetadata handles PUT /v1/metadata/{path+}
func (h *Handler) SetMetadata(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyMetadataChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	path := vars["path"]

	if path == "" {
		h.writeError(w, domain.InvalidInputError("metadata path cannot be empty", nil))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, domain.InvalidInputError("failed to read request body", nil))
		return
	}

	value := string(body)

	metadata, err := h.service.SetMetadata(path, value)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, metadata)
}

// GetMetadata handles GET /v1/metadata/{path+}
func (h *Handler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyMetadataChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	path := vars["path"]

	if path == "" {
		h.writeError(w, domain.InvalidInputError("metadata path cannot be empty", nil))
		return
	}

	value, err := h.service.GetMetadataValue(path)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeText(w, http.StatusOK, value)
}

// ListMetadata handles GET /v1/metadata with prefix query parameter
func (h *Handler) ListMetadata(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyMetadataChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	opts := domain.MetadataListOptions{
		Prefix: r.URL.Query().Get("prefix"),
	}

	paths, err := h.service.ListMetadata(opts)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, paths)
}

// DeleteMetadata handles DELETE /v1/metadata/{path+}
func (h *Handler) DeleteMetadata(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyMetadataChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	path := vars["path"]

	if path == "" {
		h.writeError(w, domain.InvalidInputError("metadata path cannot be empty", nil))
		return
	}

	err := h.service.DeleteMetadata(path)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}