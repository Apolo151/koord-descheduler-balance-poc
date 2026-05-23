package descheduler

import (
	"testing"

	"koord-descheduler-balance-poc/pkg/model"
)

func TestPlanEvictionsSkipsBelowThreshold(t *testing.T) {
	snapshot := &model.ClusterSnapshot{
		Nodes: map[string]*model.Node{
			"node-a": {Name: "node-a", Allocatable: model.ResourceVector{"cpu": 10, "memory": 10}, Pods: []string{"default/pod-a"}},
		},
		Pods: map[string]*model.Pod{
			"default/pod-a": {Name: "pod-a", Namespace: "default", Resources: model.ResourceVector{"cpu": 1, "memory": 1}, Evictable: true},
		},
	}
	cfg := model.ImbalanceConfig{Resources: []string{"cpu", "memory"}, Threshold: 0.1}
	plans := PlanEvictions(snapshot, cfg, model.EvictionPolicy{ReservationFirst: true})
	if len(plans) != 0 {
		t.Fatalf("expected no eviction plans, got %d", len(plans))
	}
}

func TestPlanEvictionsRespectsMaxEvictions(t *testing.T) {
	snapshot := &model.ClusterSnapshot{
		Nodes: map[string]*model.Node{
			"node-a": {Name: "node-a", Allocatable: model.ResourceVector{"cpu": 10, "memory": 10}, Pods: []string{"default/pod-a", "default/pod-b", "default/pod-c"}},
		},
		Pods: map[string]*model.Pod{
			"default/pod-a": {Name: "pod-a", Namespace: "default", Resources: model.ResourceVector{"cpu": 9, "memory": 1}, Evictable: true},
			"default/pod-b": {Name: "pod-b", Namespace: "default", Resources: model.ResourceVector{"cpu": 0, "memory": 4}, Evictable: true},
			"default/pod-c": {Name: "pod-c", Namespace: "default", Resources: model.ResourceVector{"cpu": 0, "memory": 0}, Evictable: true},
		},
	}
	cfg := model.ImbalanceConfig{Resources: []string{"cpu", "memory"}, Threshold: 0.1}
	policy := model.EvictionPolicy{ReservationFirst: true, MaxEvictionsPerNode: 2}
	plans := PlanEvictions(snapshot, cfg, policy)
	if len(plans) != 1 {
		t.Fatalf("expected 1 eviction plan, got %d", len(plans))
	}
	plan := plans[0]
	if len(plan.Candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(plan.Candidates))
	}
	if len(plan.Chosen) != 2 {
		t.Fatalf("expected 2 chosen, got %d", len(plan.Chosen))
	}
	if plan.Chosen[0].PodID != "default/pod-a" || plan.Chosen[1].PodID != "default/pod-c" {
		t.Fatalf("unexpected chosen order %v", plan.Chosen)
	}
}
