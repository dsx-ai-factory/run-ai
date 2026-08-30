package postgres

type IsHealthyer interface {
	IsHealthy() (bool, error)
}
