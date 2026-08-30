package utils

import "github.com/google/uuid"

type UuidGenerator interface {
	Get(idx int) string
}

// for production, we use a real UUID generator, but for testing we will replace
// it with a mock UUID generator
type realUuidGenerator struct {
	UuidGenerator
}

func (_ *realUuidGenerator) Get(_ int) string {
	return uuid.New().String()
}

// ActiveUuidGenerator is the actual UUID generator that is being used
var ActiveUuidGenerator UuidGenerator = &realUuidGenerator{}
