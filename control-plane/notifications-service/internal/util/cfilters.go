package util

import (
	"github.com/rs/zerolog/log"
)

func FilterMap[KEY comparable, VAL any](m map[KEY]VAL, keepPredicate func(key KEY, val VAL) bool) map[KEY]VAL {
	ret := map[KEY]VAL{}
	if len(m) == 0 {
		return ret
	}

	for key, val := range m {
		if keepPredicate(key, val) {
			ret[key] = val
		}
	}
	return ret
}

func FilterMapByKeys[KEY comparable, VAL any](m map[KEY]VAL, keys ...KEY) map[KEY]VAL {
	if len(keys) == 0 {
		return m
	}

	ret := map[KEY]VAL{}
	for _, key := range keys {
		if val, found := m[key]; found {
			ret[key] = val
		}
	}
	return ret
}

func FilterSlice[VAL any](slice []VAL, keepPredicate func(idx int, val VAL) bool) []VAL {
	ret := []VAL{}
	if len(slice) == 0 {
		return ret
	}

	for idx, val := range slice {
		if keepPredicate(idx, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

func MigrateSlice[FROM, TO any](slice []FROM, migrate func(idx int, val FROM) (TO, error)) ([]TO, error) {
	ret := []TO{}
	if len(slice) == 0 {
		return ret, nil
	}

	for idx, val := range slice {
		m, err := migrate(idx, val)
		if err != nil {
			return nil, err
		}
		ret = append(ret, m)
	}
	return ret, nil
}

func MigrateSliceToMap[FROM, TO any, KEY comparable](slice []FROM, migrate func(idx int, val FROM) (KEY, TO, error)) map[KEY]TO {
	ret := map[KEY]TO{}
	if len(slice) == 0 {
		return ret
	}

	for idx, val := range slice {
		toKey, toVal, err := migrate(idx, val)
		if err != nil {
			log.Err(err).Msgf("failed migrating an item at idx %d", idx)
		} else {
			ret[toKey] = toVal
		}
	}
	return ret
}
