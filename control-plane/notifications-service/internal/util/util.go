package util

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"github.com/xyproto/jpath"
)

func ListToSet[T comparable](list []T) map[T]struct{} {
	result := make(map[T]struct{})
	for _, item := range list {
		result[item] = struct{}{}
	}
	return result
}

func AggregateRegexList(list []string) *regexp.Regexp {
	regexStr := strings.Join(list, "|")
	return regexp.MustCompile(regexStr)
}

func StringPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ReadJsonConfig(jsonStr string, v *viper.Viper) error {
	v.SetConfigType("json")
	return v.ReadConfig(strings.NewReader(jsonStr))
}

func MapToJPath(m map[string]interface{}) (*jpath.Node, error) {
	mStr, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return jpath.New(mStr)
}
