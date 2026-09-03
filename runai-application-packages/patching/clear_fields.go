/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patching

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PaesslerAG/jsonpath"
)

var exprFieldAndIndex = `^([^\[\]]+)(?:\[(\d+)\])?$`
var reFieldAndIndex = regexp.MustCompile(exprFieldAndIndex)

type ClearFieldsOptions struct {
	IgnoreNonExisting bool
}

// SpecReference is needed for cleaning arrays at the end of the process
// see more explanations in the functions below
type SpecReference struct {
	Parent    map[string]interface{}
	FieldName string
}

// ClearFields receives a json payload such as
//
//	  spec:
//	    image: ubuntu
//	    environmentVariables:
//	       - name: env-1
//	         value: value-1
//	And a set of json paths such as
//	   spec.image
//	   spec.environmentVariables[0]
//	And it removed from the payload all the entries that match the xpaths
//	There is an option to ignore non-existent elements otherwise error should return on non-existing elements
func ClearFields(spec map[string]interface{}, jsonPaths []string, options ClearFieldsOptions) error {
	arraysToClean := make([]*SpecReference, 0)

	for _, jsonPath := range jsonPaths {
		arrayToClean, pathErr := removeByJsonPath(spec, jsonPath, options)
		if pathErr != nil {
			return pathErr
		}

		if arrayToClean != nil {
			arraysToClean = append(arraysToClean, arrayToClean)
		}
	}

	//
	//   arrays entries such as spec.environment[1] are not immediately removed, but are marked
	//   for removed. the reason is that if there are 5 entries and user gives us two paths
	//   to remove items #1 and #3, and we remove at once:
	//       item#0
	//       item#1   <-- this one removed
	//       item#2
	//       item#3   <-- this one is still to removed
	//       item#4
	//   now that we removed #1, removing #3 will effectively remove #4 which is not good. so
	//   what we do is to mark both of them to be removed and in this loop we remove them completely
	//
	for _, specPointer := range arraysToClean {
		arrayToClean := specPointer.Parent[specPointer.FieldName].([]interface{})

		cleanedArray := make([]interface{}, 0, len(arrayToClean))

		for _, item := range arrayToClean {
			if item != nil {
				cleanedArray = append(cleanedArray, item)
			}
		}

		specPointer.Parent[specPointer.FieldName] = cleanedArray
	}

	return nil
}

// we support the following jsonPaths:
//
//	spec -> flat field
//	spec.image -> flat field
//	spec.compute.gpuDevicesRequest -> flat field
//	spec.compute.extendedResources[0] -> remove the first instance entirely
//	spec.compute.extendedResources[0].quantity -> remove the quantity from first instance of
//	master.image -> flat field from master spec
func removeByJsonPath(specAsJson map[string]interface{}, path string, options ClearFieldsOptions) (*SpecReference, error) {
	pathParts := strings.Split(path, ".")
	numParts := len(pathParts)

	var parentAsMap map[string]interface{}
	var lastPart string

	if numParts == 1 {
		parentAsMap = specAsJson
		lastPart = pathParts[0]
	} else {
		var ok bool

		parentRef := strings.Join(pathParts[0:numParts-1], ".")
		lastPart = pathParts[numParts-1]

		parent, pathErr := jsonpath.Get(parentRef, specAsJson)
		if pathErr != nil {
			return nil, errorOrIgnore(options, fmt.Errorf("path %s not found: %v", path, pathErr))
		}

		parentAsMap, ok = parent.(map[string]interface{})
		if !ok {
			return nil, errorOrIgnore(options, fmt.Errorf("path %s does not point to a field in the spec", path))
		}
	}

	fieldName, index, err := splitFieldAndIndex(path, lastPart)
	if err != nil {
		return nil, err
	}

	field, ok := parentAsMap[fieldName]
	if !ok {
		return nil, errorOrIgnore(options, fmt.Errorf("%s: field %s does not exist", path, fieldName))
	}

	// flat field - remove from the spec and we're done
	if index < 0 {
		delete(parentAsMap, fieldName)
		return nil, nil
	}

	// array, need to mark for deletion
	fieldAsArray, ok := field.([]interface{})
	if !ok {
		return nil, fmt.Errorf("field %s is not itemized", fieldName)
	}

	if index >= len(fieldAsArray) {
		return nil, errorOrIgnore(options, fmt.Errorf("%s[%d] exceeds the number of instances %d", fieldName, index, len(fieldAsArray)))
	}

	//  note that we do not remove the instance yet, because there may be removal of multiple instances
	//  on the same array. only at the end we will clean them
	fieldAsArray[index] = nil

	return &SpecReference{
		Parent:    parentAsMap,
		FieldName: fieldName,
	}, nil
}

// possible values
//
//	xxx -> returns xxx, -1
//	xxx[2] -> return xxx, 2
func splitFieldAndIndex(path string, str string) (string, int, error) {
	matches := reFieldAndIndex.FindStringSubmatch(str)

	if len(matches) == 0 {
		return "", -1, fmt.Errorf("invalid path '%s': expecting last part to be field[index], index>=0", path)
	}

	// If no index part is found, default to -1
	index := -1
	if matches[2] != "" {
		// Parse the index if it exists. no need for error checking
		// because the regular expression already verified the validity
		_, _ = fmt.Sscanf(matches[2], "%d", &index)
	}

	return matches[1], index, nil
}

func errorOrIgnore(options ClearFieldsOptions, error error) error {
	if options.IgnoreNonExisting {
		return nil
	}
	return error
}
