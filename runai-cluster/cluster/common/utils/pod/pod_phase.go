package pod

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/utils/structs"
	v1 "k8s.io/api/core/v1"
)

type RunaiPodPhase int

const (
	RunaiPhasePending RunaiPodPhase = iota
	RunaiPhaseInitializing
	RunaiPhaseReady
	RunaiPhaseSuccess
	RunaiPhaseFailed
	RunaiPhaseUnknown

	EmptyStr                  string = ""
	RunaiPhasePendingStr             = "Pending"
	RunaiPhaseInitializingStr        = "Initializing"
	RunaiPhaseReadyStr               = "Ready"
	RunaiPhaseSuccessStr             = "Success"
	RunaiPhaseFailedStr              = "Failed"
	RunaiPhaseUnknownStr             = "Unknown"
)

var PhaseEnumStringMap = structs.NewMap(map[RunaiPodPhase]string{
	RunaiPhasePending:      RunaiPhasePendingStr,
	RunaiPhaseInitializing: RunaiPhaseInitializingStr,
	RunaiPhaseReady:        RunaiPhaseReadyStr,
	RunaiPhaseSuccess:      RunaiPhaseSuccessStr,
	RunaiPhaseFailed:       RunaiPhaseFailedStr,
	RunaiPhaseUnknown:      RunaiPhaseUnknownStr,
})

type Extended struct {
	pod               *v1.Pod
	runaiPhase        RunaiPodPhase
	runaiPhaseMessage string
	conditionMap      map[v1.PodConditionType]*v1.PodCondition
}

type ExtendedList []Extended

func NewRunaiPodPhaseExtended(pod *v1.Pod) *Extended {
	return &Extended{
		pod:        pod,
		runaiPhase: RunaiPhaseUnknown,
	}
}

func (p *Extended) populateConditionMap() {
	if p.pod.Status.Conditions != nil {
		p.conditionMap = make(map[v1.PodConditionType]*v1.PodCondition)
		for index, condition := range p.pod.Status.Conditions {
			p.conditionMap[condition.Type] = &p.pod.Status.Conditions[index]
		}
	}
}

func (p *Extended) GetPod() *v1.Pod {
	return p.pod
}

func (p *Extended) GetPhase() RunaiPodPhase {
	return p.runaiPhase
}

func (p *Extended) GetPhaseMessage() string {
	return p.runaiPhaseMessage
}

func (p *Extended) IsPodActive() bool {
	return p.pod.Status.Phase == v1.PodRunning
}

func (e ExtendedList) Len() int {
	return len(e)
}

func (e ExtendedList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e ExtendedList) Less(i, j int) bool {
	return e[i].pod.Name < e[j].pod.Name
}

func GetPhaseStr(value RunaiPodPhase) string {
	val, _ := PhaseEnumStringMap.GetP(value)
	return val
}
func GetPhaseEnum(value string) RunaiPodPhase {
	val, ok := PhaseEnumStringMap.GetS(value)
	if !ok {
		return RunaiPhaseUnknown
	}
	return val
}
