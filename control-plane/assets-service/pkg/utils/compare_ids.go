package utils

type RequestedVsExisting struct {
	Requested bool
	Existing  bool
}

type RequestedVsExistingMap map[string]RequestedVsExisting

func CompareIds(existing []string, requested []string) RequestedVsExistingMap {
	result := make(RequestedVsExistingMap)
	for _, id := range requested {
		result[id] = RequestedVsExisting{Requested: true}
	}
	for _, id := range existing {
		_, inMap := result[id]
		if inMap {
			result[id] = RequestedVsExisting{Requested: true, Existing: true}
		} else {
			result[id] = RequestedVsExisting{Existing: true}
		}
	}
	return result
}
