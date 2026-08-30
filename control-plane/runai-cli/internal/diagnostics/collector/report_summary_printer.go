package collector

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"

	"github.com/charmbracelet/bubbles/table"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	"modernc.org/mathutil"
)

const defaultColumnWidth = 10

type reportSummaryPrinter struct {
	noPrevious        bool
	reportByNamespace map[string]*NamespaceReport
}

func newReportSummaryPrinter(reportByNamespaces map[string]*NamespaceReport, noPrevious bool) *reportSummaryPrinter {
	return &reportSummaryPrinter{
		noPrevious:        noPrevious,
		reportByNamespace: reportByNamespaces,
	}
}

func (rsp *reportSummaryPrinter) Subject() string {
	if rsp.noPrevious {
		return "Containers [current]:"
	}

	return "Containers [current, previous]:"
}

func (rsp *reportSummaryPrinter) Columns() []table.Column {
	return []table.Column{
		{Title: "Namespace", Width: rsp.maxNamespaceSize() + 3},
		{Title: "Total", Width: defaultColumnWidth},
		{Title: "Succeeded", Width: defaultColumnWidth},
		{Title: "Failed", Width: defaultColumnWidth},
	}
}

func (rsp *reportSummaryPrinter) Data() (data [][]string) {
	for namespace, report := range rsp.reportByNamespace {
		data = append(data, []string{namespace, strconv.Itoa(report.SuccessCount + report.FailCount), strconv.Itoa(report.SuccessCount), strconv.Itoa(report.FailCount)})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i][0] < data[j][0]
	})
	return data
}

func (rsp *reportSummaryPrinter) SummaryTable() string {
	st, err := uitable.NewTableDisplay(rsp).Generate()
	if err != nil {
		config.GetInstance().GetLogger().Error(fmt.Sprintf("failed to build namespaces summary %+v", err))
		return "failed to build namespaces summary"
	}
	return st
}

func (rsp *reportSummaryPrinter) maxNamespaceSize() int {
	namespaceSize := defaultColumnWidth
	for namespace := range rsp.reportByNamespace {
		namespaceSize = mathutil.Max(namespaceSize, len(namespace))
	}
	return namespaceSize
}
