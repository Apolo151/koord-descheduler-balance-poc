package binpack

import "koord-descheduler-balance-poc/pkg/model"

func ScoreCandidateRemoval(snapshot *model.ClusterSnapshot, cfg model.BinpackConfig, podID, nodeName string, selected map[string]struct{}) float64 {
	if snapshot == nil {
		return 0
	}

	before := float64(countNodesWithTargetPods(snapshot, cfg, selected))
	exclude := make(map[string]struct{}, len(selected)+1)
	for id := range selected {
		exclude[id] = struct{}{}
	}
	if podID != "" {
		exclude[podID] = struct{}{}
	}
	_ = nodeName
	after := float64(countNodesWithTargetPods(snapshot, cfg, exclude))
	return before - after
}
