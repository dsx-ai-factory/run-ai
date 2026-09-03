/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package generator_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/secret-generator/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/secret-generator/pkg/generator"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// stringPtr returns a pointer to the given string value.
func stringPtr(s string) *string {
	return &s
}

var _ = Describe("SecretGenerator", func() {
	var (
		ctx        context.Context
		fakeClient client.Client
		cfg        *config.ConfigData
		tempDir    string
		configFile string
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Create temp directory for config files
		var err error
		tempDir, err = os.MkdirTemp("", "secret-generator-test")
		Expect(err).NotTo(HaveOccurred())

		configFile = filepath.Join(tempDir, "secrets.json")

		// Default config
		cfg = &config.ConfigData{
			Namespace:        "test-namespace",
			ConfigFilePath:   configFile,
			FullnameOverride: "runai-backend",
			LogLevel:         "Info",
			DryRun:           false,
		}
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	// Helper function to create fake client with objects
	setupFakeClient := func(objects ...runtime.Object) {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(objects...).
			Build()
	}

	// Helper function to write generator config
	writeGeneratorConfig := func(genConfig generator.GeneratorConfig) {
		data, err := json.Marshal(genConfig)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(configFile, data, 0644)).To(Succeed())
	}

	Describe("NewSecretGenerator", func() {
		It("should create a new SecretGenerator with the provided config", func() {
			setupFakeClient()
			sg := generator.NewSecretGenerator(fakeClient, cfg)
			Expect(sg).NotTo(BeNil())
		})
	})

	Describe("Run", func() {
		Context("when no secrets are configured", func() {
			It("should succeed with empty configuration", func() {
				setupFakeClient()
				writeGeneratorConfig(generator.GeneratorConfig{Secrets: []generator.SecretSpec{}})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when config file does not exist", func() {
			It("should return an error", func() {
				setupFakeClient()

				cfg.ConfigFilePath = "/nonexistent/path/secrets.json"
				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to load generator config"))
			})
		})

		Context("when config file is invalid JSON", func() {
			It("should return an error", func() {
				setupFakeClient()

				Expect(os.WriteFile(configFile, []byte("invalid json"), 0644)).To(Succeed())

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to load generator config"))
			})
		})

		Context("fresh install - no existing secrets", func() {
			It("should create new secrets with generated values", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password", Length: 32},
								{Name: "api-key", Type: "alphanumeric", Length: 24},
								{Name: "username", Type: "static", ChartValue: stringPtr("admin")},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify secret was created
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Verify keys exist
				Expect(secret.Data).To(HaveKey("password"))
				Expect(secret.Data).To(HaveKey("api-key"))
				Expect(secret.Data).To(HaveKey("username"))

				// Verify static value
				Expect(string(secret.Data["username"])).To(Equal("admin"))

				// Verify password length
				Expect(len(secret.Data["password"])).To(Equal(32))

				// Verify api-key length
				Expect(len(secret.Data["api-key"])).To(Equal(24))

				// Verify labels
				Expect(secret.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", "secret-generator"))
				Expect(secret.Labels).To(HaveKeyWithValue("app.kubernetes.io/part-of", "runai-backend"))

				// Verify no owner references (secrets should persist independently)
				Expect(secret.OwnerReferences).To(BeEmpty())
			})

			It("should create multiple secrets", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "secret-1",
							Keys: []generator.SecretKeySpec{
								{Name: "key1", Type: "password", Length: 16},
							},
						},
						{
							Name: "secret-2",
							Keys: []generator.SecretKeySpec{
								{Name: "key2", Type: "uuid"},
							},
						},
						{
							Name: "secret-3",
							Keys: []generator.SecretKeySpec{
								{Name: "key3", Type: "hex", Length: 32},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify all secrets were created
				for _, name := range []string{"secret-1", "secret-2", "secret-3"} {
					secret := &corev1.Secret{}
					err = fakeClient.Get(ctx, client.ObjectKey{
						Namespace: "test-namespace",
						Name:      name,
					}, secret)
					Expect(err).NotTo(HaveOccurred())
				}
			})
		})

		Context("CopyFrom functionality", func() {
			It("should copy value from one key to another", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret-with-copy",
							Keys: []generator.SecretKeySpec{
								{Name: "source-key", Type: "alphanumeric", Length: 32},
								{Name: "copied-key", CopyFrom: "source-key"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify secret was created
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret-with-copy",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Verify both keys exist
				Expect(secret.Data).To(HaveKey("source-key"))
				Expect(secret.Data).To(HaveKey("copied-key"))

				// Verify copied value matches source
				Expect(string(secret.Data["copied-key"])).To(Equal(string(secret.Data["source-key"])))

				// Verify source value has correct length
				Expect(len(secret.Data["source-key"])).To(Equal(32))
			})

			It("should support multiple keys copying from the same source", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "nats-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "pubsub-password", Type: "alphanumeric", Length: 32},
								{Name: "cache-password", CopyFrom: "pubsub-password"},
								{Name: "another-copy", CopyFrom: "pubsub-password"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify secret was created
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "nats-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Verify all keys have the same value
				sourceValue := string(secret.Data["pubsub-password"])
				Expect(string(secret.Data["cache-password"])).To(Equal(sourceValue))
				Expect(string(secret.Data["another-copy"])).To(Equal(sourceValue))
			})

			It("should return error when CopyFrom references itself", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "self-ref-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "self-ref", CopyFrom: "self-ref"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no resolver in chain could resolve key self-ref"))
			})

			It("should return error when CopyFrom creates a cycle", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "cycle-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "key-a", CopyFrom: "key-b"},
								{Name: "key-b", CopyFrom: "key-a"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no resolver in chain could resolve key key-a"))
			})

			It("should return error when CopyFrom references non-existent key", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "invalid-copy-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "valid-key", Type: "password", Length: 16},
								{Name: "invalid-copy", CopyFrom: "non-existent-key"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("resolve"))
			})

			It("should work with static values as source", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "static-copy-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "username", Type: "static", ChartValue: stringPtr("admin")},
								{Name: "username-copy", CopyFrom: "username"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify secret was created
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "static-copy-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Verify both keys have the static value
				Expect(string(secret.Data["username"])).To(Equal("admin"))
				Expect(string(secret.Data["username-copy"])).To(Equal("admin"))
			})
		})

		Context("upgrade - generator-managed secret already exists", func() {
			It("should preserve existing secret without modification when all keys present", func() {
				existingSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "existing-secret",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"app.kubernetes.io/managed-by": "secret-generator",
						},
					},
					Data: map[string][]byte{
						"password": []byte("original-password-value"),
					},
				}
				setupFakeClient(existingSecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "existing-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password", Length: 32},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify secret still has original value
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "existing-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(secret.Data["password"])).To(Equal("original-password-value"))
			})

			It("should reconcile existing secret by adding missing keys on upgrade", func() {
				existingSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "existing-secret",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"app.kubernetes.io/managed-by": "secret-generator",
						},
					},
					Data: map[string][]byte{
						"password": []byte("original-password-value"),
					},
				}
				setupFakeClient(existingSecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "existing-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password", Length: 32},
								{Name: "new-api-key", Type: "alphanumeric", Length: 24},
								{Name: "new-static", Type: "static", ChartValue: stringPtr("static-value")},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify secret was updated
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "existing-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Original key preserved
				Expect(string(secret.Data["password"])).To(Equal("original-password-value"))

				// New keys added
				Expect(secret.Data).To(HaveKey("new-api-key"))
				Expect(len(secret.Data["new-api-key"])).To(Equal(24))
				Expect(string(secret.Data["new-api-key"])).To(MatchRegexp(`^[a-zA-Z0-9]+$`))
				Expect(string(secret.Data["new-static"])).To(Equal("static-value"))
			})

			It("should reconcile existing secret with CopyFrom referencing existing key", func() {
				existingSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "existing-secret",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"app.kubernetes.io/managed-by": "secret-generator",
						},
					},
					Data: map[string][]byte{
						"AUTH_TOKEN": []byte("existing-token-value"),
					},
				}
				setupFakeClient(existingSecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "existing-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
								{Name: "cache-password", CopyFrom: "AUTH_TOKEN"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "existing-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Original key preserved
				Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("existing-token-value"))

				// CopyFrom key should copy from existing key
				Expect(string(secret.Data["cache-password"])).To(Equal("existing-token-value"))
			})
		})

		Context("upgrade - migration from legacy Helm-managed secret", func() {
			It("should migrate values from legacy secret with same key names", func() {
				legacySecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "legacy-secret",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"app.kubernetes.io/managed-by": "Helm",
						},
					},
					Data: map[string][]byte{
						"password": []byte("legacy-password"),
						"api-key":  []byte("legacy-api-key"),
					},
				}
				setupFakeClient(legacySecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:             "new-secret",
							LegacySecretName: "legacy-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
								{Name: "api-key", Type: "alphanumeric"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify new secret was created with migrated values
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "new-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(secret.Data["password"])).To(Equal("legacy-password"))
				Expect(string(secret.Data["api-key"])).To(Equal("legacy-api-key"))
			})

			It("should migrate values with different legacy key names", func() {
				legacySecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "legacy-secret",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"old-password-key": []byte("migrated-password"),
						"old-api-key":      []byte("migrated-api-key"),
					},
				}
				setupFakeClient(legacySecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:             "new-secret",
							LegacySecretName: "legacy-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", LegacyKeyName: "old-password-key", Type: "password"},
								{Name: "api-key", LegacyKeyName: "old-api-key", Type: "alphanumeric"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify new secret has values under new key names
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "new-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(secret.Data["password"])).To(Equal("migrated-password"))
				Expect(string(secret.Data["api-key"])).To(Equal("migrated-api-key"))
			})

			It("should generate new values for missing keys in legacy secret", func() {
				legacySecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "legacy-secret",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"password": []byte("existing-password"),
						// missing "new-key"
					},
				}
				setupFakeClient(legacySecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:             "new-secret",
							LegacySecretName: "legacy-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
								{Name: "new-key", Type: "alphanumeric", Length: 20},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify new secret
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "new-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Existing key should be migrated
				Expect(string(secret.Data["password"])).To(Equal("existing-password"))

				// New key should be generated with correct length
				Expect(secret.Data).To(HaveKey("new-key"))
				Expect(len(secret.Data["new-key"])).To(Equal(20))
			})

			It("should resolve copyFrom keys during migration", func() {
				legacySecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "legacy-secret",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"AUTH_TOKEN": []byte("legacy-token-value"),
					},
				}
				setupFakeClient(legacySecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:             "new-secret",
							LegacySecretName: "legacy-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
								{Name: "pubsub-password", CopyFrom: "AUTH_TOKEN"},
								{Name: "cache-password", CopyFrom: "AUTH_TOKEN"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify new secret was created
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "new-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// AUTH_TOKEN should be migrated from legacy
				Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("legacy-token-value"))

				// copyFrom keys should have the same value as AUTH_TOKEN
				Expect(string(secret.Data["pubsub-password"])).To(Equal("legacy-token-value"))
				Expect(string(secret.Data["cache-password"])).To(Equal("legacy-token-value"))
			})

			It("should use static chartValue when key is missing from legacy secret and type is static", func() {
				legacySecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "legacy-secret",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"password": []byte("existing-password"),
					},
				}
				setupFakeClient(legacySecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:             "new-secret",
							LegacySecretName: "legacy-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
								{Name: "username", Type: "static", ChartValue: stringPtr("admin")},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "new-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(secret.Data["username"])).To(Equal("admin"))
			})
		})

		Context("existingSecret pattern - skip generation", func() {
			It("should generate secret when existingSecretName matches the generated name", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:               "my-secret",
							ExistingSecretName: "my-secret", // Same as Name - should generate
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password", Length: 24},
								{Name: "api-key", Type: "alphanumeric", Length: 16},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify secret WAS created (not skipped)
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "my-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Verify keys exist and have correct lengths
				Expect(secret.Data).To(HaveKey("password"))
				Expect(secret.Data).To(HaveKey("api-key"))
				Expect(len(secret.Data["password"])).To(Equal(24))
				Expect(len(secret.Data["api-key"])).To(Equal(16))

				// Verify it's managed by the generator
				Expect(secret.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", "secret-generator"))
			})

			It("should skip generation when existingSecret is found without validating keys", func() {
				externalSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "external-secret",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"password": []byte("external-password"),
						"api-key":  []byte("external-api-key"),
					},
				}
				setupFakeClient(externalSecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:               "my-secret",
							ExistingSecretName: "external-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password"},
								{Name: "api-key"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify no new secret was created
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "my-secret",
				}, secret)
				Expect(err).To(HaveOccurred()) // Should not exist
			})

			It("should skip generation even when existingSecret is missing some keys", func() {
				externalSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "external-secret",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"password": []byte("external-password"),
						// missing "api-key" - user takes ownership of secret contents
					},
				}
				setupFakeClient(externalSecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:               "my-secret",
							ExistingSecretName: "external-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password"},
								{Name: "api-key"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should fail when existingSecret is not found", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:               "my-secret",
							ExistingSecretName: "nonexistent-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("existingSecret"))
			})
		})

		Context("dry-run mode", func() {
			It("should not create secrets in dry-run mode", func() {
				setupFakeClient()
				cfg.DryRun = true

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify secret was NOT created
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).To(HaveOccurred()) // Should not exist
			})
		})

		Context("different value types", func() {
			It("should generate password type correctly", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "pwd", Type: "password", Length: 16},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				pwd := string(secret.Data["pwd"])
				Expect(len(pwd)).To(Equal(16))
			})

			It("should generate alphanumeric type correctly", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "key", Type: "alphanumeric", Length: 20},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				key := string(secret.Data["key"])
				Expect(len(key)).To(Equal(20))
				Expect(key).To(MatchRegexp(`^[a-zA-Z0-9]+$`))
			})

			It("should generate hex type correctly", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "key", Type: "hex", Length: 32},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				key := string(secret.Data["key"])
				Expect(len(key)).To(Equal(32))
				Expect(key).To(MatchRegexp(`^[a-f0-9]+$`))
			})

			It("should generate UUID type correctly", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "id", Type: "uuid"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				id := string(secret.Data["id"])
				// UUID v4 format
				Expect(id).To(MatchRegexp(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`))
			})

			It("should use default type (password) when type is not specified", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "key", Length: 24}, // No type specified
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				key := string(secret.Data["key"])
				Expect(len(key)).To(Equal(24))
			})

			It("should use static chartValue when provided", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "username", Type: "static", ChartValue: stringPtr("my-static-user")},
								{Name: "password", Type: "password"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(secret.Data["username"])).To(Equal("my-static-user"))
			})

			It("should use default length when not specified", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "key", Type: "alphanumeric"}, // No length specified
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Default length is 32
				Expect(len(secret.Data["key"])).To(Equal(32))
			})
		})

		Context("metadata and labels", func() {
			It("should not set owner references to allow persistence", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Verify no owner references - secrets should persist independently across helm operations
				Expect(secret.OwnerReferences).To(BeEmpty())
			})

			It("should set correct labels", func() {
				setupFakeClient()
				cfg.FullnameOverride = "my-custom-release"

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				Expect(secret.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", "secret-generator"))
				Expect(secret.Labels).To(HaveKeyWithValue("app.kubernetes.io/part-of", "my-custom-release"))
			})
		})

		Context("mixed scenarios", func() {
			It("should handle a mix of new, existing, migrated, and external secrets", func() {
				// Existing generator-managed secret (should be preserved)
				existingSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "existing-secret",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"app.kubernetes.io/managed-by": "secret-generator",
						},
					},
					Data: map[string][]byte{
						"password": []byte("preserved-password"),
					},
				}

				// Legacy secret for migration
				legacySecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "legacy-secret",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"password": []byte("migrated-password"),
					},
				}

				// External secret (for existingSecret pattern)
				externalSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "external-secret",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"password": []byte("external-password"),
					},
				}

				setupFakeClient(existingSecret, legacySecret, externalSecret)

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "existing-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
							},
						},
						{
							Name:             "migrated-secret",
							LegacySecretName: "legacy-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
							},
						},
						{
							Name:               "skip-secret",
							ExistingSecretName: "external-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password"},
							},
						},
						{
							Name: "new-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify existing secret was preserved
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "existing-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(secret.Data["password"])).To(Equal("preserved-password"))

				// Verify migrated secret was created with migrated value
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "migrated-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(secret.Data["password"])).To(Equal("migrated-password"))

				// Verify skip-secret was NOT created (using external)
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "skip-secret",
				}, secret)
				Expect(err).To(HaveOccurred())

				// Verify new secret was created
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "new-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())
				Expect(secret.Data).To(HaveKey("password"))
			})
		})

		Context("empty configuration", func() {
			It("should succeed with empty secrets list", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("error handling", func() {
			It("should return error when one of multiple secrets fails", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "good-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
							},
						},
						{
							Name:               "bad-secret",
							ExistingSecretName: "nonexistent-external-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("existingSecret"))

				// But the good secret should still have been created
				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "good-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle invalid value type gracefully (defaults to password)", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "key", Type: "unknown-type", Length: 16},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())
				// Unknown type defaults to password
				Expect(len(secret.Data["key"])).To(Equal(16))
			})
		})

		Context("secret type", func() {
			It("should create secrets with Opaque type", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				Expect(secret.Type).To(Equal(corev1.SecretTypeOpaque))
			})
		})

		Context("minimum password length enforcement", func() {
			It("should enforce minimum password length of 8", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "short-pwd", Type: "password", Length: 4},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Minimum enforced to 8
				Expect(len(secret.Data["short-pwd"])).To(BeNumerically(">=", 8))
			})
		})

		Context("namespace handling", func() {
			It("should create secrets in the configured namespace", func() {
				cfg.Namespace = "custom-namespace"
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "test-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "custom-namespace",
					Name:      "test-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())
				Expect(secret.Namespace).To(Equal("custom-namespace"))
			})
		})

		Context("multiple specs targeting the same secret", func() {
			It("should create secret from first spec and reconcile keys from subsequent specs", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:             "shared-secret",
							LegacySecretName: "legacy-shared",
							Keys: []generator.SecretKeySpec{
								{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
							},
						},
						{
							Name:             "shared-secret",
							LegacySecretName: "legacy-shared",
							Keys: []generator.SecretKeySpec{
								{Name: "pubsub-password", CopyFrom: "AUTH_TOKEN"},
								{Name: "pubsub-user", Type: "static", ChartValue: stringPtr("nats_client")},
							},
						},
						{
							Name:             "shared-secret",
							LegacySecretName: "legacy-shared",
							Keys: []generator.SecretKeySpec{
								{Name: "cache-password", CopyFrom: "AUTH_TOKEN"},
							},
						},
						{
							Name:             "shared-secret",
							LegacySecretName: "legacy-shared",
							Keys: []generator.SecretKeySpec{
								{Name: "extra-password", Type: "alphanumeric", Length: 24},
								{Name: "extra-user", Type: "static", ChartValue: stringPtr("")},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "shared-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				Expect(secret.Data).To(HaveLen(6))
				Expect(secret.Data["AUTH_TOKEN"]).NotTo(BeEmpty())
				Expect(len(secret.Data["AUTH_TOKEN"])).To(Equal(32))
				// copyFrom keys should match AUTH_TOKEN
				Expect(string(secret.Data["pubsub-password"])).To(Equal(string(secret.Data["AUTH_TOKEN"])))
				Expect(string(secret.Data["cache-password"])).To(Equal(string(secret.Data["AUTH_TOKEN"])))
				// static values
				Expect(string(secret.Data["pubsub-user"])).To(Equal("nats_client"))
				Expect(string(secret.Data["extra-user"])).To(Equal(""))
				// independently generated key
				Expect(secret.Data["extra-password"]).NotTo(BeEmpty())
				Expect(len(secret.Data["extra-password"])).To(Equal(24))
			})

			It("should migrate from legacy secret on first spec and reconcile keys from subsequent specs", func() {
				setupFakeClient()

				legacySecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "legacy-shared",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"AUTH_TOKEN":          []byte("migrated-token-value"),
						"subscriber-password": []byte("migrated-subscriber-pw"),
					},
				}
				Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:             "shared-secret",
							LegacySecretName: "legacy-shared",
							Keys: []generator.SecretKeySpec{
								{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
							},
						},
						{
							Name:             "shared-secret",
							LegacySecretName: "legacy-shared",
							Keys: []generator.SecretKeySpec{
								{Name: "pubsub-password", CopyFrom: "AUTH_TOKEN"},
							},
						},
						{
							Name:             "shared-secret",
							LegacySecretName: "legacy-shared",
							Keys: []generator.SecretKeySpec{
								// Regular generated key that exists in the legacy secret --
								// must be migrated from legacy, not regenerated randomly.
								{Name: "subscriber-password", Type: "alphanumeric", Length: 24},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "shared-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				Expect(secret.Data).To(HaveLen(3))
				// First spec migrated AUTH_TOKEN from legacy
				Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("migrated-token-value"))
				// Second spec reconciled: copyFrom resolves from existing AUTH_TOKEN
				Expect(string(secret.Data["pubsub-password"])).To(Equal("migrated-token-value"))
				// Third spec reconciled: subscriber-password must come from legacy, not random
				Expect(string(secret.Data["subscriber-password"])).To(Equal("migrated-subscriber-pw"))
			})

			It("should reconcile missing keys from different legacy secrets during migration", func() {
				setupFakeClient()

				// Each spec has a different legacy secret with its own data
				legacyFirst := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "legacy-first",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"AUTH_TOKEN": []byte("original-auth-token"),
					},
				}
				legacySecond := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "legacy-second",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"subscriber-password": []byte("original-subscriber-pw"),
						"subscriber-user":     []byte("original-subscriber-user"),
					},
				}
				Expect(fakeClient.Create(ctx, legacyFirst)).To(Succeed())
				Expect(fakeClient.Create(ctx, legacySecond)).To(Succeed())

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name:             "shared-secret",
							LegacySecretName: "legacy-first",
							Keys: []generator.SecretKeySpec{
								{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
							},
						},
						{
							Name:             "shared-secret",
							LegacySecretName: "legacy-second",
							Keys: []generator.SecretKeySpec{
								{Name: "subscriber-password", Type: "alphanumeric", Length: 24},
								{Name: "subscriber-user", Type: "alphanumeric", Length: 24},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "shared-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				Expect(secret.Data).To(HaveLen(3))
				// First spec migrated AUTH_TOKEN from legacy-first
				Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("original-auth-token"))
				// Second spec reconciled: missing keys pulled from legacy-second, not regenerated
				Expect(string(secret.Data["subscriber-password"])).To(Equal("original-subscriber-pw"))
				Expect(string(secret.Data["subscriber-user"])).To(Equal("original-subscriber-user"))
			})

			It("should skip individual spec when its existingSecret is set while others reconcile", func() {
				setupFakeClient()

				externalSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-external-secret",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"external-key": []byte("external-value"),
					},
				}
				Expect(fakeClient.Create(ctx, externalSecret)).To(Succeed())

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "shared-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
							},
						},
						{
							Name:               "shared-secret",
							ExistingSecretName: "my-external-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "pubsub-password", Type: "password", Length: 24},
							},
						},
						{
							Name: "shared-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "cache-password", CopyFrom: "AUTH_TOKEN"},
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "shared-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// First spec created AUTH_TOKEN, second was skipped (existingSecret),
				// third reconciled cache-password
				Expect(secret.Data).To(HaveKey("AUTH_TOKEN"))
				Expect(secret.Data).To(HaveKey("cache-password"))
				Expect(secret.Data).NotTo(HaveKey("pubsub-password"))
			})
		})

		Context("multiple keys in single secret", func() {
			It("should handle secrets with many keys of different types", func() {
				setupFakeClient()

				writeGeneratorConfig(generator.GeneratorConfig{
					Secrets: []generator.SecretSpec{
						{
							Name: "multi-key-secret",
							Keys: []generator.SecretKeySpec{
								{Name: "password", Type: "password", Length: 32},
								{Name: "api-key", Type: "alphanumeric", Length: 24},
								{Name: "secret-hex", Type: "hex", Length: 64},
								{Name: "client-id", Type: "uuid"},
								{Name: "username", Type: "static", ChartValue: stringPtr("admin")},
								{Name: "default-key"}, // defaults to password with default length
							},
						},
					},
				})

				sg := generator.NewSecretGenerator(fakeClient, cfg)
				err := sg.Run(ctx)
				Expect(err).NotTo(HaveOccurred())

				secret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{
					Namespace: "test-namespace",
					Name:      "multi-key-secret",
				}, secret)
				Expect(err).NotTo(HaveOccurred())

				// Verify all keys exist with correct characteristics
				Expect(secret.Data).To(HaveLen(6))
				Expect(len(secret.Data["password"])).To(Equal(32))
				Expect(len(secret.Data["api-key"])).To(Equal(24))
				Expect(string(secret.Data["api-key"])).To(MatchRegexp(`^[a-zA-Z0-9]+$`))
				Expect(len(secret.Data["secret-hex"])).To(Equal(64))
				Expect(string(secret.Data["secret-hex"])).To(MatchRegexp(`^[a-f0-9]+$`))
				Expect(string(secret.Data["client-id"])).To(MatchRegexp(`^[a-f0-9-]+$`))
				Expect(string(secret.Data["username"])).To(Equal("admin"))
				Expect(len(secret.Data["default-key"])).To(Equal(32)) // Default length
			})
		})
	})
})
