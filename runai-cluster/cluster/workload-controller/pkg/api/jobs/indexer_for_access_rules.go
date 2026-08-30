package jobs

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func IndexAccessRules(mgr manager.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &v1alpha1.AccessRule{},
		common.PrivilegedWorkloadCrudPermissionsField, common.PrivilegedWorkloadCrudPermissionsIndexer)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed indexing accessrule field %s",
			common.PrivilegedWorkloadCrudPermissionsField))
		return err
	}

	return nil
}
