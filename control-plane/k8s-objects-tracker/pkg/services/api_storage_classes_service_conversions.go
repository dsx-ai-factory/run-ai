package services

import (
	"encoding/json"
	"fmt"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/google/uuid"
	authclient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
)

func storageClassesDbToApi(objects []*models.Object) (*api.StorageClasses, error) {
	result := &api.StorageClasses{
		Items: make([]api.StorageClass, len(objects)),
	}
	for i, object := range objects {
		o, err := storageClassDbToApi(object)
		if err != nil {
			return nil, err
		}
		result.Items[i] = *o
	}
	return result, nil
}

func storageClassesDbToApiV2(storageClasses []*models.StorageClassWithOptionalConfig) (*api.StorageClassesV2, error) {
	result := &api.StorageClassesV2{
		Items: make([]api.StorageClassV2, len(storageClasses)),
	}

	for i, storageClass := range storageClasses {
		converted, err := storageClassDbToApiV2(storageClass)
		if err != nil {
			return nil, err
		}
		result.Items[i] = *converted
	}
	return result, nil
}

func storageClassDbToApi(obj *models.Object) (*api.StorageClass, error) {
	sc := &api.StorageClass{
		StorageClassName: *obj.Name,
		ClusterId:        obj.ClusterId.String(),
		Spec:             api.StorageClassSpec{},
	}

	if err := json.Unmarshal([]byte(obj.Spec), &sc.Spec); err != nil {
		return nil, fmt.Errorf("failed unmarshaling spec: %v", err)
	}
	return sc, nil
}

func storageClassDbToApiV2(stc *models.StorageClassWithOptionalConfig) (*api.StorageClassV2, error) {
	sc := &api.StorageClassV2{
		StorageClassName: convutils.Deref(stc.Object.Name, ""),
		ClusterId:        stc.ClusterId.String(),
		Spec:             api.StorageClassSpec{},
	}

	if stc.StorageClassConfiguration != nil {
		sc.Permissions = stc.Permissions
	}

	if err := json.Unmarshal([]byte(stc.Spec), &sc.Spec); err != nil {
		return nil, fmt.Errorf("failed unmarshaling spec: %v", err)
	}

	return sc, nil
}

func storageClassesApiToDb(tenantId int32, scs api.StorageClasses) ([]*models.Object, error) {
	result := make([]*models.Object, len(scs.Items))
	for i, sc := range scs.Items {
		var err error
		result[i], err = storageClassApiToDb(tenantId, sc)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func storageClassApiToDb(tenantId int32, sc api.StorageClass) (*models.Object, error) {
	spec, err := json.Marshal(sc.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling spec: %v", err)
	}

	clusterUuid, err := uuid.Parse(sc.ClusterId)
	if err != nil {
		return nil, fmt.Errorf("invalid cluster ID format: %v", err)
	}

	return &models.Object{
		ObjectFilters: models.ObjectFilters{
			Type:      models.ObjectTypeStorageClass,
			Name:      &sc.StorageClassName,
			TenantId:  tenantId,
			ClusterId: clusterUuid,
			ScopeType: authclient.SCOPETYPE_CLUSTER,
		},
		Spec: string(spec),
	}, nil
}

func v1StorageClassToApi(o *api.V1StorageClass, clusterUuid string) *api.StorageClass {
	return &api.StorageClass{
		StorageClassName: o.StorageClassName,
		ClusterId:        clusterUuid,
		Spec:             api.StorageClassSpec{IsDefault: o.IsDefault},
	}
}

func storageClassesToNameList(o *api.StorageClasses) []string {
	result := make([]string, len(o.Items))
	for i, object := range o.Items {
		result[i] = object.StorageClassName
	}
	return result
}

func storageClassToV1Api(o *api.StorageClass) *api.V1StorageClass {
	return &api.V1StorageClass{
		StorageClassName: o.StorageClassName,
		IsDefault:        o.Spec.IsDefault,
	}
}
