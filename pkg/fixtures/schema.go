package fixtures

import "koord-descheduler-balance-poc/pkg/model"

type Scenario struct {
	Metadata        map[string]string     `json:"metadata" yaml:"metadata"`
	Cluster         Cluster               `json:"cluster" yaml:"cluster"`
	ImbalanceConfig model.ImbalanceConfig `json:"imbalanceConfig" yaml:"imbalanceConfig"`
	EvictionPolicy  model.EvictionPolicy  `json:"evictionPolicy" yaml:"evictionPolicy"`
	BinpackConfig   model.BinpackConfig   `json:"binpackConfig" yaml:"binpackConfig"`
	Events          Events                `json:"events" yaml:"events"`
}

type Cluster struct {
	Nodes map[string]*model.Node `json:"nodes" yaml:"nodes"`
	Pods  map[string]*model.Pod  `json:"pods" yaml:"pods"`
}

type Events struct {
	ScaleDown *ScaleDownEvent `json:"scaleDown" yaml:"scaleDown"`
}

type ScaleDownEvent struct {
	TargetWorkload string            `json:"targetWorkload" yaml:"targetWorkload"`
	Selector       map[string]string `json:"selector" yaml:"selector"`
	Victims        int               `json:"victims" yaml:"victims"`
}
