package internal

func isUriContained(existingUriSlice []string, uri string) bool {
	for _, v := range existingUriSlice {
		if v == uri {
			return true
		}
	}
	return false
}
