package converter

// NodePoolIdentifiers groups the configurable string identifiers needed when
// converting a workload's affinity/labels/annotations/project into a list of
// requested node pool names.
type NodePoolIdentifiers struct {
	// NodePoolAssignmentLabelKey is the label KEY written on a workload-side
	// object (PodGroup, Pod, Queue) to record which node pool that object is
	// assigned to.
	NodePoolAssignmentLabelKey string

	// DefaultNodepoolName is the name of the implicit "default" node pool —
	// the pool that holds workloads with no explicit node pool label. The
	// converter returns this name as the fallback when no other source yields
	// a node pool list.
	DefaultNodepoolName string

	// UnexistingNodepoolSentinel is the label VALUE (placed under
	// NodePoolAssignmentLabelKey) that marks a PodGroup as created but not
	// yet assigned to any real node pool. The converter treats this value as
	// "no assignment" rather than as a real pool name.
	UnexistingNodepoolSentinel string

	// AnnotationNodepoolsKey is the annotation KEY (on Pods) whose value is a
	// space-separated list of explicitly requested node pool names. When
	// present, it takes precedence over affinity-, label-, and project-based
	// resolution.
	AnnotationNodepoolsKey string
}
