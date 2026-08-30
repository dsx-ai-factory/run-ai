package utils

import (
	"encoding/json"
	"strings"

	jsonq "github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
)

// JsonqDelim is the default jsonQ delimiter is "." which is ok for out synthetic resources.
// note that runai job can contain elements like "resources.gpu.limit". if this is ever
// used for json representation of runaijob - choose a different dlimiter.
const JsonqDelim = "."

// LoadJson loads any object to a generic map
func LoadJson(anyObject interface{}) (map[string]interface{}, error) {

	//
	//   convert the object to json string
	//
	srJsonBytes, err := json.MarshalIndent(anyObject, "", "\t")
	if err != nil {
		return nil, err
	}

	//
	//  load the string to a generic map.
	//
	jsonMap := make(map[string]interface{})
	decoder := json.NewDecoder(strings.NewReader(string(srJsonBytes)))

	//
	//  we want to use numbers rather than the mighty default float64, which is not accepted by some
	//  json libraries
	//
	decoder.UseNumber()

	err = decoder.Decode(&jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonMap, nil
}

// NewJsonQ load an object to JsonQ
func NewJsonQ(anObject interface{}) (*jsonq.JSONQ, error) {

	jsonMap, err := LoadJson(anObject)
	if err != nil {
		return nil, err
	}
	jq := jsonq.New(jsonq.SetSeparator(JsonqDelim)).FromInterface(jsonMap)
	if jq.Error() != nil {
		return nil, jq.Error()
	}

	return jq, nil
}
