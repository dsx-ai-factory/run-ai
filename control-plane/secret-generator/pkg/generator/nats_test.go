/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package generator_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/secret-generator/config"
	"github.com/run-ai/runai/control-plane/secret-generator/pkg/generator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("NATS Secret Generation", func() {
	var (
		ctx        context.Context
		fakeClient client.Client
		gen        *generator.SecretGenerator
		namespace  string
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespace = "test-namespace"

		// Create a fake Kubernetes client with scheme
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		cfg := &config.ConfigData{
			Namespace:      namespace,
			ConfigFilePath: "/tmp/test-config.json",
			DryRun:         false,
		}

		gen = generator.NewSecretGenerator(fakeClient, cfg)
	})

	Context("When creating a new NATS secret", func() {
		It("should generate AUTH_TOKEN", func() {
			spec := generator.SecretSpec{
				Name:              "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("created"))
			Expect(result.SecretName).To(Equal("runai-backend-gen-cache-pubsub"))

			// Verify secret was created with correct keys
			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())
			Expect(secret.Data).To(HaveKey("AUTH_TOKEN"))
			Expect(len(secret.Data["AUTH_TOKEN"])).To(Equal(32))
		})
	})

	Context("When migrating from legacy NATS secret", func() {
		It("should extract token from nats-server.conf", func() {
			// Create legacy secret with nats-server.conf
			legacyConfig := `
server_name: nats-0
listen: 0.0.0.0:4222
http: 0.0.0.0:8222
# Authorization for client connections
authorization {
  token: "my-secret-token-123"
  timeout: 1
}
# Clustering definition
cluster {
  name: "nats"
  listen: 0.0.0.0:6222
}
`
			legacySecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(legacyConfig),
				},
			}
			Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:              "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("migrated"))
			Expect(result.Message).To(ContainSubstring("migrated from legacy secret"))

			// Verify migrated secret has extracted values
			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())
			Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("my-secret-token-123"))
		})

		It("should handle token without quotes", func() {
			legacyConfig := `
authorization {
  token: unquoted-token
  timeout: 1
}
`
			legacySecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(legacyConfig),
				},
			}
			Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:              "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("migrated"))

			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())
			Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("unquoted-token"))
		})

		It("should fail when legacy secret has no nats-server.conf", func() {
			// Legacy secret without nats-server.conf key
			legacySecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"some-other-key": []byte("value"),
				},
			}
			Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:              "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(HaveOccurred())
			Expect(result.Error.Error()).To(ContainSubstring("missing expected key nats-server.conf"))
		})

		It("should extract token when config contains unresolvable env var references", func() {
			// This is the exact production scenario: $NATS_SERVER_NAME is only set inside
			// the NATS server process, not in the secret-generator container. The old
			// conf.Parse-based implementation hard-failed on undefined $VAR references,
			// causing the entire token extraction to return "". The regex-based fix
			// extracts the quoted token regardless of other env var references in the file.
			legacyConfig := `
server_name: $NATS_SERVER_NAME
listen: 0.0.0.0:4222
http: 0.0.0.0:8222
# Authorization for client connections
authorization {
  token: "QHKPb2eYWtfgRjg"
  timeout:  1
}
# Clustering definition
cluster {
  name: "nats"
  listen: 0.0.0.0:6222
  authorization {
    user: "nats_cluster"
    password: "unY8FCs2nQ"
    timeout: 1
  }
  routes = [
    nats://nats_cluster:unY8FCs2nQ@runai-backend-nats-0.runai-backend-nats-headless:6222
  ]
}
jetstream: enabled
jetstream {
  store_dir: /data
  max_memory_store: 1G
  max_file_store: 3Gi
}
`
			legacySecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(legacyConfig),
				},
			}
			Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:             "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("migrated"))

			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())
			Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("QHKPb2eYWtfgRjg"))
		})

		It("should fail when nats-server.conf exists but token cannot be extracted", func() {
			// Config with authorization block but no token field
			legacyConfig := `
authorization {
  # No token field
  timeout: 1
}
`
			legacySecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(legacyConfig),
				},
			}
			Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:              "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(HaveOccurred())
			Expect(result.Error.Error()).To(ContainSubstring("no authorization token could be extracted"))
		})
	})

	Context("When NATS secret already exists", func() {
		It("should preserve existing secret on upgrade", func() {
			existingSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-gen-cache-pubsub",
					Namespace: namespace,
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "secret-generator",
					},
				},
				Data: map[string][]byte{
					"AUTH_TOKEN": []byte("existing-token"),
				},
			}
			Expect(fakeClient.Create(ctx, existingSecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:              "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("preserved"))
			Expect(result.Message).To(ContainSubstring("already exists"))

			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())
			Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("existing-token"))
		})

		It("should preserve even when legacy NATS secret has unparseable env var references", func() {
			existingSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-gen-cache-pubsub",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"AUTH_TOKEN": []byte("existing-token"),
				},
			}
			Expect(fakeClient.Create(ctx, existingSecret)).To(Succeed())

			legacySecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(`
server_name: $NATS_SERVER_NAME
listen: 0.0.0.0:4222
authorization {
  token: $AUTH_TOKEN
  timeout: 1
}
cluster {
  name: "nats"
  listen: 0.0.0.0:6222
}`),
				},
			}
			Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:              "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("preserved"))
			Expect(string(result.Message)).To(ContainSubstring("already exists"))
		})
	})

	Context("Token extraction edge cases", func() {
		It("should fail gracefully when token is itself an env var reference", func() {
			// If the token value is an unresolvable $VAR (unquoted), we can't extract
			// the real value so migration should fail with a clear error rather than
			// silently producing a "$VARNAME" string as the token.
			legacyConfig := `
server_name: $NATS_SERVER_NAME
authorization {
  token: $AUTH_TOKEN
  timeout: 1
}
`
			legacySecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(legacyConfig),
				},
			}
			Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:             "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(HaveOccurred())
			Expect(result.Error.Error()).To(ContainSubstring("no authorization token could be extracted"))
		})

		It("should extract token even when cluster authorization block appears first", func() {
			// The cluster section contains a nested authorization block (user/password,
			// no token). If the cluster block appears before the top-level authorization
			// block in the file, a naive first-match approach would pick the wrong block
			// and return "". The implementation iterates all blocks and returns the first
			// one that actually contains a token field.
			legacyConfig := `
cluster {
  name: "nats"
  authorization {
    user: "nats_cluster"
    password: "cluster-password"
    timeout: 1
  }
}
authorization {
  token: "correct-client-token"
  timeout: 1
}
`
			legacySecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(legacyConfig),
				},
			}
			Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:             "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("migrated"))

			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())
			Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("correct-client-token"))
		})

		It("should handle complex nats-server.conf with multiple blocks", func() {
			legacyConfig := `
# Server configuration
server_name: nats-0
port: 4222
max_connections: 1000

# Authorization
authorization {
  timeout: 1
  token: "complex-token-xyz"
}

# Jetstream
jetstream {
  store_dir: /data
  max_memory_store: 1G
}

# Cluster configuration
cluster {
  name: "nats-cluster"
  port: 6222
  routes: [
    nats://nats-1:6222
  ]
}
`
			legacySecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(legacyConfig),
				},
			}
			Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:              "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("migrated"))

			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())
			Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("complex-token-xyz"))
		})
	})

	Context("NATS secret with CopyFrom", func() {
		It("should copy AUTH_TOKEN to pubsub-password and cache-password", func() {
			spec := generator.SecretSpec{
				Name:              "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
					{Name: "pubsub-password", CopyFrom: "AUTH_TOKEN"},
					{Name: "cache-password", CopyFrom: "AUTH_TOKEN"},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("created"))

			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())

			// Verify all keys exist
			Expect(secret.Data).To(HaveKey("AUTH_TOKEN"))
			Expect(secret.Data).To(HaveKey("pubsub-password"))
			Expect(secret.Data).To(HaveKey("cache-password"))

			// Verify copies match AUTH_TOKEN
			authToken := string(secret.Data["AUTH_TOKEN"])
			Expect(string(secret.Data["pubsub-password"])).To(Equal(authToken))
			Expect(string(secret.Data["cache-password"])).To(Equal(authToken))

			// Verify lengths
			Expect(len(secret.Data["AUTH_TOKEN"])).To(Equal(32))
		})

		It("should copy extracted token to cache-password during migration", func() {
			legacyConfig := `
authorization {
  token: "migrated-token-abc"
  timeout:  1
}
cluster {
  name: "nats"
  listen: 0.0.0.0:6222
}`

			legacySecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(legacyConfig),
				},
			}
			Expect(fakeClient.Create(ctx, legacySecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:              "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
					{Name: "pubsub-password", CopyFrom: "AUTH_TOKEN"},
					{Name: "cache-password", CopyFrom: "AUTH_TOKEN"},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("migrated"))

			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())

			// Verify migrated token was used
			Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("migrated-token-abc"))

			// Verify copies match AUTH_TOKEN
			Expect(string(secret.Data["pubsub-password"])).To(Equal("migrated-token-abc"))
			Expect(string(secret.Data["cache-password"])).To(Equal("migrated-token-abc"))
		})
	})

	Context("When migrating from multiple legacy secrets", func() {
		It("should migrate from the first found legacy secret only", func() {
			// Create legacy NATS secret with nats-server.conf (first in list, will be used)
			natsConfig := `
authorization {
  token: "nats-auth-token-123"
}

cluster {
  name: "test-cluster"
}`
			legacyNatsSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(natsConfig),
				},
			}

			// Create legacy cache secret (will be ignored since nats is found first)
			legacyCacheSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-cache",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"cache-password": []byte("legacy-cache-pass-789"),
				},
			}

			// Create legacy pubsub secret (will be ignored since nats is found first)
			legacyPubsubSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-pubsub",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"pubsub-password": []byte("legacy-pubsub-pass-xyz"),
				},
			}

			Expect(fakeClient.Create(ctx, legacyNatsSecret)).To(Succeed())
			Expect(fakeClient.Create(ctx, legacyCacheSecret)).To(Succeed())
			Expect(fakeClient.Create(ctx, legacyPubsubSecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:             "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
					{Name: "pubsub-password", CopyFrom: "AUTH_TOKEN"},
					{Name: "pubsub-user", Type: "static", ChartValue: stringPtr("test-user")},
					{Name: "cache-password", CopyFrom: "AUTH_TOKEN"},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("migrated"))
			Expect(result.SecretName).To(Equal("runai-backend-gen-cache-pubsub"))
			Expect(result.Message).To(ContainSubstring("runai-backend-nats"))

			// Verify the secret was created with values from NATS secret only
			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())

			// Verify AUTH_TOKEN came from NATS config
			Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("nats-auth-token-123"))
			// Verify copies match AUTH_TOKEN (not from cache/pubsub secrets)
			Expect(string(secret.Data["pubsub-password"])).To(Equal("nats-auth-token-123"))
			Expect(string(secret.Data["cache-password"])).To(Equal("nats-auth-token-123"))
			// Verify static value
			Expect(string(secret.Data["pubsub-user"])).To(Equal("test-user"))
		})

		It("should handle missing legacy secrets gracefully", func() {
			// Only create NATS secret, cache and pubsub don't exist
			natsConfig := `
authorization {
  token: "nats-only-token"
}`
			legacyNatsSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-backend-nats",
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"nats-server.conf": []byte(natsConfig),
				},
			}

			Expect(fakeClient.Create(ctx, legacyNatsSecret)).To(Succeed())

			spec := generator.SecretSpec{
				Name:             "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
					{Name: "pubsub-password", CopyFrom: "AUTH_TOKEN"},
					{Name: "cache-password", CopyFrom: "AUTH_TOKEN"},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("migrated"))

			// Verify the secret was created with token from NATS
			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())

			Expect(string(secret.Data["AUTH_TOKEN"])).To(Equal("nats-only-token"))
			// Copies should match AUTH_TOKEN
			Expect(string(secret.Data["pubsub-password"])).To(Equal("nats-only-token"))
			Expect(string(secret.Data["cache-password"])).To(Equal("nats-only-token"))
		})

		It("should generate new values when no legacy secrets exist", func() {
			spec := generator.SecretSpec{
				Name:             "runai-backend-gen-cache-pubsub",
				LegacySecretName: "runai-backend-nats",
				Keys: []generator.SecretKeySpec{
					{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
				},
			}

			result := gen.ProcessSecretSpec(ctx, spec)

			Expect(result.Error).To(BeNil())
			Expect(result.Action).To(Equal("created"))

			// Verify all values were generated
			secret := &corev1.Secret{}
			err := fakeClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      "runai-backend-gen-cache-pubsub",
			}, secret)
			Expect(err).To(BeNil())

			Expect(secret.Data["AUTH_TOKEN"]).NotTo(BeEmpty())
			Expect(len(secret.Data["AUTH_TOKEN"])).To(Equal(32))
		})

	})
})
