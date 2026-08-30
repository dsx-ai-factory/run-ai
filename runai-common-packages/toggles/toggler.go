package toggles

import "context"

// Toggler is the interface for checking feature flag state.
type Toggler interface {
	IsEnabled(ctx context.Context, name string) bool
	Close() error
}

// dummyToggler is used when the feature flag system is disabled (e.g. self-hosted).
// It always returns true so all features are enabled by default.
type dummyToggler struct{}

func (dt *dummyToggler) IsEnabled(ctx context.Context, name string) bool {
	return true
}

func (dt *dummyToggler) Close() error {
	return nil
}

// failedToggler is used when Init fails or there is no communication with Unleash.
// It always returns false so all features are disabled by default.
type failedToggler struct{}

func (ft *failedToggler) IsEnabled(ctx context.Context, name string) bool {
	return false
}

func (ft *failedToggler) Close() error {
	return nil
}
