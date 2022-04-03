package util

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func HasKey(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}
