package field_groups

import "context"

// NodePoolNameLister supplies the node pool names offered as policy options.
//
// This is an interface rather than a direct read of the NodePool CRD on purpose.
// Reading the kai.resources NodePool CRD means importing its API package, which
// pulls KAI-scheduler's pkg/apis/kai/v1 and from there its scheduler internals - code
// written against controller-runtime >= 0.23. The control plane services compile this
// package for its field structs and pin controller-runtime to 0.22.4 via a replace
// directive, so they must not compile any of it.
//
// Only names cross this boundary, so nothing about the CRD shape leaks in. The
// implementation lives in workload-common/nodepools, which the control plane never
// compiles.
type NodePoolNameLister interface {
	// ListSchedulableNodePoolNames returns the names of the node pools a workload may
	// be submitted to, excluding any that are being deleted.
	ListSchedulableNodePoolNames(ctx context.Context) ([]string, error)
}
