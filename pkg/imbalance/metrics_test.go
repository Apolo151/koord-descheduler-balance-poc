package imbalance

import (
	"math"
	"testing"

	"koord-descheduler-balance-poc/pkg/model"
)

func TestComputeRatiosZeroAllocatable(t *testing.T) {
	alloc := model.ResourceVector{"cpu": 10, "memory": 0}
	req := model.ResourceVector{"cpu": 5, "memory": 3}
	ratios, used := ComputeRatios(alloc, req, []string{"cpu", "memory"})
	if len(used) != 1 || used[0] != "cpu" {
		t.Fatalf("expected used [cpu], got %v", used)
	}
	if ratios["cpu"] != 0.5 {
		t.Fatalf("expected cpu ratio 0.5, got %v", ratios["cpu"])
	}
	if _, ok := ratios["memory"]; ok {
		t.Fatalf("expected memory ratio to be skipped")
	}
}

func TestMeanStdDev(t *testing.T) {
	ratios := map[string]float64{"cpu": 0.9, "memory": 0.5}
	mean, stdDev := MeanStdDev(ratios, []string{"cpu", "memory"})
	if !almostEqual(mean, 0.7) {
		t.Fatalf("expected mean 0.7, got %v", mean)
	}
	if !almostEqual(stdDev, 0.2) {
		t.Fatalf("expected stddev 0.2, got %v", stdDev)
	}
}

func TestMeanStdDevSkipsMissingResources(t *testing.T) {
	ratios := map[string]float64{"cpu": 0.5}
	mean, stdDev := MeanStdDev(ratios, []string{"cpu", "memory"})
	if !almostEqual(mean, 0.5) {
		t.Fatalf("expected mean 0.5, got %v", mean)
	}
	if !almostEqual(stdDev, 0.0) {
		t.Fatalf("expected stddev 0.0, got %v", stdDev)
	}
}

func TestDeltaStdDevForPod(t *testing.T) {
	snapshot := &model.ClusterSnapshot{
		Nodes: map[string]*model.Node{
			"node-a": {
				Name:        "node-a",
				Allocatable: model.ResourceVector{"cpu": 10, "memory": 10},
				Pods:        []string{"default/pod-a", "default/pod-b"},
			},
		},
		Pods: map[string]*model.Pod{
			"default/pod-a": {Name: "pod-a", Namespace: "default", Resources: model.ResourceVector{"cpu": 9, "memory": 1}},
			"default/pod-b": {Name: "pod-b", Namespace: "default", Resources: model.ResourceVector{"cpu": 0, "memory": 4}},
		},
	}
	cfg := model.ImbalanceConfig{Resources: []string{"cpu", "memory"}}
	node := snapshot.Nodes["node-a"]

	deltaA, _ := DeltaStdDevForPod(node, snapshot.Pods["default/pod-a"], snapshot, cfg)
	deltaB, _ := DeltaStdDevForPod(node, snapshot.Pods["default/pod-b"], snapshot, cfg)
	if !almostEqual(deltaA, 0.0) {
		t.Fatalf("expected delta 0.0 for pod-a, got %v", deltaA)
	}
	if !almostEqual(deltaB, -0.2) {
		t.Fatalf("expected delta -0.2 for pod-b, got %v", deltaB)
	}
}

func TestDeltaStdDevForPodNilInputs(t *testing.T) {
	delta, stdDev := DeltaStdDevForPod(nil, nil, nil, model.ImbalanceConfig{})
	if delta != 0 || stdDev != 0 {
		t.Fatalf("expected zero delta/stddev for nil inputs, got %v/%v", delta, stdDev)
	}
}

func TestScoreNodeUsesAllocatableKeys(t *testing.T) {
	snapshot := &model.ClusterSnapshot{
		Nodes: map[string]*model.Node{
			"node-a": {Name: "node-a", Allocatable: model.ResourceVector{"cpu": 10, "memory": 10}, Pods: []string{"default/pod-a"}},
		},
		Pods: map[string]*model.Pod{
			"default/pod-a": {Name: "pod-a", Namespace: "default", Resources: model.ResourceVector{"cpu": 2, "memory": 4}},
		},
	}
	score := ScoreNode(snapshot.Nodes["node-a"], snapshot, model.ImbalanceConfig{})
	if !almostEqual(score.Mean, 0.3) {
		t.Fatalf("expected mean 0.3, got %v", score.Mean)
	}
	if !almostEqual(score.StdDev, 0.1) {
		t.Fatalf("expected stddev 0.1, got %v", score.StdDev)
	}
}

func TestSortCandidates(t *testing.T) {
	candidates := []model.EvictionCandidate{
		{PodID: "b", NodeName: "node-a", DeltaStdDev: 0.1},
		{PodID: "a", NodeName: "node-a", DeltaStdDev: 0.1},
		{PodID: "c", NodeName: "node-b", DeltaStdDev: 0.2},
	}
	SortCandidates(candidates)
	if candidates[0].PodID != "c" {
		t.Fatalf("expected highest delta first, got %s", candidates[0].PodID)
	}
	if candidates[1].PodID != "a" || candidates[2].PodID != "b" {
		t.Fatalf("expected tie-break by pod ID, got %v", candidates)
	}
}

func almostEqual(value, expected float64) bool {
	return math.Abs(value-expected) < 1e-6
}
