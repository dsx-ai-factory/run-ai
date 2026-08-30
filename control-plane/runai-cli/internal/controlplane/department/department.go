package department

import (
	"fmt"
	"sync"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/department/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/department/model"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/department/service"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

const (
	TimerList = "department_list_timer"
)

var departmentClient service.DepartmentClient
var departmentService *service.DepartmentService

var getDepartmentServiceOnce sync.Once

func SetClient(dc service.DepartmentClient) {
	departmentClient = dc
	departmentService = service.NewDepartmentService(departmentClient)
}

func GetDepartmentService(domain, token string) (*service.DepartmentService, error) {
	var err error
	getDepartmentServiceOnce.Do(func() {
		if departmentClient != nil {
			return
		}
		departmentClient, err = client.NewDepartmentClient(domain, token)
		if err != nil {
			return
		}
		departmentService = service.NewDepartmentService(departmentClient)
	})
	if err != nil {
		return nil, err
	}
	return departmentService, nil
}

func GetDepartmentList(txn *transaction.Txn, domain, token string, filters []model.Filter, limit, offset int32) (*service.BriefDepartmentHandler, error) {
	var err error
	ptnId := txn.AddTimer(TimerList)
	defer txn.StopTimer(ptnId)

	ds, err := GetDepartmentService(domain, token)
	if err != nil {
		return nil, err
	}
	return ds.GetDepartmentList(txn.GetCtx(), filters, limit, offset)
}

// GetDepartmentId returns the department ID by the department name
func GetDepartmentId(txn *transaction.Txn, domain, token, clusterUuid, departmentName string) (string, error) {
	if departmentName == "" {
		return "", fmt.Errorf(consts.ErrorNoDepartmentName)
	}

	filters := []model.Filter{
		*model.NewClusterIDFilter(clusterUuid),
		*model.NewDepartmentFilter(departmentName),
	}
	handler, err := GetDepartmentList(txn, domain, token, filters, 1, 0)
	if err != nil {
		return "", err
	}

	for _, department := range handler.GetList() {
		if department.Name == departmentName {
			return department.Id, nil
		}
	}

	return "", fmt.Errorf(consts.ErrorNoDepartmentFound, departmentName)
}
