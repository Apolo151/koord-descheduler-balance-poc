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

func TestSelectVictimsSelectorLabels(t *testing.T) {
	snapshot := &model.ClusterSnapshot{
		Nodes: map[string]*model.Node{
			"node-a": {Name: "node-a", Allocatable: model.ResourceVector{"cpu": 10, "memory": 10}, Pods: []string{"default/app-a"}},
			"node-b": {Name: "node-b", Allocatable: model.ResourceVector{"cpu": 10, "memory": 10}, Pods: []string{"default/app-b"}},
			"node-c": {Name: "node-c", Allocatable: model.ResourceVector{"cpu": 10, "memory": 10}, Pods: []string{"default/app-c"}},
		},
		Pods: map[string]*model.Pod{
			"default/app-a": {Name: "app-a", Namespace: "default", Labels: map[string]string{"app": "demo", "tier": "api"}},
			"default/app-b": {Name: "app-b", Namespace: "default", Labels: map[string]string{"app": "demo", "tier": "api"}},
			"default/app-c": {Name: "app-c", Namespace: "default", Labels: map[string]string{"app": "demo", "tier": "api"}},
		},
	}
	cfg := model.BinpackConfig{Selector: map[string]string{"app": "demo", "tier": "api"}, Victims: 2}
	decision := SelectVictims(snapshot, cfg)
	if len(decision.Victims) != 2 {
		t.Fatalf("expected 2 victims, got %v", decision.Victims)
	}
	if decision.Victims[0] != "default/app-a" || decision.Victims[1] != "default/app-b" {
		t.Fatalf("unexpected victim order %v", decision.Victims)
	}
}

func TestScoreCandidateRemovalNoChange(t *testing.T) {
	snapshot := &model.ClusterSnapshot{
		Nodes: map[string]*model.Node{
			"node-a": {Name: "node-a", Pods: []string{"default/app-a", "default/app-b"}},
		},
		Pods: map[string]*model.Pod{
			"default/app-a": {Name: "app-a", Namespace: "default", Owner: "app"},
			"default/app-b": {Name: "app-b", Namespace: "default", Owner: "app"},
		},
	}
	cfg := model.BinpackConfig{TargetWorkload: "app", Victims: 1}
	score := ScoreCandidateRemoval(snapshot, cfg, "default/app-a", "node-a", nil)
	if score != 0 {
		t.Fatalf("expected score 0 when node still has target pod, got %v", score)
	}
}

func TestMatchesTargetOwnerAndSelector(t *testing.T) {
	pod := &model.Pod{Owner: "app", Labels: map[string]string{"tier": "api"}}
	if !matchesTarget(pod, model.BinpackConfig{TargetWorkload: "app"}) {
		t.Fatalf("expected owner match")
	}
	if matchesTarget(pod, model.BinpackConfig{TargetWorkload: "other"}) {
		t.Fatalf("expected owner mismatch")
	}
	if !matchesTarget(pod, model.BinpackConfig{Selector: map[string]string{"tier": "api"}}) {
		t.Fatalf("expected selector match")
	}
	if matchesTarget(pod, model.BinpackConfig{Selector: map[string]string{"tier": "db"}}) {
		t.Fatalf("expected selector mismatch")
	}
}
