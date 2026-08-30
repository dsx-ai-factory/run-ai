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
	pwdCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_password.*`
	pwdJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_password on assets_metadata.id = assets_password.asset_id`
	pwdScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	pwdGet        = pwdCols + ` ` + pwdJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + pwdScope + ` AND private_use_only = ? AND assets_metadata.deleted_at IS NULL`
	pwdGetPrivUse = pwdCols + ` ` + pwdJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + pwdScope + ` AND assets_metadata.deleted_at IS NULL`
	pwdGetDeleted = pwdCols + ` ` + pwdJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + pwdScope + ` AND private_use_only = ?`
	pwdList       = pwdCols + ` ` + pwdJoins + ` WHERE assets_metadata.kind = ? AND ` + pwdScope + ` AND private_use_only = ? AND assets_metadata.deleted_at IS NULL`
	pwdVisible    = pwdCols + ` ` + pwdJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + pwdScope + ` AND private_use_only = ? AND assets_metadata.deleted_at IS NULL`
)

func pwdRow() []dbmodel.PasswordFullRecord { return []dbmodel.PasswordFullRecord{{}} }

func TestGetPasswordById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(pwdGet, "p-1", "password", dbtesting.AnyArg(), dbtesting.AnyArg(), false).Returns(pwdRow())
	got, apiErr := dao.GetPasswordById(auth.MockAuthInfo(), "p-1", dao.CredentialsOptions{})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestGetPasswordById_IncludePrivateUse(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(pwdGetPrivUse, "p-1", "password", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(pwdRow())
	got, apiErr := dao.GetPasswordById(auth.MockAuthInfo(), "p-1", dao.CredentialsOptions{IncludePrivateUsePasswords: true})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestGetPasswordById_IncludeDeleted(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(pwdGetDeleted, "p-1", "password", dbtesting.AnyArg(), dbtesting.AnyArg(), false).Returns(pwdRow())
	got, apiErr := dao.GetPasswordById(auth.MockAuthInfo(), "p-1", dao.CredentialsOptions{IncludeDeleted: true})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListPasswords(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(pwdList, "password", dbtesting.AnyArg(), dbtesting.AnyArg(), false).Returns(pwdRow())
	got, apiErr := dao.ListPasswords(auth.MockAuthInfo(), nil, dao.CredentialsOptions{})
	require.Nil(t, apiErr)
	require.Len(t, got.Entries, 1)
}

func TestGetPasswordByIdVisible(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(pwdVisible, "p-1", "password", dbtesting.AnyArg(), dbtesting.AnyArg(), true).Returns(pwdRow())
	got, apiErr := dao.GetPasswordByIdVisible(auth.MockAuthInfo(), "p-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

var _ = api.ASSETKIND_PASSWORD

const (
	pwdInsert     = `INSERT INTO assets_password (asset_id,private_use_only,existing_secret_name,user,password,key_id) VALUES (?,?,?,?,?,?)`
	pwdCountGit   = `SELECT count(*) FROM assets_git JOIN assets_metadata ON assets_git.credential_id = assets_metadata.id WHERE assets_git.credential_id = ? AND assets_metadata.kind = ?`
	pwdCountGit2  = `SELECT count(*) FROM assets_git WHERE assets_git.password_asset_id = ?`
	pwdCountSecV  = `SELECT count(*) FROM assets_secret_volume WHERE assets_secret_volume.credential_asset_id = ?`
	pwdHardDelKey = `DELETE FROM assets_password WHERE assets_password.asset_id = ?`
)

func newPwdCreateReq() api.PasswordCreationRequest {
	return api.PasswordCreationRequest{
		Meta: api.AssetCreationRequest{Name: "my-password", Scope: api.SCOPE_PROJECT},
		Spec: api.PasswordCreationSpec{User: strptr("u"), Password: strptr("p")},
	}
}

func pwdExpectNotInUse(m dbtesting.Mocker) {
	m.ExpectQuery(pwdCountGit, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(pwdCountGit2, dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(pwdCountSecV, dbtesting.AnyArg()).Returns(zero())
	m.ExpectQuery(sqlCountEnv, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(zero())
}

func TestCreatePassword(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-password", "password")
	m.ExpectExec(pwdInsert, dbtesting.AnyArg(), false, dbtesting.AnyArg(), "u", "p", int32(1)).Returns(1)
	m.ExpectTransaction().Commit()

	got, apiErr := dao.CreatePassword(auth.MockAuthInfo(), newPwdCreateReq(), 1, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeletePassword_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	pwdExpectNotInUse(m)
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(pwdHardDelKey, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()

	apiErr := dao.DeletePassword(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_PASSWORD}, dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestUpdatePassword(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(pwdGet, dbtesting.AnyArg(), "password", dbtesting.AnyArg(), dbtesting.AnyArg(), false).Returns(pwdRow())
	m.ExpectTransaction().Commit()
	got, apiErr := dao.UpdatePassword(auth.MockAuthInfo(), "p-1", api.PasswordUpdateRequest{Meta: api.CredentialUpdateRequest{Description: strptr("d")}}, 1)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeletePassword_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	pwdExpectNotInUse(m)
	expectSoftDeleteAssetMeta(m, "password")
	m.ExpectTransaction().Commit()
	apiErr := dao.DeletePassword(auth.MockAuthInfo(), &api.AssetMeta{Id: "asset-1", Kind: api.ASSETKIND_PASSWORD}, dao.SoftDelete)
	require.Nil(t, apiErr)
}
