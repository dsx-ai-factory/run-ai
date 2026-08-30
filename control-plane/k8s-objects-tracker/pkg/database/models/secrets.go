package models

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"

	"k8s.io/utils/ptr"

	"github.com/google/uuid"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm"
)

const (
	SecretsTable          = "secrets"
	SecretNameColumn      = "secret_name"
	SecretTypeColumn      = "secret_type"
	SecretOriginColumn    = "origin"
	secretTenantIdColumn  = "tenant_id"
	secretClusterIdColumn = "cluster_id"
	secretScopeTypeColumn = "scope_type"
	secretScopeIdColumn   = "scope_id"
	SecretPrivateColumn   = "private"

	// GenericSecretType is only used for adding secrets of old clusters, which do not report secret type
	// for those clusters - list secret of any type will return the same set of secrets, with their "real" type mixed
	GenericSecretType = "all"
)

type SecretFilters struct {
	TenantId   int32
	ClusterId  *string
	SecretName *string
	SecretType *string
	Origin     *string
	ScopeQuery *ScopeQueryParams `gorm:"embedded"`
}

func (s *SecretFilters) Clone() *SecretFilters {
	return &SecretFilters{
		TenantId:   s.TenantId,
		ClusterId:  s.ClusterId,
		SecretName: s.SecretName,
		SecretType: s.SecretType,
		Origin:     s.Origin,
		ScopeQuery: s.ScopeQuery,
	}
}

func (s *SecretFilters) Apply(ctx context.Context, db orm.DB, includeAncestors bool, includeDescendants bool) (orm.DB, *api.Error) {
	// we need all the filters except the scope query that we handle separately
	clonedFilters := s.Clone()
	clonedFilters.ScopeQuery = nil

	// filter by type we have to modify the original filter and add the "all" generic type
	// that old cluster stores in the table. "all" is valid for any secret type.
	filterByType := make([]string, 0)
	if clonedFilters.SecretType != nil {
		filterByType = append(filterByType, *clonedFilters.SecretType)
		filterByType = append(filterByType, GenericSecretType)

		db = db.Where(fmt.Sprintf("%s in ?", SecretTypeColumn), filterByType)

		clonedFilters.SecretType = nil
	}

	db = db.Where(clonedFilters)

	if s.ScopeQuery != nil {
		scopeQuery, args, err := s.ScopeQuery.BuildScopeQuery(ctx, includeAncestors, includeDescendants)
		if err != nil {
			return nil, err
		}
		db = db.Where(scopeQuery, args...)
	}

	return db, nil
}

func (s *SecretFilters) TableName() string {
	return SecretsTable
}

func NewSecret(filters *SecretFilters) *Secret {
	var result Secret
	if filters != nil {
		result.SecretFilters = *filters
	}

	return &result
}

type Secret struct {
	ID            uuid.UUID `gorm:"primaryKey"`
	SecretFilters `gorm:"embedded"`
	Private       *bool
}

func (s *Secret) ScopeParams() *Secret {
	return &Secret{
		SecretFilters: SecretFilters{
			TenantId:   s.TenantId,
			ClusterId:  s.ClusterId,
			ScopeQuery: s.ScopeQuery,
			Origin:     s.Origin,
		},
	}
}

func (s *Secret) BeforeCreate(_ *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}

func (s *Secret) TableName() string {
	return SecretsTable
}

func (s *Secret) TenantIdColumn() string {
	return secretTenantIdColumn
}

func (s *Secret) ClusterIdColumn() string {
	return secretClusterIdColumn
}

func (s *Secret) ScopeTypeColumn() string {
	return secretScopeTypeColumn
}

func (s *Secret) ScopeIdColumn() string {
	return secretScopeIdColumn
}

func (s *Secret) ErrorDetails() string {
	var scopeType string
	var scopeId string

	if s.ScopeQuery != nil {
		scopeType = s.ScopeQuery.ScopeType
		scopeId = s.ScopeQuery.ScopeId
	}

	return fmt.Sprintf("tenant[%d] cluster[%s] %s[%s] %s secret[%s/%s] private[%v]",
		s.TenantId,
		ptr.Deref(s.ClusterId, ""),
		scopeType,
		scopeId,
		ptr.Deref(s.Origin, ""),
		ptr.Deref(s.SecretName, ""),
		ptr.Deref(s.SecretType, ""),
		s.Private)
}

func (s *Secret) ErrorMessage(title string) string {
	return fmt.Sprintf("failed to %s secret %s", title, s.ErrorDetails())
}
