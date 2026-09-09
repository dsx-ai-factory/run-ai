package conjungo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// Options is used to determine the behavior of a merge.
// It also holds the collection of functions used to determine merge behavior of various types.
// Always use NewOptions() to generate options and then modify as needed.
type Options struct {
	// Overwrite a target value with source value even if it already exists
	Overwrite bool

	// Unexported fields on a struct can not be set. When a struct contains an unexported
	// field, the default behavior is to treat the entire struct as a single entity and
	// replace according to Overwrite settings. If this is enabled, an error will be thrown instead.
	//
	// Note: this is used by the default mergeStruct function, and may not apply if that is
	// overwritten with a custom function. Custom struct merge functions should consider
	// using this value as well.
	ErrorOnUnexported bool

	// A set of default and customizable functions that define how values are merged
	// Use the following to define custom merge behavior
	//		Options.SetTypeMergeFunc(t reflect.Type, mf MergeFunc)
	//		Options.SetKindMergeFunc(k reflect.Kind, mf MergeFunc)
	//		Options.SetDefaultMergeFunc(mf MergeFunc)
	mergeFuncs *funcSelector

	// To be used by merge functions to pass values down into recursive calls freely
	Context context.Context
}

// NewOptions generates default Options. Overwrite is set to true, and a set of
// default merge function definitions are added.
func NewOptions() *Options {
	return &Options{
		Overwrite:  true,
		mergeFuncs: newFuncSelector(),
	}
}

// SetTypeMergeFunc is used to define a custom merge func that will be used to merge two
// items of a particular type. Accepts the reflect.Type representation of the type and
// the MergeFunc to merge it.
// This is useful for defining specific merge behavior of things such as specific struct types
func (o *Options) SetTypeMergeFunc(t reflect.Type, mf MergeFunc) {
	o.mergeFuncs.setTypeMergeFunc(t, mf)
}

// SetKindMergeFunc is used to define a custom merge func that will be used to merge two
// items of a particular kind. Accepts reflect.Kind and the MergeFunc to merge it.
// This is useful for defining more general merge behavior, for instance
// merge all maps or structs in a particular way.
// A default merge behavior is predefined for map, slice and struct when using NewOptions()
func (o *Options) SetKindMergeFunc(k reflect.Kind, mf MergeFunc) {
	o.mergeFuncs.setKindMergeFunc(k, mf)
}

// SetDefaultMergeFunc is used to define a default merge func that will be used as a fallback
// when there is no specific merge behavior defined for a given item.
// If using NewOptions(), a very basic default merge function is predefined which will
// return the source in overwrite mode and the target otherwise.
// Use this to define custom default behavior when the simple case is not sufficient.
func (o *Options) SetDefaultMergeFunc(mf MergeFunc) {
	o.mergeFuncs.setDefaultMergeFunc(mf)
}

var valType = reflect.TypeOf(reflect.Value{})

// Merge the given source onto the given target following the options given. The target value
// must be a pointer. If opt is nil, defaults will be used. If an error occurs during
// the merge process the target will be unmodified. Merge will accept any two entities,
// as long as their types are the same.
// See Options and MergeFunc for further customization possibilities.
func Merge(target, source interface{}, opt *Options) error {
	vT := reflect.ValueOf(target)
	vS := reflect.ValueOf(source)

	if target != nil && vT.Type() == valType {
		vT = vT.Interface().(reflect.Value)
	}
	if source != nil && vS.Type() == valType {
		vS = vS.Interface().(reflect.Value)
	}

	if vT.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer")
	}

	if !reflect.Indirect(vT).IsValid() {
		return errors.New("target can not be zero value")
	}

	// use defaults if none are provided
	if opt == nil {
		opt = NewOptions()
	}

	if opt.mergeFuncs == nil {
		return errors.New("invalid options, use NewOptions() to generate and then modify as needed")
	}

	//make a copy here so if there is an error mid way, the target stays in tact
	cp := vT.Elem()

	merged, err := merge(cp, reflect.Indirect(vS), opt)
	if err != nil {
		return err
	}

	if !isSettable(vT.Elem(), merged) {
		return fmt.Errorf("Merge failed: expected merged result to be %v but got %v",
			vT.Elem().Type(), merged.Type())
	}

	vT.Elem().Set(merged)
	return nil
}

func isSettable(t, s reflect.Value) bool {
	if t.Kind() != reflect.Interface && t.Type() != s.Type() {
		return false
	}

	return true
}

func merge(valT, valS reflect.Value, opt *Options) (reflect.Value, error) {
	// if source is nil, skip
	if isEmpty(valS) {
		return valT, nil
	}

	// if target is nil write to it
	if isEmpty(valT) {
		return valS, nil
	}

	// get to the real type
	if valT.Kind() == reflect.Interface || valS.Kind() == reflect.Interface {
		valT = reflect.ValueOf(valT.Interface())
		valS = reflect.ValueOf(valS.Interface())
	}

	// if types do not match, bail
	if valT.Type() != valS.Type() {
		return reflect.Value{}, fmt.Errorf("types do not match: %v, %v", valT.Type(), valS.Type())
	}

	// look for a merge function
	f := opt.mergeFuncs.getFunc(valT)
	val, err := f(valT, valS, opt)
	if err != nil {
		return reflect.Value{}, err
	}

	return val, nil
}

func isEmpty(val reflect.Value) bool {
	// is zero value
	if !val.IsValid() {
		return true
	}

	switch val.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface, reflect.Slice:
		if val.IsNil() {
			return true
		}
	}

	return false
}
