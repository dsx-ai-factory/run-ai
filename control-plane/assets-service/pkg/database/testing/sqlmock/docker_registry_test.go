package sqlmock

import (
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	drCols    = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_docker_registry.*`
	drJoins   = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_docker_registry on assets_metadata.id = assets_docker_registry.asset_id`
	drScope   = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`
	drPrivate = `(COALESCE(assets_metadata.private, FALSE) = false OR assets_metadata.created_by = ?)`

	drGet        = drCols + ` ` + drJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + drScope + ` AND ` + drPrivate + ` AND assets_metadata.deleted_at IS NULL`
	drGetDeleted = drCols + ` ` + drJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + drScope + ` AND ` + drPrivate
	drList       = drCols + ` ` + drJoins + ` WHERE assets_metadata.kind = ? AND ` + drScope + ` AND assets_metadata.deleted_at IS NULL`
	drListDel    = drCols + ` ` + drJoins + ` WHERE assets_metadata.kind = ? AND ` + drScope

	drInsert  = `INSERT INTO assets_docker_registry (asset_id,existing_secret_name,user,password,url,key_id) VALUES (?,?,?,?,?,?)`
	drHardDel = `DELETE FROM assets_docker_registry WHERE assets_docker_registry.asset_id = ?`
)

func drRow() []dbmodel.DockerRegistryFullRecord { return []dbmodel.DockerRegistryFullRecord{{}} }

func newDRCreateReq() api.DockerRegistryCreationRequest {
	return api.DockerRegistryCreationRequest{
		Meta: api.AssetCreationRequest{Name: "my-dr", Scope: api.SCOPE_PROJECT},
		Spec: api.DockerRegistryCreationSpec{User: strptr("u"), Password: strptr("p"), Url: strptr("url")},
	}
}

func TestGetDockerRegistryById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(drGet, "d-1", "dockerRegistry", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(drRow())
	got, apiErr := dao.GetDockerRegistryById(auth.MockAuthInfo(), "d-1", dao.CredentialsOptions{})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestGetDockerRegistryById_IncludeDeleted(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(drGetDeleted, "d-1", "dockerRegistry", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(drRow())
	got, apiErr := dao.GetDockerRegistryById(auth.MockAuthInfo(), "d-1", dao.CredentialsOptions{IncludeDeleted: true})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListDockerRegistries(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(drList, "dockerRegistry", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(drRow())
	got, apiErr := dao.ListDockerRegistries(auth.MockAuthInfo(), nil, dao.CredentialsOptions{})
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestListDockerRegistries_IncludeDeleted(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(drListDel, "dockerRegistry", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(drRow())
	_, apiErr := dao.ListDockerRegistries(auth.MockAuthInfo(), nil, dao.CredentialsOptions{IncludeDeleted: true})
	require.Nil(t, apiErr)
}

func TestCreateDockerRegistry(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-dr", "dockerRegistry")
	m.ExpectExec(drInsert, dbtesting.AnyArg(), dbtesting.AnyArg(), "u", "p", "url", int32(1)).Returns(1)
	m.ExpectTransaction().Commit()
	got, apiErr := dao.CreateDockerRegistry(auth.MockAuthInfo(), newDRCreateReq(), 1, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteDockerRegistry_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(drHardDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteDockerRegistry(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_DOCKER_REGISTRY}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestUpdateDockerRegistry(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectTransaction().Commit()
	m.ExpectQuery(drGet, dbtesting.AnyArg(), "dockerRegistry", dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(drRow())
	got, apiErr := dao.UpdateDockerRegistry(auth.MockAuthInfo(), "d-1", api.DockerRegistryUpdateRequest{Meta: api.CredentialUpdateRequest{Description: strptr("d")}}, 1)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteDockerRegistry_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteAssetMeta(m, "dockerRegistry")
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteDockerRegistry(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_DOCKER_REGISTRY}, dao.SoftDelete)
	require.Nil(t, apiErr)
}
