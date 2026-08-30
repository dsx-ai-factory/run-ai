package gojsonq

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/run-ai/runai/runai-common-packages/logging"
)

// JsonqDelim - the default jsonQ delimiter is "." which is ok for out synthetic resources.
// note that runai job can contain elements like "resources.gpu.limit". if this is ever
// used for json representation of runaijob - choose a different delimiter.
const JsonqDelim = "."

// LoadJson loads any object to a generic map
func LoadJson(anyObject interface{}) (map[string]interface{}, error) {
	//
	//   convert the object to json string
	//
	arJsonBytes, err := json.MarshalIndent(anyObject, "", "\t")
	if err != nil {
		return nil, err
	}

	return LoadJsonBytes(arJsonBytes)
}

func ToBytes(jsonMap map[string]interface{}) []byte {
	bytes, err := json.Marshal(jsonMap)
	if err != nil {
		logging.Errorf("failed to convert json to bytes: %v", err)
	}
	return bytes
}

func LoadJsonBytes(arJsonBytes []byte) (map[string]interface{}, error) {
	//
	//  load the string to a generic map.
	//
	jsonMap := make(map[string]interface{})
	decoder := json.NewDecoder(strings.NewReader(string(arJsonBytes)))

	//
	//  we want to use numbers rather than the mighty default float64, which is not accepted by some
	//  json libraries
	//
	decoder.UseNumber()

	err := decoder.Decode(&jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonMap, nil
}

// NewJsonQ loads an object to JsonQ
func NewJsonQ(anObject interface{}) (*JSONQ, error) {

	jsonMap, err := LoadJson(anObject)
	if err != nil {
		return nil, err
	}
	jq := New(SetSeparator(JsonqDelim)).FromInterface(jsonMap)
	if jq.Error() != nil {
		return nil, jq.Error()
	}

	return jq, nil
}

// CleanMap "cleans" entries which are null or having no keys
// for example:
//
//	environment:
//	  image: null
//	  environmentVariables: {}
//	  imagePullPolicy: Always
//
// after clean should look like this:
//
//	environment:
//	  imagePullPolicy: always
//
// because the image and environmentVariables are empty
func CleanMap(aMap map[string]interface{}) {
	for key, value := range aMap {
		switch v := value.(type) {
		case nil:
			delete(aMap, key)
		case map[string]interface{}:
			CleanMap(v)
			if len(v) == 0 {
				delete(aMap, key)
			}
		case []interface{}:
			cleaned := v[:0] // reuse backing array
			for _, elem := range v {
				if em, ok := elem.(map[string]interface{}); ok {
					CleanMap(em)
					if len(em) == 0 {
						continue // drop emptied element
					}
				} else if elem == nil {
					continue // drop nil element
				}
				cleaned = append(cleaned, elem)
			}

			if len(cleaned) == 0 {
				delete(aMap, key) // optional: also drop empty slices
			} else {
				aMap[key] = cleaned
			}
		}
	}
}

// NormalizeNumbers walks v and rewrites every numeric value (float64, float32
// and the signed/unsigned integer kinds) as a json.Number, mirroring the
// representation that LoadJson (decoder.UseNumber()) produces.
//
// It exists because some producers - notably the standard json decoder without
// UseNumber and the gojq expression engine - emit float64 for JSON numbers.
// When such a value is later merged against data that WAS loaded with
// UseNumber, strict type-equality checks (e.g. conjungo's merge) fail with
// "types do not match: json.Number, float64". Running NormalizeNumbers over the
// float64-tainted side restores the json.Number invariant the rest of the code
// relies on.
//
// Maps and slices are converted in place; the (possibly replaced) value is also
// returned so scalar inputs can be reassigned by the caller:
//
//	v = NormalizeNumbers(v)
func NormalizeNumbers(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, e := range val {
			val[k] = NormalizeNumbers(e)
		}
		return val
	case []interface{}:
		for i, e := range val {
			val[i] = NormalizeNumbers(e)
		}
		return val
	case json.Number:
		return val
	case float32, float64:
		if n, ok := floatToJSONNumber(val); ok {
			return n
		}
		// Non-finite floats (NaN, +/-Inf) have no JSON number representation, so
		// leave them untouched rather than emit an invalid json.Number.
		return v
	case int, int8, int16, int32, int64:
		return json.Number(strconv.FormatInt(reflect.ValueOf(val).Int(), 10))
	case uint, uint8, uint16, uint32, uint64:
		return json.Number(strconv.FormatUint(reflect.ValueOf(val).Uint(), 10))
	default:
		return v
	}
}

// floatToJSONNumber marshals f (a float32 or float64) the same way encoding/json
// does, preserving the value's original precision so the resulting json.Number
// literal is identical to what decoder.UseNumber() would have produced - e.g.
// float32(0.1) becomes "0.1", not the widened "0.10000000149011612". It reports
// ok=false for the non-finite values (math.NaN(), math.Inf()), which have no JSON
// number representation, so the caller can leave them untouched.
func floatToJSONNumber(f interface{}) (json.Number, bool) {
	var f64 float64
	switch n := f.(type) {
	case float32:
		f64 = float64(n)
	case float64:
		f64 = n
	default:
		return "", false
	}

	if math.IsNaN(f64) || math.IsInf(f64, 0) {
		return "", false
	}

	// Marshal the original typed value (not the widened float64) so a float32
	// keeps its shortest 32-bit round-trip form.
	b, err := json.Marshal(f)
	if err != nil {
		return "", false
	}
	return json.Number(b), true
}

// RemoveIdenticalKeys remove from destMap all the keys that also exist in sourceMap
// for example
//
//	destMap:
//	  environmentVariables:
//	    AAA: value-of-a
//	    BBB; value-of-b
//	sourceMap:
//	    BBB: another-value-of-b
//
// the result will be //
//
//	destMap:
//	  AAA: value-of-a
//
// and this works recursively on nested keys as well
func RemoveIdenticalKeys(destMap, sourceMap map[string]interface{}) {
	for key := range sourceMap {
		if nestedMap, ok := sourceMap[key].(map[string]interface{}); ok {
			if nestedMap1, ok := destMap[key].(map[string]interface{}); ok {
				RemoveIdenticalKeys(nestedMap1, nestedMap)
			}
		} else {
			delete(destMap, key)
		}
	}
}
