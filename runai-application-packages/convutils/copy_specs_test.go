package convutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T {
	return &v
}

// SupersetSpec is a complete spec with all the fields
type SupersetSpec struct {
	Image    *string       `json:"image,omitempty"`
	Compute  *ComputeSpec  `json:"compute,omitempty"`
	Security *SecuritySpec `json:"security,omitempty"`
	Storage  *StorageSpec  `json:"storage,omitempty"`
}

// PartialSpec has only part of the fields
type PartialSpec struct {
	Image   *string         `json:"image,omitempty"`
	Compute *ComputeSpec    `json:"compute,omitempty"`
	Storage *PartialStorage `json:"storage,omitempty"`
}

type ComputeSpec struct {
	GpuDevicesRequest *int32 `json:"gpuDevicesRequest,omitempty"`
}

type SecuritySpec struct {
	Capabilities []string `json:"capabilities,omitempty"`
}

type HostPathInstance struct {
	Path *string `json:"path,omitempty"`
}

type NfsInstance struct {
	Server *string `json:"server,omitempty"`
	Path   *string `json:"path,omitempty"`
}

type StorageSpec struct {
	HostPath []HostPathInstance `json:"hostPath,omitempty"`
	Nfs      []NfsInstance      `json:"nfs,omitempty"`
}

type PartialStorage struct {
	HostPath []HostPathInstance `json:"hostPath,omitempty"`
}

func fullSuperset() SupersetSpec {
	return SupersetSpec{
		Image:    ptr("my-image"),
		Compute:  &ComputeSpec{GpuDevicesRequest: ptr(int32(3))},
		Security: &SecuritySpec{Capabilities: []string{"AUDIT_CONTROL", "KILL"}},
		Storage: &StorageSpec{
			HostPath: []HostPathInstance{{Path: ptr("my-host-path")}},
			Nfs:      []NfsInstance{{Server: ptr("nfs.example.com"), Path: ptr("/exports/data")}},
		},
	}
}

func TestCopySpecsSupersetToSuperset(t *testing.T) {
	src := fullSuperset()

	var dest SupersetSpec
	assert.NoError(t, CopySpecs(&dest, &src))

	assert.Equal(t, src, dest)
}

func TestCopySpecsSupersetToPartial(t *testing.T) {
	src := fullSuperset()

	var partial PartialSpec
	assert.NoError(t, CopySpecs(&partial, &src))

	// Only the fields that exist on PartialSpec are carried over; the rest are dropped.
	assert.Equal(t, src.Image, partial.Image)
	assert.Equal(t, src.Compute, partial.Compute)
	// Nested subset: PartialStorage keeps HostPath, drops Nfs.
	assert.Equal(t, src.Storage.HostPath, partial.Storage.HostPath)
}

func TestCopySpecsPartialToSuperset(t *testing.T) {
	partial := PartialSpec{
		Image:   ptr("my-image"),
		Compute: &ComputeSpec{GpuDevicesRequest: ptr(int32(3))},
		Storage: &PartialStorage{HostPath: []HostPathInstance{{Path: ptr("my-host-path")}}},
	}

	var dest SupersetSpec
	assert.NoError(t, CopySpecs(&dest, &partial))

	assert.Equal(t, partial.Image, dest.Image)
	assert.Equal(t, partial.Compute, dest.Compute)
	// PartialStorage's HostPath lands on the superset's storage; Nfs stays empty.
	assert.Equal(t, partial.Storage.HostPath, dest.Storage.HostPath)
	assert.Nil(t, dest.Storage.Nfs)
	// Fields absent from PartialSpec stay zero-valued on the superset.
	assert.Nil(t, dest.Security)
}

func TestIsEmpty(t *testing.T) {
	assert.True(t, IsEmpty(SupersetSpec{}))
	assert.False(t, IsEmpty(SupersetSpec{Image: ptr("my-image")}))
}
