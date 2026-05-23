package descheduler

import (
	"sort"

	"koord-descheduler-balance-poc/pkg/imbalance"
	"koord-descheduler-balance-poc/pkg/model"
)

func PlanEvictions(snapshot *model.ClusterSnapshot, cfg model.ImbalanceConfig, policy model.EvictionPolicy) []model.EvictionPlan {
	if snapshot == nil || snapshot.Nodes == nil {
		return nil
	}
	plans := make([]model.EvictionPlan, 0)

	nodeNames := make([]string, 0, len(snapshot.Nodes))
	for name := range snapshot.Nodes {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	for _, nodeName := range nodeNames {
		node := snapshot.Nodes[nodeName]
		if node == nil {
			continue
		}
		score := imbalance.ScoreNode(node, snapshot, cfg)
		if score.StdDev < cfg.Threshold {
			continue
		}

		evictable := FilterEvictablePods(node, snapshot, policy)
		candidates := make([]model.EvictionCandidate, 0, len(evictable))
		for _, pod := range evictable {
			delta, finalStd := imbalance.DeltaStdDevForPod(node, pod, snapshot, cfg)
			candidate := model.EvictionCandidate{
				PodID:       model.PodKey(pod.Namespace, pod.Name),
				NodeName:    node.Name,
				DeltaStdDev: delta,
				FinalStdDev: finalStd,
				PolicyScore: ApplyReservationFirstScore(pod, policy),
			}
			candidates = append(candidates, candidate)
		}

		imbalance.SortCandidates(candidates)
		chosen := candidates
		if policy.MaxEvictionsPerNode > 0 && len(chosen) > policy.MaxEvictionsPerNode {
			chosen = chosen[:policy.MaxEvictionsPerNode]
		}

		plans = append(plans, model.EvictionPlan{
			NodeName:   node.Name,
			Candidates: candidates,
			Chosen:     chosen,
		})
	}

	return plans
}
