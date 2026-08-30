package datavolume

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	invalidFormatMsg = "invalid datavolume format '%s': required format is name=NAME,mountpath=PATH"
)

type ParsedDataVolume struct {
	Name      string
	MountPath string
}

func ParseDataVolumeList(dvList []string) ([]ParsedDataVolume, error) {
	var parsedList = make([]ParsedDataVolume, 0)
	for _, dv := range dvList {
		parsed, err := ParseDataVolume(dv)
		if err != nil {
			return nil, err
		}
		parsedList = append(parsedList, *parsed)
	}
	return parsedList, nil
}

func ParseDataVolume(dv string) (*ParsedDataVolume, error) {
	invalidFormatErr := fmt.Errorf(invalidFormatMsg, dv)

	dp := NewParser(dv)
	err := dp.Parse()
	if err != nil {
		return nil, invalidFormatErr
	}

	err = parser.CheckMandatoryFields(dp, Name, MountPath)
	if err != nil {
		return nil, invalidFormatErr
	}

	return &ParsedDataVolume{
		Name:      strings.TrimSpace(*dp.Get(Name)),
		MountPath: strings.TrimSpace(*dp.Get(MountPath)),
	}, nil
}

func NewDataVolumeInstanceList(parsedList []ParsedDataVolume, nameToIdMap map[string]string) ([]openapi.DataVolumeInstance, error) {
	var instanceList = make([]openapi.DataVolumeInstance, 0)
	for _, parsed := range parsedList {
		id, ok := nameToIdMap[parsed.Name]
		if !ok {
			return nil, fmt.Errorf("data volume '%s' not found", parsed.Name)
		}
		instance := openapi.NewDataVolumeInstance()
		instance.SetId(id)
		instance.SetMountPath(parsed.MountPath)
		instanceList = append(instanceList, *instance)
	}
	return instanceList, nil
}
