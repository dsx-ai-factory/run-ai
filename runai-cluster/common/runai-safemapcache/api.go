package safemapcache

//go:generate mockery --name Cache
type Cache[K any, V any] interface {
	Store(key K, value V)
	Load(key K) (value V, found bool, valid bool)
	Delete(key K)
	Clear()
	Destroy()
	Active() bool
}
