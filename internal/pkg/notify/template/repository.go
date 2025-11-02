package template

import (
	"context"
	"fmt"
)

// ITemplateRepository defines the interface for template storage
type ITemplateRepository interface {
	// Create creates a new template
	Create(ctx context.Context, template *Template) error

	// Get retrieves a template by ID
	Get(ctx context.Context, id string) (*Template, error)

	// GetByNameAndType retrieves a template by name and type
	GetByNameAndType(ctx context.Context, name string, templateType TemplateType) (*Template, error)

	// List lists all templates with optional filtering
	List(ctx context.Context, filter *TemplateFilter) ([]*Template, error)

	// Update updates an existing template
	Update(ctx context.Context, template *Template) error

	// Delete deletes a template by ID
	Delete(ctx context.Context, id string) error

	// ListByType lists templates by type
	ListByType(ctx context.Context, templateType TemplateType) ([]*Template, error)

	// ListByChannel lists templates by channel
	ListByChannel(ctx context.Context, channel string) ([]*Template, error)
}

// TemplateFilter represents filter criteria for listing templates
type TemplateFilter struct {
	Type    TemplateType
	Channel string
	Name    string
	Limit   int
	Offset  int
}

// InMemoryTemplateRepository implements ITemplateRepository using in-memory storage
// This is for testing and development. Use a database implementation in production.
type InMemoryTemplateRepository struct {
	templates map[string]*Template
}

// NewInMemoryTemplateRepository creates a new in-memory template repository
func NewInMemoryTemplateRepository() *InMemoryTemplateRepository {
	return &InMemoryTemplateRepository{
		templates: make(map[string]*Template),
	}
}

// Create creates a new template
func (r *InMemoryTemplateRepository) Create(ctx context.Context, template *Template) error {
	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}

	if _, exists := r.templates[template.ID]; exists {
		return fmt.Errorf("template with ID %s already exists", template.ID)
	}

	r.templates[template.ID] = template
	return nil
}

// Get retrieves a template by ID
func (r *InMemoryTemplateRepository) Get(ctx context.Context, id string) (*Template, error) {
	template, exists := r.templates[id]
	if !exists {
		return nil, fmt.Errorf("template with ID %s not found", id)
	}
	return template, nil
}

// GetByNameAndType retrieves a template by name and type
func (r *InMemoryTemplateRepository) GetByNameAndType(ctx context.Context, name string, templateType TemplateType) (*Template, error) {
	for _, tmpl := range r.templates {
		if tmpl.Name == name && tmpl.Type == templateType {
			return tmpl, nil
		}
	}
	return nil, fmt.Errorf("template with name %s and type %s not found", name, templateType)
}

// List lists all templates with optional filtering
func (r *InMemoryTemplateRepository) List(ctx context.Context, filter *TemplateFilter) ([]*Template, error) {
	var result []*Template

	for _, tmpl := range r.templates {
		if filter != nil {
			if filter.Type != "" && tmpl.Type != filter.Type {
				continue
			}
			if filter.Channel != "" && tmpl.Channel != filter.Channel {
				continue
			}
			if filter.Name != "" && tmpl.Name != filter.Name {
				continue
			}
		}
		result = append(result, tmpl)
	}

	// Apply pagination
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		end := start + filter.Limit
		if start >= len(result) {
			return []*Template{}, nil
		}
		if end > len(result) {
			end = len(result)
		}
		result = result[start:end]
	}

	return result, nil
}

// Update updates an existing template
func (r *InMemoryTemplateRepository) Update(ctx context.Context, template *Template) error {
	if _, exists := r.templates[template.ID]; !exists {
		return fmt.Errorf("template with ID %s not found", template.ID)
	}

	r.templates[template.ID] = template
	return nil
}

// Delete deletes a template by ID
func (r *InMemoryTemplateRepository) Delete(ctx context.Context, id string) error {
	if _, exists := r.templates[id]; !exists {
		return fmt.Errorf("template with ID %s not found", id)
	}

	delete(r.templates, id)
	return nil
}

// ListByType lists templates by type
func (r *InMemoryTemplateRepository) ListByType(ctx context.Context, templateType TemplateType) ([]*Template, error) {
	return r.List(ctx, &TemplateFilter{Type: templateType})
}

// ListByChannel lists templates by channel
func (r *InMemoryTemplateRepository) ListByChannel(ctx context.Context, channel string) ([]*Template, error) {
	return r.List(ctx, &TemplateFilter{Channel: channel})
}
