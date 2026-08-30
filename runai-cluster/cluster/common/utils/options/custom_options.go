package options

import (
	"fmt"
	"regexp"
	"strings"
)

type ArrayFlags []string

func (i *ArrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *ArrayFlags) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	compiledRegExp := regexp.MustCompile("\\s*,\\s*")
	valueRemovedWhiteSpaces := compiledRegExp.ReplaceAllString(value, ",")
	*i = strings.Split(valueRemovedWhiteSpaces, ",")
	return nil
}
