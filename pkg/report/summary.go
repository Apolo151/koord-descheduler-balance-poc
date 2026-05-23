package report

import (
	"fmt"
	"sort"
	"strings"

	"koord-descheduler-balance-poc/pkg/model"
)

type Result struct {
	Scores  []model.NodeImbalanceScore `json:"scores" yaml:"scores"`
	Plans   []model.EvictionPlan       `json:"evictionPlans" yaml:"evictionPlans"`
	Binpack *model.BinpackDecision     `json:"binpack" yaml:"binpack"`
}

func RenderSummary(result Result) string {
	var builder strings.Builder
	builder.WriteString("Summary\n")
	builder.WriteString("=======\n")
	builder.WriteString(fmt.Sprintf("Nodes scored: %d\n", len(result.Scores)))
	builder.WriteString(fmt.Sprintf("Eviction plans: %d\n", len(result.Plans)))
	if result.Binpack != nil {
		builder.WriteString(fmt.Sprintf("Binpack victims: %d\n", len(result.Binpack.Victims)))
	}
	builder.WriteString("\n")

	if len(result.Scores) > 0 {
		builder.WriteString("Node imbalance\n")
		builder.WriteString("--------------\n")
		scores := append([]model.NodeImbalanceScore{}, result.Scores...)
		sort.SliceStable(scores, func(i, j int) bool {
			return scores[i].NodeName < scores[j].NodeName
		})
		for _, score := range scores {
			builder.WriteString(fmt.Sprintf("- %s: mean=%.4f std=%.4f\n", score.NodeName, score.Mean, score.StdDev))
		}
		builder.WriteString("\n")
	}

	if len(result.Plans) > 0 {
		builder.WriteString("Eviction plans\n")
		builder.WriteString("--------------\n")
		plans := append([]model.EvictionPlan{}, result.Plans...)
		sort.SliceStable(plans, func(i, j int) bool {
			return plans[i].NodeName < plans[j].NodeName
		})
		for _, plan := range plans {
			builder.WriteString(fmt.Sprintf("- %s: candidates=%d chosen=%d\n", plan.NodeName, len(plan.Candidates), len(plan.Chosen)))
		}
		builder.WriteString("\n")
	}

	if result.Binpack != nil {
		builder.WriteString("Binpack decision\n")
		builder.WriteString("----------------\n")
		builder.WriteString(fmt.Sprintf("Victims: %s\n", strings.Join(result.Binpack.Victims, ", ")))
		builder.WriteString(fmt.Sprintf("Nodes remaining: %s\n", strings.Join(result.Binpack.NodeSet, ", ")))
	}

	return builder.String()
}
