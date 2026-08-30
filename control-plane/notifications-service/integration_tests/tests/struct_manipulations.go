package tests

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

func StructToMap(obj interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}
	err := mapstructure.Decode(obj, &newMap)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return newMap
}
