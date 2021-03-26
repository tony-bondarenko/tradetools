package tinkoff

import (
	"fmt"
)

type ClientConfiguration struct {
	token string
}

func createClientConfig(configuration interface{}) (*ClientConfiguration, error) {
	if configurationMap, ok := configuration.(map[string]interface{}); ok {
		if token, ok := configurationMap["token"]; ok {
			if tokenString, ok := token.(string); ok {
				return &ClientConfiguration{tokenString}, nil
			}
			return nil, fmt.Errorf("token has incorrect type")
		}
		return nil, fmt.Errorf("token is missing in configuration")
	}
	return nil, fmt.Errorf("unknown configuration format")
}
