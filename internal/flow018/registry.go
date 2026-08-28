package flow018

import (
	"sort"

	"campusqa/internal/model"
)

type Registry struct {
	services map[string]*Service
}

func NewRegistry() *Registry {
	return &Registry{services: make(map[string]*Service)}
}

func (r *Registry) Register(name string, service *Service) bool {
	if name == "" || service == nil {
		return false
	}
	if _, exists := r.services[name]; exists {
		return false
	}
	r.services[name] = service
	return true
}

func (r *Registry) Get(name string) (*Service, bool) {
	service, ok := r.services[name]
	return service, ok
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) Health() map[string]string {
	result := make(map[string]string)
	for name, service := range r.services {
		if service == nil || service.store == nil {
			result[name] = "unavailable"
		} else {
			result[name] = "ready"
		}
	}
	return result
}

func CompletedWorkflow(workflow model.Workflow, step string) model.Workflow {
	for _, existing := range workflow.Completed {
		if existing == step {
			return workflow
		}
	}
	for _, planned := range workflow.Steps {
		if planned == step {
			workflow.Completed = append(workflow.Completed, step)
			break
		}
	}
	workflow.Ready = len(workflow.Completed) == len(workflow.Steps)
	return workflow
}
