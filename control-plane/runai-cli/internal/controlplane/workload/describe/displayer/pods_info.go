package displayer

type PodsInfo struct {
	Pods        []Pod `json:"pods"`
	RunningPods int32 `json:"runningPods"`
}
