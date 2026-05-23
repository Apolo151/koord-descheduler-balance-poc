package model

type Pod struct {
	Name          string            `json:"name" yaml:"name"`
	Namespace     string            `json:"namespace" yaml:"namespace"`
	Owner         string            `json:"owner" yaml:"owner"`
	Resources     ResourceVector    `json:"resources" yaml:"resources"`
	Labels        map[string]string `json:"labels" yaml:"labels"`
	Evictable     bool              `json:"evictable" yaml:"evictable"`
	PriorityClass string            `json:"priorityClass" yaml:"priorityClass"`
}
