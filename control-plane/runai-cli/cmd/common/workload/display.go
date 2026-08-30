package workload

import (
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	controlApi "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	table2 "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	"github.com/spf13/cobra"
)

func DisplayGetWorkloadError(err error, cmd *cobra.Command, baseErrorMsg string) error {
	multiWorkloadsErr := controlApi.MultiWorkloadsError{}
	if errors.As(err, &multiWorkloadsErr) {
		multiWorkloadTable := controlApi.NewMultiWorkloadsErrorTable(multiWorkloadsErr)
		table := table2.NewTableDisplay(multiWorkloadTable)
		output, tableErr := table.Generate()
		if tableErr != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, baseErrorMsg, tableErr)
		}
		tableErr = fmt.Errorf(consts.MultiWorkloadTableDisplay, output) //nolint:staticcheck // user facing message
		return cmdtxn.CmdErrorPrintAndSet(cmd, baseErrorMsg, tableErr)
	}
	return cmdtxn.CmdErrorPrintAndSet(cmd, baseErrorMsg, err)
}
