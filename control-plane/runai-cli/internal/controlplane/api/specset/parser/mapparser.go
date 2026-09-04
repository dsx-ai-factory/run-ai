/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parser

import (
	"fmt"
	"log"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	mapset "github.com/deckarep/golang-set/v2"
)

const (
	CommaSignStr = ","
	EqualSignStr = "="
)

type MapParser struct {
	name                       string
	strict                     bool
	strMap                     map[string]string
	dic                        mapset.Set[string]
	allowValueContainEqualSign bool
}

func NewStrictMapParser(name string, dictionary mapset.Set[string]) *MapParser {
	return &MapParser{name: name, strict: true, dic: dictionary}
}

func NewLooseMapParser(name string, dictionary mapset.Set[string]) *MapParser {
	return &MapParser{name: name, strict: false, dic: dictionary}
}

func NewStrictMapParserAllowValueContainEqual(name string, dictionary mapset.Set[string]) *MapParser {
	return &MapParser{name: name, strict: true, dic: dictionary, allowValueContainEqualSign: true}
}

func NewLooseMapParserAllowValueContainEqual(name string, dictionary mapset.Set[string]) *MapParser {
	return &MapParser{name: name, strict: false, dic: dictionary, allowValueContainEqualSign: true}
}

func (mp *MapParser) parse(value string) error {
	var err error

	if mp.dic == nil {
		log.Fatal(ErrParserNoDic)
		return ErrParserNoDic
	}
	if value == "" {
		return ErrParseValueIsEmpty
	}

	mp.strMap = make(map[string]string)
	attList := strings.Split(value, CommaSignStr)
	for _, att := range attList {
		equalCounter := strings.Count(att, EqualSignStr)
		switch equalCounter {
		case 0:
			err = mp.setKeyTrue(att)
		case 1:
			err = mp.setKeyValue(att)
		default:
			if !mp.allowValueContainEqualSign {
				mp.strMap = nil
				return fmt.Errorf(consts.ErrorAttributeMultipleEqualSign, att)
			}
			err = mp.setKeyValue(att)
		}

		if err != nil {
			return fmt.Errorf("%s, value: %s", err.Error(), value)
		}
	}
	return mp.checkCardinality()
}

func (mp *MapParser) checkCardinality() error {
	if mp.strict && mp.dic.Cardinality() != len(mp.strMap) {
		var missingList []string
		iter := mp.dic.Iterator()
		for i := range iter.C {
			_, ok := mp.strMap[i]
			if !ok {
				missingList = append(missingList, i)
			}
		}
		return fmt.Errorf("dictionary must contain all keys: %s", missingList)
	}
	return nil
}

func (mp *MapParser) GetMap() map[string]string {
	return mp.strMap
}

func (mp *MapParser) Get(key string) *string {
	value, ok := mp.strMap[key]
	if ok {
		return &value
	}
	return nil
}

func (mp *MapParser) setKeyValue(att string) error {
	att = strings.TrimSpace(att)

	entry := strings.SplitN(att, EqualSignStr, 2)
	if mp.dic.Contains(entry[0]) {
		mp.strMap[entry[0]] = entry[1]
	} else {
		return fmt.Errorf("parser (%s), dictionary does not contain key: %s", mp.name, entry[0])
	}
	return nil
}

func (mp *MapParser) setKeyTrue(att string) error {
	if mp.dic.Contains(att) {
		mp.strMap[att] = ""
	} else if mp.strict {
		return fmt.Errorf("parser (%s), dictionary does not contain key: %s", mp.name, att)
	}
	return nil
}
