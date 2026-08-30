package source

import (
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type KartaConfiguration struct {
	Category       string
	Priority       string
	Preemptibility openapi.WorkloadTypePreemptibility
	Kartas         []models.Karta
}
