package main

// getMapFieldStringValue get value from a map[sting]any map.
// And convert it into string type. Return empty if the conversion failed.
// The keys should all exist as they are popluated by github, to simple the
// code on unnecessary error handling, a empty string is returned.
func getMapFieldStringValue(m map[string]any, key string) string {
	v, ok := m[key].(string)
	if !ok {
		v = ""
	}
	return v
}
