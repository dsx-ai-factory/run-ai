package page

type DuplicatorMapper interface {
	ShouldAdd(id string) bool
	ReachedLimit() bool
}
