package topology

import (
	"fmt"
	"strconv"
)

// MaxSegmentTopologySubgroups limits the number of KAI subgroups generated
// for one segmented workload.
const MaxSegmentTopologySubgroups int64 = 10_000

// ParseSegmentSize parses the KAI segment-size annotation value.
func ParseSegmentSize(value string) (int64, error) {
	segmentSize, err := strconv.ParseInt(value, 10, 32)
	if err != nil || segmentSize < 1 {
		return 0, fmt.Errorf("must be a positive integer")
	}
	return segmentSize, nil
}

// SegmentCount validates the segment size and returns the number of subgroups
// needed for the configured number of workers.
func SegmentCount(workerReplicas, segmentSize int64) (int64, error) {
	if workerReplicas < 0 {
		return 0, fmt.Errorf("worker replicas must be non-negative")
	}
	if segmentSize < 1 {
		return 0, fmt.Errorf("segment size must be a positive integer")
	}

	// Avoid workerReplicas+segmentSize-1 here: callers can pass int32-sized
	// values and that ceiling-division form can overflow before being widened.
	segmentCount := workerReplicas / segmentSize
	if workerReplicas%segmentSize != 0 {
		segmentCount++
	}
	if segmentCount > MaxSegmentTopologySubgroups {
		return 0, fmt.Errorf(
			"segment topology-aware scheduling would create %d segments, exceeding the maximum of %d",
			segmentCount,
			MaxSegmentTopologySubgroups,
		)
	}

	return segmentCount, nil
}
