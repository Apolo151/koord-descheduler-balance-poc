package descheduler

import (
	"testing"

	"koord-descheduler-balance-poc/pkg/model"
)

func TestFilterEvictablePods(t *testing.T) {
	snapshot := &model.ClusterSnapshot{
		Nodes: map[string]*model.Node{
			"node-a": {Name: "node-a", Pods: []string{"kube-system/sys", "default/pod-a", "default/pod-b"}},
		},
		Pods: map[string]*model.Pod{
			"kube-system/sys": {Name: "sys", Namespace: "kube-system", Evictable: true},
			"default/pod-a":   {Name: "pod-a", Namespace: "default", Evictable: true},
			"default/pod-b":   {Name: "pod-b", Namespace: "default", Evictable: false},
		},
	}

	policy := model.EvictionPolicy{ReservationFirst: true}
	pods := FilterEvictablePods(snapshot.Nodes["node-a"], snapshot, policy)
	if len(pods) != 1 || pods[0].Name != "pod-a" {
		t.Fatalf("expected only pod-a evictable, got %v", pods)
	}

	policy.ProtectedLabels = map[string]string{"protected": "true"}
	snapshot.Pods["default/pod-a"].Labels = map[string]string{"protected": "true"}
	pods = FilterEvictablePods(snapshot.Nodes["node-a"], snapshot, policy)
	if len(pods) != 0 {
		t.Fatalf("expected no evictable pods with protected label, got %v", pods)
	}
}

func TestFilterEvictablePodsAllowDenyPrecedence(t *testing.T) {
	snapshot := &model.ClusterSnapshot{
		Nodes: map[string]*model.Node{
			"node-a": {Name: "node-a", Pods: []string{"team-a/pod-a", "default/pod-b", "kube-system/sys"}},
		},
		Pods: map[string]*model.Pod{
			"team-a/pod-a":    {Name: "pod-a", Namespace: "team-a", Evictable: true},
			"default/pod-b":   {Name: "pod-b", Namespace: "default", Evictable: true},
			"kube-system/sys": {Name: "sys", Namespace: "kube-system", Evictable: true},
		},
	}

	policy := model.EvictionPolicy{
		ReservationFirst: true,
		AllowNamespaces:  []string{"team-a", "default"},
		DenyNamespaces:   []string{"kube-system", "default"},
	}
	pods := FilterEvictablePods(snapshot.Nodes["node-a"], snapshot, policy)
	if len(pods) != 1 || pods[0].Namespace != "team-a" {
		t.Fatalf("expected only team-a pod allowed, got %v", pods)
	}
}

func TestApplyReservationFirstScore(t *testing.T) {
	policy := model.EvictionPolicy{ReservationFirst: true}
	if ApplyReservationFirstScore(&model.Pod{Evictable: false}, policy) != -1 {
		t.Fatalf("expected non-evictable pod to score -1")
	}
	if ApplyReservationFirstScore(&model.Pod{Evictable: true}, policy) != 0 {
		t.Fatalf("expected evictable pod to score 0")
	}
	if ApplyReservationFirstScore(&model.Pod{Evictable: false}, model.EvictionPolicy{}) != 0 {
		t.Fatalf("expected reservationFirst disabled to score 0")
	}
}
