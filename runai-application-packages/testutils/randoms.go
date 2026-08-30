package testutils

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////////////////
// RandomValues utility functions maintain a set of random values to be used by testing.
////////////////////////////////////////////////////////////////////////////////////////////

// we hold a thread-safe counter that increases every time we want new random number
var (
	counter      = uint(rand.Int31n(100)) //nolint:gosec // G404: non-cryptographic test data
	counterMutex sync.Mutex
)

func nextRandom() uint {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	counter += uint(rand.Int31n(99) + 1) //nolint:gosec // G404: non-cryptographic test data
	return counter
}

var randomTenantId = uint(rand.Int31n(2048) + 1) //nolint:gosec // G404: non-cryptographic test data

// RandomValues is a set of random values, so tests can use them
type RandomValues struct {
	TenantId     uint
	TenantIdStr  string
	ClusterId    string
	DepartmentId string
	ProjectId    uint
	ProjectIdStr string

	IsForbidden bool // is the requesting API forbidden from seeing this?
}

func NewRandomValues() *RandomValues {
	result := RandomValues{
		TenantId:     randomTenantId,
		ClusterId:    uuid.New().String(),
		DepartmentId: fmt.Sprintf("%d", nextRandom()),
		ProjectId:    nextRandom(),
	}

	result.TenantIdStr = fmt.Sprintf("%v", result.TenantId)
	result.ProjectIdStr = fmt.Sprintf("%v", result.ProjectId)

	return &result
}

func (r RandomValues) Clone() *RandomValues {
	return &RandomValues{
		TenantId:     r.TenantId,
		TenantIdStr:  r.TenantIdStr,
		ProjectId:    r.ProjectId,
		ProjectIdStr: r.ProjectIdStr,
		DepartmentId: r.DepartmentId,
		ClusterId:    r.ClusterId,
		IsForbidden:  r.IsForbidden,
	}
}

// DifferentTenant returns the set of values with different tenant ID than whatever was originally chosen
func (r *RandomValues) CloneWithDifferentTenant() *RandomValues {
	result := r.Clone()
	result.TenantId = nextRandom()
	result.TenantIdStr = fmt.Sprintf("%v", result.TenantId)
	return result.CloneWithDifferentCluster()
}

// CloneWithDifferentCluster returns the set of values with different cluster ID than whatever was originally chosen
func (r *RandomValues) CloneWithDifferentCluster() *RandomValues {
	result := r.Clone()
	result.ClusterId = uuid.New().String()
	return result.CloneWithDifferentDepartment()
}

func (r *RandomValues) CloneWithSpecificCluster(clusterId string) *RandomValues {
	result := r.Clone()
	result.ClusterId = clusterId
	return result.CloneWithDifferentDepartment()
}

// CloneWithDifferentDepartment returns the set of values with different department ID than whatever was originally chosen
func (r *RandomValues) CloneWithDifferentDepartment() *RandomValues {
	result := r.Clone()
	result.DepartmentId = fmt.Sprintf("%d", nextRandom())
	return result.CloneWithDifferentProject()
}

// CloneWithDifferentProject returns the set of values with different cluster ID than whatever was originally chosen
func (r *RandomValues) CloneWithDifferentProject() *RandomValues {
	result := r.Clone()
	result.ProjectId = nextRandom()
	result.ProjectIdStr = fmt.Sprintf("%v", result.ProjectId)
	return result
}

func (r *RandomValues) CloneWithSpecificProject(projectId uint) *RandomValues {
	result := r.Clone()
	result.ProjectId = projectId
	result.ProjectIdStr = fmt.Sprintf("%v", result.ProjectId)
	return result
}

// Forbidden marks this set of values as forbidden for the API
func (r *RandomValues) Forbidden() *RandomValues {
	r.IsForbidden = true
	return r
}

// Allowed marks this set of values as forbidden for the API
func (r *RandomValues) Allowed() *RandomValues {
	r.IsForbidden = false
	return r
}

func (r *RandomValues) ProjectIdParam() string {
	return fmt.Sprintf("projectId=%s", r.ProjectIdStr)
}
