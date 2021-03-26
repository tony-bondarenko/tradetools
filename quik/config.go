package quik

import (
	"fmt"
)

type ClientConfiguration struct {
	account string
	host    string
	ports   PortConfiguration
}

type PortConfiguration struct {
	cmd      string
	callback string
}

func createClientConfig(configuration interface{}) (*ClientConfiguration, error) {
	var configurationMap map[string]interface{}
	var ok bool
	var err error
	if configurationMap, ok = configuration.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unknown configuration format")
	}

	config := new(ClientConfiguration)
	config.ports = PortConfiguration{}

	if config.account, err = getInterfaceMapStringValue(configurationMap, "account"); err != nil {
		return nil, err
	}

	if config.host, err = getInterfaceMapStringValue(configurationMap, "host"); err != nil {
		return nil, err
	}

	var portConfigurationMap map[string]interface{}
	if portConfigurationMap, ok = configurationMap["port"].(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unknown configuration format")
	}

	if config.ports.cmd, err = getInterfaceMapStringValue(portConfigurationMap, "cmd"); err != nil {
		return nil, fmt.Errorf("unknown configuration format")
	}

	if config.ports.callback, err = getInterfaceMapStringValue(portConfigurationMap, "callback"); err != nil {
		return nil, fmt.Errorf("unknown configuration format")
	}

	return config, nil
}
