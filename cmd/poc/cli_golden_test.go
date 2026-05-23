package main

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"koord-descheduler-balance-poc/pkg/binpack"
	"koord-descheduler-balance-poc/pkg/descheduler"
	"koord-descheduler-balance-poc/pkg/fixtures"
	"koord-descheduler-balance-poc/pkg/imbalance"
	"koord-descheduler-balance-poc/pkg/model"
	"koord-descheduler-balance-poc/pkg/report"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files")

func TestSummaryGoldenScenarios(t *testing.T) {
	cases := []struct {
		scenario string
		golden   string
	}{
		{scenario: "basic-imbalance.yaml", golden: "basic-imbalance.txt"},
		{scenario: "scale-down-binpack.yaml", golden: "scale-down-binpack.txt"},
		{scenario: "boundary-threshold.yaml", golden: "boundary-threshold.txt"},
		{scenario: "max-evictions.yaml", golden: "max-evictions.txt"},
		{scenario: "allow-deny-namespaces.yaml", golden: "allow-deny-namespaces.txt"},
		{scenario: "selector-binpack.yaml", golden: "selector-binpack.txt"},
		{scenario: "no-evictable.yaml", golden: "no-evictable.txt"},
		{scenario: "zero-allocatable.yaml", golden: "zero-allocatable.txt"},
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.scenario, func(t *testing.T) {
			assertGoldenSummary(t, testCase.scenario, testCase.golden)
		})
	}
}

func assertGoldenSummary(t *testing.T, scenarioFile, goldenFile string) {
	root := filepath.Join("..", "..")
	scenarioPath := filepath.Join(root, "testdata", "scenarios", scenarioFile)
	goldenPath := filepath.Join(root, "testdata", "golden", goldenFile)
	htmlDir := filepath.Join(root, "ui")

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
	if err := writeHTMLArtifact(htmlDir, scenarioPath, scenario.Metadata, snapshot, result); err != nil {
		t.Fatalf("write html artifact: %v", err)
	}
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	normalizedSummary := normalizeSummary(summary)
	normalizedGolden := normalizeSummary(string(golden))

	if *updateGolden {
		if err := os.WriteFile(goldenPath, []byte(normalizedSummary+"\n"), 0644); err != nil {
			t.Fatalf("update golden: %v", err)
		}
		return
	}

	if normalizedSummary != normalizedGolden {
		t.Fatalf("summary mismatch\nexpected:\n%s\n\nactual:\n%s", normalizedGolden, normalizedSummary)
	}
}

func writeHTMLArtifact(outputDir string, scenarioPath string, metadata map[string]string, snapshot *model.ClusterSnapshot, result report.Result) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	name := scenarioTitle(scenarioPath, metadata)
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(scenarioPath), filepath.Ext(scenarioPath))
	}
	name = sanitizeFilename(name)
	if name == "" {
		name = "scenario"
	}

	path := filepath.Join(outputDir, "test-"+name+".html")
	visual, err := report.RenderHTML(snapshot, result, report.HTMLMetadata{
		Title:       name,
		Subtitle:    "Golden test visualization",
		ScenarioRef: filepath.Base(scenarioPath),
	})
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(visual), 0644)
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

func normalizeSummary(value string) string {
	return strings.TrimRight(value, " \n\t\r")
}
