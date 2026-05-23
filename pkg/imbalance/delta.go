package imbalance

import "koord-descheduler-balance-poc/pkg/model"

func DeltaStdDevForPod(node *model.Node, pod *model.Pod, snapshot *model.ClusterSnapshot, cfg model.ImbalanceConfig) (float64, float64) {
	if node == nil || pod == nil || snapshot == nil {
		return 0, 0
	}
	current := ScoreNode(node, snapshot, cfg)
	requested := snapshot.RequestedOnNode(node.Name)
	hypothetical := requested.Sub(pod.Resources)
	ratios, used := ComputeRatios(node.Allocatable, hypothetical, cfg.Resources)
	_, hypotheticalStd := MeanStdDev(ratios, used)
	delta := current.StdDev - hypotheticalStd
	return delta, hypotheticalStd
}
