package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"koord-descheduler-balance-poc/pkg/binpack"
	"koord-descheduler-balance-poc/pkg/descheduler"
	"koord-descheduler-balance-poc/pkg/fixtures"
	"koord-descheduler-balance-poc/pkg/imbalance"
	"koord-descheduler-balance-poc/pkg/model"
	"koord-descheduler-balance-poc/pkg/report"
)

func main() {
	scenarioPath := flag.String("scenario", "", "path to scenario YAML/JSON")
	outJSON := flag.String("out-json", "", "output JSON file (default: skip)")
	outSummary := flag.String("out-summary", "", "output summary file (default: stdout)")
	flag.Parse()

	if *scenarioPath == "" {
		fmt.Fprintln(os.Stderr, "--scenario is required")
		os.Exit(2)
	}

	scenario, err := fixtures.LoadScenario(*scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load scenario: %v\n", err)
		os.Exit(1)
	}
	if err := fixtures.ValidateScenario(scenario); err != nil {
		fmt.Fprintf(os.Stderr, "validate scenario: %v\n", err)
		os.Exit(1)
	}

	snapshot, imbalanceCfg, evictionPolicy, binpackCfg := fixtures.ToSnapshot(scenario)
	result := report.Result{
		Scores:  computeScores(snapshot, imbalanceCfg),
		Plans:   descheduler.PlanEvictions(snapshot, imbalanceCfg, evictionPolicy),
		Binpack: nil,
	}

	if binpackCfg.Victims > 0 && (binpackCfg.TargetWorkload != "" || len(binpackCfg.Selector) > 0) {
		decision := binpack.SelectVictims(snapshot, binpackCfg)
		result.Binpack = &decision
	}

	if *outJSON != "" {
		if err := writeJSON(*outJSON, result); err != nil {
			fmt.Fprintf(os.Stderr, "write json: %v\n", err)
			os.Exit(1)
		}
	}

	summary := report.RenderSummary(result)
	if *outSummary != "" {
		if err := os.WriteFile(*outSummary, []byte(summary), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write summary: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(summary)
	}
}

func writeJSON(path string, result report.Result) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func computeScores(snapshot *model.ClusterSnapshot, cfg model.ImbalanceConfig) []model.NodeImbalanceScore {
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
