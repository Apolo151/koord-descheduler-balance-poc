package imbalance

import (
	"sort"

	"koord-descheduler-balance-poc/pkg/model"
)

func SortCandidates(candidates []model.EvictionCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].DeltaStdDev == candidates[j].DeltaStdDev {
			if candidates[i].NodeName == candidates[j].NodeName {
				return candidates[i].PodID < candidates[j].PodID
			}
			return candidates[i].NodeName < candidates[j].NodeName
		}
		return candidates[i].DeltaStdDev > candidates[j].DeltaStdDev
	})
}
