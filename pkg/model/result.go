package model

type NodeImbalanceScore struct {
	NodeName string             `json:"nodeName" yaml:"nodeName"`
	Ratios   map[string]float64 `json:"ratios" yaml:"ratios"`
	Mean     float64            `json:"mean" yaml:"mean"`
	StdDev   float64            `json:"stdDev" yaml:"stdDev"`
}

type EvictionCandidate struct {
	PodID       string  `json:"podId" yaml:"podId"`
	NodeName    string  `json:"nodeName" yaml:"nodeName"`
	DeltaStdDev float64 `json:"deltaStdDev" yaml:"deltaStdDev"`
	FinalStdDev float64 `json:"finalStdDev" yaml:"finalStdDev"`
	PolicyScore float64 `json:"policyScore" yaml:"policyScore"`
}

type EvictionPlan struct {
	NodeName   string              `json:"nodeName" yaml:"nodeName"`
	Candidates []EvictionCandidate `json:"candidates" yaml:"candidates"`
	Chosen     []EvictionCandidate `json:"chosen" yaml:"chosen"`
}

type BinpackDecision struct {
	Victims     []string `json:"victims" yaml:"victims"`
	ScoreBefore float64  `json:"scoreBefore" yaml:"scoreBefore"`
	ScoreAfter  float64  `json:"scoreAfter" yaml:"scoreAfter"`
	NodeSet     []string `json:"nodeSet" yaml:"nodeSet"`
}
