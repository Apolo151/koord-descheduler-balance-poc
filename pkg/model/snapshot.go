package model

type ClusterSnapshot struct {
	Nodes map[string]*Node `json:"nodes" yaml:"nodes"`
	Pods  map[string]*Pod  `json:"pods" yaml:"pods"`
}

func PodKey(namespace, name string) string {
	if namespace == "" {
		return name
	}
	return namespace + "/" + name
}

func (cs *ClusterSnapshot) GetPod(id string) *Pod {
	if cs == nil || cs.Pods == nil {
		return nil
	}
	return cs.Pods[id]
}

func (cs *ClusterSnapshot) RequestedOnNode(nodeName string) ResourceVector {
	requested := ResourceVector{}
	if cs == nil || cs.Nodes == nil || cs.Pods == nil {
		return requested
	}
	node := cs.Nodes[nodeName]
	if node == nil {
		return requested
	}
	for _, podID := range node.Pods {
		pod := cs.Pods[podID]
		if pod == nil {
			continue
		}
		requested = requested.Add(pod.Resources)
	}
	return requested
}

func (cs *ClusterSnapshot) FindPodsByOwner(owner string) []*Pod {
	if cs == nil || cs.Pods == nil {
		return nil
	}
	result := make([]*Pod, 0)
	for _, pod := range cs.Pods {
		if pod != nil && pod.Owner == owner {
			result = append(result, pod)
		}
	}
	return result
}

func (cs *ClusterSnapshot) FindPodsBySelector(selector map[string]string) []*Pod {
	if cs == nil || cs.Pods == nil {
		return nil
	}
	if len(selector) == 0 {
		return nil
	}
	result := make([]*Pod, 0)
	for _, pod := range cs.Pods {
		if pod == nil {
			continue
		}
		match := true
		for key, value := range selector {
			if pod.Labels == nil || pod.Labels[key] != value {
				match = false
				break
			}
		}
		if match {
			result = append(result, pod)
		}
	}
	return result
}
