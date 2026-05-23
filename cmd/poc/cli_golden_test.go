package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"koord-descheduler-balance-poc/pkg/binpack"
	"koord-descheduler-balance-poc/pkg/descheduler"
	"koord-descheduler-balance-poc/pkg/fixtures"
	"koord-descheduler-balance-poc/pkg/imbalance"
	"koord-descheduler-balance-poc/pkg/model"
	"koord-descheduler-balance-poc/pkg/report"
)

func TestSummaryGoldenBasicImbalance(t *testing.T) {
	assertGoldenSummary(t, "basic-imbalance.yaml", "basic-imbalance.txt")
}

func TestSummaryGoldenScaleDown(t *testing.T) {
	assertGoldenSummary(t, "scale-down-binpack.yaml", "scale-down-binpack.txt")
}

func assertGoldenSummary(t *testing.T, scenarioFile, goldenFile string) {
	root := filepath.Join("..", "..")
	scenarioPath := filepath.Join(root, "testdata", "scenarios", scenarioFile)
	goldenPath := filepath.Join(root, "testdata", "golden", goldenFile)

	scenario, err := fixtures.LoadScenario(scenarioPath)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	if err := fixtures.ValidateScenario(scenario); err != nil {
		t.Fatalf("validate scenario: %v", err)
	}

	snapshot, imbalanceCfg, evictionPolicy, binpackCfg := fixtures.ToSnapshot(scenario)
	result := report.Result{
		Scores:  computeScoresForTest(snapshot, imbalanceCfg),
		Plans:   descheduler.PlanEvictions(snapshot, imbalanceCfg, evictionPolicy),
		Binpack: nil,
	}
	if binpackCfg.Victims > 0 && (binpackCfg.TargetWorkload != "" || len(binpackCfg.Selector) > 0) {
		decision := binpack.SelectVictims(snapshot, binpackCfg)
		result.Binpack = &decision
	}

	summary := report.RenderSummary(result)
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if summary != string(golden) {
		t.Fatalf("summary mismatch\nexpected:\n%s\n\nactual:\n%s", string(golden), summary)
	}
}

func computeScoresForTest(snapshot *model.ClusterSnapshot, cfg model.ImbalanceConfig) []model.NodeImbalanceScore {
	if snapshot == nil || snapshot.Nodes == nil {
		return nil
	}
	nodeNames := make([]string, 0, len(snapshot.Nodes))
	for name := range snapshot.Nodes {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	scores := make([]model.NodeImbalanceScore, 0, len(nodeNames))
	for _, nodeName := range nodeNames {
		node := snapshot.Nodes[nodeName]
		if node == nil {
			continue
		}
		scores = append(scores, imbalance.ScoreNode(node, snapshot, cfg))
	}
	return scores
}
