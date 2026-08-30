package encryption

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"sync"
	"time"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"k8s.io/apimachinery/pkg/api/errors"
)

var (
	mutex          sync.Mutex
	perTenantMutex = map[int32]*sync.Mutex{}
)

const (
	KeyBitSize = 4096
)

type TenantEncryptionKey struct {
	Id      int32 // unique identifier of each key
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
	Created time.Time
}

type TenantEncryptionKeys struct {
	ActiveKeyId int32 // we always have one active key and optional obsolete keys
	Keys        map[int32]TenantEncryptionKey
}

func NewTenantEncryptionKeys() *TenantEncryptionKeys {
	return &TenantEncryptionKeys{
		ActiveKeyId: -1,
		Keys:        make(map[int32]TenantEncryptionKey),
	}
}

func GetTenantEncryptionKeys(ctx context.Context) (*TenantEncryptionKeys, *api.Error) {
	tenantId := auth.TenantIdOf(ctx)

	//
	//   we do not want two threads racing over the secret of the same tenant
	//
	tenantMutex := getTenantMutex(tenantId)
	(*tenantMutex).Lock()
	defer (*tenantMutex).Unlock()

	tenantKeys, err := LoadKeysFromSecret(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	if tenantKeys.ActiveKeyId >= 0 {
		return tenantKeys, nil
	}

	return CreateNewKeysForTenant(ctx, tenantId, tenantKeys)
}

func CreateNewKeysForTenant(ctx context.Context, tenantId int32, existingKeys *TenantEncryptionKeys) (*TenantEncryptionKeys, *api.Error) {
	privKey, err := rsa.GenerateKey(rand.Reader, KeyBitSize)
	if err != nil {
		return existingKeys, api.NewError(http.StatusServiceUnavailable, "failed to generate key", err)
	}

	existingKeys.ActiveKeyId += 1 // set the new key to be the active key

	newKeys := TenantEncryptionKey{
		Id:      existingKeys.ActiveKeyId,
		Private: privKey,
		Public:  &privKey.PublicKey,
		Created: time.Now(),
	}
	existingKeys.Keys[existingKeys.ActiveKeyId] = newKeys

	return existingKeys, StoreTenantKeysInSecret(ctx, tenantId, existingKeys)
}

func getTenantMutex(tenantId int32) *sync.Mutex {
	mutex.Lock()
	defer mutex.Unlock()

	tenantMutex, ok := perTenantMutex[tenantId]
	if !ok {
		tenantMutex = &sync.Mutex{}
		perTenantMutex[tenantId] = tenantMutex
	}
	return tenantMutex
}

// Check if cluster error is a notFound error
func notFound(err error) bool {
	if err == nil {
		return false
	}
	switch realError := err.(type) {
	case *errors.StatusError:
		return realError.ErrStatus.Code == http.StatusNotFound
	}
	return false
}

func getSecretNamespace() string {
	assetsConfig := runai_config.Get[*config.AssetsServiceConfig]()
	return assetsConfig.PodNamespace
}
