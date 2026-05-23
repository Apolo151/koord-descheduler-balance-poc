package binpack

import (
	"sort"

	"koord-descheduler-balance-poc/pkg/model"
)

func SelectVictims(snapshot *model.ClusterSnapshot, cfg model.BinpackConfig) model.BinpackDecision {
	decision := model.BinpackDecision{
		Victims:     []string{},
		ScoreBefore: 0,
		ScoreAfter:  0,
		NodeSet:     []string{},
	}
	if snapshot == nil || cfg.Victims <= 0 {
		return decision
	}
	if snapshot.Nodes == nil || snapshot.Pods == nil {
		return decision
	}

	candidates := collectTargetPods(snapshot, cfg)
	if len(candidates) == 0 {
		return decision
	}

	podNode := buildPodNodeIndex(snapshot)
	before := float64(countNodesWithTargetPods(snapshot, cfg, nil))
	decision.ScoreBefore = before

	selected := make(map[string]struct{})
	for len(decision.Victims) < cfg.Victims {
		bestScore := -1.0
		bestPodID := ""
		bestNode := ""
		for _, pod := range candidates {
			podID := model.PodKey(pod.Namespace, pod.Name)
			if _, chosen := selected[podID]; chosen {
				continue
			}
			nodeName := podNode[podID]
			score := ScoreCandidateRemoval(snapshot, cfg, podID, nodeName, selected)
			if score > bestScore {
				bestScore = score
				bestPodID = podID
				bestNode = nodeName
				continue
			}
			if score == bestScore {
				if nodeName < bestNode || (nodeName == bestNode && podID < bestPodID) {
					bestPodID = podID
					bestNode = nodeName
				}
			}
		}

		if bestPodID == "" {
			break
		}
		decision.Victims = append(decision.Victims, bestPodID)
		selected[bestPodID] = struct{}{}
	}

	decision.NodeSet = RemainingNodeSet(snapshot, cfg, decision.Victims)
	decision.ScoreAfter = float64(len(decision.NodeSet))
	return decision
}

func RemainingNodeSet(snapshot *model.ClusterSnapshot, cfg model.BinpackConfig, victims []string) []string {
	if snapshot == nil || snapshot.Nodes == nil || snapshot.Pods == nil {
		return nil
	}
	exclude := make(map[string]struct{}, len(victims))
	for _, id := range victims {
		exclude[id] = struct{}{}
	}

	remaining := make(map[string]struct{})
	for _, node := range snapshot.Nodes {
		if node == nil {
			continue
		}
		for _, podID := range node.Pods {
			if _, drop := exclude[podID]; drop {
				continue
			}
			pod := snapshot.Pods[podID]
			if pod == nil || !matchesTarget(pod, cfg) {
				continue
			}
			remaining[node.Name] = struct{}{}
			break
		}
	}

	nodeNames := make([]string, 0, len(remaining))
	for name := range remaining {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)
	return nodeNames
}

func collectTargetPods(snapshot *model.ClusterSnapshot, cfg model.BinpackConfig) []*model.Pod {
	result := make([]*model.Pod, 0)
	for _, pod := range snapshot.Pods {
		if pod == nil {
			continue
		}
		if matchesTarget(pod, cfg) {
			result = append(result, pod)
		}
	}
	return result
}

func matchesTarget(pod *model.Pod, cfg model.BinpackConfig) bool {
	if pod == nil {
		return false
	}
	if cfg.TargetWorkload != "" && pod.Owner != cfg.TargetWorkload {
		return false
	}
	if len(cfg.Selector) > 0 {
		for key, value := range cfg.Selector {
			if pod.Labels == nil || pod.Labels[key] != value {
				return false
			}
		}
	}
	return true
}

func buildPodNodeIndex(snapshot *model.ClusterSnapshot) map[string]string {
	index := make(map[string]string)
	for _, node := range snapshot.Nodes {
		if node == nil {
			continue
		}
		for _, podID := range node.Pods {
			index[podID] = node.Name
		}
	}
	return index
}

func countNodesWithTargetPods(snapshot *model.ClusterSnapshot, cfg model.BinpackConfig, exclude map[string]struct{}) int {
	if snapshot == nil || snapshot.Nodes == nil || snapshot.Pods == nil {
		return 0
	}
	count := 0
	for _, node := range snapshot.Nodes {
		if node == nil {
			continue
		}
		for _, podID := range node.Pods {
			if exclude != nil {
				if _, drop := exclude[podID]; drop {
					continue
				}
			}
			pod := snapshot.Pods[podID]
			if pod == nil || !matchesTarget(pod, cfg) {
				continue
			}
			count++
			break
		}
	}
	return count
}
