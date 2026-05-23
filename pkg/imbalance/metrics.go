package imbalance

import (
	"math"
	"sort"

	"koord-descheduler-balance-poc/pkg/model"
)

func ComputeRatios(allocatable, requested model.ResourceVector, resources []string) (map[string]float64, []string) {
	ratios := make(map[string]float64)
	used := make([]string, 0)

	resourceList := resources
	if len(resourceList) == 0 {
		keys := allocatable.Keys()
		if len(keys) == 0 {
			keys = requested.Keys()
		}
		resourceList = keys
	}

	for _, resource := range resourceList {
		alloc := allocatable.Get(resource)
		if alloc <= 0 {
			continue
		}
		req := requested.Get(resource)
		ratios[resource] = req / alloc
		used = append(used, resource)
	}

	sort.Strings(used)
	return ratios, used
}

func MeanStdDev(ratios map[string]float64, resources []string) (float64, float64) {
	if len(ratios) == 0 {
		return 0, 0
	}

	resourceList := resources
	if len(resourceList) == 0 {
		resourceList = make([]string, 0, len(ratios))
		for key := range ratios {
			resourceList = append(resourceList, key)
		}
		sort.Strings(resourceList)
	}

	count := 0
	sum := 0.0
	for _, resource := range resourceList {
		value, ok := ratios[resource]
		if !ok {
			continue
		}
		sum += value
		count++
	}
	if count == 0 {
		return 0, 0
	}

	mean := sum / float64(count)
	variance := 0.0
	for _, resource := range resourceList {
		value, ok := ratios[resource]
		if !ok {
			continue
		}
		diff := value - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(count))
	return mean, stdDev
}

func ScoreNode(node *model.Node, snapshot *model.ClusterSnapshot, cfg model.ImbalanceConfig) model.NodeImbalanceScore {
	result := model.NodeImbalanceScore{
		NodeName: node.Name,
		Ratios:   map[string]float64{},
		Mean:     0,
		StdDev:   0,
	}
	if node == nil || snapshot == nil {
		return result
	}
	requested := snapshot.RequestedOnNode(node.Name)
	ratios, used := ComputeRatios(node.Allocatable, requested, cfg.Resources)
	mean, stdDev := MeanStdDev(ratios, used)
	result.Ratios = ratios
	result.Mean = mean
	result.StdDev = stdDev
	return result
}
