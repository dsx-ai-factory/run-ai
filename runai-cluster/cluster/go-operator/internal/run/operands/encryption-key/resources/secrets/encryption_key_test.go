package secrets

import (
	"context"
	"crypto/aes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiinternal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testNamespace = "runai"

func newRunaiConfig(clusterUUID string) runv1.RunaiConfig {
	return runv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace},
		Spec: runv1.RunaiConfigSpec{
			Internal: &runaiinternal.Internal{
				Global: &runaiinternal.Global{
					ControlPlane: &runaiinternal.ControlPlane{
						ClusterUUID: clusterUUID,
					},
				},
			},
		},
	}
}

func newFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("register corev1 scheme: %v", err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func objectToSecret(t *testing.T, obj client.Object) *corev1.Secret {
	t.Helper()
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		t.Fatalf("expected *corev1.Secret, got %T", obj)
	}
	return secret
}

// materializedKey returns the effective key bytes that end up in Secret.Data
// after Kubernetes merges StringData into Data on write. The fake client does
// not run that merge for us, so we do it ourselves.
func materializedKey(secret *corev1.Secret) []byte {
	if b, ok := secret.Data[constants.RunaiAdsEncryptionKeyName]; ok && len(b) > 0 {
		return b
	}
	if s, ok := secret.StringData[constants.RunaiAdsEncryptionKeyName]; ok {
		return []byte(s)
	}
	return nil
}

// TestObject_GeneratesRandomKeyIndependentOfClusterUUID is the core regression
// test for the vulnerability: two calls with the same ClusterUUID must produce
// two different keys, and neither key may equal sha256(ClusterUUID)[:32]-hex or
// any function of ClusterUUID.
func TestObject_GeneratesRandomKeyIndependentOfClusterUUID(t *testing.T) {
	const clusterUUID = "11111111-2222-3333-4444-555555555555"
	runaiConfig := newRunaiConfig(clusterUUID)

	got1 := objectToSecret(t, mustGenerate(t, runaiConfig))
	got2 := objectToSecret(t, mustGenerate(t, runaiConfig))

	key1 := materializedKey(got1)
	key2 := materializedKey(got2)

	if len(key1) != aesKeySize {
		t.Fatalf("key length = %d, want %d", len(key1), aesKeySize)
	}
	if len(key2) != aesKeySize {
		t.Fatalf("second key length = %d, want %d", len(key2), aesKeySize)
	}

	if string(key1) == string(key2) {
		t.Fatalf("two independent reconciles produced the same key %q — key is not random", key1)
	}

	hash := sha256.Sum256([]byte(clusterUUID))
	derived := hex.EncodeToString(hash[:])[:aesKeySize]
	if string(key1) == derived {
		t.Fatalf("generated key equals sha256(ClusterUUID)[:32]-hex — deterministic derivation is back")
	}
}

// TestObject_KeyIsValidAES256Material asserts the key we ship works as an
// AES-256 cipher key (aes.NewCipher requires len == 16, 24, or 32) and
// contains no NUL bytes (which would be truncated when delivered through
// SecretKeyRef -> env var at container exec time).
func TestObject_KeyIsValidAES256Material(t *testing.T) {
	runaiConfig := newRunaiConfig("any-uuid")

	secret := objectToSecret(t, mustGenerate(t, runaiConfig))
	key := materializedKey(secret)

	if _, err := aes.NewCipher(key); err != nil {
		t.Fatalf("generated key is not a valid AES key: %v", err)
	}
	for i, b := range key {
		if b == 0x00 {
			t.Fatalf("generated key contains NUL byte at index %d — will be truncated by execve()", i)
		}
	}
}

// TestObject_PreservesExistingKey verifies the sticky-key behavior: if a
// valid-length key is already persisted (e.g. from a prior reconcile, or from
// an already-deployed cluster), we do NOT overwrite it. Rotating in-place
// would render every already-encrypted personal-credential payload
// undecryptable and break every workload that consumes them.
func TestObject_PreservesExistingKey(t *testing.T) {
	runaiConfig := newRunaiConfig("uuid-does-not-matter")

	existingKey := make([]byte, aesKeySize)
	for i := range existingKey {
		existingKey[i] = byte('A' + (i % 26))
	}
	existing := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunaiAdsSecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			constants.RunaiAdsEncryptionKeyName: existingKey,
		},
	}

	kbClient := newFakeClient(t, existing)
	res := NewEncryptionKeySecretResource("encryption-key", kbClient, runaiConfig)

	obj, err := res.Object(context.Background())
	if err != nil {
		t.Fatalf("Object() error: %v", err)
	}
	got := materializedKey(objectToSecret(t, obj))
	if string(got) != string(existingKey) {
		t.Fatalf("existing key was overwritten:\n  had:  %q\n  got:  %q", existingKey, got)
	}
}

func mustGenerate(t *testing.T, runaiConfig runv1.RunaiConfig) client.Object {
	t.Helper()
	kbClient := newFakeClient(t)
	res := NewEncryptionKeySecretResource("encryption-key", kbClient, runaiConfig)

	obj, err := res.Object(context.Background())
	if err != nil {
		t.Fatalf("Object() error: %v", err)
	}
	return obj
}
