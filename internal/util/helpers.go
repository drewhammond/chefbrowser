package util

// MakeJSONPath will take a JSON object and return flattened JSON paths
func MakeJSONPath(obj map[string]interface{}, prefix string) map[string]interface{} {
	attrs := make(map[string]interface{})
	var nested map[string]interface{}

	for k, v := range obj {
		switch V := v.(type) {
		default:
			attrs[prefix+"."+k] = V
		case map[string]interface{}:
			nested = MakeJSONPath(V, prefix+"."+k)
			for k2, v2 := range nested {
				attrs[k2] = v2
			}
		}
	}

	return attrs
}
