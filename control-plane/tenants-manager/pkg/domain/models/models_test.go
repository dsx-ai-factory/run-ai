package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/tenants-manager/generated/pkg/api"
	"gorm.io/gorm"
)

var _ = Describe("Validate mapper sanity", Ordered, func() {
	It("checks tenant mapper", func() {
		tenant := Tenant{
			Model: gorm.Model{
				ID:        10,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Uuid:         uuid.New().String(),
			Name:         "test",
			DisplayName:  "Test",
			Status:       TenantCreating,
			ContractType: "test",
		}
		res := api.Tenant{
			Id:          int32(tenant.ID),
			Uuid:        tenant.Uuid,
			Name:        tenant.Name,
			DisplayName: tenant.DisplayName,
			Status:      string(tenant.Status),
			CreatedAt:   tenant.CreatedAt,
			UpdatedAt:   tenant.UpdatedAt,
		}
		Expect(res).To(Equal(tenant.ToApiModel(context.Background())))
	})

	It("when mimir id exists use it", func() {
		tenant := Tenant{
			Model: gorm.Model{
				ID:        10,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Uuid:         uuid.New().String(),
			MimirId:      "123",
			Name:         "test",
			DisplayName:  "Test",
			Status:       TenantCreating,
			ContractType: "test",
		}
		res := api.Tenant{
			Id:          int32(tenant.ID),
			Uuid:        tenant.Uuid,
			MimirId:     "123",
			Name:        tenant.Name,
			DisplayName: tenant.DisplayName,
			Status:      string(tenant.Status),
			CreatedAt:   tenant.CreatedAt,
			UpdatedAt:   tenant.UpdatedAt,
		}
		Expect(res).To(Equal(tenant.ToApiModel(context.Background())))
	})

	It("when mimir id is missing use id", func() {
		tenant := Tenant{
			Model: gorm.Model{
				ID:        10,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Uuid:         uuid.New().String(),
			Name:         "test",
			DisplayName:  "Test",
			Status:       TenantCreating,
			ContractType: "test",
		}
		res := api.Tenant{
			Id:          int32(tenant.ID),
			Uuid:        tenant.Uuid,
			MimirId:     "10",
			Name:        tenant.Name,
			DisplayName: tenant.DisplayName,
			Status:      string(tenant.Status),
			CreatedAt:   tenant.CreatedAt,
			UpdatedAt:   tenant.UpdatedAt,
		}
		Expect(res).To(Equal(tenant.ToApiModel(context.Background())))
	})
})
