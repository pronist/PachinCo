package client

func TransformArrayMap(array interface{}) []map[string]interface{} {
	var arrayMap []map[string]interface{}

	if array, ok := array.([]interface{}); ok {
		for _, item := range array {
			if item, ok := item.(map[string]interface{}); ok {
				arrayMap = append(arrayMap, item)
			}
		}
	}

	return arrayMap
}