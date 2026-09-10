package conjungo

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("newFuncSelector", func() {
	var (
		fs *funcSelector
	)

	BeforeEach(func() {
		fs = newFuncSelector()
	})

	It("has the type mergefuncs map", func() {
		Expect(fs.typeFuncs).ToNot(BeNil())
	})

	It("has the correct kind mergeFuncs map", func() {
		mapMerge, mapOk := fs.kindFuncs[reflect.Map]
		Expect(mapOk).To(BeTrue())
		Expect(mapMerge).ToNot(BeNil())

		sliceMerge, sliceOK := fs.kindFuncs[reflect.Slice]
		Expect(sliceOK).To(BeTrue())
		Expect(sliceMerge).ToNot(BeNil())

		structMerge, structOK := fs.kindFuncs[reflect.Struct]
		Expect(structOK).To(BeTrue())
		Expect(structMerge).ToNot(BeNil())
	})

	It("has default mergeFunc", func() {
		Expect(fs.defaultFunc).ToNot(BeNil())
	})
})

var _ = Describe("Set Merge Func", func() {
	type TestKey struct{}

	var (
		fs *funcSelector
	)

	BeforeEach(func() {
		fs = newFuncSelector()
	})

	Context("Type Func", func() {
		It("adds the func correctly", func() {
			stubReturns := "uniqe string"
			t := reflect.TypeOf(TestKey{})
			fs.setTypeMergeFunc(t, newMergeFuncStub(stubReturns))
			returned, _ := fs.typeFuncs[t](reflect.Value{}, reflect.Value{}, NewOptions())
			Expect(returned.Interface()).To(Equal(stubReturns))
		})
	})

	Context("Kind Func", func() {
		It("adds the func correctly", func() {
			stubReturns := "uniqe string"
			k := reflect.TypeOf(TestKey{}).Kind()
			fs.setKindMergeFunc(k, newMergeFuncStub(stubReturns))
			returned, _ := fs.kindFuncs[k](reflect.Value{}, reflect.Value{}, NewOptions())

			Expect(returned.Interface()).To(Equal(stubReturns))
		})
	})

	Context("Default Func", func() {
		It("adds the func correctly", func() {
			stubReturns := "uniqe string"
			fs.setDefaultMergeFunc(newMergeFuncStub(stubReturns))
			returned, _ := fs.defaultFunc(reflect.Value{}, reflect.Value{}, NewOptions())

			Expect(returned.Interface()).To(Equal(stubReturns))
		})
	})

	Context("nil merge func maps", func() {
		It("does not panic", func() {
			fs = &funcSelector{}
			stubReturns := "uniqe string"
			f := newMergeFuncStub(stubReturns)
			t := reflect.TypeOf(TestKey{})
			k := reflect.TypeOf(TestKey{}).Kind()

			fs.setTypeMergeFunc(t, f)
			fs.setKindMergeFunc(k, f)
			fs.setDefaultMergeFunc(f)
		})
	})
})

var _ = Describe("GetFunc", func() {
	type TestKey struct{}

	var (
		fs  *funcSelector
		key *TestKey
	)

	const (
		typeStubReturns    = "type"
		kindStubReturns    = "kind"
		defaultStubReturns = "default"
	)

	BeforeEach(func() {
		fs = &funcSelector{}
		key = &TestKey{}
	})

	Context("Type Func is defined", func() {
		BeforeEach(func() {
			fs.setTypeMergeFunc(reflect.TypeOf(key), newMergeFuncStub(typeStubReturns))
		})

		It("gets the func", func() {
			f := fs.getFunc(reflect.ValueOf(key))
			returned, _ := f(reflect.Value{}, reflect.Value{}, NewOptions())
			Expect(returned.Interface()).To(Equal(typeStubReturns))
		})

		Context("kind func is also defined", func() {
			It("choses the type func", func() {
				fs.setKindMergeFunc(reflect.TypeOf(key).Kind(), newMergeFuncStub(kindStubReturns))
				f := fs.getFunc(reflect.ValueOf(key))
				returned, _ := f(reflect.Value{}, reflect.Value{}, NewOptions())
				Expect(returned.Interface()).To(Equal(typeStubReturns))
			})
		})
	})

	Context("no type func defined", func() {
		Context("kind func is defined", func() {
			It("choses the kind func", func() {
				fs.setKindMergeFunc(reflect.TypeOf(key).Kind(), newMergeFuncStub(kindStubReturns))
				f := fs.getFunc(reflect.ValueOf(key))
				returned, _ := f(reflect.Value{}, reflect.Value{}, NewOptions())
				Expect(returned.Interface()).To(Equal(kindStubReturns))
			})
		})

		Context("no kind func defined", func() {
			Context("default func defined", func() {
				It("choses the default func", func() {
					fs.setDefaultMergeFunc(newMergeFuncStub(defaultStubReturns))
					f := fs.getFunc(reflect.ValueOf(key))
					returned, _ := f(reflect.Value{}, reflect.Value{}, NewOptions())
					Expect(returned.Interface()).To(Equal(defaultStubReturns))
				})
			})

			Context("no default func defined", func() {
				It("choses the global default func", func() {
					f := fs.getFunc(reflect.ValueOf(key))
					returned, _ := f(reflect.ValueOf("a"), reflect.ValueOf("b"), NewOptions())
					Expect(returned.Interface()).To(Equal("b"))
				})
			})
		})
	})

	Context("no merge funcs defined", func() {
		It("returns defaultMergeFunc", func() {
			f := fs.getFunc(reflect.ValueOf(key))
			Expect(f).ToNot(BeNil())
			merged, _ := f(reflect.ValueOf("a"), reflect.ValueOf("b"), NewOptions())
			Expect(merged.Interface()).To(Equal("b"))
		})
	})
})

