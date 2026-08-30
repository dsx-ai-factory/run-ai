package structs

type Map[T comparable, K comparable] struct {
	a map[T]K
	b map[K]T
}

// Lookup a value based on the Primary key type
func (t *Map[T, K]) GetP(key T) (K, bool) {
	val, ok := t.a[key]
	return val, ok
}

// Return a value based on the Secondary key type
func (t *Map[T, K]) GetS(key K) (T, bool) {
	val, ok := t.b[key]
	return val, ok
}

// Create a new association between the primary key and secondary key
func (t *Map[T, K]) Associate(key T, value K) {
	t.a[key] = value
	t.b[value] = key
}

// Create a new map based on an existing map
// The key on this map will be considered the primary key and the value type the secondary key
func NewMap[T comparable, K comparable](input map[T]K) *Map[T, K] {
	maps := Map[T, K]{
		a: make(map[T]K),
		b: make(map[K]T),
	}

	for key, value := range input {
		maps.a[key] = value
		maps.b[value] = key
	}

	return &maps
}
