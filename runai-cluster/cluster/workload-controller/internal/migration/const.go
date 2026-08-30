package migration

const (
	SuspendPatchString = `{"spec":{"active":{"value": false }}}`
)

var (
	paginationLimit int64 = 500
)

func setLimit(l int64) {
	paginationLimit = l
}
