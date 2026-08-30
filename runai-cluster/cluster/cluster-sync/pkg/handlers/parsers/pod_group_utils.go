package parsers

import (
	"sort"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
)

func IsPodGroupUnscheduleable(podGroup *kaiv2alpha2.PodGroup) bool {
	if len(podGroup.Status.SchedulingConditions) > 0 {
		for _, schedulingCondition := range podGroup.Status.SchedulingConditions {
			if schedulingCondition.Type == kaiv2alpha2.UnschedulableOnNodePool {
				return true
			}
		}
	}
	return false
}

func sortPodGroupsByName(podGroups []*kaiv2alpha2.PodGroup) {
	sort.Slice(podGroups, func(i, j int) bool {
		return podGroups[i].Name < podGroups[j].Name
	})
}
