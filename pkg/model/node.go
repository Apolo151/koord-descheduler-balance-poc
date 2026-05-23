package model

type Node struct {
	Name        string            `json:"name" yaml:"name"`
	Allocatable ResourceVector    `json:"allocatable" yaml:"allocatable"`
	Pods        []string          `json:"pods" yaml:"pods"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Taints      []string          `json:"taints" yaml:"taints"`
}
