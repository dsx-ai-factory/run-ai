package clients

import (
	"encoding/json"
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func GetSpecByTrainingName(wsName string) (*openapi.TrainingSpecSpec, error) {
	value, ok := specMap[wsName]
	if !ok {
		return nil, fmt.Errorf("workspace name: %s is not in map", wsName)
	}
	var byteArr []byte
	byteArr, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	wlSpec := openapi.TrainingSpecSpec{}
	err = json.Unmarshal(byteArr, &wlSpec)
	if err != nil {
		return nil, err
	}

	return &wlSpec, nil
}
