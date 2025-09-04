# DirtCloud Web Console

A simple web-based interface for managing DirtCloud resources using HTMX for dynamic interactions.

## Features

The web console provides BREAD (Browse, Read, Edit, Add, Delete) operations for all DirtCloud resources:

### Projects
- **Browse**: View all projects in a table format
- **Read**: View project details 
- **Edit**: Update project name
- **Add**: Create new projects
- **Delete**: Remove projects (with validation - cannot delete projects with existing instances)

### Instances
- **Browse**: View all instances with their specifications
- **Read**: View instance details including CPU, memory, image, and status
- **Edit**: Update instance configuration
- **Add**: Create new instances (requires selecting a project)
- **Delete**: Remove instances

### Metadata
- **Browse**: View all metadata key-value pairs with optional prefix filtering
- **Read**: View metadata values
- **Edit**: Update metadata values (path is read-only)
- **Add**: Create new metadata entries
- **Delete**: Remove metadata entries

## Access

The web console is available at:
- **Dashboard**: `http://localhost:8080/web/`
- **Projects**: `http://localhost:8080/web/projects`
- **Instances**: `http://localhost:8080/web/instances` 
- **Metadata**: `http://localhost:8080/web/metadata`

## Technology

- **Backend**: Go with Gorilla Mux router
- **Frontend**: HTML with HTMX for dynamic interactions
- **Styling**: Embedded CSS with Bootstrap-inspired classes
- **Modals**: JavaScript-based modal dialogs for forms

## Features

- **Real-time updates**: HTMX provides seamless updates without page refreshes
- **Form validation**: Client-side and server-side validation
- **Confirmation dialogs**: Prevents accidental deletions
- **Business logic enforcement**: Respects API constraints (e.g., cannot delete projects with instances)
- **Responsive design**: Works on desktop and mobile devices

## Usage

1. Navigate to `http://localhost:8080/web/` in your browser
2. Use the navigation links to switch between resource types
3. Click "New [Resource]" to create resources
4. Click "Edit" to modify existing resources
5. Click "Delete" to remove resources (with confirmation)

The interface updates dynamically using HTMX, providing a smooth single-page application experience.