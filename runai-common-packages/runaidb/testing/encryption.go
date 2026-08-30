package testing

import (
	"context"
	"reflect"

	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/gormext"
	"gorm.io/gorm/schema"
)

const (
	encryptedSerializerName = "encrypted"
	testEncryptionKey       = "0123456789abcdef"
)

func EncryptForTest(t TestingT, plaintext string) string {
	serializer, err := gormext.NewSymmetricEncryptionSerializer(testEncryptionKey)
	if err != nil {
		t.Errorf("runaidb testing: failed to build encryption serializer: %v", err)
		return ""
	}
	value, err := serializer.Value(context.Background(), nil, reflect.Value{}, plaintext)
	if err != nil {
		t.Errorf("runaidb testing: failed to encrypt value: %v", err)
		return ""
	}
	ciphertext, _ := value.(string)
	return ciphertext
}

func registerEncryptionSerializer(t TestingT) {
	serializer, err := gormext.NewSymmetricEncryptionSerializer(testEncryptionKey)
	if err != nil {
		t.Errorf("runaidb testing: failed to build encryption serializer: %v", err)
		return
	}
	schema.RegisterSerializer(encryptedSerializerName, serializer)
}