func newMergeFuncStub(s string) MergeFunc {
	return func(reflect.Value, reflect.Value, *Options) (reflect.Value, error) {
		return reflect.ValueOf(s), nil
	}
}

var _ = Describe("defaultMergeFunc", func() {
	var opts *Options

	BeforeEach(func() {
		opts = &Options{}
	})

	Context("overwrite true", func() {
		It("returns source", func() {
			opts.Overwrite = true
			merged, err := defaultMergeFunc(reflect.ValueOf("a"), reflect.ValueOf("b"), opts)
			Expect(err).ToNot(HaveOccurred())
			Expect(merged.Interface()).To(Equal("b"))
		})
	})

	Context("overwrite false", func() {
		It("returns target", func() {
			opts.Overwrite = false
			merged, err := defaultMergeFunc(reflect.ValueOf("a"), reflect.ValueOf("b"), opts)
			Expect(err).ToNot(HaveOccurred())
			Expect(merged.Interface()).To(Equal("a"))
		})
	})

	Context("never errors", func() {
		It("doesnt error", func() {
			_, err := defaultMergeFunc(reflect.ValueOf(nil), reflect.ValueOf(nil), NewOptions())
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("mergeMap", func() {
	var (
		targetMap, sourceMap       map[string]interface{}
		targetMapVal, sourceMapVal reflect.Value
	)

	BeforeEach(func() {
		targetMap = map[string]interface{}{
			"A": "wrong",
			"B": 1,
			"C": map[string]interface{}{"foo": "unchanged", "bar": "orig"},
			"D": []interface{}{"unchanged", 0},
		}

		sourceMap = map[string]interface{}{
			"A": "correct",
			"B": 2,
			"C": map[string]interface{}{"bar": "newVal", "baz": "added"},
			"D": []interface{}{"added", 1},
		}
	})

	JustBeforeEach(func() {
		targetMapVal = reflect.ValueOf(targetMap)
		sourceMapVal = reflect.ValueOf(sourceMap)
	})

	Context("happy path smoke test", func() {
		It("merges correctly", func() {
			merged, err := mergeMap(targetMapVal, sourceMapVal, NewOptions())

			Expect(err).ToNot(HaveOccurred())

			mergedMap, ok := merged.Interface().(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(mergedMap["A"]).To(Equal("correct"))
			Expect(mergedMap["B"]).To(Equal(2))

			subMap, ok := mergedMap["C"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(subMap["foo"]).To(Equal("unchanged"))
			Expect(subMap["bar"]).To(Equal("newVal"))
			Expect(subMap["baz"]).To(Equal("added"))

			subSlice, ok := mergedMap["D"]
			Expect(ok).To(BeTrue())
			Expect(subSlice).To(And(
				ContainElement("unchanged"),
				ContainElement(0),
				ContainElement("added"),
				ContainElement(1),
			))
		})
	})

	Context("overwrite is true", func() {
		It("overwrites", func() {
			opt := NewOptions()
			opt.Overwrite = true
			merged, err := mergeMap(targetMapVal, sourceMapVal, opt)

			Expect(err).ToNot(HaveOccurred())

			mergedMap, ok := merged.Interface().(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(mergedMap["A"]).To(Equal("correct"))
			Expect(mergedMap["B"]).To(Equal(2))

			subMap, ok := mergedMap["C"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(subMap["foo"]).To(Equal("unchanged"))
			Expect(subMap["bar"]).To(Equal("newVal"))
			Expect(subMap["baz"]).To(Equal("added"))
		})
	})

	Context("overwrite is false", func() {
		It("doesnt overwrite", func() {
			opt := NewOptions()
			opt.Overwrite = false
			merged, err := mergeMap(targetMapVal, sourceMapVal, opt)

			Expect(err).ToNot(HaveOccurred())

			mergedMap, ok := merged.Interface().(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(mergedMap["A"]).To(Equal("wrong"))
			Expect(mergedMap["B"]).To(Equal(1))

			subMap, ok := mergedMap["C"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(subMap["foo"]).To(Equal("unchanged"))
			Expect(subMap["bar"]).To(Equal("orig"))
			Expect(subMap["baz"]).To(Equal("added"))
		})
	})

	Context("typed maps", func() {
		It("selects correct func and merges", func() {
			t := map[int]string{1: "old", 2: "keep"}
			s := map[int]string{1: "new"}

			mergedVal, err := merge(reflect.ValueOf(t), reflect.ValueOf(s), NewOptions())
			Expect(err).ToNot(HaveOccurred())

			mergedMap, ok := mergedVal.Interface().(map[int]string)
			Expect(ok).To(BeTrue())
			Expect(mergedMap[1]).To(Equal("new"))
			Expect(mergedMap[2]).To(Equal("keep"))
		})

		Context("mergeFunc returns wrong type", func() {
			type Bar struct{}

			It("returns an error", func() {
				t := map[int]Bar{1: {}}
				s := map[int]Bar{1: {}}

				opts := NewOptions()
				// define a merge func that returns wrong type
				opts.mergeFuncs.setTypeMergeFunc(
					reflect.TypeOf(Bar{}),
					func(t, s reflect.Value, o *Options) (reflect.Value, error) {
						return reflect.ValueOf("a string"), nil
					},
				)

				_, err := merge(reflect.ValueOf(t), reflect.ValueOf(s), opts)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("value of type string is not assignable to type conjungo.Bar"))
			})
		})
	})

	Context("empty target", func() {
		BeforeEach(func() {
			targetMap = map[string]interface{}{}
		})

		It("equals source", func() {
			merged, err := mergeMap(targetMapVal, sourceMapVal, NewOptions())

			Expect(err).ToNot(HaveOccurred())
			Expect(merged.Interface()).To(Equal(sourceMap))
		})
	})

	Context("empty source", func() {
		BeforeEach(func() {
			sourceMap = map[string]interface{}{}
		})

		It("equals target", func() {
			merged, err := mergeMap(targetMapVal, sourceMapVal, NewOptions())

			Expect(err).ToNot(HaveOccurred())
			Expect(merged.Interface()).To(Equal(targetMap))
		})
	})

	Context("nils", func() {
		// these are called via merge() because that is what does the nil checks
		By("call via merge()")

		Context("nil value target", func() {
			It("equals source", func() {
				targetMapVal = reflect.Value{}
				merged, err := merge(targetMapVal, sourceMapVal, NewOptions())

				Expect(err).ToNot(HaveOccurred())
				Expect(merged.Interface()).To(Equal(sourceMap))
			})
		})

		Context("nil map target", func() {
			It("equals source", func() {
				var m map[string]interface{}
				targetMapVal = reflect.ValueOf(m)
				merged, err := merge(targetMapVal, sourceMapVal, NewOptions())

				Expect(err).ToNot(HaveOccurred())
				Expect(merged.Interface()).To(Equal(sourceMap))
			})
		})

		Context("nil source", func() {
			JustBeforeEach(func() {
				sourceMapVal = reflect.Value{}
			})

			It("equals target", func() {
				merged, err := merge(targetMapVal, sourceMapVal, NewOptions())

				Expect(err).ToNot(HaveOccurred())
				Expect(merged.Interface()).To(Equal(targetMap))
			})
		})
	})

	Context("non-map type", func() {
		It("errors", func() {
			_, err := mergeMap(reflect.ValueOf("foo"), sourceMapVal, NewOptions())

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("got non-map type"))
		})
	})

	Context("mismatched field types", func() {
		It("errors", func() {
			targetMap["A"] = 0
			_, err := mergeMap(targetMapVal, sourceMapVal, NewOptions())

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("key 'A': types do not match: int, string"))
		})

		Context("submap mismatch", func() {
			It("errors", func() {
				targetMap["C"] = map[string]interface{}{"bar": 0, "baz": "added"}
				_, err := mergeMap(targetMapVal, sourceMapVal, NewOptions())

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("key 'C': key 'bar': types do not match: int, string"))
			})
		})
	})

	Context("mergeFunc returns different type", func() {
		type Bar struct{}

		It("if interface, allows to be set", func() {
			targetMap["E"] = Bar{}
			sourceMap["E"] = Bar{}

			opts := NewOptions()
			// define a merge func that returns wrong type
			opts.mergeFuncs.setTypeMergeFunc(
				reflect.TypeOf(Bar{}),
				func(t, s reflect.Value, o *Options) (reflect.Value, error) {
					return reflect.ValueOf("a string"), nil
				},
			)

			_, err := mergeMap(targetMapVal, sourceMapVal, opts)

			Expect(err).ToNot(HaveOccurred())
		})
	})

})

var _ = Describe("mergeSlice", func() {
	var (
		targetSlice, sourceSlice       []interface{}
		targetSliceVal, sourceSliceVal reflect.Value
	)

	BeforeEach(func() {
		targetSlice = []interface{}{3.6, "unchanged", 0}
		sourceSlice = []interface{}{1, "added", true}
	})

	JustBeforeEach(func() {
		targetSliceVal = reflect.ValueOf(targetSlice)
		sourceSliceVal = reflect.ValueOf(sourceSlice)
	})

	Context("two populated slices", func() {
		It("merges them", func() {
			merged, err := mergeSlice(targetSliceVal, sourceSliceVal, NewOptions())
			Expect(err).ToNot(HaveOccurred())

			mergedSlice, ok := merged.Interface().([]interface{})
			Expect(ok).To(BeTrue())

			Expect(mergedSlice).To(And(
				ContainElement("unchanged"),
				ContainElement(0),
				ContainElement("added"),
				ContainElement(1),
				ContainElement(3.6),
				ContainElement(true),
			))
		})
	})

	Context("target slice is empty", func() {
		BeforeEach(func() {
			targetSlice = []interface{}{}
		})

		It("equals source", func() {
			merged, err := mergeSlice(targetSliceVal, sourceSliceVal, NewOptions())
			Expect(err).ToNot(HaveOccurred())

			mergedSlice, ok := merged.Interface().([]interface{})
			Expect(ok).To(BeTrue())

			Expect(len(mergedSlice)).To(Equal(len(sourceSlice)))
			Expect(mergedSlice).To(And(
				ContainElement("added"),
				ContainElement(1),
				ContainElement(true),
			))
			Expect(mergedSlice).ToNot(And(
				ContainElement("unchanged"),
				ContainElement(0),
				ContainElement(3.6),
			))
		})
	})

	Context("source slice is empty", func() {
		BeforeEach(func() {
			sourceSlice = []interface{}{}
		})

		It("equals target", func() {
			merged, err := mergeSlice(targetSliceVal, sourceSliceVal, NewOptions())
			Expect(err).ToNot(HaveOccurred())

			mergedSlice, ok := merged.Interface().([]interface{})
			Expect(ok).To(BeTrue())

			Expect(len(mergedSlice)).To(Equal(len(targetSlice)))
			Expect(mergedSlice).To(And(
				ContainElement("unchanged"),
				ContainElement(0),
				ContainElement(3.6),
			))
			Expect(mergedSlice).ToNot(And(
				ContainElement("added"),
				ContainElement(1),
				ContainElement(true),
			))
		})
	})

	Context("both slices are empty", func() {
		BeforeEach(func() {
			targetSlice = []interface{}{}
			sourceSlice = []interface{}{}
		})

		It("returns empty slice", func() {
			merged, err := mergeSlice(targetSliceVal, sourceSliceVal, NewOptions())
			Expect(err).ToNot(HaveOccurred())

			mergedSlice, ok := merged.Interface().([]interface{})
			Expect(ok).To(BeTrue())

			Expect(mergedSlice).To(BeEmpty())
		})
	})

	Context("nil values", func() {
		// these are called via merge() because that is what does the nil checks
		By("call via merge()")

		Context("target val is nil", func() {
			JustBeforeEach(func() {
				targetSliceVal = reflect.Value{}
			})

			It("equals source", func() {
				merged, err := merge(targetSliceVal, sourceSliceVal, NewOptions())
				Expect(err).ToNot(HaveOccurred())

				mergedSlice, ok := merged.Interface().([]interface{})
				Expect(ok).To(BeTrue())

				Expect(len(mergedSlice)).To(Equal(len(sourceSlice)))
				Expect(mergedSlice).To(And(
					ContainElement("added"),
					ContainElement(1),
					ContainElement(true),
				))
				Expect(mergedSlice).ToNot(And(
					ContainElement("unchanged"),
					ContainElement(0),
					ContainElement(3.6),
				))
			})
		})

		Context("nil slice target", func() {
			It("equals source", func() {
				var m []interface{}
				targetSliceVal = reflect.ValueOf(m)
				merged, err := merge(targetSliceVal, sourceSliceVal, NewOptions())

				Expect(err).ToNot(HaveOccurred())
				Expect(merged.Interface()).To(Equal(sourceSlice))
			})
		})

		Context("source slice is nil", func() {
			JustBeforeEach(func() {
				sourceSliceVal = reflect.Value{}
			})

			It("equals target", func() {
				merged, err := merge(targetSliceVal, sourceSliceVal, NewOptions())
				Expect(err).ToNot(HaveOccurred())

				mergedSlice, ok := merged.Interface().([]interface{})
				Expect(ok).To(BeTrue())

				Expect(len(mergedSlice)).To(Equal(len(targetSlice)))
				Expect(mergedSlice).To(And(
					ContainElement("unchanged"),
					ContainElement(0),
					ContainElement(3.6),
				))
				Expect(mergedSlice).ToNot(And(
					ContainElement("added"),
					ContainElement(1),
					ContainElement(true),
				))
			})
		})

		Context("both slices are nil", func() {
			JustBeforeEach(func() {
				targetSliceVal = reflect.Value{}
				sourceSliceVal = reflect.Value{}
			})

			It("returns empty slice", func() {
				merged, err := merge(targetSliceVal, sourceSliceVal, NewOptions())
				Expect(err).ToNot(HaveOccurred())

				Expect(merged.IsValid()).ToNot(BeTrue())
			})
		})

	})

	Context("slice types are different", func() {
		It("errors", func() {
			_, err := mergeSlice(reflect.ValueOf([]int{}), reflect.ValueOf([]string{}), NewOptions())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("slices must have same type"))
		})
	})
})

var _ = Describe("mergeStruct", func() {
	type Foo struct {
		Name    string
		Size    int
		Special bool
	}

	var (
		targetStruct, sourceStruct       Foo
		targetStructVal, sourceStructVal reflect.Value
	)

	JustBeforeEach(func() {
		targetStructVal = reflect.ValueOf(targetStruct)
		sourceStructVal = reflect.ValueOf(sourceStruct)
	})

	Context("two populated structs", func() {
		BeforeEach(func() {
			targetStruct = Foo{
				Name:    "target",
				Size:    2,
				Special: false,
			}

			sourceStruct = Foo{
				Name:    "source",
				Size:    4,
				Special: true,
			}
		})

		It("merges correctly", func() {
			merged, err := mergeStruct(targetStructVal, sourceStructVal, NewOptions())
			Expect(err).ToNot(HaveOccurred())
			Expect(merged).ToNot(BeNil())

			mergedStruct, ok := merged.Interface().(Foo)
			Expect(ok).To(BeTrue())
			Expect(mergedStruct.Name).To(Equal(sourceStruct.Name))
			Expect(mergedStruct.Size).To(Equal(sourceStruct.Size))
			Expect(mergedStruct.Special).To(Equal(sourceStruct.Special))
			Expect(targetStruct.Name).To(Equal("target"))
		})

		DescribeTable("pointer combinations",
			func(f func() (reflect.Value, error)) {
				merged, err := f()
				Expect(err).ToNot(HaveOccurred())
				Expect(merged).ToNot(BeNil())
				mergedStruct, ok := merged.Interface().(Foo)
				Expect(ok).To(BeTrue())
				Expect(mergedStruct.Name).To(Equal(sourceStruct.Name))
				Expect(mergedStruct.Size).To(Equal(sourceStruct.Size))
				Expect(mergedStruct.Special).To(Equal(sourceStruct.Special))
				Expect(targetStruct.Name).To(Equal("target"))
			},
			Entry(
				"pointer: T:n S:n",
				func() (reflect.Value, error) {
					return mergeStruct(reflect.ValueOf(targetStruct), reflect.ValueOf(sourceStruct), NewOptions())
				},
			),
			Entry(
				"pointer: T:y S:y",
				func() (reflect.Value, error) {
					return mergeStruct(reflect.ValueOf(&targetStruct), reflect.ValueOf(&sourceStruct), NewOptions())
				},
			),
			Entry(
				"pointer: T:y S:n",
				func() (reflect.Value, error) {
					return mergeStruct(reflect.ValueOf(&targetStruct), reflect.ValueOf(sourceStruct), NewOptions())
				},
			),
			Entry(
				"pointer: T:n S:y",
				func() (reflect.Value, error) {
					return mergeStruct(reflect.ValueOf(targetStruct), reflect.ValueOf(&sourceStruct), NewOptions())
				},
			),
		)
	})

	Context("partially populated", func() {
		Context("target is empty", func() {
			BeforeEach(func() {
				targetStruct = Foo{}
			})

			It("returns source", func() {
				merged, err := mergeStruct(targetStructVal, sourceStructVal, NewOptions())
				Expect(err).ToNot(HaveOccurred())
				Expect(merged).ToNot(BeNil())
				mergedStruct, ok := merged.Interface().(Foo)
				Expect(ok).To(BeTrue())
				Expect(mergedStruct.Name).To(Equal(sourceStruct.Name))
				Expect(mergedStruct.Size).To(Equal(sourceStruct.Size))
				Expect(mergedStruct.Special).To(Equal(sourceStruct.Special))
			})
		})

		Context("source is empty", func() {
			var emptyFoo Foo

			BeforeEach(func() {
				emptyFoo = Foo{}
				sourceStruct = emptyFoo
			})

			It("returns empty", func() {
				merged, err := mergeStruct(targetStructVal, sourceStructVal, NewOptions())
				Expect(err).ToNot(HaveOccurred())
				Expect(merged).ToNot(BeNil())
				mergedStruct, ok := merged.Interface().(Foo)
				Expect(ok).To(BeTrue())
				Expect(mergedStruct.Name).To(Equal(emptyFoo.Name))
				Expect(mergedStruct.Size).To(Equal(emptyFoo.Size))
				Expect(mergedStruct.Special).To(Equal(emptyFoo.Special))
			})
		})
	})

	Context("fields that need to be merged", func() {
		Context("slice fields", func() {
			type Baz struct {
				Slice []interface{}
			}

			var targetBaz, sourceBaz Baz
			var targetBazVal, sourceBazVal reflect.Value

			BeforeEach(func() {
				targetBaz = Baz{
					Slice: []interface{}{"unchanged", 0},
				}
				sourceBaz = Baz{
					Slice: []interface{}{"added", 1},
				}
			})

			JustBeforeEach(func() {
				targetBazVal = reflect.ValueOf(targetBaz)
				sourceBazVal = reflect.ValueOf(sourceBaz)
			})

			It("merges them", func() {
				merged, err := mergeStruct(targetBazVal, sourceBazVal, NewOptions())
				Expect(err).ToNot(HaveOccurred())
				mergedStruct, ok := merged.Interface().(Baz)
				Expect(ok).To(BeTrue())
				Expect(mergedStruct.Slice).To(And(
					ContainElement("unchanged"),
					ContainElement("added"),
					ContainElement(0),
					ContainElement(1),
				))
			})
		})

		Context("pointer fields", func() {
			type Baz struct {
				Ptr   *string
				Slice *[]interface{}
			}

			var targetBaz, sourceBaz Baz
			var targetBazVal, sourceBazVal reflect.Value

			BeforeEach(func() {
				t := "target"
				s := "source"
				targetBaz = Baz{
					Ptr:   &t,
					Slice: &[]interface{}{"unchanged", 0},
				}
				sourceBaz = Baz{
					Ptr:   &s,
					Slice: &[]interface{}{"added", 1},
				}
			})

			JustBeforeEach(func() {
				targetBazVal = reflect.ValueOf(targetBaz)
				sourceBazVal = reflect.ValueOf(sourceBaz)
			})

			It("handles them properly", func() {
				merged, err := mergeStruct(targetBazVal, sourceBazVal, NewOptions())
				Expect(err).ToNot(HaveOccurred())
				mergedStruct, ok := merged.Interface().(Baz)
				Expect(ok).To(BeTrue())
				Expect(*mergedStruct.Ptr).To(Equal("source"))
				Expect(*mergedStruct.Slice).To(And(
					ContainElement("added"),
					ContainElement(1),
				))
			})

			Context("source field is nil", func() {
				BeforeEach(func() {
					sourceBaz.Ptr = nil
					sourceBaz.Slice = nil
				})

				It("handles the nil", func() {
					By("with overwrite")

					merged, err := mergeStruct(targetBazVal, sourceBazVal, NewOptions())
					Expect(err).ToNot(HaveOccurred())
					mergedStruct, ok := merged.Interface().(Baz)
					Expect(ok).To(BeTrue())
					Expect(*mergedStruct.Ptr).To(Equal("target"))
					Expect(*mergedStruct.Slice).To(And(
						ContainElement("unchanged"),
						ContainElement(0),
					))
				})
			})
		})
	})

	Context("invalid merge value", func() {
		type Baz struct {
			Bar *Foo
		}

		var targetBaz, sourceBaz Baz
		var targetBazVal, sourceBazVal reflect.Value

		Context("invalid returned from merge", func() {
			BeforeEach(func() {
				targetBaz = Baz{
					Bar: &Foo{Name: "target"},
				}

				sourceBaz = Baz{
					Bar: &Foo{Name: "source"},
				}
			})

			JustBeforeEach(func() {
				targetBazVal = reflect.ValueOf(targetBaz)
				sourceBazVal = reflect.ValueOf(sourceBaz)
			})

			It("merges them using default", func() {
				opt := NewOptions()
				opt.mergeFuncs.setTypeMergeFunc(reflect.TypeOf(&Foo{}),
					func(t, s reflect.Value, o *Options) (reflect.Value, error) {
						return reflect.ValueOf(nil), nil
					},
				)

				merged, err := mergeStruct(targetBazVal, sourceBazVal, opt)

				Expect(err).ToNot(HaveOccurred())

				mBaz, ok := merged.Interface().(Baz)
				Expect(ok).To(BeTrue())
				Expect(mBaz.Bar.Name).To(Equal("source"))
			})
		})
	})

	// These are tested through the merge() func because that is what protects against panics
	Context("invalid entries", func() {
		var opt *Options

		BeforeEach(func() {
			opt = NewOptions()
			opt.mergeFuncs.setKindMergeFunc(reflect.Struct, mergeStruct)
		})

		Context("nils", func() {
			Context("target nil", func() {
				It("return source", func() {
					merged, err := merge(reflect.ValueOf(nil), reflect.ValueOf(sourceStruct), opt)
					Expect(err).ToNot(HaveOccurred())
					Expect(merged).ToNot(BeNil())
					mergedStruct, ok := merged.Interface().(Foo)
					Expect(ok).To(BeTrue())
					Expect(mergedStruct.Name).To(Equal(sourceStruct.Name))
					Expect(mergedStruct.Size).To(Equal(sourceStruct.Size))
					Expect(mergedStruct.Special).To(Equal(sourceStruct.Special))
				})
			})

			Context("source nil", func() {
				It("return target", func() {
					merged, err := merge(reflect.ValueOf(targetStruct), reflect.ValueOf(nil), opt)
					Expect(err).ToNot(HaveOccurred())
					Expect(merged).ToNot(BeNil())
					mergedStruct, ok := merged.Interface().(Foo)
					Expect(ok).To(BeTrue())
					Expect(mergedStruct.Name).To(Equal(targetStruct.Name))
					Expect(mergedStruct.Size).To(Equal(targetStruct.Size))
					Expect(mergedStruct.Special).To(Equal(targetStruct.Special))
				})

			})
		})

		Context("struct types do not match", func() {
			It("errors", func() {
				type Bar struct {
					Baz float64
				}
				merged, err := merge(reflect.ValueOf(targetStruct), reflect.ValueOf(Bar{}), opt)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("types do not match: conjungo.Foo, conjungo.Bar"))
				Expect(merged.IsValid()).To(BeFalse())
			})
		})
	})

	Context("non-struct type", func() {
		It("errors", func() {
			merged, err := mergeStruct(targetStructVal, reflect.ValueOf("a string"), NewOptions())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("got non-struct kind (tagret: struct; source: string)"))
			Expect(merged.IsValid()).ToNot(BeTrue())
		})
	})

	Context("private fields on struct", func() {
		type Baz struct {
			Public  string
			private string
		}
		var targetBaz, sourceBaz Baz

		BeforeEach(func() {
			targetBaz = Baz{
				Public:  "target",
				private: "target",
			}
			sourceBaz = Baz{
				Public:  "source",
				private: "source",
			}
		})

		var targetBazVal, sourceBazVal reflect.Value

		JustBeforeEach(func() {
			targetBazVal = reflect.ValueOf(targetBaz)
			sourceBazVal = reflect.ValueOf(sourceBaz)
		})

		Context("ErrOnUnexported is true", func() {
			It("errors", func() {
				opt := NewOptions()
				opt.ErrorOnUnexported = true
				merged, err := mergeStruct(targetBazVal, sourceBazVal, opt)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("struct of type Baz has unexported field: private"))
				Expect(merged.IsValid()).ToNot(BeTrue())
			})
		})

		Context("ErrOnUnexported is false", func() {
			var opt *Options

			BeforeEach(func() {
				opt = NewOptions()
				opt.ErrorOnUnexported = false
			})

			Context("overwrite is true", func() {
				It("replaces the whole struct", func() {
					opt.Overwrite = true
					merged, err := mergeStruct(targetBazVal, sourceBazVal, opt)
					Expect(err).ToNot(HaveOccurred())
					Expect(merged).To(Equal(sourceBazVal))
				})
			})

			Context("overwrite is true", func() {
				It("replaces the whole struct", func() {
					opt.Overwrite = false
					merged, err := mergeStruct(targetBazVal, sourceBazVal, opt)
					Expect(err).ToNot(HaveOccurred())
					Expect(merged).To(Equal(targetBazVal))
				})
			})
		})
	})

	Context("can not merge interface field containing different types", func() {
		type Baz struct {
			Foo interface{}
		}
		var targetBaz, sourceBaz Baz

		BeforeEach(func() {
			targetBaz = Baz{
				Foo: "string",
			}
			sourceBaz = Baz{
				Foo: 1,
			}
		})

		var targetBazVal, sourceBazVal reflect.Value

		JustBeforeEach(func() {
			targetBazVal = reflect.ValueOf(targetBaz)
			sourceBazVal = reflect.ValueOf(sourceBaz)
		})

		It("returns error", func() {
			_, err := mergeStruct(targetBazVal, sourceBazVal, NewOptions())
			Expect(err).To(HaveOccurred())
		})
	})

	Context("merge error on field", func() {
		type badType string

		type Baz struct {
			Foo badType
		}
		var targetBaz, sourceBaz Baz

		BeforeEach(func() {
			targetBaz = Baz{
				Foo: "bad",
			}
			sourceBaz = Baz{
				Foo: "blah",
			}
		})

		var targetBazVal, sourceBazVal reflect.Value

		JustBeforeEach(func() {
			targetBazVal = reflect.ValueOf(targetBaz)
			sourceBazVal = reflect.ValueOf(sourceBaz)
		})

		It("returns error", func() {
			opt := NewOptions()
			opt.mergeFuncs.setTypeMergeFunc(reflect.TypeOf(targetBaz.Foo), erroringMergeFunc)

			merged, err := mergeStruct(targetBazVal, sourceBazVal, opt)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to merge field `Baz.Foo`: returns error"))
			Expect(merged.IsValid()).ToNot(BeTrue())
		})
	})

	Context("receives wrong type on merge", func() {
		type Baz struct {
			Bar string
		}
		var (
			targetBaz, sourceBaz Baz
			opt                  *Options
		)

		BeforeEach(func() {
			opt = NewOptions()
			// merge func for string returns an int. wont be able to set the field to an int
			opt.mergeFuncs.setKindMergeFunc(
				reflect.String,
				func(t, s reflect.Value, o *Options) (reflect.Value, error) {
					return reflect.ValueOf(0), nil
				},
			)

			targetBaz = Baz{
				Bar: "target",
			}
			sourceBaz = Baz{
				Bar: "source",
			}
		})

		var targetBazVal, sourceBazVal reflect.Value

		JustBeforeEach(func() {
			targetBazVal = reflect.ValueOf(targetBaz)
			sourceBazVal = reflect.ValueOf(sourceBaz)
		})

		It("errors", func() {
			merged, err := mergeStruct(targetBazVal, sourceBazVal, opt)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("types dont match string <> int"))
			Expect(merged.IsValid()).ToNot(BeTrue())
		})
	})
})
