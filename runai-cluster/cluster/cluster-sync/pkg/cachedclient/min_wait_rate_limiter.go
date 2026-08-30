package cachedclient

import (
	"sync"
	"time"
)

type MinWaitRateLimiter[T comparable] struct {
	lastTimestamp sync.Map // Using sync.Map for concurrent access to timestamps

	// delay - the delay between two handles of the same object
	delay time.Duration
}

// NewMinWaitRateLimiter creates a new ClusterSyncRateLimiter with the specified delay
func NewMinWaitRateLimiter[T comparable](delay time.Duration) *MinWaitRateLimiter[T] {
	return &MinWaitRateLimiter[T]{
		lastTimestamp: sync.Map{},
		delay:         delay,
	}
}

// When calculates when the item can be handled again, based on the delay
func (r *MinWaitRateLimiter[T]) When(item T) time.Duration {
	// Retrieve the last timestamp for the item
	if lastTimestamp, ok := r.lastTimestamp.Load(item); ok {
		// If the item is present, calculate the delay
		now := time.Now()
		lastHandledAgo := now.Sub(lastTimestamp.(time.Time))

		// If the last handled timestamp is older than the delay, return 0 for immediate processing
		if lastHandledAgo >= r.delay {
			return 0
		}

		// Otherwise, return the remaining time before the item can be handled again
		when := r.delay - lastHandledAgo
		if when > r.delay {
			return r.delay
		}
		return when
	}

	// If the item doesn't exist, treat it as a new item and update its timestamp
	r.updateLastTimestampToNow(item)
	return 0
}

// Forget removes the item from the map
func (r *MinWaitRateLimiter[T]) Forget(item T) {
	r.lastTimestamp.Delete(item)
}

// NumRequeues returns the number of times an item has been requeued (always 0 for this case)
func (r *MinWaitRateLimiter[T]) NumRequeues(_ T) int {
	return 0
}

// updateLastTimestampToNow sets the current timestamp for the item
func (r *MinWaitRateLimiter[T]) updateLastTimestampToNow(item T) {
	r.lastTimestamp.Store(item, time.Now())
}

// UpdateLastTimestampToNowIfNotForgotten updates the last timestamp of the given item to now,
// if the item is not forgotten.
// If the item is forgotten - it means it was deleted, so we don't need to continue and call the handler.
// returns - should continue with the handler or not.
func (r *MinWaitRateLimiter[T]) UpdateLastTimestampToNowIfNotForgotten(item interface{}) bool {
	if _, ok := r.lastTimestamp.Load(item); ok {
		r.lastTimestamp.Store(item, time.Now())
		return true
	}
	return false
}
