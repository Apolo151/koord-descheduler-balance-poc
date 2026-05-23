package fixtures

import (
	"os"
	"path/filepath"
	"testing"

	"koord-descheduler-balance-poc/pkg/model"
)

func TestLoadScenarioYAMLAndJSON(t *testing.T) {
	dir := t.TempDir()

	yamlPath := filepath.Join(dir, "scenario.yaml")
	yamlData := "metadata:\n  name: load-yaml\ncluster:\n  pods:\n    default/pod-a:\n      resources:\n        cpu: 1\n        memory: 1\n      evictable: true\n  nodes:\n    node-a:\n      allocatable:\n        cpu: 2\n        memory: 2\n      pods:\n        - default/pod-a\nimbalanceConfig:\n  resources: [cpu, memory]\n  threshold: 0.1\nevictionPolicy:\n  reservationFirst: true\nbinpackConfig:\n  victims: 0\nevents: {}\n"
	if err := os.WriteFile(yamlPath, []byte(yamlData), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	scenario, err := LoadScenario(yamlPath)
	if err != nil {
		t.Fatalf("load yaml: %v", err)
	}
	if err := ValidateScenario(scenario); err != nil {
		t.Fatalf("validate yaml: %v", err)
	}

	jsonPath := filepath.Join(dir, "scenario.json")
	jsonData := "{\n  \"cluster\": {\n    \"pods\": {\n      \"default/pod-a\": {\n        \"resources\": {\"cpu\": 1, \"memory\": 1},\n        \"evictable\": true\n      }\n    },\n    \"nodes\": {\n      \"node-a\": {\n        \"allocatable\": {\"cpu\": 2, \"memory\": 2},\n        \"pods\": [\"default/pod-a\"]\n      }\n    }\n  },\n  \"imbalanceConfig\": {\n    \"resources\": [\"cpu\", \"memory\"],\n    \"threshold\": 0.1\n  },\n  \"evictionPolicy\": {\n    \"reservationFirst\": true\n  },\n  \"binpackConfig\": {\n    \"victims\": 0\n  }\n}\n"
	if err := os.WriteFile(jsonPath, []byte(jsonData), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	scenario, err = LoadScenario(jsonPath)
	if err != nil {
		t.Fatalf("load json: %v", err)
	}
	if err := ValidateScenario(scenario); err != nil {
		t.Fatalf("validate json: %v", err)
	}
}

func TestValidateScenarioErrors(t *testing.T) {
	cases := []struct {
		name     string
		scenario Scenario
	}{
		{
			name:     "missing-nodes",
			scenario: Scenario{Cluster: Cluster{Pods: map[string]*model.Pod{"default/pod-a": {Resources: model.ResourceVector{"cpu": 1}}}}},
		},
		{
			name:     "missing-pods",
			scenario: Scenario{Cluster: Cluster{Nodes: map[string]*model.Node{"node-a": {Allocatable: model.ResourceVector{"cpu": 1}}}}},
		},
		{
			name: "missing-pod-reference",
			scenario: Scenario{Cluster: Cluster{
				Pods:  map[string]*model.Pod{"default/pod-a": {Resources: model.ResourceVector{"cpu": 1}}},
				Nodes: map[string]*model.Node{"node-a": {Allocatable: model.ResourceVector{"cpu": 1}, Pods: []string{"default/pod-b"}}},
			}},
		},
		{
			name: "duplicate-pod-assignment",
			scenario: Scenario{Cluster: Cluster{
				Pods: map[string]*model.Pod{"default/pod-a": {Resources: model.ResourceVector{"cpu": 1}}},
				Nodes: map[string]*model.Node{
					"node-a": {Allocatable: model.ResourceVector{"cpu": 1}, Pods: []string{"default/pod-a"}},
					"node-b": {Allocatable: model.ResourceVector{"cpu": 1}, Pods: []string{"default/pod-a"}},
				},
			}},
		},
		{
			name: "gpu-fractional",
			scenario: Scenario{Cluster: Cluster{
				Pods:  map[string]*model.Pod{"default/gpu": {Resources: model.ResourceVector{"nvidia.com/gpu": 0.5}}},
				Nodes: map[string]*model.Node{"node-a": {Allocatable: model.ResourceVector{"nvidia.com/gpu": 1}, Pods: []string{"default/gpu"}}},
			}},
		},
	}

	for _, testCase := range cases {
		if err := ValidateScenario(testCase.scenario); err == nil {
			t.Fatalf("expected error for %s", testCase.name)
		}
	}
}

func TestToSnapshotAppliesScaleDownOverrides(t *testing.T) {
	scenario := Scenario{
		Cluster: Cluster{
			Pods: map[string]*model.Pod{
				"default/pod-a": {Resources: model.ResourceVector{"cpu": 1}},
			},
			Nodes: map[string]*model.Node{
				"node-a": {Allocatable: model.ResourceVector{"cpu": 2}, Pods: []string{"default/pod-a"}},
			},
		},
		BinpackConfig: model.BinpackConfig{TargetWorkload: "app", Victims: 1},
		Events:        Events{ScaleDown: &ScaleDownEvent{TargetWorkload: "override", Selector: map[string]string{"app": "demo"}, Victims: 2}},
	}
	_, _, _, binpackCfg := ToSnapshot(scenario)
	if binpackCfg.TargetWorkload != "override" {
		t.Fatalf("expected targetWorkload override, got %q", binpackCfg.TargetWorkload)
	}
	if len(binpackCfg.Selector) != 1 || binpackCfg.Selector["app"] != "demo" {
		t.Fatalf("expected selector override, got %v", binpackCfg.Selector)
	}
	if binpackCfg.Victims != 2 {
		t.Fatalf("expected victims override, got %d", binpackCfg.Victims)
	}
}
