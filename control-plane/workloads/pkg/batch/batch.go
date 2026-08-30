package batch

import openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

func NewBatchResponseFailure(id string, message string, code int32) openapi.BatchResponseFailureItem {
	return openapi.BatchResponseFailureItem{
		Id:      id,
		Message: message,
		Code:    code,
	}
}

// BatchIds splits a list of ids into batches of the given size.
func BatchIds(ids []string, batchSize int) [][]string {
	var batches [][]string
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batches = append(batches, ids[i:end])
	}
	return batches
}
