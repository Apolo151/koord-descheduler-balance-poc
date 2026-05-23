package fixtures

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"koord-descheduler-balance-poc/pkg/model"
)

func LoadScenario(path string) (Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, err
	}

	var scenario Scenario
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" {
		if err := json.Unmarshal(data, &scenario); err != nil {
			return Scenario{}, err
		}
		return scenario, nil
	}
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &scenario); err != nil {
			return Scenario{}, err
		}
		return scenario, nil
	}

	if err := json.Unmarshal(data, &scenario); err == nil {
		return scenario, nil
	}
	if err := yaml.Unmarshal(data, &scenario); err == nil {
		return scenario, nil
	}

	return Scenario{}, fmt.Errorf("unable to parse scenario: %s", path)
}

func ValidateScenario(s Scenario) error {
	if len(s.Cluster.Nodes) == 0 {
		return errors.New("cluster.nodes is required")
	}
	if len(s.Cluster.Pods) == 0 {
		return errors.New("cluster.pods is required")
	}

	for podID, pod := range s.Cluster.Pods {
		if pod == nil {
			return fmt.Errorf("pod %q is null", podID)
		}
		name, namespace := parsePodKey(podID)
		if pod.Name == "" {
			pod.Name = name
		}
		if pod.Namespace == "" {
			pod.Namespace = namespace
		}
		if err := validateResourceVector(pod.Resources); err != nil {
			return fmt.Errorf("pod %q: %w", podID, err)
		}
	}

	seen := make(map[string]string)
	for nodeName, node := range s.Cluster.Nodes {
		if node == nil {
			return fmt.Errorf("node %q is null", nodeName)
		}
		if node.Name == "" {
			node.Name = nodeName
		}
		if err := validateResourceVector(node.Allocatable); err != nil {
			return fmt.Errorf("node %q: %w", nodeName, err)
		}
		for _, podID := range node.Pods {
			if _, ok := s.Cluster.Pods[podID]; !ok {
				return fmt.Errorf("node %q references missing pod %q", node.Name, podID)
			}
			if prev, exists := seen[podID]; exists {
				return fmt.Errorf("pod %q assigned to multiple nodes (%s, %s)", podID, prev, node.Name)
			}
			seen[podID] = node.Name
		}
	}

	return nil
}

func ToSnapshot(s Scenario) (*model.ClusterSnapshot, model.ImbalanceConfig, model.EvictionPolicy, model.BinpackConfig) {
	snapshot := &model.ClusterSnapshot{
		Nodes: make(map[string]*model.Node),
		Pods:  make(map[string]*model.Pod),
	}

	for podID, pod := range s.Cluster.Pods {
		if pod == nil {
			continue
		}
		name, namespace := parsePodKey(podID)
		if pod.Name == "" {
			pod.Name = name
		}
		if pod.Namespace == "" {
			pod.Namespace = namespace
		}
		snapshot.Pods[podID] = pod
	}

	for nodeName, node := range s.Cluster.Nodes {
		if node == nil {
			continue
		}
		if node.Name == "" {
			node.Name = nodeName
		}
		snapshot.Nodes[node.Name] = node
	}

	binpack := s.BinpackConfig
	if s.Events.ScaleDown != nil {
		if s.Events.ScaleDown.TargetWorkload != "" {
			binpack.TargetWorkload = s.Events.ScaleDown.TargetWorkload
		}
		if len(s.Events.ScaleDown.Selector) > 0 {
			binpack.Selector = s.Events.ScaleDown.Selector
		}
		if s.Events.ScaleDown.Victims > 0 {
			binpack.Victims = s.Events.ScaleDown.Victims
		}
	}

	return snapshot, s.ImbalanceConfig, s.EvictionPolicy, binpack
}

func validateResourceVector(rv model.ResourceVector) error {
	if rv == nil {
		return nil
	}
	for key, value := range rv {
		if strings.Contains(strings.ToLower(key), "gpu") {
			if math.Trunc(value) != value {
				return fmt.Errorf("resource %q must be integer", key)
			}
		}
	}
	return nil
}

func parsePodKey(podID string) (string, string) {
	parts := strings.SplitN(podID, "/", 2)
	if len(parts) == 2 {
		return parts[1], parts[0]
	}
	return podID, ""
}
