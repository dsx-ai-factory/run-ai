package clients

import (
	"encoding/json"
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func GetSpecByDistributedName(wsName string) (*openapi.DistributedSpec, error) {
	value, ok := specMap[wsName]
	if !ok {
		return nil, fmt.Errorf("workspace name: %s is not in map", wsName)
	}
	var byteArr []byte
	byteArr, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	wlSpec := openapi.DistributedSpec{}
	err = json.Unmarshal(byteArr, &wlSpec)
	if err != nil {
		return nil, err
	}

	return &wlSpec, nil
}
