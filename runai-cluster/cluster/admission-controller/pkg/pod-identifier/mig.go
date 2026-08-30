package pod_identifier

import (
	"strings"
)

const migResourcePrefix = "nvidia.com/mig"

func HasMigDeviceRequest(annotations map[string]string) bool {
	for annotation := range annotations {
		if strings.HasPrefix(annotation, migResourcePrefix) {
			return true
		}
	}
	return false
}
