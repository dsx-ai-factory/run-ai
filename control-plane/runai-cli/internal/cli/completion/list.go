package completion

import (
	"fmt"
	"sort"
	"strings"
)

type StringBased interface {
	~string // ~string means "underlying type is string"
}

// returns the input after the last comma, for autocomplete it
// returns the previous input, to chain it to the available completions
// return the split input, will help us to create selected columns map.
func ParseColumnsForCompletion(toComplete string) (string, string, []string) {
	var prefix string
	allColumnsSplit := strings.Split(toComplete, ",")
	currentColumn := allColumnsSplit[len(allColumnsSplit)-1]
	lastCommaIndex := strings.LastIndex(toComplete, ",")
	if lastCommaIndex != -1 {
		prefix = toComplete[:lastCommaIndex+1]
	}
	return currentColumn, prefix, allColumnsSplit
}

func CreateSelectedColumnsMap(splitInput []string) map[string]bool {
	selectedColumnsMap := make(map[string]bool)
	for _, column := range splitInput {
		trimmed := strings.TrimSpace(column)
		if trimmed != "" {
			selectedColumnsMap[strings.ToLower(trimmed)] = true
		}

	}
	return selectedColumnsMap
}

func CreateSortedColumnsCompletions[T StringBased](currentInput string, previousInput string, selectedColumnsMap map[string]bool, listColumnMap []T, lowercaseToPascalCase map[T]string) []string {
	completions := make([]string, 0, len(listColumnMap))
	for _, columnKey := range listColumnMap {
		if !selectedColumnsMap[strings.ToLower(string(columnKey))] {
			if currentInput == "" || strings.HasPrefix(string(columnKey), strings.ToLower(currentInput)) {
				completions = append(completions, fmt.Sprintf("%s%s", previousInput, lowercaseToPascalCase[columnKey]))
			}
		}
	}
	sort.Strings(completions)
	return completions
}
