/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestDmodel — the Ginkgo suite runner for this package — is defined in
// get_asset_id_test.go and picks up every `var _ = Describe(...)` in the
// package.

// unmarshalableChannelPayload is an intentionally-non-json-encodable value
// used to verify that ToJsonBMap surfaces json.Marshal failures unchanged.
// A channel is the smallest thing encoding/json refuses to serialize.
func unmarshalableChannelPayload() any {
	type broken struct {
		Ch chan int `json:"ch"`
	}
	return broken{Ch: make(chan int)}
}

var _ = Describe("JsonBMapFromString.Value()", func() {
	It("returns the JSON literal \"null\" for a nil map (skips json.Marshal)", func() {
		var m JsonBMapFromString
		v, err := m.Value()
		Expect(err).To(BeNil())
		Expect(v).To(Equal("null"))
	})

	It("returns \"{}\" for an empty non-nil map", func() {
		m := JsonBMapFromString{}
		v, err := m.Value()
		Expect(err).To(BeNil())
		Expect(v).To(Equal("{}"))
	})

	It("serializes a non-empty map to JSON", func() {
		m := JsonBMapFromString{"k": "v", "n": float64(42)}
		v, err := m.Value()
		Expect(err).To(BeNil())

		// Round-trip through json.Unmarshal to compare structurally — Go's
		// json.Marshal doesn't guarantee key order on maps, so comparing the
		// raw string would be flaky.
		out := map[string]any{}
		Expect(json.Unmarshal([]byte(v.(string)), &out)).To(BeNil())
		Expect(out).To(Equal(map[string]any{"k": "v", "n": float64(42)}))
	})
})

var _ = Describe("JsonBMapFromString.Scan()", func() {
	// Scan uses gojsonq.LoadJsonBytes which decodes JSON numbers as
	// json.Number (i.e. Decoder.UseNumber()), not float64. That contract
	// matters downstream — dbmodel.FromJsonBMap round-trips these values
	// into typed swagger structs, and json.Number preserves precision for
	// int64 / uint64 fields that float64 would silently round.
	It("populates the map from a JSON []byte payload (numbers as json.Number)", func() {
		var m JsonBMapFromString
		Expect(m.Scan([]byte(`{"a":1,"b":"x"}`))).To(BeNil())
		Expect(m).To(Equal(JsonBMapFromString{"a": json.Number("1"), "b": "x"}))
	})

	It("populates the map from a JSON string payload (numbers as json.Number)", func() {
		var m JsonBMapFromString
		Expect(m.Scan(`{"a":1,"b":"x"}`)).To(BeNil())
		Expect(m).To(Equal(JsonBMapFromString{"a": json.Number("1"), "b": "x"}))
	})

	It("returns an explanatory error for an unsupported scan type", func() {
		var m JsonBMapFromString
		err := m.Scan(42)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("invalid scan type"))
	})

	It("propagates the parse error for malformed JSON", func() {
		var m JsonBMapFromString
		err := m.Scan([]byte(`{not: valid json`))
		Expect(err).NotTo(BeNil())
	})
})

var _ = Describe("ToJsonBMap()", func() {
	type testStruct struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	It("marshals a typed struct with json tags into a map", func() {
		m, err := ToJsonBMap(testStruct{Name: "gpu", Count: 3})
		Expect(err).To(BeNil())
		Expect(m).To(Equal(JsonBMapFromString{"name": "gpu", "count": float64(3)}))
	})

	It("marshals a pointer to a struct identically to its value", func() {
		v := testStruct{Name: "gpu", Count: 3}
		mFromVal, errVal := ToJsonBMap(v)
		mFromPtr, errPtr := ToJsonBMap(&v)
		Expect(errVal).To(BeNil())
		Expect(errPtr).To(BeNil())
		Expect(mFromPtr).To(Equal(mFromVal))
	})

	It("round-trips an existing map[string]any unchanged (modulo json number type)", func() {
		in := map[string]any{"k": "v", "nested": map[string]any{"a": true}}
		m, err := ToJsonBMap(in)
		Expect(err).To(BeNil())
		Expect(m).To(Equal(JsonBMapFromString(in)))
	})

	It("returns nil map when the input marshals to JSON null (e.g. untyped nil)", func() {
		// json.Marshal(nil) yields "null" and unmarshalling "null" into a
		// map leaves the map at its zero value (nil). Not an error.
		m, err := ToJsonBMap(nil)
		Expect(err).To(BeNil())
		Expect(m).To(BeNil())
	})

	It("returns a marshal error for values encoding/json refuses to serialize", func() {
		_, err := ToJsonBMap(unmarshalableChannelPayload())
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("marshal value to json"))
	})

	It("returns an unmarshal error when the JSON is not a JSON object (e.g. array input)", func() {
		// map[string]any cannot receive a JSON array — this is the failure
		// mode a caller would hit if they accidentally passed a slice.
		_, err := ToJsonBMap([]int{1, 2, 3})
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("unmarshal json"))
	})
})

var _ = Describe("FromJsonBMap()", func() {
	type testStruct struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	It("materializes a typed struct from a map", func() {
		src := JsonBMapFromString{"name": "gpu", "count": float64(3)}
		var out testStruct
		Expect(FromJsonBMap(src, &out)).To(BeNil())
		Expect(out).To(Equal(testStruct{Name: "gpu", Count: 3}))
	})

	It("round-trips: ToJsonBMap then FromJsonBMap reconstructs the original struct", func() {
		orig := testStruct{Name: "gpu", Count: 3}
		mid, err := ToJsonBMap(orig)
		Expect(err).To(BeNil())

		var out testStruct
		Expect(FromJsonBMap(mid, &out)).To(BeNil())
		Expect(out).To(Equal(orig))
	})

	It("leaves the destination at its zero value when the source is a nil map (unmarshals JSON null)", func() {
		out := testStruct{Name: "prefilled", Count: 99}
		// json.Marshal(nil map) is "null"; json.Unmarshal("null", &struct{})
		// is a no-op — the destination is not zeroed. Verified here so
		// callers relying on FromJsonBMap don't assume otherwise.
		Expect(FromJsonBMap(nil, &out)).To(BeNil())
		Expect(out).To(Equal(testStruct{Name: "prefilled", Count: 99}))
	})

	It("returns an unmarshal error when the destination type is incompatible", func() {
		src := JsonBMapFromString{"name": "gpu"}
		var out int
		err := FromJsonBMap(src, &out)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("unmarshal json"))
	})
})
