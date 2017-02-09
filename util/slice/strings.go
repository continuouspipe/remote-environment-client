package slice

func RemoveDuplicateString(strings []string) []string {
	var uniqueStrings []string
	for _, val := range strings {
		if ContainString(val, uniqueStrings) == false {
			uniqueStrings = append(uniqueStrings, val)
		}
	}
	return uniqueStrings
}

func ContainString(item string, array []string) bool {
	for _, val := range array {
		if val == item {
			return true
		}
	}
	return false
}
