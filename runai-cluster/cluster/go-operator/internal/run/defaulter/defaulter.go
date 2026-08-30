package defaulter

// Defaulter is an interface types must implement to default their fields
type Defaulter interface {
	// SetDefaultsWhereNeeded iterates over all struct fields and sets defaults for missing values
	SetDefaultsWhereNeeded()
}
