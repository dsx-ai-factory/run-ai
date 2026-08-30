package apis

import "context"

type HardTenantDeletionAPIServicer interface {
	HardDelete(ctx context.Context, tenantId string) error
}
