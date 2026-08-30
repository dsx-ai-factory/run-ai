package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseStorageClassToServerScheme(t *testing.T) {
	allowVolumeExpansion := true
	storageClass := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-storage-class",
			Annotations: map[string]string{
				IsDefaultClassLabel: "true",
			},
		},
		Provisioner:          "test-provisioner",
		AllowVolumeExpansion: &allowVolumeExpansion,
	}

	result := ParseStorageClassToServerScheme(storageClass)

	// Check all fields are correctly mapped
	assert.Equal(t, storageClass.Name, result.StorageClassName)
	assert.Equal(t, storageClass.Provisioner, *result.Spec.Provisioner)
	assert.Equal(t, storageClass.AllowVolumeExpansion, result.Spec.AllowVolumeExpansion) // Changed this line
	assert.True(t, result.Spec.IsDefault)
}
