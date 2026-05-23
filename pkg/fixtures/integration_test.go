package fixtures

import (
	"path/filepath"
	"testing"

	"koord-descheduler-balance-poc/pkg/binpack"
	"koord-descheduler-balance-poc/pkg/descheduler"
)

func TestScenarioBasicImbalance(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "scenarios", "basic-imbalance.yaml")
	scenario, err := LoadScenario(path)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	if err := ValidateScenario(scenario); err != nil {
		t.Fatalf("validate scenario: %v", err)
	}

	snapshot, imbalanceCfg, evictionPolicy, _ := ToSnapshot(scenario)
	plans := descheduler.PlanEvictions(snapshot, imbalanceCfg, evictionPolicy)
	if len(plans) != 1 {
		t.Fatalf("expected 1 eviction plan, got %d", len(plans))
	}
	plan := plans[0]
	if len(plan.Candidates) != 2 || len(plan.Chosen) != 1 {
		t.Fatalf("expected 2 candidates and 1 chosen, got %d/%d", len(plan.Candidates), len(plan.Chosen))
	}
	if plan.Chosen[0].PodID != "default/pod-a" {
		t.Fatalf("expected default/pod-a chosen, got %s", plan.Chosen[0].PodID)
	}
}

func TestScenarioScaleDownBinpack(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "scenarios", "scale-down-binpack.yaml")
	scenario, err := LoadScenario(path)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	if err := ValidateScenario(scenario); err != nil {
		t.Fatalf("validate scenario: %v", err)
	}

	snapshot, _, _, binpackCfg := ToSnapshot(scenario)
	decision := binpack.SelectVictims(snapshot, binpackCfg)
	if len(decision.Victims) != 1 || decision.Victims[0] != "default/app-a" {
		t.Fatalf("expected victim default/app-a, got %v", decision.Victims)
	}
	if decision.ScoreBefore != 3 || decision.ScoreAfter != 2 {
		t.Fatalf("expected scores 3->2, got %.0f->%.0f", decision.ScoreBefore, decision.ScoreAfter)
	}
}

func TestScenarioNoScaleDown(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "scenarios", "no-scale-down.yaml")
	scenario, err := LoadScenario(path)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	if err := ValidateScenario(scenario); err != nil {
		t.Fatalf("validate scenario: %v", err)
	}

	_, _, _, binpackCfg := ToSnapshot(scenario)
	if binpackCfg.Victims != 0 {
		t.Fatalf("expected no binpack victims, got %d", binpackCfg.Victims)
	}
}

func TestScenarioBoundaryThreshold(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "scenarios", "boundary-threshold.yaml")
	scenario, err := LoadScenario(path)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	if err := ValidateScenario(scenario); err != nil {
		t.Fatalf("validate scenario: %v", err)
	}

	snapshot, imbalanceCfg, evictionPolicy, _ := ToSnapshot(scenario)
	plans := descheduler.PlanEvictions(snapshot, imbalanceCfg, evictionPolicy)
	if len(plans) != 1 {
		t.Fatalf("expected 1 eviction plan, got %d", len(plans))
	}
	plan := plans[0]
	if len(plan.Candidates) != 2 || len(plan.Chosen) != 1 {
		t.Fatalf("expected 2 candidates and 1 chosen, got %d/%d", len(plan.Candidates), len(plan.Chosen))
	}
}

func TestScenarioMaxEvictions(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "scenarios", "max-evictions.yaml")
	scenario, err := LoadScenario(path)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	if err := ValidateScenario(scenario); err != nil {
		t.Fatalf("validate scenario: %v", err)
	}

	snapshot, imbalanceCfg, evictionPolicy, _ := ToSnapshot(scenario)
	plans := descheduler.PlanEvictions(snapshot, imbalanceCfg, evictionPolicy)
	if len(plans) != 1 {
		t.Fatalf("expected 1 eviction plan, got %d", len(plans))
	}
	plan := plans[0]
	if len(plan.Candidates) != 3 || len(plan.Chosen) != 2 {
		t.Fatalf("expected 3 candidates and 2 chosen, got %d/%d", len(plan.Candidates), len(plan.Chosen))
	}
}

func TestScenarioSelectorBinpack(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "scenarios", "selector-binpack.yaml")
	scenario, err := LoadScenario(path)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	if err := ValidateScenario(scenario); err != nil {
		t.Fatalf("validate scenario: %v", err)
	}

	snapshot, _, _, binpackCfg := ToSnapshot(scenario)
	decision := binpack.SelectVictims(snapshot, binpackCfg)
	if len(decision.Victims) != 2 {
		t.Fatalf("expected 2 victims, got %v", decision.Victims)
	}
	if decision.Victims[0] != "default/app-a" || decision.Victims[1] != "default/app-b" {
		t.Fatalf("unexpected victim order %v", decision.Victims)
	}
	if decision.ScoreBefore != 3 || decision.ScoreAfter != 1 {
		t.Fatalf("expected scores 3->1, got %.0f->%.0f", decision.ScoreBefore, decision.ScoreAfter)
	}
}
