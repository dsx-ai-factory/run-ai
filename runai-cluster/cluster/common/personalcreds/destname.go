package personalcreds

import "fmt"

// DestName is the deterministic Secret name in the workload namespace that the
// personalcreds plugin materializes for a given (runner, source) pair. The CP
// composer and the cluster plugin compute this identically.
func DestName(runner, sourceSecret string) string {
	return fmt.Sprintf("pc-%s-%s", runner, sourceSecret)
}
