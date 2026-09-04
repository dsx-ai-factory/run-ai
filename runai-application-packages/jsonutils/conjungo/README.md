<img align="right" src="images/gopher-merge.png">

# conjungo

[![LICENSE](https://img.shields.io/badge/license-MIT-orange.svg)](LICENSE)
[![Golang](https://img.shields.io/badge/Golang-v1.7-blue.svg)](https://golang.org/dl/)
[![Godocs](https://img.shields.io/badge/golang-documentation-blue.svg)](https://godoc.org/github.com/InVisionApp/conjungo)
[![Go Report Card](https://goreportcard.com/badge/github.com/InVisionApp/conjungo)](https://goreportcard.com/report/github.com/InVisionApp/conjungo)
[![Travis Build Status](https://travis-ci.com/InVisionApp/conjungo.svg?token=KosA43m1X3ikri8JEukQ&branch=master)](https://travis-ci.com/InVisionApp/conjungo) 
[![codecov](https://codecov.io/gh/InVisionApp/conjungo/branch/master/graph/badge.svg?token=lesB1PUEtL)](https://codecov.io/gh/InVisionApp/conjungo)

A merge utility designed for flexibility and customizability. The library has a
single simple point of entry that works out of the box for most basic use cases. 
From there, customizations can be made to the way two items are merged to fit 
your specific needs.

Merge any two things of the same type, including maps, slices, structs, and even 
basic types like string and int. By default, the target value will be overwritten 
by the source. If the overwrite option is turned off, only new values in source 
that do not already exist in target will be added.  

If you would like to change the way two items of a particular type get merged,
custom merge functions can be defined for any type or kind (see below).  

## Why Conjungo?

The definition of Conjungo:  
> I.v. a., to bind together, connect, join, unite (very freq. in all perr. and species of composition); constr. with cum, inter se, the dat., or the acc. only; trop. also with ad.

Reference: [Latin Dictionary...](http://www.perseus.tufts.edu/hopper/text?doc=Perseus:text:1999.04.0059:entry=conjungo)

There are other merge libraries written in go, but none of them have the 
flexibility of this one. If you simply need to merge two things, a default 
set of merge functions are defined for merging maps, slices, and structs like 
most other libraries. But, if the way these default functions are defined does 
not meet your needs, Conjungo provides the ability to define your own merge 
functions. For instance, the default behavior when merging two integers is to 
replace the target with the source, but if you'd like to redefine that, you 
can write a custom merge function that is used when assessing integers. A custom 
function could add the two integers and return the result, or return the larger 
of the two integers. You could define a custom merge function for a specific struct 
type in your code, and define how that gets merged. The customizability of how 
things get merged is the focus of Conjungo. 

The need for this library arose when we were merging large custom structs. We 
found that there was no single library that merged all the parts of the struct 
in the way that we needed. We had struct fields that were pointers to sub structs 
and maps that needed to be followed instead of simply replaced. We had slices 
that needed to be appended but also deduped. Conjungo solves these types of 
problems by allowing custom functions to be defined to handle each type.

## Setup
To get **conjungo**:
```sh
go get github.com/InVisionApp/conjungo
```

We recommend that you vendor it within your project. We chose to use govendor.

```sh 
govendor fetch github.com/InVisionApp/conjungo
```

## Usage
### Simple Merge
Merge two structs together:
```go
type Foo struct {
	Name    string
	Size    int
	Special bool
	SubMap  map[string]string
}

targetStruct := Foo{
	Name:    "target",
	Size:    2,
	Special: false,
	SubMap:  map[string]string{"foo": "unchanged", "bar": "orig"},
}

sourceStruct := Foo{
	Name:    "source",
	Size:    4,
	Special: true,
	SubMap:  map[string]string{"bar": "newVal", "safe": "added"},
}

err := conjungo.Merge(&targetStruct, sourceStruct, nil)
if err != nil {
	log.Error(err)
}
```
results in:
```json
{
  "Name": "source",
  "Size": 4,
  "Special": true,
  "SubMap": {
    "bar": "newVal",
    "foo": "unchanged",
    "safe": "added"
  }
}
```

### Options
**Overwrite** `bool`  
If true, overwrite a target value with source value even if it already exists

**ErrorOnUnexported** `bool`  
Unexported fields on a struct can not be set. When a struct contains an unexported
field, the default behavior is to treat the entire struct as a single entity and
replace according to Overwrite settings.  
If this is enabled, an error will be thrown instead.

### Custom Merge Functions
#### Define a custom merge function for a type:
```go
opts := conjungo.NewOptions()
opts.MergeFuncs.SetTypeMergeFunc(
	reflect.TypeOf(0),
	// merge two 'int' types by adding them together
	func(t, s reflect.Value, o *conjungo.Options) (reflect.Value, error) {
		iT, _ := t.Interface().(int)
		iS, _ := s.Interface().(int)
		return reflect.ValueOf(iT + iS), nil
	},
)

x := 1
y := 2

err := conjungo.Merge(&x, y, opts)
if err != nil {
	log.Error(err)
}

// x == 3
```

#### Define a custom merge function for a kind:
```go
opts := conjungo.NewOptions()
opts.MergeFuncs.SetKindMergeFunc(
	reflect.TypeOf(struct{}{}).Kind(),
	// merge two 'struct' kinds by replacing the target with the source
	// provides a mechanism to set override = true for just structs
	func(t, s reflect.Value, o *conjungo.Options) (reflect.Value, error) {
		return s, nil
	},
)
```

#### Define a custom merge function for a struct type:
```go
	type Foo struct {
		Name string
		Size int
	}

	target := Foo{
		Name: "bar",
		Size: 25,
	}

	source := Foo{
		Name: "baz",
		Size: 35,
	}

	opts := conjungo.NewOptions()
	opts.MergeFuncs.SetTypeMergeFunc(
		reflect.TypeOf(Foo{}),
		// merge two 'int' types by adding them together
		func(t, s reflect.Value, o *conjungo.Options) (reflect.Value, error) {
			tFoo := t.Interface().(Foo)
			sFoo := s.Interface().(Foo)

			// names are merged by concatenating them
			tFoo.Name = tFoo.Name + "." + sFoo.Name
			// sizes are merged by averaging them
			tFoo.Size = (tFoo.Size + sFoo.Size) / 2

			return reflect.ValueOf(tFoo), nil
		},
	)
```

#### Define a custom type and a function to merge it:
```go
type jsonString string

var targetJSON jsonString = `
{
  "a": "wrong",
  "b": 1,
  "c": {"bar": "orig", "foo": "unchanged"},
}`

var sourceJSON jsonString = `
{
  "a": "correct",
  "b": 2,
  "c": {"bar": "newVal", "safe": "added"},
}`

opts := conjungo.NewOptions()
opts.MergeFuncs.SetTypeMergeFunc(
	reflect.TypeOf(jsonString("")),
	// merge two json strings by unmarshalling them to maps
	func(t, s reflect.Value, o *conjungo.Options) (reflect.Value, error) {
		targetStr, _ := t.Interface().(jsonString)
		sourceStr, _ := s.Interface().(jsonString)

		targetMap := map[string]interface{}{}
		if err := json.Unmarshal([]byte(targetStr), &targetMap); err != nil {
			return reflect.Value{}, err
		}

		sourceMap := map[string]interface{}{}
		if err := json.Unmarshal([]byte(sourceStr), &sourceMap); err != nil {
			return reflect.Value{}, err
		}

		err := conjungo.Merge(&targetMap, sourceMap, o)
		if err != nil {
			return reflect.Value{}, err
		}

		mergedJSON, err := json.Marshal(targetMap)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(jsonString(mergedJSON)), nil
	},
)
```

See [working examples](_example/main.go) for more details.
