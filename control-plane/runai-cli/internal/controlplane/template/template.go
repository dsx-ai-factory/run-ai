package template

import (
	"errors"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/template/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/template/describe"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/template/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

const (
	TimerList = "template_list_timer"
)

var (
	ErrTemplateNameMissing = errors.New("template name is missing")

	lister list.TemplateLister
)

// SetClient sets the template lister client
func SetClient(templateLister list.TemplateLister) {
	lister = templateLister
}

func GetTemplateList(txn *transaction.Txn, domain, token, clusterId string, params list.GetListParams) (*list.Output, error) {
	ptnId := txn.AddTimer(TimerList)
	defer txn.StopTimer(ptnId)

	templateService, err := getTemplateListService(domain, token, clusterId)
	if err != nil {
		return nil, err
	}
	return templateService.GetTemplateList(txn.GetCtx(), params)
}

// GetId fetches template ID by name, respecting scope hierarchy.
// Validates the template supports the given workload type.
func GetId(txn *transaction.Txn, domain, token, templateName, clusterId, projectId string, workloadType spectype.Workload) (string, error) {
	if templateName == "" {
		return "", ErrTemplateNameMissing
	}

	templateService, err := getTemplateDescribeService(domain, token, clusterId)
	if err != nil {
		return "", err
	}

	template, err := templateService.FindByName(txn.GetCtx(), templateName, projectId, workloadType.APIString())
	if err != nil {
		return "", err
	}

	return template.Meta.GetId(), nil
}

func getTemplateDescribeService(domain, token, clusterId string) (*describe.Service, error) {
	// Use mock if injected (for testing)
	if lister != nil {
		if describeLister, ok := lister.(describe.Lister); ok {
			return describe.NewService(describeLister, clusterId), nil
		}
	}

	// Create real client
	templateClient, err := client.NewClient(domain, token)
	if err != nil {
		return nil, err
	}

	return describe.NewService(templateClient, clusterId), nil
}

func getTemplateListService(domain, token, clusterId string) (*list.Service, error) {
	// Use mock if injected (for testing)
	if lister != nil {
		return list.NewService(lister, clusterId), nil
	}

	// Create real client
	templateClient, err := client.NewClient(domain, token)
	if err != nil {
		return nil, err
	}

	return list.NewService(templateClient, clusterId), nil
}
