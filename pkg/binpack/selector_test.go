package binpack

import (
	"testing"

	"koord-descheduler-balance-poc/pkg/model"
)

func TestSelectVictimsDeterministic(t *testing.T) {
	snapshot := &model.ClusterSnapshot{
		Nodes: map[string]*model.Node{
			"node-a": {Name: "node-a", Allocatable: model.ResourceVector{"cpu": 10, "memory": 10}, Pods: []string{"default/app-a"}},
			"node-b": {Name: "node-b", Allocatable: model.ResourceVector{"cpu": 10, "memory": 10}, Pods: []string{"default/app-b"}},
			"node-c": {Name: "node-c", Allocatable: model.ResourceVector{"cpu": 10, "memory": 10}, Pods: []string{"default/app-c"}},
		},
		Pods: map[string]*model.Pod{
			"default/app-a": {Name: "app-a", Namespace: "default", Owner: "app", Resources: model.ResourceVector{"cpu": 2, "memory": 2}},
			"default/app-b": {Name: "app-b", Namespace: "default", Owner: "app", Resources: model.ResourceVector{"cpu": 2, "memory": 2}},
			"default/app-c": {Name: "app-c", Namespace: "default", Owner: "app", Resources: model.ResourceVector{"cpu": 2, "memory": 2}},
		},
	}
	cfg := model.BinpackConfig{TargetWorkload: "app", Victims: 1}
	decision := SelectVictims(snapshot, cfg)
	if len(decision.Victims) != 1 || decision.Victims[0] != "default/app-a" {
		t.Fatalf("expected victim default/app-a, got %v", decision.Victims)
	}

	remaining := RemainingNodeSet(snapshot, cfg, decision.Victims)
	if len(remaining) != 2 || remaining[0] != "node-b" || remaining[1] != "node-c" {
		t.Fatalf("unexpected remaining node set %v", remaining)
	}
}
