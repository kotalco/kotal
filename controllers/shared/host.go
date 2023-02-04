package shared

// Host returns localhost of toggle is not enabled
// otherwise returns the wildcard address
func Host(toggle bool) string {
	if toggle {
		return "0.0.0.0"
	} else {
		return "127.0.0.1"
	}
}
