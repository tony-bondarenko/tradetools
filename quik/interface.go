package quik

import "fmt"

func getInterfaceMapStringValue(configuration map[string]interface{}, key string) (string, error) {
	if value, ok := configuration[key].(string); ok {
		return value, nil
	}
	return "", fmt.Errorf("incorrect interface type")
}
