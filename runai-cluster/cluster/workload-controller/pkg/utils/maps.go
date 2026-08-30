package utils

func ConcatStringMaps(maps ...map[string]string) map[string]string {
	ret := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			ret[k] = v
		}
	}
	return ret
}

func GetMapKeys[K comparable, T any](mapObj map[K]T) []K {
	keyList := make([]K, 0, len(mapObj))
	for k := range mapObj {
		keyList = append(keyList, k)
	}
	return keyList
}
