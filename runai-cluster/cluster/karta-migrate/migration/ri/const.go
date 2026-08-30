// Package ri groups operations that act on the legacy ResourceInterface
// CRD/CRs (existence check, cleanup). Constants for the RI side live here.
package ri

const (
	CRDName = "resourceinterfaces.optimization.nvidia.com"

	// Finalizer is the legacy RI finalizer that karta-migrate strips before
	// deleting an RI.
	Finalizer = "rid.kai.bolt/finalize"

	// The migration writes these labels on the Karta it creates, so it is
	// possible to trace a Karta back to the RI it came from.
	GroupLabel     = "run.ai/resource-interface-group"
	VersionLabel   = "run.ai/resource-interface-version"
	KindLabel      = "run.ai/resource-interface-kind"
	UpdatedAtLabel = "run.ai/resource-interface-updated-at"
)
