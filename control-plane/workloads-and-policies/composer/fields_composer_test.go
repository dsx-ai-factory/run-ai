package composer

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var _ = Describe("fields composer test", func() {
	Context("copy int", func() {
		It("should copy correctly", func() {
			var src = 5
			var wrongSrc = 6.5
			var target int64

			fc := FieldsComposer{}

			err := fc.CopyInt(&src, &target, intToIntTest, nil)()
			Expect(err).To(BeNil())
			Expect(target).To(Equal(int64(5)))

			err = fc.CopyInt(&wrongSrc, &target, intToIntTest, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("unsupported type *float64 for int field"))
		})
	})

	Context("copy string to map", func() {
		It("should copy correctly", func() {
			var target map[string]string

			expectedMap := make(map[string]string)

			fc := FieldsComposer{}

			err := fc.CopyStringToMap(nil, &target, mapTtoMapTest, "test1")()
			Expect(err).To(BeNil())
			Expect(target).To(BeEmpty())

			err = fc.CopyStringToMap(convutils.Ref("test2val"), &target, mapTtoMapTest, "test2")()
			Expect(err).To(BeNil())
			expectedMap["test2"] = "test2val"
			testutils.Expect(target).ToEqual(expectedMap)

			err = fc.CopyStringToMap(convutils.Ref("test2replaced"), &target, mapTtoMapTest, "test2")()
			Expect(err).To(BeNil())
			expectedMap["test2"] = "test2replaced"
			testutils.Expect(target).ToEqual(expectedMap)

			err = fc.CopyStringToMap(convutils.Ref("test3val"), &target, mapTtoMapTest, "test3")()
			Expect(err).To(BeNil())
			expectedMap["test3"] = "test3val"
			testutils.Expect(target).ToEqual(expectedMap)
		})
	})

	Context("env variables errors", func() {
		It("should report errors correctly", func() {
			fc := FieldsComposer{
				Options: OptionsForPolicy(),
			}

			err := fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name:      *workloadsApi.NewNullableString(convutils.Ref("test1")),
					ConfigMap: *workloadsApi.NewNullableEnvironmentVariableConfigMap(&workloadsApi.EnvironmentVariableConfigMap{}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required configMap name is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					ConfigMap: *workloadsApi.NewNullableEnvironmentVariableConfigMap(&workloadsApi.EnvironmentVariableConfigMap{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required configMap key is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name:   *workloadsApi.NewNullableString(convutils.Ref("test1")),
					Secret: *workloadsApi.NewNullableEnvironmentVariableSecret(&workloadsApi.EnvironmentVariableSecret{}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required secret name is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					Secret: *workloadsApi.NewNullableEnvironmentVariableSecret(&workloadsApi.EnvironmentVariableSecret{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required secret key is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name:        *workloadsApi.NewNullableString(convutils.Ref("test1")),
					PodFieldRef: *workloadsApi.NewNullableEnvironmentVariablePodFieldReference(&workloadsApi.EnvironmentVariablePodFieldReference{}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required filedRef path is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name:           *workloadsApi.NewNullableString(convutils.Ref("test1")),
					UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(&workloadsApi.EnvironmentVariableUserCredential{}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required user credential name is not provided"))

			err = fc.ComposeEnvironmentVariables(nil, []workloadsApi.EnvironmentVariable{
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(&workloadsApi.EnvironmentVariableUserCredential{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
			}, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': required user credential key is not provided"))
		})
	})

	Context("env variables cluster support", func() {
		It("should report errors correctly", func() {
			vars := []workloadsApi.EnvironmentVariable{
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					ConfigMap: *workloadsApi.NewNullableEnvironmentVariableConfigMap(&workloadsApi.EnvironmentVariableConfigMap{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
						Key:  *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					Secret: *workloadsApi.NewNullableEnvironmentVariableSecret(&workloadsApi.EnvironmentVariableSecret{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
						Key:  *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					PodFieldRef: *workloadsApi.NewNullableEnvironmentVariablePodFieldReference(&workloadsApi.EnvironmentVariablePodFieldReference{
						Path: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
				{
					Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
					UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(&workloadsApi.EnvironmentVariableUserCredential{
						Name: *workloadsApi.NewNullableString(convutils.Ref("test1")),
						Key:  *workloadsApi.NewNullableString(convutils.Ref("test1")),
					}),
				},
			}

			fc := FieldsComposer{
				Options: FieldsComposerOptions{
					UserCredentialsSupported: false,
				},
			}
			err := fc.ComposeEnvironmentVariables(nil, vars, nil, nil, nil, nil)()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("environmentVariables 'test1': userCredential is not supported"))
		})
	})
})

func intToIntTest(src *int64, target any, cookie any) *workloadsApi.Error {
	targetPtr, ok := target.(*int64)
	Expect(ok).To(BeTrue())
	*targetPtr = *src
	return nil
}

func mapTtoMapTest(src map[string]*string, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*map[string]string)
	Expect(ok).To(BeTrue())

	if *targetPtr == nil {
		*targetPtr = make(map[string]string)
	}

	for key, value := range src {
		if value != nil {
			(*targetPtr)[key] = *value
		}
	}

	return nil
}
