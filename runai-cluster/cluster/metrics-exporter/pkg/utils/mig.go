package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	A100memory40gbDeviceName = "A100-40GB"
	A100memory80gbDeviceName = "A100-80GB"

	a100             = "A100"
	memory40GbString = "40GB"
	memory80GbString = "80GB"
	h100             = "H100"

	memory40GBValue = float64(40)
	memory80GBValue = float64(80)

	migDeviceNamePattern                           = `^nvidia.com/mig-(\d+)g\.(\d+)gb$`
	successMatchedCountForMigDeviceSearchInPattern = 3
)

type MigDeviceInfo struct {
	Name          string `json:"name"`
	Position      int    `json:"position"`
	GPUAmount     int    `json:"gpuamount,omitempty"`
	DeviceUUID    string `json:"deviceUUID"`
	GpuInstanceId int    `json:"gpuinstanceid"`
	GPUIndex      string `json:"gpuindex"`
	GPUProduct    string `json:"gputype,omitempty"`
}

func GetMigDeviceFamilies() []string {
	return []string{a100, h100}
}

func GetGBMemoryInMigProduct(labelValue string) (float64, error) {
	if strings.Contains(labelValue, a100) {
		if strings.Contains(labelValue, memory40GbString) {
			return memory40GBValue, nil
		}
		if strings.Contains(labelValue, memory80GbString) {
			return memory80GBValue, nil
		}
	}
	if strings.Contains(labelValue, h100) {
		return memory80GBValue, nil
	}
	return -1, fmt.Errorf("unknown device memory for device %s", labelValue)
}

func ExtractGpuAndMemoryFromMigResourceName(migResourceName string) (int, int, error) {
	matches := regexp.MustCompile(migDeviceNamePattern).FindStringSubmatch(migResourceName)
	if len(matches) < successMatchedCountForMigDeviceSearchInPattern {
		return -1, -1, fmt.Errorf("failed to extract gpu/memory from %v", migResourceName)
	}

	gpu, err := strconv.Atoi(matches[1])
	if err != nil {
		return -1, -1, fmt.Errorf("failed parsing %v to integer", matches[2])
	}
	mem, err := strconv.Atoi(matches[2])
	if err != nil {
		return -1, -1, fmt.Errorf("failed parsing %v to integer", matches[2])
	}

	return gpu, mem, nil
}
