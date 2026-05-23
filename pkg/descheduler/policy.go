package descheduler

import "koord-descheduler-balance-poc/pkg/model"

func FilterEvictablePods(node *model.Node, snapshot *model.ClusterSnapshot, policy model.EvictionPolicy) []*model.Pod {
	if node == nil || snapshot == nil || snapshot.Pods == nil {
		return nil
	}

	deny := policy.DenyNamespaces
	if len(deny) == 0 {
		deny = []string{"kube-system"}
	}
	denySet := make(map[string]struct{}, len(deny))
	for _, name := range deny {
		denySet[name] = struct{}{}
	}
	allowSet := make(map[string]struct{}, len(policy.AllowNamespaces))
	for _, name := range policy.AllowNamespaces {
		allowSet[name] = struct{}{}
	}

	result := make([]*model.Pod, 0)
	for _, podID := range node.Pods {
		pod := snapshot.Pods[podID]
		if pod == nil {
			continue
		}
		if _, denied := denySet[pod.Namespace]; denied {
			continue
		}
		if len(allowSet) > 0 {
			if _, ok := allowSet[pod.Namespace]; !ok {
				continue
			}
		}
		if !pod.Evictable && policy.ReservationFirst {
			continue
		}
		if isProtectedByLabels(pod, policy.ProtectedLabels) {
			continue
		}
		result = append(result, pod)
	}

	return result
}

func ApplyReservationFirstScore(pod *model.Pod, policy model.EvictionPolicy) float64 {
	if pod == nil || !policy.ReservationFirst {
		return 0
	}
	if !pod.Evictable {
		return -1
	}
	return 0
}

func isProtectedByLabels(pod *model.Pod, protected map[string]string) bool {
	if pod == nil || len(protected) == 0 {
		return false
	}
	for key, value := range protected {
		if pod.Labels == nil {
			return false
		}
		if pod.Labels[key] != value {
			return false
		}
	}
	return true
}
