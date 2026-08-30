package network

import (
	"fmt"
	"strconv"
)

// delete this file when deleting the bc submit
const (
	LowestNodePort  = 30000
	HighestNodePort = 32767
)

func ParsePort(strPort string, isNodePort bool) (int, error) {
	num, err := strconv.Atoi(strPort)
	if err != nil {
		return 0, fmt.Errorf("invalid port number %s. Use only numbers", strPort)
	}

	legalLowestPort := lowestPort
	legalHighestPort := highestPort
	if isNodePort {
		legalLowestPort = LowestNodePort
		legalHighestPort = HighestNodePort
	}

	if num < legalLowestPort || num > legalHighestPort {
		return 0, fmt.Errorf("invalid port number %s. Valid port numbers are between %d and %d", strPort, legalLowestPort, legalHighestPort)
	}

	return num, nil
}
