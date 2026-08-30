package specset

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Submit Tests")
}

var _ = Describe("Submit Function Test", Ordered, func() {
	Context("Submit environment validation", Ordered, func() {
		It("valid 2 flag - success", func() {

			envList := []string{"env=test", "hash=MTc0MjM3OTExNA=="}
			env, err := GenerateEnvironmentList(envList)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(2))
			Expect(*env[0].Name.Get()).To(Equal("env"))
			Expect(*env[0].Value.Get()).To(Equal("test"))
			Expect(*env[1].Name.Get()).To(Equal("hash"))
			Expect(*env[1].Value.Get()).To(Equal("MTc0MjM3OTExNA=="))
		})

		It("validate double equals success", func() {
			envList := []string{"env==test", "env2=env2"}
			env, err := GenerateEnvironmentList(envList)
			Expect(err).To(BeNil())
			Expect(*env[0].Name.Get()).To(Equal("env"))
			Expect(*env[0].Value.Get()).To(Equal("=test"))
			Expect(*env[1].Name.Get()).To(Equal("env2"))
			Expect(*env[1].Value.Get()).To(Equal("env2"))
		})
		It("no equals should fail", func() {
			envList := []string{"env test", "env2=env2"}
			_, err := GenerateFieldRefList(envList)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("value (env test) does not contain `=` sign"))
		})
	})
	Context("Submit field ref validation", Ordered, func() {
		It("valid 2 flag - success", func() {
			envList := []string{"env=test", "env2=test2"}
			env, err := GenerateFieldRefList(envList)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(2))
			Expect(*env[0].Name.Get()).To(Equal("env"))
			Expect(env[0].PodFieldRef.Get().GetPath()).To(Equal("test"))
			Expect(*env[1].Name.Get()).To(Equal("env2"))
			Expect(env[1].PodFieldRef.Get().GetPath()).To(Equal("test2"))
		})

		It("validate double equals success", func() {
			envList := []string{"env==test", "env2=env2"}
			env, err := GenerateFieldRefList(envList)
			Expect(err).To(BeNil())
			Expect(*env[0].Name.Get()).To(Equal("env"))
			Expect(env[0].PodFieldRef.Get().GetPath()).To(Equal("=test"))
			Expect(*env[1].Name.Get()).To(Equal("env2"))
			Expect(env[1].PodFieldRef.Get().GetPath()).To(Equal("env2"))
		})
		It("no equals should fail", func() {
			envList := []string{"env test", "env2=env2"}
			_, err := GenerateFieldRefList(envList)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("value (env test) does not contain `=` sign"))
		})
	})
	Context("image-pull-my-credentials parsing", Ordered, func() {
		It("single dockerRegistry credential with type=dockerRegistry,name=myregistry-creds", func() {
			vars := []string{"type=dockerRegistry,name=myregistry-creds"}
			secrets, err := GenerateImagePullSecretsList(vars)
			Expect(err).To(BeNil())
			Expect(secrets).To(HaveLen(1))
			Expect(secrets[0].GetName()).To(Equal("myregistry-creds"))
			Expect(secrets[0].GetUserCredential()).To(BeTrue())
		})

		It("single ngcApiKey credential with type=ngcApiKey,name=ngc-api-creds", func() {
			vars := []string{"type=ngcApiKey,name=ngc-api-creds"}
			secrets, err := GenerateImagePullSecretsList(vars)
			Expect(err).To(BeNil())
			Expect(secrets).To(HaveLen(1))
			Expect(secrets[0].GetName()).To(Equal("ngc-api-creds"))
			Expect(secrets[0].GetUserCredential()).To(BeTrue())
		})

		It("credential name with hyphens and numbers", func() {
			vars := []string{"type=dockerRegistry,name=my-registry-123"}
			secrets, err := GenerateImagePullSecretsList(vars)
			Expect(err).To(BeNil())
			Expect(secrets).To(HaveLen(1))
			Expect(secrets[0].GetName()).To(Equal("my-registry-123"))
		})

		It("value without type= and name= fails", func() {
			vars := []string{"myregistry-creds"}
			_, err := GenerateImagePullSecretsList(vars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("dictionary does not contain key"))
		})

		It("missing type fails", func() {
			vars := []string{"name=myregistry"}
			_, err := GenerateImagePullSecretsList(vars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("dictionary must contain all keys"))
			Expect(err.Error()).To(ContainSubstring("type"))
		})

		It("missing name fails", func() {
			vars := []string{"type=dockerRegistry"}
			_, err := GenerateImagePullSecretsList(vars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("dictionary must contain all keys"))
			Expect(err.Error()).To(ContainSubstring("name"))
		})

		It("invalid type fails", func() {
			vars := []string{"type=invalid,name=creds"}
			_, err := GenerateImagePullSecretsList(vars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("unsupported credential type"))
			Expect(err.Error()).To(ContainSubstring("invalid"))
			Expect(err.Error()).To(ContainSubstring("dockerRegistry"))
			Expect(err.Error()).To(ContainSubstring("ngcApiKey"))
		})

		It("name= with empty value fails", func() {
			vars := []string{"type=dockerRegistry,name="}
			_, err := GenerateImagePullSecretsList(vars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("required flag(s) not set: name"))
		})

		It("unsupported key fails", func() {
			for _, tc := range []struct {
				input      string
				errContain string
			}{
				{"type=dockerRegistry,name=test,secret=extra", "secret"},
				{"type=dockerRegistry,name=test,key=extra", "key"},
				{"type=dockerRegistry,nme=myregistry", "nme"},
			} {
				_, err := GenerateImagePullSecretsList([]string{tc.input})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("dictionary does not contain key"))
				Expect(err.Error()).To(ContainSubstring(tc.errContain))
			}
		})

		It("name= specified twice in one entry uses last value", func() {
			vars := []string{"type=dockerRegistry,name=cred1,name=cred2"}
			secrets, err := GenerateImagePullSecretsList(vars)
			Expect(err).To(BeNil())
			Expect(secrets).To(HaveLen(1))
			Expect(secrets[0].GetName()).To(Equal("cred2"))
		})

		It("empty string as credential value fails", func() {
			vars := []string{""}
			_, err := GenerateImagePullSecretsList(vars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("value is an empty string"))
		})

		It("dockerRegistry and ngcApiKey credentials", func() {
			vars := []string{"type=dockerRegistry,name=docker1", "type=ngcApiKey,name=ngc1"}
			secrets, err := GenerateImagePullSecretsList(vars)
			Expect(err).To(BeNil())
			Expect(secrets).To(HaveLen(2))
			Expect(secrets[0].GetName()).To(Equal("docker1"))
			Expect(secrets[0].GetUserCredential()).To(BeTrue())
			Expect(secrets[1].GetName()).To(Equal("ngc1"))
			Expect(secrets[1].GetUserCredential()).To(BeTrue())
		})

		It("invalid entry fails", func() {
			vars := []string{"invalid-no-equals"}
			_, err := GenerateImagePullSecretsList(vars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("dictionary does not contain key"))
		})

		It("unsupported key in entry fails", func() {
			vars := []string{"type=dockerRegistry,name=valid-docker", "foo=bar"}
			_, err := GenerateImagePullSecretsList(vars)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("dictionary does not contain key"))
			Expect(err.Error()).To(ContainSubstring("foo"))
		})
	})
	Context("Submit secret env validation", Ordered, func() {
		It("multiple secret envs - success", func() {
			envList := []string{"SECRET1=secret1,key=key1", "SECRET2=secret2,key=key2"}
			env, err := GenerateSecretEnvList(envList)
			Expect(err).To(BeNil())
			Expect(len(env)).To(Equal(2))
			Expect(*env[0].Name.Get()).To(Equal("SECRET1"))
			Expect(*env[0].Secret.Get().Name.Get()).To(Equal("secret1"))
			Expect(*env[0].Secret.Get().Key.Get()).To(Equal("key1"))
			Expect(*env[1].Name.Get()).To(Equal("SECRET2"))
			Expect(*env[1].Secret.Get().Name.Get()).To(Equal("secret2"))
			Expect(*env[1].Secret.Get().Key.Get()).To(Equal("key2"))
		})

		It("missing key should fail", func() {
			envList := []string{"MY_SECRET=my-secret-name"}
			_, err := GenerateSecretEnvList(envList)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("expected format: secret-name,key=secret-key"))
		})

		It("empty secret name should fail", func() {
			envList := []string{"MY_SECRET=,key=secret-key"}
			_, err := GenerateSecretEnvList(envList)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("secret name cannot be empty"))
		})

		It("no equals should fail", func() {
			envList := []string{"MY_SECRET secret,key=key"}
			_, err := GenerateSecretEnvList(envList)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid env-secret format"))
		})
	})
})
