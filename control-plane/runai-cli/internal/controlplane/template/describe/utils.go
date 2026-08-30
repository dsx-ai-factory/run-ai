package describe

import "encoding/json"

// omitEmpty removes empty arrays, maps, and strings from nested structures
func omitEmpty(data interface{}) interface{} {
	b, _ := json.Marshal(data)
	var m interface{}
	json.Unmarshal(b, &m)
	return clean(m)
}

func clean(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		res := make(map[string]interface{})
		for k, v := range val {
			if c := clean(v); c != nil {
				res[k] = c
			}
		}
		if len(res) == 0 {
			return nil
		}
		return res
	case []interface{}:
		if len(val) == 0 {
			return nil
		}
		res := make([]interface{}, 0, len(val))
		for _, item := range val {
			if c := clean(item); c != nil {
				res = append(res, c)
			}
		}
		if len(res) == 0 {
			return nil
		}
		return res
	case string:
		if val == "" {
			return nil
		}
		return val
	default:
		return val
	}
}
