package encryption

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
)

const SecretMask = "********"

// Encrypt data using the active key. returns the encrypted data as base64 string
func Encrypt(ctx context.Context, keys *TenantEncryptionKeys, data *string, label string) (*string, *api.Error) {
	activeKey := keys.Keys[keys.ActiveKeyId]

	hash := sha512.New()
	encrypted, err := rsa.EncryptOAEP(hash, rand.Reader, activeKey.Public, []byte(*data), nil)
	if err != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to encrypt %s", label), err)
	}

	return convutils.Ref(base64.StdEncoding.EncodeToString(encrypted)), nil
}

// Decrypt should be used to decrypt data which was encrypted using key with id keyId. note that for encryption
// the active key is always used, but for decryption we need to decrypt with the key which was used for encryption
func Decrypt(ctx context.Context, keys *TenantEncryptionKeys, keyId int32, encrypted *string, label string) (string, *api.Error) {
	activeKey, ok := keys.Keys[keyId]
	if !ok {
		return "", api.NewError(http.StatusPreconditionFailed, fmt.Sprintf("cannot decrypt %s: key with id %d is not found", label, keyId), nil)
	}

	decoded, err := base64.StdEncoding.DecodeString(*encrypted)
	if err != nil {
		return "", api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to decode base64 containing encrypted %s", label), err)
	}

	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, activeKey.Private, decoded, nil)
	if err != nil {
		return "", api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to decrypt %s", label), err)
	}

	return string(plaintext), nil
}

func EncryptOAEPChunks(ctx context.Context, keys *TenantEncryptionKeys, data *string, label string) (*string, *api.Error) {
	public := keys.Keys[keys.ActiveKeyId].Public

	hash := sha512.New()

	msg := []byte(*data)

	msgLen := len(msg)
	step := public.Size() - 2*hash.Size() - 2
	var encryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		chunk := msg[start:finish]

		encryptedBlockBytes, err := rsa.EncryptOAEP(hash, rand.Reader, public, chunk, nil)
		if err != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to encrypt %s", label), err)
		}

		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}

	return convutils.Ref(base64.StdEncoding.EncodeToString(encryptedBytes)), nil
}

func DecryptOAEPChunks(ctx context.Context, keys *TenantEncryptionKeys, keyId int32, encrypted *string, label string) (string, *api.Error) {
	activeKey, ok := keys.Keys[keyId]
	private := activeKey.Private

	if !ok {
		return "", api.NewError(http.StatusPreconditionFailed, fmt.Sprintf("cannot decrypt %s: key with id %d is not found", label, keyId), nil)
	}

	decoded, err := base64.StdEncoding.DecodeString(*encrypted)
	if err != nil {
		return "", api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to decode base64 containing encrypted %s", label), err)
	}

	msgLen := len(decoded)
	step := private.Size()
	var decryptedBytes []byte

	hash := sha512.New()

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		chunk := decoded[start:finish]

		decryptedBlockBytes, err := rsa.DecryptOAEP(hash, rand.Reader, private, chunk, nil)
		if err != nil {
			return "", api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to decrypt %s", label), err)
		}

		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}

	return string(decryptedBytes), nil
}
