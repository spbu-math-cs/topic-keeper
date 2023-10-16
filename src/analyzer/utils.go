package main

func contains(m []string, e string) bool {
	for _, elem := range m {
		if elem == e {
			return true
		}
	}
	return false
}
