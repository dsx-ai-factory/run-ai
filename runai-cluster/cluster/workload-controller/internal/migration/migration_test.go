/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"time"
)

const (
	configName      = "test-config"
	configNamespace = "test-namespace"
	migratorName    = "test-migrator"
)

var _ = Describe("WorkloadMigration", func() {
	var (
		workloadMigration *WorkloadMigration
		fakeClient        client.Client
		mockCtrl          *gomock.Controller
		mockMigrator      *mocks.Migrator
		mockServer        *mocks.WebhookServer
	)

	BeforeEach(func() {
		fakeClient = fake.NewClientBuilder().Build()
		mockCtrl = gomock.NewController(GinkgoT())

		mockMigrator = &mocks.Migrator{}
		mockMigrator.On("RequiresWebhooks").Return(false)
		mockMigrator.On("String").Return(migratorName)

		workloadMigration = &WorkloadMigration{
			k8sClient:                fakeClient,
			migrationSteps:           []Migrator{mockMigrator},
			migrationConfigMapName:   configName,
			migrationConfigNamespace: configNamespace,
			webserverGraceDelay:      1 * time.Millisecond,
		}

		mockServer = &mocks.WebhookServer{}
		mockServer.On("Start", mock.Anything).Return(nil).Once()
		mockServer.On("StartedChecker", mock.Anything).Return(healthz.Ping).Once()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("when running migration steps", func() {
		It("should return nil if migration steps are successful", func() {
			mockMigrator.On("Migrate", mock.Anything).Return(nil)
			err := workloadMigration.Run(context.Background(), mockServer)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should update config map", func() {
			mockMigrator.On("Migrate", mock.Anything).Return(nil)
			err := workloadMigration.Run(context.Background(), mockServer)
			Expect(err).NotTo(HaveOccurred())
			configMap, err := workloadMigration.GetMigrationConfigMap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap.Data[migratorName]).To(Equal("true"))
		})
		It("should return an error if migration steps fail", func() {
			mockMigrator.On("Migrate", mock.Anything).Return(errors.New("migration error"))
			err := workloadMigration.Run(context.Background(), mockServer)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when webhook server is required", func() {
		BeforeEach(func() {
			mockMigrator.On("RequiresWebhooks").Unset()
			mockMigrator.On("RequiresWebhooks").Return(true)
		})

		It("webhook server is starting", func() {
			mockMigrator.On("Migrate", mock.Anything).Return(nil)
			err := workloadMigration.Run(context.Background(), mockServer)
			Expect(err).NotTo(HaveOccurred())
		})

		It("webhook server fails to start - timeout", func() {
			var failingChecker healthz.Checker
			failingChecker = func(r *http.Request) error { return errors.New("i'm a failure") }
			mockServer.On("StartedChecker", mock.Anything).Unset()
			mockServer.On("StartedChecker", mock.Anything).Return(failingChecker).Once()

			mockMigrator.On("Migrate", mock.Anything).Return(nil)

			workloadMigration.webserverGraceDelay = 1 * time.Second

			err := workloadMigration.Run(context.Background(), mockServer)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(webserverGracePeriodTimeoutMessage))
		})

		It("webhook server fails to start - err", func() {
			startErr := errors.New("i'm a failure")
			mockMigrator.On("Migrate", mock.Anything).Return(startErr)

			err := workloadMigration.Run(context.Background(), mockServer)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(startErr))
		})
	})

	Context("when getting migration config map", func() {
		It("should return error if config map does not exist", func() {
			_, err := workloadMigration.GetMigrationConfigMap(context.Background())
			Expect(err).To(HaveOccurred())
		})
		Context("when config map exists", func() {

			It("should return config map if it exists", func() {

				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      configName,
						Namespace: configNamespace,
					},
				}
				Expect(fakeClient.Create(context.Background(), configMap)).To(Succeed())
				retrievedConfigMap, err := workloadMigration.GetMigrationConfigMap(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(retrievedConfigMap.Name).To(Equal(configMap.Name))
				Expect(retrievedConfigMap.Namespace).To(Equal(configMap.Namespace))
			})
		})
		It("should run migration steps if config map exists", func() {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configName,
					Namespace: configNamespace,
				},
			}
			Expect(fakeClient.Create(context.Background(), configMap)).To(Succeed())
			mockMigrator.On("Migrate", mock.Anything).Return(nil)
			err := workloadMigration.Run(context.Background(), mockServer)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should not run migration steps if config map exists and migration step is done", func() {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configName,
					Namespace: configNamespace,
				},
				Data: map[string]string{
					migratorName: "true",
				},
			}
			Expect(fakeClient.Create(context.Background(), configMap)).To(Succeed())
			err := workloadMigration.Run(context.Background(), mockServer)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
