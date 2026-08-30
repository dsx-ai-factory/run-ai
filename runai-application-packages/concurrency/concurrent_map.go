package concurrency

import "github.com/puzpuzpuz/xsync/v2"

// ConcurrentMap is a generic type on top of xsync.Map. in the code we use this map extensively, as it allows
// read/write with no locking, but its value is defined as interface{} which is not comfortable to work with,
// that's why we use this generic type.
type ConcurrentMap[T any] struct {
	theMap       *xsync.Map
	defaultValue T
}

func NewConcurrentMap[T any](defaultValue T) ConcurrentMap[T] {
	return ConcurrentMap[T]{
		theMap:       xsync.NewMap(),
		defaultValue: defaultValue, // for the case of key not found, where we need to return default value + boolean
	}
}

func (cm ConcurrentMap[T]) Load(key string) (T, bool) {
	val, ok := cm.theMap.Load(key)
	if !ok {
		return cm.defaultValue, ok
	} else {
		return val.(T), ok
	}
}

func (cm ConcurrentMap[T]) Store(key string, val T) {
	cm.theMap.Store(key, val)
}

func (cm ConcurrentMap[T]) Delete(key string) {
	cm.theMap.Delete(key)
}

func (cm ConcurrentMap[T]) Size() int {
	return cm.theMap.Size()
}

func (cm ConcurrentMap[T]) Range(f func(key string, val T) bool) {
	cm.theMap.Range(func(key string, val interface{}) bool {
		return f(key, val.(T))
	})
}
