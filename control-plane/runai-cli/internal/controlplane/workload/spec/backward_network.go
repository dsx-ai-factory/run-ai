package spec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/network"
)

const (
	nodePortServiceType    = "nodeport"
	externalUrlServiceType = "external-url"
	portParam              = "port"
	customUrlParam         = "custom-url"
)

type ExternalUrlArg struct {
	port      string
	customUrl string
}

type ServiceTypeArgsType struct {
	nodePorts       []string
	externalUrlArgs []ExternalUrlArg
}

// parseKeyValueParams gets an array of strings of the form "key1=value1,key1=value2,key2=value3"
// and returns a map of the form "key1": ["value1", "value2"], "key2": ["value3"]
func parseKeyValueParams(keyValueStrArray []string) (map[string][]string, error) {
	params := make(map[string][]string)
	for _, keyValueItem := range keyValueStrArray {
		keyValue := strings.Split(keyValueItem, "=")
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("invalid service-type parameter format '%s', use key=value", keyValueItem)
		}
		key := keyValue[0]
		value := keyValue[1]
		params[key] = append(params[key], value)
	}

	return params, nil
}

// handleServiceTypeNodePort takes the params relevant to the nodeport case and fills in the ServiceTypeArgsType struct
func handleServiceTypeNodePort(serviceTypeArgs *ServiceTypeArgsType, params map[string][]string, portsArg []string) error {
	if len(serviceTypeArgs.nodePorts) != 0 {
		return fmt.Errorf("service-type 'nodeport' provided more than once, use a single one and specify multiple ports if needed")
	}
	ports, ok := params[portParam]
	if !ok {
		// port not found in string, use portsArg
		if len(portsArg) == 0 {
			return fmt.Errorf("port not provided for service-type 'nodeport', use either --port or --service-type=nodeport,port=<port>")
		}
		ports = portsArg
	} else {
		// port found, make sure portsArg wasn't used as well
		if len(portsArg) != 0 {
			return fmt.Errorf("port provided for service-type 'nodeport' both in --service-type and --port, use either of them")
		}
	}

	// verify no invalid params were passed
	delete(params, portParam)
	if len(params) != 0 {
		return fmt.Errorf("invalid params provided for service-type 'nodeport': %s", reflect.ValueOf(params).MapKeys())
	}

	serviceTypeArgs.nodePorts = ports

	return nil
}

// handleServiceTypeExternalUrl takes the params relevant to the external-url case and fills in the ServiceTypeArgsType struct
func handleServiceTypeExternalUrl(serviceTypeArgs *ServiceTypeArgsType, params map[string][]string, portsArg []string, customUrlArg string) error {
	var externalUrlArg ExternalUrlArg

	ports, ok := params[portParam]
	if !ok {
		// port not found in string, use portsArg
		if len(portsArg) == 0 {
			return fmt.Errorf("port not provided for service-type 'external-url', use either --port or --service-type=external-url,port=<port>")
		}
		ports = portsArg
	} else {
		// port found, make sure portsArg wasn't used as well
		if len(portsArg) != 0 {
			return fmt.Errorf("port provided for service-type 'external-url' both in --service-type and --port, use either of them")
		}
	}
	// make sure only a single port was provided
	if len(ports) != 1 {
		return fmt.Errorf("multiple ports provided for service-type 'external-url': %s, only one is required", ports)
	}

	externalUrlArg.port = ports[0]

	customUrl, ok := params[customUrlParam]
	if !ok {
		// custom-url not found in string, use customUrlArg, if available
		if customUrlArg != "" {
			externalUrlArg.customUrl = customUrlArg
		}
	} else {
		// custom-url found, make sure customUrlArg wasn't used as well
		if customUrlArg != "" {
			return fmt.Errorf("custom-url provided for service-type 'external-url' both in --service-type and --custom-url, use either of them")
		}
		// make sure only a single custom-url was provided
		if len(customUrl) != 1 {
			return fmt.Errorf("multiple custom-urls provided for service-type 'external-url': %s, only one is required", customUrl)
		}

		externalUrlArg.customUrl = customUrl[0]
	}

	// verify no invalid params were passed
	delete(params, portParam)
	delete(params, customUrlParam)
	if len(params) != 0 {
		return fmt.Errorf("invalid params provided for service-type 'external-url': %s", reflect.ValueOf(params).MapKeys())
	}

	serviceTypeArgs.externalUrlArgs = append(serviceTypeArgs.externalUrlArgs, externalUrlArg)

	return nil
}

// ParseServiceTypeParams parses the given service types, along with optional port arguments and a custom URL argument,
// and returns a map of service types mapped to their corresponding parameters.
// an error is returned if the input is in an invalid format or if required parameters are missing.
// sample inputs:
// serviceTypes=["nodeport"] portsArg=["30000:3000", "4000"]
// serviceTypes=["nodeport,port=30000:3000", "external-url,port=3000,custom-url=http://my.customurl.com/myapi"]
// serviceTypes=["external-url,port=3000,custom-url=http://my.customurl.com/myapi", "external-url,port=3001"]
func ParseServiceTypeParams(serviceTypes, portsArg []string, customUrlArg string) (*ServiceTypeArgsType, error) {
	var serviceTypeArgs ServiceTypeArgsType

	// do not allow port arg without service-type
	if len(portsArg) != 0 && len(serviceTypes) == 0 {
		return nil, fmt.Errorf("port provided without any service-type")
	}

	for _, serviceType := range serviceTypes {
		// separate the service-type name and its params
		parts := strings.Split(serviceType, ",")
		if len(parts) == 0 {
			continue // this means we got an empty string, which is not possible in the cli
		}

		serviceType := strings.ToLower(parts[0])
		serviceParams := parts[1:]

		// parse the params, which should be in the format of key=value
		params, err := parseKeyValueParams(serviceParams)
		if err != nil {
			return nil, err
		}

		// handle each service-type according to its format and validation rules

		switch serviceType {
		case nodePortServiceType:
			if err := handleServiceTypeNodePort(&serviceTypeArgs, params, portsArg); err != nil {
				return nil, err
			}
		case externalUrlServiceType:
			if err := handleServiceTypeExternalUrl(&serviceTypeArgs, params, portsArg, customUrlArg); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown service type '%s'", serviceType)
		}
	}

	// do not allow custom-url arg without service-type external-url
	if len(customUrlArg) != 0 && len(serviceTypeArgs.externalUrlArgs) == 0 {
		return nil, fmt.Errorf("custom-url provided without a service-type 'external-url'")
	}

	return &serviceTypeArgs, nil
}

// parseServicePortLiterals returns host and container ports of a service
func parseServicePortLiterals(portStr string, isNodePort bool) (hostPort int, containerPort int, err error) {
	ports := strings.Split(portStr, ":")
	if (len(ports) != 1 && len(ports) != 2) ||
		(len(ports) == 2 && (ports[0] == "" || ports[1] == "")) {
		return 0, 0,
			fmt.Errorf("invalid port. Use one of the following formats:" +
				" 'host-port:container-port' or 'container-port'")
	}

	if len(ports) == 1 {
		containerPort, err = network.ParsePort(ports[0], false)
		if err != nil {
			return 0, 0, err
		}
	} else {
		hostPort, err = network.ParsePort(ports[0], isNodePort)
		if err != nil {
			return 0, 0, err
		}
		containerPort, err = network.ParsePort(ports[1], false)
		if err != nil {
			return 0, 0, err
		}
	}

	return hostPort, containerPort, nil
}
