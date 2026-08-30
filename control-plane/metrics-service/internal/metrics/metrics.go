package metrics

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

func stringToSemver(v string) (*version.Version, error) {
	version, err := version.NewSemver(v)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version: %s with err: %v", v, err)
	}
	return version, nil
}
