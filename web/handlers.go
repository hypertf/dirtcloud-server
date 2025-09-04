package web

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nicolas/dirtcloud/domain"
	"github.com/nicolas/dirtcloud/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

// Dashboard shows the main dashboard
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>DirtCloud Console</title>
    <script src="https://unpkg.com/htmx.org@1.9.6"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .nav { margin-bottom: 20px; }
        .nav a { margin-right: 20px; text-decoration: none; color: #007bff; }
        .nav a:hover { text-decoration: underline; }
        .content { margin-top: 20px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .btn { padding: 8px 16px; margin: 4px; background: #007bff; color: white; border: none; cursor: pointer; }
        .btn:hover { background: #0056b3; }
        .btn-danger { background: #dc3545; }
        .btn-danger:hover { background: #c82333; }
        .form-group { margin: 10px 0; }
        .form-group label { display: block; margin-bottom: 5px; }
        .form-group input, .form-group select { width: 100%; padding: 8px; border: 1px solid #ddd; }
        .modal { display: none; position: fixed; z-index: 1; left: 0; top: 0; width: 100%; height: 100%; background-color: rgba(0,0,0,0.4); }
        .modal-content { background-color: #fefefe; margin: 15% auto; padding: 20px; border: 1px solid #888; width: 50%; }
        .close { color: #aaa; float: right; font-size: 28px; font-weight: bold; cursor: pointer; }
        .close:hover { color: black; }
    </style>
</head>
<body>
    <h1>DirtCloud Console</h1>
    <div class="nav">
        <a href="#" hx-get="/web/projects" hx-target="#content">Projects</a>
        <a href="#" hx-get="/web/instances" hx-target="#content">Instances</a>
        <a href="#" hx-get="/web/metadata" hx-target="#content">Metadata</a>
    </div>
    <div id="content" class="content">
        <p>Welcome to DirtCloud Console. Select a resource type from the navigation above.</p>
    </div>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}

// Projects handlers
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(domain.ProjectListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
<div>
    <h2>Projects</h2>
    <button class="btn" hx-get="/web/projects/new" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">New Project</button>
    <table>
        <thead>
            <tr>
                <th>ID</th>
                <th>Name</th>
                <th>Created At</th>
                <th>Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range .}}
            <tr>
                <td>{{.ID}}</td>
                <td>{{.Name}}</td>
                <td>{{.CreatedAt.Format "2006-01-02 15:04:05"}}</td>
                <td>
                    <button class="btn" hx-get="/web/projects/{{.ID}}/edit" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">Edit</button>
                    <button class="btn btn-danger" hx-delete="/web/projects/{{.ID}}" hx-target="closest tr" hx-confirm="Are you sure?">Delete</button>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>

<!-- Modal -->
<div id="modal" class="modal">
    <div class="modal-content">
        <span class="close" onclick="document.getElementById('modal').style.display='none'">&times;</span>
        <div id="modal-content"></div>
    </div>
</div>
`

	t := template.Must(template.New("projects").Parse(tmpl))
	if err := t.Execute(w, projects); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) NewProjectForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<h3>New Project</h3>
<form hx-post="/web/projects" hx-target="#content" hx-on-success="document.getElementById('modal').style.display='none'">
    <div class="form-group">
        <label for="name">Name:</label>
        <input type="text" id="name" name="name" required>
    </div>
    <button type="submit" class="btn">Create</button>
    <button type="button" class="btn" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
</form>
`
	w.Write([]byte(tmpl))
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req := domain.CreateProjectRequest{
		Name: r.FormValue("name"),
	}

	_, err := h.service.CreateProject(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated projects list
	h.ListProjects(w, r)
}

func (h *Handler) EditProjectForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	project, err := h.service.GetProject(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	tmpl := `
<h3>Edit Project</h3>
<form hx-put="/web/projects/{{.ID}}" hx-target="#content" hx-on-success="document.getElementById('modal').style.display='none'">
    <div class="form-group">
        <label for="name">Name:</label>
        <input type="text" id="name" name="name" value="{{.Name}}" required>
    </div>
    <button type="submit" class="btn">Update</button>
    <button type="button" class="btn" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
</form>
`
	t := template.Must(template.New("edit-project").Parse(tmpl))
	if err := t.Execute(w, project); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req := domain.UpdateProjectRequest{
		Name: r.FormValue("name"),
	}

	_, err := h.service.UpdateProject(id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated projects list
	h.ListProjects(w, r)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteProject(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Instances handlers
func (h *Handler) ListInstances(w http.ResponseWriter, r *http.Request) {
	instances, err := h.service.ListInstances(domain.InstanceListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get projects for the dropdown
	projects, err := h.service.ListProjects(domain.ProjectListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
<div>
    <h2>Instances</h2>
    <button class="btn" hx-get="/web/instances/new" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">New Instance</button>
    <table>
        <thead>
            <tr>
                <th>ID</th>
                <th>Project ID</th>
                <th>Name</th>
                <th>CPU</th>
                <th>Memory (MB)</th>
                <th>Image</th>
                <th>Status</th>
                <th>Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range .Instances}}
            <tr>
                <td>{{.ID}}</td>
                <td>{{.ProjectID}}</td>
                <td>{{.Name}}</td>
                <td>{{.CPU}}</td>
                <td>{{.MemoryMB}}</td>
                <td>{{.Image}}</td>
                <td>{{.Status}}</td>
                <td>
                    <button class="btn" hx-get="/web/instances/{{.ID}}/edit" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">Edit</button>
                    <button class="btn btn-danger" hx-delete="/web/instances/{{.ID}}" hx-target="closest tr" hx-confirm="Are you sure?">Delete</button>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>

<!-- Modal -->
<div id="modal" class="modal">
    <div class="modal-content">
        <span class="close" onclick="document.getElementById('modal').style.display='none'">&times;</span>
        <div id="modal-content"></div>
    </div>
</div>
`

	data := struct {
		Instances []*domain.Instance
		Projects  []*domain.Project
	}{
		Instances: instances,
		Projects:  projects,
	}

	t := template.Must(template.New("instances").Parse(tmpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) NewInstanceForm(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(domain.ProjectListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
<h3>New Instance</h3>
<form hx-post="/web/instances" hx-target="#content" hx-on-success="document.getElementById('modal').style.display='none'">
    <div class="form-group">
        <label for="project_id">Project:</label>
        <select id="project_id" name="project_id" required>
            <option value="">Select a project</option>
            {{range .}}
            <option value="{{.ID}}">{{.Name}}</option>
            {{end}}
        </select>
    </div>
    <div class="form-group">
        <label for="name">Name:</label>
        <input type="text" id="name" name="name" required>
    </div>
    <div class="form-group">
        <label for="cpu">CPU:</label>
        <input type="number" id="cpu" name="cpu" value="1" required>
    </div>
    <div class="form-group">
        <label for="memory_mb">Memory (MB):</label>
        <input type="number" id="memory_mb" name="memory_mb" value="512" required>
    </div>
    <div class="form-group">
        <label for="image">Image:</label>
        <input type="text" id="image" name="image" value="ubuntu:20.04" required>
    </div>
    <div class="form-group">
        <label for="status">Status:</label>
        <select id="status" name="status">
            <option value="running">Running</option>
            <option value="stopped">Stopped</option>
        </select>
    </div>
    <button type="submit" class="btn">Create</button>
    <button type="button" class="btn" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
</form>
`
	t := template.Must(template.New("new-instance").Parse(tmpl))
	if err := t.Execute(w, projects); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cpu, _ := strconv.Atoi(r.FormValue("cpu"))
	memoryMB, _ := strconv.Atoi(r.FormValue("memory_mb"))

	req := domain.CreateInstanceRequest{
		ProjectID: r.FormValue("project_id"),
		Name:      r.FormValue("name"),
		CPU:       cpu,
		MemoryMB:  memoryMB,
		Image:     r.FormValue("image"),
		Status:    r.FormValue("status"),
	}

	_, err := h.service.CreateInstance(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated instances list
	h.ListInstances(w, r)
}

func (h *Handler) EditInstanceForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	instance, err := h.service.GetInstance(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	projects, err := h.service.ListProjects(domain.ProjectListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
<h3>Edit Instance</h3>
<form hx-put="/web/instances/{{.Instance.ID}}" hx-target="#content" hx-on-success="document.getElementById('modal').style.display='none'">
    <div class="form-group">
        <label for="project_id">Project:</label>
        <select id="project_id" name="project_id" required>
            {{range .Projects}}
            <option value="{{.ID}}" {{if eq .ID $.Instance.ProjectID}}selected{{end}}>{{.Name}}</option>
            {{end}}
        </select>
    </div>
    <div class="form-group">
        <label for="name">Name:</label>
        <input type="text" id="name" name="name" value="{{.Instance.Name}}" required>
    </div>
    <div class="form-group">
        <label for="cpu">CPU:</label>
        <input type="number" id="cpu" name="cpu" value="{{.Instance.CPU}}" required>
    </div>
    <div class="form-group">
        <label for="memory_mb">Memory (MB):</label>
        <input type="number" id="memory_mb" name="memory_mb" value="{{.Instance.MemoryMB}}" required>
    </div>
    <div class="form-group">
        <label for="image">Image:</label>
        <input type="text" id="image" name="image" value="{{.Instance.Image}}" required>
    </div>
    <div class="form-group">
        <label for="status">Status:</label>
        <select id="status" name="status">
            <option value="running" {{if eq .Instance.Status "running"}}selected{{end}}>Running</option>
            <option value="stopped" {{if eq .Instance.Status "stopped"}}selected{{end}}>Stopped</option>
        </select>
    </div>
    <button type="submit" class="btn">Update</button>
    <button type="button" class="btn" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
</form>
`

	data := struct {
		Instance domain.Instance
		Projects []*domain.Project
	}{
		Instance: *instance,
		Projects: projects,
	}

	t := template.Must(template.New("edit-instance").Parse(tmpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	cpuStr := r.FormValue("cpu")
	cpu, _ := strconv.Atoi(cpuStr)
	memoryMBStr := r.FormValue("memory_mb")
	memoryMB, _ := strconv.Atoi(memoryMBStr)
	image := r.FormValue("image")
	status := r.FormValue("status")

	req := domain.UpdateInstanceRequest{
		Name:     &name,
		CPU:      &cpu,
		MemoryMB: &memoryMB,
		Image:    &image,
		Status:   &status,
	}

	_, err := h.service.UpdateInstance(id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated instances list
	h.ListInstances(w, r)
}

func (h *Handler) DeleteInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteInstance(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Metadata handlers
func (h *Handler) ListMetadata(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	paths, err := h.service.ListMetadata(domain.MetadataListOptions{Prefix: prefix})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get full metadata objects for each path
	var metadata []domain.Metadata
	for _, path := range paths {
		meta, err := h.service.GetMetadata(path)
		if err != nil {
			continue // Skip if metadata was deleted between list and get
		}
		metadata = append(metadata, *meta)
	}

	tmpl := `
<div>
    <h2>Metadata</h2>
    <div class="form-group">
        <label for="prefix-filter">Filter by prefix:</label>
        <input type="text" id="prefix-filter" name="prefix" hx-get="/web/metadata" hx-target="#content" hx-trigger="input changed delay:500ms" value="{{.Prefix}}">
    </div>
    <button class="btn" hx-get="/web/metadata/new" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">New Metadata</button>
    <table>
        <thead>
            <tr>
                <th>Path</th>
                <th>Value</th>
                <th>Updated At</th>
                <th>Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range .Metadata}}
            <tr>
                <td>{{.Path}}</td>
                <td>{{.Value}}</td>
                <td>{{.UpdatedAt.Format "2006-01-02 15:04:05"}}</td>
                <td>
                    <button class="btn" hx-get="/web/metadata/edit?path={{.Path}}" hx-target="#modal-content" onclick="document.getElementById('modal').style.display='block'">Edit</button>
                    <button class="btn btn-danger" hx-delete="/web/metadata/delete?path={{.Path}}" hx-target="closest tr" hx-confirm="Are you sure?">Delete</button>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>

<!-- Modal -->
<div id="modal" class="modal">
    <div class="modal-content">
        <span class="close" onclick="document.getElementById('modal').style.display='none'">&times;</span>
        <div id="modal-content"></div>
    </div>
</div>
`

	data := struct {
		Metadata []domain.Metadata
		Prefix   string
	}{
		Metadata: metadata,
		Prefix:   prefix,
	}

	t := template.Must(template.New("metadata").Parse(tmpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) NewMetadataForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<h3>New Metadata</h3>
<form hx-post="/web/metadata" hx-target="#content" hx-on-success="document.getElementById('modal').style.display='none'">
    <div class="form-group">
        <label for="path">Path:</label>
        <input type="text" id="path" name="path" required>
    </div>
    <div class="form-group">
        <label for="value">Value:</label>
        <textarea id="value" name="value" rows="4" style="width: 100%; padding: 8px; border: 1px solid #ddd;" required></textarea>
    </div>
    <button type="submit" class="btn">Create</button>
    <button type="button" class="btn" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
</form>
`
	w.Write([]byte(tmpl))
}

func (h *Handler) CreateMetadata(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	path := r.FormValue("path")
	value := r.FormValue("value")

	if _, err := h.service.SetMetadata(path, value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated metadata list
	h.ListMetadata(w, r)
}

func (h *Handler) EditMetadataForm(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}

	metadata, err := h.service.GetMetadata(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	tmpl := `
<h3>Edit Metadata</h3>
<form hx-put="/web/metadata/update" hx-target="#content" hx-on-success="document.getElementById('modal').style.display='none'">
    <div class="form-group">
        <label for="path">Path:</label>
        <input type="text" id="path" name="path" value="{{.Path}}" readonly>
    </div>
    <div class="form-group">
        <label for="value">Value:</label>
        <textarea id="value" name="value" rows="4" style="width: 100%; padding: 8px; border: 1px solid #ddd;" required>{{.Value}}</textarea>
    </div>
    <button type="submit" class="btn">Update</button>
    <button type="button" class="btn" onclick="document.getElementById('modal').style.display='none'">Cancel</button>
</form>
`
	t := template.Must(template.New("edit-metadata").Parse(tmpl))
	if err := t.Execute(w, metadata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	path := r.FormValue("path")
	value := r.FormValue("value")

	if _, err := h.service.SetMetadata(path, value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated metadata list
	h.ListMetadata(w, r)
}

func (h *Handler) DeleteMetadata(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteMetadata(path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}