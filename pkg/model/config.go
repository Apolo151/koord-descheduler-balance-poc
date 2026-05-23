package model

type ImbalanceConfig struct {
	Resources      []string           `json:"resources" yaml:"resources"`
	Threshold      float64            `json:"threshold" yaml:"threshold"`
	Weights        map[string]float64 `json:"weights" yaml:"weights"`
	MinAllocatable map[string]float64 `json:"minAllocatable" yaml:"minAllocatable"`
}

type EvictionPolicy struct {
	ReservationFirst    bool              `json:"reservationFirst" yaml:"reservationFirst"`
	AllowNamespaces     []string          `json:"allowNamespaces" yaml:"allowNamespaces"`
	DenyNamespaces      []string          `json:"denyNamespaces" yaml:"denyNamespaces"`
	ProtectedLabels     map[string]string `json:"protectedLabels" yaml:"protectedLabels"`
	MaxEvictionsPerNode int               `json:"maxEvictionsPerNode" yaml:"maxEvictionsPerNode"`
}

type BinpackConfig struct {
	TargetWorkload string             `json:"targetWorkload" yaml:"targetWorkload"`
	Selector       map[string]string  `json:"selector" yaml:"selector"`
	Victims        int                `json:"victims" yaml:"victims"`
	TieBreaker     string             `json:"tieBreaker" yaml:"tieBreaker"`
	BalanceWeights map[string]float64 `json:"balanceWeights" yaml:"balanceWeights"`
}
