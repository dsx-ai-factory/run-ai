package patch

import (
	v1 "k8s.io/api/core/v1"
)

func AddEnvVarToContainer(container *v1.Container, envVar v1.EnvVar) {
	updatedEnvs := make([]v1.EnvVar, 0)
	for _, variable := range container.Env {
		if variable.Name != envVar.Name {
			updatedEnvs = append(updatedEnvs, variable)
		}
	}
	updatedEnvs = append(updatedEnvs, envVar)
	container.Env = updatedEnvs
}
