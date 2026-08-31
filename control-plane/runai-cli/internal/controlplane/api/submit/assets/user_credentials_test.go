/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package assets

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Credentials Environment Variables", func() {
	envVarsFromCredentialVars := func(credVars []string) ([]openapi.EnvironmentVariable, error) {
		var result []openapi.EnvironmentVariable
		specFunc := func(v []openapi.EnvironmentVariable) {
			result = v
		}
		err := specset.SetEnvironmentVariables(
			[]string{}, []string{}, []string{},
			credVars,
			specFunc,
		)
		return result, err
	}

	Context("env-my-credentials genericSecret (type=genericSecret)", Ordered, func() {
		It("should parse type=genericSecret with name, credential-name, and key", func() {
			credVars := []string{"type=genericSecret,name=KEY1,credential-name=secret1,key=key1"}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(1))
			Expect(*env[0].Name.Get()).To(Equal("KEY1"))
			Expect(env[0].UserCredential.IsSet()).To(BeTrue())
			Expect(*env[0].UserCredential.Get().Name.Get()).To(Equal("secret1"))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal("key1"))
		})

		It("should parse multiple generic entries", func() {
			credVars := []string{
				"type=genericSecret,name=KEY1,credential-name=secret1,key=key1",
				"type=genericSecret,name=KEY2,credential-name=secret1,key=key2",
			}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(2))
			Expect(*env[0].Name.Get()).To(Equal("KEY1"))
			Expect(*env[0].UserCredential.Get().Name.Get()).To(Equal("secret1"))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal("key1"))
			Expect(*env[1].Name.Get()).To(Equal("KEY2"))
			Expect(*env[1].UserCredential.Get().Name.Get()).To(Equal("secret1"))
			Expect(*env[1].UserCredential.Get().Key.Get()).To(Equal("key2"))
		})

		It("should accept type,name,credential-name,key in any order", func() {
			credVars := []string{"key=mykey,type=genericSecret,name=MY_VAR,credential-name=mysecret"}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(1))
			Expect(*env[0].Name.Get()).To(Equal("MY_VAR"))
			Expect(*env[0].UserCredential.Get().Name.Get()).To(Equal("mysecret"))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal("mykey"))
		})

		It("should preserve equals signs inside key value", func() {
			credVars := []string{"type=genericSecret,name=KEY1,credential-name=my-secret,key=complex=value"}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(1))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal("complex=value"))
		})

		It("should return empty list for empty cred vars", func() {
			env, err := envVarsFromCredentialVars([]string{})
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(0))
		})

		It("should fail when key is missing for generic", func() {
			credVars := []string{"type=genericSecret,name=KEY1,credential-name=secret1"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("key is required for type=genericSecret"))
		})

		It("should fail when credential-name is missing", func() {
			credVars := []string{"type=genericSecret,name=KEY1,key=key1"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("credential-name"))
		})

		It("should fail when name is missing", func() {
			credVars := []string{"type=genericSecret,credential-name=secret1,key=key1"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("name"))
		})

		It("should fail when type is missing", func() {
			credVars := []string{"name=KEY1,credential-name=secret1,key=key1"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("type"))
		})

		It("should fail when a value is empty (name=)", func() {
			credVars := []string{"type=genericSecret,name=,credential-name=secret1,key=key1"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("missing mandatory fields"))
			Expect(err.Error()).To(ContainSubstring("name"))
		})

		It("should reject unrecognized keys", func() {
			credVars := []string{"type=genericSecret,name=KEY1,credential-name=secret1,key=key1,extra=value"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("dictionary does not contain key"))
			Expect(err.Error()).To(ContainSubstring("extra"))
		})

		It("duplicate keys use last value (as rest of CLI)", func() {
			credVars := []string{"type=genericSecret,name=KEY1,name=KEY2,credential-name=secret1,key=key1"}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(env).To(HaveLen(1))
			Expect(*env[0].Name.Get()).To(Equal("KEY2"))
		})

		It("should fail when a part has no equals sign", func() {
			credVars := []string{"type=genericSecret,name=KEY1,credential-name secret1,key=key1"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("missing mandatory fields"))
			Expect(err.Error()).To(ContainSubstring("credential-name"))
		})

		It("should reject invalid type", func() {
			credVars := []string{"type=foo,name=KEY1,credential-name=secret1,key=key1"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid type"))
			Expect(err.Error()).To(ContainSubstring("dockerRegistry"))
			Expect(err.Error()).To(ContainSubstring("genericSecret"))
			Expect(err.Error()).To(ContainSubstring("ngcApiKey"))
		})
	})

	Context("env-my-credentials dockerRegistry (type=dockerRegistry)", Ordered, func() {
		It("should parse type=dockerRegistry with name and credential-name and use .dockerconfigjson key", func() {
			credVars := []string{"type=dockerRegistry,name=DOCKER_CONFIG,credential-name=docker-creds"}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(1))
			Expect(*env[0].Name.Get()).To(Equal("DOCKER_CONFIG"))
			Expect(env[0].UserCredential.IsSet()).To(BeTrue())
			Expect(*env[0].UserCredential.Get().Name.Get()).To(Equal("docker-creds"))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal(".dockerconfigjson"))
		})

		It("should parse multiple docker entries", func() {
			credVars := []string{
				"type=dockerRegistry,name=DOCKER_CFG1,credential-name=docker-creds-1",
				"type=dockerRegistry,name=DOCKER_CFG2,credential-name=docker-creds-2",
			}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(2))
			Expect(*env[0].UserCredential.Get().Name.Get()).To(Equal("docker-creds-1"))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal(".dockerconfigjson"))
			Expect(*env[1].UserCredential.Get().Name.Get()).To(Equal("docker-creds-2"))
			Expect(*env[1].UserCredential.Get().Key.Get()).To(Equal(".dockerconfigjson"))
		})

		It("should accept credential-name and name in any order", func() {
			credVars := []string{"credential-name=my-secret,type=dockerRegistry,name=MY_VAR"}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(1))
			Expect(*env[0].Name.Get()).To(Equal("MY_VAR"))
			Expect(*env[0].UserCredential.Get().Name.Get()).To(Equal("my-secret"))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal(".dockerconfigjson"))
		})

		It("should fail when key is set for type=dockerRegistry", func() {
			credVars := []string{"type=dockerRegistry,name=MY_VAR,credential-name=my-secret,key=ignored"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("key is not allowed for type=dockerRegistry"))
		})

		It("should fail when credential-name is missing", func() {
			credVars := []string{"type=dockerRegistry,name=MY_VAR"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("credential-name"))
		})

		It("should fail when name is missing", func() {
			credVars := []string{"type=dockerRegistry,credential-name=my-secret"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("name"))
		})

		It("should fail when a value is empty", func() {
			credVars := []string{"type=dockerRegistry,name=MY_VAR,credential-name="}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("missing mandatory fields"))
			Expect(err.Error()).To(ContainSubstring("credential-name"))
		})

		It("should fail on completely empty string input", func() {
			credVars := []string{""}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("value is an empty string"))
		})
	})

	Context("env-my-credentials ngcApiKey (type=ngcApiKey)", Ordered, func() {
		It("should parse type=ngcApiKey with name and credential-name and use NGC_API_KEY key", func() {
			credVars := []string{"type=ngcApiKey,name=NGC_KEY,credential-name=ngc-creds"}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(1))
			Expect(*env[0].Name.Get()).To(Equal("NGC_KEY"))
			Expect(env[0].UserCredential.IsSet()).To(BeTrue())
			Expect(*env[0].UserCredential.Get().Name.Get()).To(Equal("ngc-creds"))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal("NGC_API_KEY"))
		})

		It("should parse multiple ngc-api-key entries", func() {
			credVars := []string{
				"type=ngcApiKey,name=NGC_KEY1,credential-name=ngc1",
				"type=ngcApiKey,name=NGC_KEY2,credential-name=ngc2",
			}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(2))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal("NGC_API_KEY"))
			Expect(*env[1].UserCredential.Get().Key.Get()).To(Equal("NGC_API_KEY"))
		})

		It("should accept credential-name and name in any order", func() {
			credVars := []string{"credential-name=ngc-secret,type=ngcApiKey,name=MY_NGC"}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(1))
			Expect(*env[0].Name.Get()).To(Equal("MY_NGC"))
			Expect(*env[0].UserCredential.Get().Name.Get()).To(Equal("ngc-secret"))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal("NGC_API_KEY"))
		})

		It("should fail when key is set for type=ngcApiKey", func() {
			credVars := []string{"type=ngcApiKey,name=MY_NGC,credential-name=ngc-creds,key=ignored"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("key is not allowed for type=ngcApiKey"))
		})

		It("should fail when credential-name is missing", func() {
			credVars := []string{"type=ngcApiKey,name=MY_NGC"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("credential-name"))
		})

		It("should fail when name is missing", func() {
			credVars := []string{"type=ngcApiKey,credential-name=ngc-creds"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("name"))
		})

		It("should fail when a value is empty", func() {
			credVars := []string{"type=ngcApiKey,name=MY_NGC,credential-name="}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("missing mandatory fields"))
			Expect(err.Error()).To(ContainSubstring("credential-name"))
		})

		It("should fail when a part has no equals sign", func() {
			credVars := []string{"type=ngcApiKey,nameNGC,credential-name=ngc-creds"}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("missing mandatory fields"))
			Expect(err.Error()).To(ContainSubstring("name"))
		})
	})

	Context("SetEnvironmentVariables aggregation with --env-my-credentials", Ordered, func() {
		It("should parse full submit-style env-my-credentials (ngc x2, docker x2, generic x2)", func() {
			credVars := []string{
				"type=ngcApiKey,name=NGC_KEY1,credential-name=ngc1",
				"type=ngcApiKey,name=NGC_KEY2,credential-name=ngc2",
				"type=dockerRegistry,name=DOCKER_CFG1,credential-name=dr1",
				"type=dockerRegistry,name=DOCKER_CFG2,credential-name=dr2",
				"type=genericSecret,name=GEN_VAR1,credential-name=gs1,key=key1",
				"type=genericSecret,name=GEN_VAR2,credential-name=gs2,key=key1",
			}
			env, err := envVarsFromCredentialVars(credVars)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(6))

			// Implementation appends generic, then docker, then ngc (not input order)
			Expect(*env[0].Name.Get()).To(Equal("GEN_VAR1"))
			Expect(*env[0].UserCredential.Get().Name.Get()).To(Equal("gs1"))
			Expect(*env[0].UserCredential.Get().Key.Get()).To(Equal("key1"))

			Expect(*env[1].Name.Get()).To(Equal("GEN_VAR2"))
			Expect(*env[1].UserCredential.Get().Name.Get()).To(Equal("gs2"))
			Expect(*env[1].UserCredential.Get().Key.Get()).To(Equal("key1"))

			Expect(*env[2].Name.Get()).To(Equal("DOCKER_CFG1"))
			Expect(*env[2].UserCredential.Get().Name.Get()).To(Equal("dr1"))
			Expect(*env[2].UserCredential.Get().Key.Get()).To(Equal(".dockerconfigjson"))

			Expect(*env[3].Name.Get()).To(Equal("DOCKER_CFG2"))
			Expect(*env[3].UserCredential.Get().Name.Get()).To(Equal("dr2"))
			Expect(*env[3].UserCredential.Get().Key.Get()).To(Equal(".dockerconfigjson"))

			Expect(*env[4].Name.Get()).To(Equal("NGC_KEY1"))
			Expect(*env[4].UserCredential.Get().Name.Get()).To(Equal("ngc1"))
			Expect(*env[4].UserCredential.Get().Key.Get()).To(Equal("NGC_API_KEY"))

			Expect(*env[5].Name.Get()).To(Equal("NGC_KEY2"))
			Expect(*env[5].UserCredential.Get().Name.Get()).To(Equal("ngc2"))
			Expect(*env[5].UserCredential.Get().Key.Get()).To(Equal("NGC_API_KEY"))
		})

		It("should combine plain, field-ref, secret, and credential env vars in order", func() {
			var result []openapi.EnvironmentVariable
			specFunc := func(v []openapi.EnvironmentVariable) {
				result = v
			}

			err := specset.SetEnvironmentVariables(
				[]string{"PLAIN=value"},
				[]string{"FIELD=status.podIP"},
				[]string{"SECRET_VAR=my-secret,key=pw"},
				[]string{
					"type=genericSecret,name=GENERIC_VAR,credential-name=gen-secret,key=gen-key",
					"type=dockerRegistry,name=DOCKER_VAR,credential-name=docker-creds",
					"type=ngcApiKey,name=NGC_VAR,credential-name=ngc-creds",
				},
				specFunc,
			)
			Expect(err).To(BeNil())
			Expect(len(result)).To(Equal(6))

			Expect(*result[0].Name.Get()).To(Equal("PLAIN"))
			Expect(*result[0].Value.Get()).To(Equal("value"))

			Expect(*result[1].Name.Get()).To(Equal("FIELD"))
			Expect(result[1].PodFieldRef.IsSet()).To(BeTrue())

			Expect(*result[2].Name.Get()).To(Equal("SECRET_VAR"))
			Expect(result[2].Secret.IsSet()).To(BeTrue())

			Expect(*result[3].Name.Get()).To(Equal("GENERIC_VAR"))
			Expect(result[3].UserCredential.IsSet()).To(BeTrue())
			Expect(*result[3].UserCredential.Get().Name.Get()).To(Equal("gen-secret"))
			Expect(*result[3].UserCredential.Get().Key.Get()).To(Equal("gen-key"))

			Expect(*result[4].Name.Get()).To(Equal("DOCKER_VAR"))
			Expect(result[4].UserCredential.IsSet()).To(BeTrue())
			Expect(*result[4].UserCredential.Get().Name.Get()).To(Equal("docker-creds"))
			Expect(*result[4].UserCredential.Get().Key.Get()).To(Equal(".dockerconfigjson"))

			Expect(*result[5].Name.Get()).To(Equal("NGC_VAR"))
			Expect(result[5].UserCredential.IsSet()).To(BeTrue())
			Expect(*result[5].UserCredential.Get().Name.Get()).To(Equal("ngc-creds"))
			Expect(*result[5].UserCredential.Get().Key.Get()).To(Equal("NGC_API_KEY"))
		})

		It("should produce only plain env vars when cred list is empty", func() {
			var result []openapi.EnvironmentVariable
			specFunc := func(v []openapi.EnvironmentVariable) {
				result = v
			}

			err := specset.SetEnvironmentVariables(
				[]string{"VAR=val"},
				[]string{}, []string{},
				[]string{},
				specFunc,
			)
			Expect(err).To(BeNil())
			Expect(len(result)).To(Equal(1))
			Expect(*result[0].Name.Get()).To(Equal("VAR"))
		})

		It("should propagate error from invalid generic credential (missing key)", func() {
			specFunc := func(v []openapi.EnvironmentVariable) {}

			err := specset.SetEnvironmentVariables(
				[]string{}, []string{}, []string{},
				[]string{"type=genericSecret,name=KEY1,credential-name=sec"},
				specFunc,
			)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("key is required for type=genericSecret"))
		})

		It("should propagate error from invalid docker credential (missing credential-name)", func() {
			specFunc := func(v []openapi.EnvironmentVariable) {}

			err := specset.SetEnvironmentVariables(
				[]string{}, []string{}, []string{},
				[]string{"type=dockerRegistry,name=VAR"},
				specFunc,
			)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("credential-name"))
		})

		It("should propagate error from invalid ngc credential (missing name)", func() {
			specFunc := func(v []openapi.EnvironmentVariable) {}

			err := specset.SetEnvironmentVariables(
				[]string{}, []string{}, []string{},
				[]string{"type=ngcApiKey,credential-name=ngc-creds"},
				specFunc,
			)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("name"))
		})

		It("should fail when key is set for type=dockerRegistry", func() {
			credVars := []string{
				"type=ngcApiKey,name=NGC_KEY1,credential-name=ngc1",
				"type=dockerRegistry,name=DOCKER_CFG1,credential-name=dr1,key=notallowed",
			}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("key is not allowed for type=dockerRegistry"))
		})

		It("should fail when key is set for type=ngcApiKey", func() {
			credVars := []string{
				"type=ngcApiKey,name=NGC_KEY1,credential-name=ngc1,key=notallowed",
			}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("key is not allowed for type=ngcApiKey"))
		})

		It("should fail when key is missing for type=genericSecret", func() {
			credVars := []string{
				"type=genericSecret,name=GEN_VAR1,credential-name=gs1",
			}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("key is required for type=genericSecret"))
		})

		It("should fail on invalid type (negative case)", func() {
			credVars := []string{
				"type=invalid,name=NGC_KEY1,credential-name=ngc1",
			}
			_, err := envVarsFromCredentialVars(credVars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid type"))
		})
	})

	Context("image pull secrets written to spec", Ordered, func() {
		It("dockerRegistry and ngcApiKey credentials, spec receives both in order", func() {
			var result []openapi.ImagePullSecret
			specFunc := func(v []openapi.ImagePullSecret) {
				result = v
			}

			err := specset.SetImagePullSecrets(
				[]string{"type=dockerRegistry,name=docker-registry-1", "type=ngcApiKey,name=ngc-api-key-1"},
				specFunc,
			)
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))
			Expect(result[0].GetName()).To(Equal("docker-registry-1"))
			Expect(result[0].GetUserCredential()).To(BeTrue())
			Expect(result[1].GetName()).To(Equal("ngc-api-key-1"))
			Expect(result[1].GetUserCredential()).To(BeTrue())
		})

		It("invalid credential format returns error", func() {
			specFunc := func(v []openapi.ImagePullSecret) {}

			err := specset.SetImagePullSecrets(
				[]string{"no-equals-sign"},
				specFunc,
			)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("dictionary does not contain key"))
		})

		It("unsupported key in credential returns error", func() {
			specFunc := func(v []openapi.ImagePullSecret) {}

			err := specset.SetImagePullSecrets(
				[]string{"type=dockerRegistry,name=creds,secret=not-allowed"},
				specFunc,
			)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("dictionary does not contain key"))
		})

		It("empty name value returns error", func() {
			specFunc := func(v []openapi.ImagePullSecret) {}

			err := specset.SetImagePullSecrets(
				[]string{"type=dockerRegistry,name="},
				specFunc,
			)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("required flag(s) not set: name"))
		})

		It("rejects unsupported type genericSecret", func() {
			err := specset.SetImagePullSecrets(
				[]string{"type=genericSecret,name=gs1"},
				func([]openapi.ImagePullSecret) {},
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unsupported credential type"))
			Expect(err.Error()).To(ContainSubstring("genericSecret"))
		})
	})

	Context("environment variables are aggregated", Ordered, func() {
		It("should combine plain, field-ref, and secret env vars in order", func() {
			var result []openapi.EnvironmentVariable
			specFunc := func(v []openapi.EnvironmentVariable) {
				result = v
			}

			err := specset.SetEnvironmentVariables(
				[]string{"PLAIN=value"},
				[]string{"FIELD=status.podIP"},
				[]string{"SECRET_VAR=my-secret,key=pw"},
				[]string{},
				specFunc,
			)
			Expect(err).To(BeNil())
			Expect(len(result)).To(Equal(3))

			Expect(*result[0].Name.Get()).To(Equal("PLAIN"))
			Expect(*result[0].Value.Get()).To(Equal("value"))

			Expect(*result[1].Name.Get()).To(Equal("FIELD"))
			Expect(result[1].PodFieldRef.IsSet()).To(BeTrue())

			Expect(*result[2].Name.Get()).To(Equal("SECRET_VAR"))
			Expect(result[2].Secret.IsSet()).To(BeTrue())
		})
	})
})
