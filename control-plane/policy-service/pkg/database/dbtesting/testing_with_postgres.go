package dbtesting

import (
	"fmt"

	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

const (
	syncUpdateCounter = "effective_policies_update_count"
)

func RestartUpdateCounter() {
	dao_utils.DB().Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", syncUpdateCounter))
}
